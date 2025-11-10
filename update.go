package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/dlclark/regexp2"

	"slices"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/dofusdude/ankabuffer"
	"github.com/dofusdude/doduda/ui"
	"github.com/dofusdude/doduda/unpack"
	mapping "github.com/dofusdude/dodumap"
	jsnan "github.com/xhhuango/json"
)

type HashFile struct {
	Hash         string
	Filename     string
	FriendlyName string
}

func PartitionSlice[T any](items []T, parts int) (chunks [][]T) {
	var divided [][]T

	chunkSize := (len(items) + parts - 1) / parts

	for i := 0; i < len(items); i += chunkSize {
		end := i + chunkSize

		if end > len(items) {
			end = len(items)
		}

		divided = append(divided, items[i:end])
	}

	return divided
}

// https://stackoverflow.com/questions/13422578/in-go-how-to-get-a-slice-of-values-from-a-map
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

func DownloadMountsImages(mounts *mapping.JSONGameData, bin int, hashJson *ankabuffer.Manifest, worker int, dir string, headless bool) {
	arr := Values(mounts.Mounts)
	workerSlices := PartitionSlice(arr, worker)

	wg := sync.WaitGroup{}
	for i, workerSlice := range workerSlices {
		wg.Add(1)
		go func(workerSlice []mapping.JSONGameMount, dir string, id int) {
			defer wg.Done()
			if headless {
				log.Print(ui.TitleStyle.Render("Mount Worker"), "id", id, "jobs", len(workerSlice), "state", "spawned")
			}
			DownloadMountImageWorker(hashJson, bin, "main", dir, workerSlice, headless)
			if headless {
				log.Print(ui.TitleStyle.Render("Mount Worker"), "id", id, "jobs", len(workerSlice), "state", "finished")
			}
		}(workerSlice, dir, i)
	}
	wg.Wait()
}

func DownloadMountImageWorker(manifest *ankabuffer.Manifest, bin int, fragment string, dir string, workerSlice []mapping.JSONGameMount, headless bool) {
	workerUpdates := make(chan bool, len(workerSlice))
	var feedbackWg sync.WaitGroup
	feedbackWg.Add(1)
	go func() {
		defer feedbackWg.Done()
		ui.Progress("Mount Images", len(workerSlice)*2, workerUpdates, 0, false, headless)
	}()

	for _, mount := range workerSlice {
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer func() {
				defer wg.Done()
				if isChannelClosed(workerUpdates) {
					os.Exit(1)
				}
				workerUpdates <- true
			}()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.png", mountId)
			image.FriendlyName = fmt.Sprintf("%d.png", mountId)
			outPath := filepath.Join(dir, "img", "mount")
			_ = DownloadUnpackFiles("Mount Bitmaps", bin, manifest, fragment, []HashFile{image}, dir, outPath, false, "", true, true)
		}(mount.Id, &wg, dir)

		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer func() {
				defer wg.Done()
				if isChannelClosed(workerUpdates) {
					os.Exit(1)
				}
				workerUpdates <- true
			}()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.swf", mountId)
			image.FriendlyName = fmt.Sprintf("%d.swf", mountId)
			outPath := filepath.Join(dir, "vector", "mount")
			_ = DownloadUnpackFiles("Mount Vectors", bin, manifest, fragment, []HashFile{image}, dir, outPath, false, "", true, true)
		}(mount.Id, &wg, dir)

		wg.Wait()
	}

	feedbackWg.Wait()
}

func GetLatestLauncherVersion(release string) (string, error) {
	versionResponse, err := http.Get("https://cytrus.cdn.ankama.com/cytrus.json")
	if err != nil {
		return "", fmt.Errorf("failed to fetch cytrus.json: %w", err)
	}
	defer versionResponse.Body.Close()

	versionBody, err := io.ReadAll(versionResponse.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read cytrus.json response: %w", err)
	}

	var versionJson map[string]interface{}
	err = json.Unmarshal(versionBody, &versionJson)
	if err != nil {
		return "", fmt.Errorf("failed to parse cytrus.json: %w", err)
	}

	// Safe type assertions with error checking
	games, ok := versionJson["games"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cytrus.json: 'games' field not found or invalid type")
	}

	dofus, ok := games["dofus"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cytrus.json: 'games.dofus' field not found or invalid type")
	}

	platforms, ok := dofus["platforms"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cytrus.json: 'platforms' field not found or invalid type")
	}

	windows, ok := platforms["windows"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("cytrus.json: 'windows' platform not found or invalid type")
	}

	version, ok := windows[release].(string)
	if !ok {
		return "", fmt.Errorf("cytrus.json: version for release '%s' not found or invalid type", release)
	}

	return version, nil
}

func touchFileIfNotExists(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	return nil
}

func CreateDataDirectoryStructure(dir string) {
	os.MkdirAll(fmt.Sprintf("%s/tmp/vector", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/img/item", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/img/mount", dir), os.ModePerm)

	os.MkdirAll(fmt.Sprintf("%s/vector/item", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/vector/mount", dir), os.ModePerm)

	os.MkdirAll(fmt.Sprintf("%s/languages", dir), os.ModePerm)
}

func GetReleaseManifest(version string, gameVersionType string, platform string, dir string) ([]byte, error) {
	gameHashesUrl := fmt.Sprintf("https://cytrus.cdn.ankama.com/dofus/releases/%s/%s/%s.manifest", gameVersionType, platform, version)
	hashResponse, err := http.Get(gameHashesUrl)
	if err != nil {
		log.Fatal("Could not get manifest file", err)
		return nil, err
	}

	hashBody, err := io.ReadAll(hashResponse.Body)
	if err != nil {
		log.Fatal("Could not read response body of manifest file", err)
		return nil, err
	}

	return hashBody, nil
}

func ignoresRegex(ignores []string, filename string) bool {
	for _, ignore := range ignores {
		compiled := regexp2.MustCompile(ignore, regexp2.None)
		match, _ := compiled.MatchString(filename)
		if match {
			return true
		}
	}
	return false
}

func contains(arr []string, str string) bool {
	if arr == nil {
		return false
	}
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// ported from https://stackoverflow.com/questions/10420352/converting-file-size-in-bytes-to-human-readable-string
func humanFileSize(bytes float64, decimal bool, precision int) string {
	thresh := 1024.0
	if decimal {
		thresh = 1000.0
	}

	if math.Abs(bytes) < thresh {
		return fmt.Sprintf("%d B", int(bytes))
	}

	var units []string
	if decimal {
		units = []string{"kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	} else {
		units = []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	}

	u := -1
	r := math.Pow(10, float64(precision))

	for {
		bytes /= thresh
		u++
		if math.Round(math.Abs(bytes)*r)/r < thresh || u == len(units)-1 {
			break
		}
	}

	return fmt.Sprintf("%.*f %s", precision, bytes, units[u])
}

func Download(releaseChannel string, version string, dir string, clean bool, fullGame bool, platform string, bin int, manifest string, jobs int, ignore []string, indent string, headless bool) error {
	var ankaManifest ankabuffer.Manifest
	manifestSearchPath := "manifest.json"

	var manifestWg sync.WaitGroup
	feedbacks := make(chan string)
	manifestWg.Add(1)
	go func() {
		defer manifestWg.Done()
		ui.Spinner("Manifest", feedbacks, false, headless)
	}()

	if isChannelClosed(feedbacks) {
		os.Exit(1)
	}
	feedbacks <- "⬇️"

	var manifestPath string
	if manifest == "" {
		if _, err := os.Stat(manifestSearchPath); os.IsNotExist(err) {
			manifestPath = ""
		} else {
			manifestPath, err = filepath.Abs(manifestSearchPath)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		var err error
		if _, err := os.Stat(manifest); os.IsNotExist(err) {
			log.Fatal(err)
		}
		manifestPath, err = filepath.Abs(manifest)
		if err != nil {
			log.Fatal(err)
		}
	}

	var dofusVersion string

	if manifestPath == "" || clean {
		cytrusPrefix := "6.0_"
		if version == "latest" {
			version = GetLatestLauncherVersion(releaseChannel)
		} else {
			// ATT: prefix changes with cytrus updates
			if !strings.HasPrefix(version, cytrusPrefix) {
				version = fmt.Sprintf("%s%s", cytrusPrefix, version)
			}
		}

		dofusVersion = strings.TrimPrefix(version, cytrusPrefix)

		rawManifest, err := GetReleaseManifest(version, releaseChannel, platform, dir)
		if err != nil {
			return err
		}

		feedbacks <- "parsing"
		ankaManifestPtr, err := ankabuffer.ParseManifest(rawManifest, dofusVersion)
		if err != nil {
			log.Fatal(err)
		}
		ankaManifest = *ankaManifestPtr

		marshalledBytes, err := json.Marshal(ankaManifestPtr)
		if err != nil {
			log.Fatal(err)
		}
		os.WriteFile(manifestSearchPath, marshalledBytes, os.ModePerm)
	} else {
		log.Debug("Using cached manifest")
		manifestFile, err := os.Open(manifestPath)
		if err != nil {
			log.Fatal(err)
		}
		defer manifestFile.Close()

		byteValue, err := io.ReadAll(manifestFile)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(byteValue, &ankaManifest)
		if err != nil {
			log.Fatal(err)
		}

		dofusVersion = ankaManifest.GameVersion
	}

	rawDofusMajorVersion, err := strconv.Atoi(strings.Split(dofusVersion, ".")[0])
	if err != nil {
		log.Fatal("Invalid version")
	}

	betaSuffix := ""
	if strings.Contains(releaseChannel, "beta") {
		betaSuffix = " [beta]"
	}
	feedbacks <- dofusVersion + betaSuffix

	close(feedbacks)
	manifestWg.Wait()

	if fullGame {
		var fullGameUiWg sync.WaitGroup
		feedbacks := make(chan string)
		fullGameUiWg.Add(1)
		go func() {
			defer fullGameUiWg.Done()
			ui.Spinner(ankaManifest.GameVersion, feedbacks, false, headless)
		}()

		if isChannelClosed(feedbacks) {
			os.Exit(1)
		}

		var totalSize int64
		fragmentFiles := map[string][]HashFile{}
		for _, fragment := range ankaManifest.Fragments {
			for _, fragmentFile := range fragment.Files {
				if fragmentFile.Name == "" {
					continue
				}

				totalSize += fragmentFile.Size
				hashFile := HashFile{
					Filename:     fragmentFile.Name,
					FriendlyName: fragmentFile.Name,
					Hash:         fragmentFile.Hash,
				}

				fragmentFiles[fragment.Name] = append(fragmentFiles[fragment.Name], hashFile)
			}
		}

		totalFragments := len(fragmentFiles)
		fragmentCounter := 0
		for fragmentName, files := range fragmentFiles {
			fragmentCounter++
			feedbacks <- "Fragment " + strconv.Itoa(fragmentCounter) + "/" + strconv.Itoa(totalFragments)
			err = DownloadUnpackFiles(ankaManifest.GameVersion, bin, &ankaManifest, fragmentName, files, dir, dir, false, "", headless, false)
			if err != nil {
				return err
			}
		}

		close(feedbacks)
		fullGameUiWg.Wait()

	} else {
		CreateDataDirectoryStructure(dir)

		if rawDofusMajorVersion == 3 {
			err := PullImages([]string{"stelzo/doduda-umbu:" + ARCH, "stelzo/assetstudio-cli:" + ARCH}, false, headless)
			if err != nil {
				return err
			}
		}

		// parallel if headless true
		/*worker := 1
		if headless {
			worker = 4
			}*/

		if !ignoresRegex(ignore, "data-languages") {
			if err := DownloadLanguages(&ankaManifest, bin, rawDofusMajorVersion, dir, indent, headless); err != nil {
				log.Fatal(err)
			}
		}

		if !ignoresRegex(ignore, "data-items") {
			if err := DownloadItems(&ankaManifest, bin, rawDofusMajorVersion, dir, indent, headless); err != nil {
				log.Fatal(err)
			}
		}

		if !ignoresRegex(ignore, "data-quests") {
			if err := DownloadQuests(&ankaManifest, bin, rawDofusMajorVersion, dir, indent, headless); err != nil {
				log.Fatal(err)
			}
		}

		if !ignoresRegex(ignore, "data-achievements") {
			if err := DownloadAchievements(&ankaManifest, bin, rawDofusMajorVersion, dir, indent, headless); err != nil {
				log.Fatal(err)
			}
		}

		if err := DownloadImagesLauncher(&ankaManifest, bin, jobs, rawDofusMajorVersion, dir, ignore, headless); err != nil {
			log.Fatal(err)
		}

		// mountsimages rendering only needed for Dofus 2.x
		if rawDofusMajorVersion == 2 && !ignoresRegex(ignore, "images-mounts") && !ignoresRegex(ignore, "data-items") {
			gamedata := mapping.ParseRawData(dir)
			if !headless {
				jobs = 1
			}
			DownloadMountsImages(gamedata, bin, &ankaManifest, jobs, dir, headless)
		}

		os.RemoveAll(fmt.Sprintf("%s/tmp", dir))
	}

	return nil
}

func DownloadBundle(bundleHash string) ([]byte, error) {
	url := fmt.Sprintf("https://cytrus.cdn.ankama.com/dofus/bundles/%s/%s", bundleHash[0:2], bundleHash)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bundle %s status %d", bundleHash, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// TODO category not used anymore
func UnpackUnityBundle(category string, inputPath string, outputPath string, muteSpinner bool, headless bool) error {
	if !strings.HasSuffix(inputPath, ".bundle") {
		return fmt.Errorf("invalid bundle suffix")
	}

	if !strings.HasSuffix(outputPath, ".json") {
		return fmt.Errorf("can only output to json")
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("bundle %s does not exist", inputPath)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	outputTrimmed := strings.TrimSuffix(outputPath, ".asset.json")
	outputFileName := outputTrimmed + ".json"

	inputRawFileName := filepath.Base(inputPath)
	outputRawFileName := filepath.Base(outputFileName)

	inputDir := filepath.Dir(inputPath)

	cmd := []string{"dotnet", "out/unity-bundle-unwrap.dll", path.Join("/app", "data", inputRawFileName), path.Join("/app", "data", outputRawFileName)}

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "stelzo/doduda-umbu:" + ARCH,
		Cmd:   cmd,
		Volumes: map[string]struct{}{
			"/app/data": {},
		},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/data", inputDir),
		},
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func UnpackUnityI18n(category string, inputPath string, outputPath string, muteSpinner bool, headless bool) error {
	if !strings.HasSuffix(inputPath, ".bin") {
		return fmt.Errorf("invalid suffix")
	}

	if !strings.HasSuffix(outputPath, ".json") {
		return fmt.Errorf("can only output to json")
	}

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", inputPath)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	inputRawFileName := filepath.Base(inputPath)
	outputRawFileName := filepath.Base(outputPath)

	inputDir := filepath.Dir(inputPath)

	cmd := []string{"dotnet", "out/unity-bundle-unwrap.dll", path.Join("/app", "data", inputRawFileName), path.Join("/app", "data", outputRawFileName)}

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "stelzo/doduda-umbu:" + ARCH,
		Cmd:   cmd,
		Volumes: map[string]struct{}{
			"/app/data": {},
		},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/data", inputDir),
		},
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func PullImages(images []string, muteSpinner bool, headless bool) error {
	feedbacks := make(chan string)

	var feedbackWg sync.WaitGroup
	if !muteSpinner {
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Docker", feedbacks, false, headless)
		}()
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, imageRef := range images {
		ctx := context.Background()
		feedbacks <- "⬇️ " + imageRef
		imageHandle, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
		if err != nil {
			return err
		}
		defer imageHandle.Close()

		_, err = io.ReadAll(imageHandle)
		if err != nil {
			log.Fatalf("Error finishing image handler: %v", err)
		}
	}

	close(feedbacks)
	feedbackWg.Wait()

	return nil
}

func UnpackUnityImages(inputDir string, outputDir string, muteSpinner bool, headless bool) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	imageName := "stelzo/assetstudio-cli:" + ARCH

	ctx := context.Background()
	bundles, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, bundle := range bundles {
		if bundle.IsDir() || !strings.HasSuffix(bundle.Name(), ".imagebundle") {
			continue
		}

		absInputPath := filepath.Join(inputDir, bundle.Name())

		cmd := []string{"./data", "--unity-version", "6000.0.41.58439"}

		uid := strconv.Itoa(os.Getuid())
		gid := strconv.Itoa(os.Getgid())
		user := uid + ":" + gid

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: imageName,
			Cmd:   cmd,
			User:  user,
			Volumes: map[string]struct{}{
				"/app/AssetStudio/data":     {},
				"/app/AssetStudio/ASExport": {},
			},
		}, &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/app/AssetStudio/data", absInputPath),
				fmt.Sprintf("%s:/app/AssetStudio/ASExport", outputDir)},
			AutoRemove: true,
		}, nil, nil, "")
		if err != nil {
			log.Fatal(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return err
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-statusCh:
		}

	}

	return nil
}

func Unpack(file string, dir string, destDir string, category string, indent string, muteSpinner bool, headless bool) {
	suffix := filepath.Ext(file)[1:]

	if suffix == "png" || suffix == "jpg" || suffix == "jpeg" {
		return // no need to unpack images files
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Fatal(err)
	}

	fileNoExt := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	absOutPath := filepath.Join(destDir, fileNoExt+".json")

	supportedUnpack := []string{"d2o", "d2i", "imagebundle", "bundle", "bin"}
	isSupported := slices.Contains(supportedUnpack, suffix)

	if !isSupported {
		log.Warnf("Unsupported file type for unpacking %s", suffix)
	}

	if suffix == "d2o" {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		reader, err := unpack.NewD2OReader(f)
		if err != nil {
			log.Fatal(err)
		}

		objects := reader.GetObjects()
		var marshalledBytes []byte
		if indent != "" {
			marshalledBytes, err = jsnan.MarshalIndent(objects, "", indent)
		} else {
			marshalledBytes, err = jsnan.Marshal(objects)
		}
		if err != nil {
			log.Fatal(err)
		}
		marshalledBytes = bytes.Replace(marshalledBytes, []byte("NaN"), []byte("null"), -1)

		err = os.WriteFile(absOutPath, marshalledBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if suffix == "d2i" {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		data := unpack.NewD2I(f).Read()

		var marshalledBytes []byte
		if indent != "" {
			marshalledBytes, err = jsnan.MarshalIndent(data, "", indent)
		} else {
			marshalledBytes, err = jsnan.Marshal(data)
		}
		if err != nil {
			log.Fatal(err)
		}
		marshalledBytes = bytes.Replace(marshalledBytes, []byte("NaN"), []byte("null"), -1)

		err = os.WriteFile(absOutPath, marshalledBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	if suffix == "imagebundle" {
		dir := filepath.Dir(file)
		err := UnpackUnityImages(dir, destDir, muteSpinner, headless)
		if err != nil {
			log.Fatal(err)
		}
	}

	if suffix == "bundle" {
		err := UnpackUnityBundle(category, file, absOutPath, muteSpinner, headless)
		if err != nil {
			log.Fatal(err)
		}
	}

	if suffix == "bin" {
		err := UnpackUnityI18n(category, file, absOutPath, muteSpinner, headless)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func isChannelClosed[T any](ch chan T) bool {
	select {
	case _, ok := <-ch:
		if !ok {
			return true
		}
	default:
	}

	return false
}

func convertMBToBytes(mb int) int64 {
	return int64(mb) * 1024 * 1024
}

func splitFilesIntoBins(files []ankabuffer.File, maxMBPerBin int) [][]ankabuffer.File {
	maxBytesPerBin := convertMBToBytes(maxMBPerBin)
	var bins [][]ankabuffer.File
	var currentBin []ankabuffer.File
	var currentBinSize int64

	for _, file := range files {
		if file.Size > maxBytesPerBin {
			if len(currentBin) > 0 {
				bins = append(bins, currentBin)
				currentBin = nil
				currentBinSize = 0
			}
			bins = append(bins, []ankabuffer.File{file})
			continue
		}

		if currentBinSize+file.Size > maxBytesPerBin {
			bins = append(bins, currentBin)
			currentBin = []ankabuffer.File{file}
			currentBinSize = file.Size
		} else {
			currentBin = append(currentBin, file)
			currentBinSize += file.Size
		}
	}

	if len(currentBin) > 0 {
		bins = append(bins, currentBin)
	}

	return bins
}

func DownloadUnpackFiles(title string, bin int, manifest *ankabuffer.Manifest, fragment string, toDownload []HashFile, dir string, destDir string, unpack bool, indent string, silent bool, muteSpinner bool) error {
	var filesToDownload []ankabuffer.File
	toDownloadFiltered := []HashFile{}
	for _, file := range toDownload {
		if strings.HasPrefix(file.Filename, "REGEX:") {
			regex := strings.TrimPrefix(file.Filename, "REGEX:")
			compiled := regexp.MustCompile(regex)
			for key := range manifest.Fragments[fragment].Files {
				if compiled.MatchString(key) {
					if manifest.Fragments[fragment].Files[key].Name == "" {
						continue
					}
					toDownloadFiltered = append(toDownloadFiltered, HashFile{Filename: key, Hash: file.Hash, FriendlyName: file.FriendlyName})
				}
			}
		} else {
			if manifest.Fragments[fragment].Files[file.Filename].Name == "" {
				continue
			}
			toDownloadFiltered = append(toDownloadFiltered, file)
		}
	}

	toDownload = toDownloadFiltered

	for i, file := range toDownload {
		filesToDownload = append(filesToDownload, manifest.Fragments[fragment].Files[file.Filename])
		toDownload[i].Hash = manifest.Fragments[fragment].Files[file.Filename].Hash
	}

	var filebins [][]ankabuffer.File
	if bin > 0 {
		filebins = splitFilesIntoBins(filesToDownload, bin)
	} else {
		filebins = [][]ankabuffer.File{filesToDownload}
	}

	for idx, filesToDownload := range filebins {
		innerTitle := fmt.Sprintf("%s (%d/%d)", title, idx+1, len(filebins))

		feedbacks := make(chan string)

		var feedbackWg sync.WaitGroup
		if !muteSpinner {
			feedbackWg.Add(1)
			go func() {
				defer feedbackWg.Done()
				ui.Spinner(innerTitle, feedbacks, false, silent)
			}()
		}

		bundles := ankabuffer.GetNeededBundles(filesToDownload)

		if len(bundles) == 0 && len(filesToDownload) > 0 {
			for _, file := range filesToDownload {
				log.Warn("Missing bundle for", file.Name)
			}
		}

		if len(bundles) == 0 {
			//log.Warn("No bundles to download") // TODO it seems like the files come out okay even if there are no bundles, maybe warning is not needed
			if !muteSpinner {
				close(feedbacks)
				feedbackWg.Wait()
			}
			continue
		}

		bundlesMap := ankabuffer.GetBundleHashMap(manifest)

		type DownloadedBundle struct {
			BundleHash string
			Data       []byte
		}

		bundlesBuffer := make(map[string]DownloadedBundle)

		if !muteSpinner {
			filesStr := "files"
			if len(filesToDownload) == 1 {
				filesStr = "file"
			}
			bundlesStr := "bundles"
			if len(bundles) == 1 {
				bundlesStr = "bundle"
			}
			feedbacks <- fmt.Sprintf("⬇️ %d %s (%d %s)", len(filesToDownload), filesStr, len(bundles), bundlesStr)
			close(feedbacks)
		}
		feedbackWg.Wait()

		bundleUpdates := make(chan bool, len(bundles))
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Progress(innerTitle, len(bundles)+1, bundleUpdates, 0, true, silent)
		}()

		var bundleDownloadWg sync.WaitGroup
		var bundleDownloadMu sync.Mutex

		for _, bundle := range bundles {
			bundleDownloadWg.Add(1)
			go func(bundle string) {
				defer bundleDownloadWg.Done()

				bundleData, err := DownloadBundle(bundle)
				if err != nil {
					log.Errorf("Could not download bundle %s: %s\n", bundle, err)
					return
				}

				bundleDownloadMu.Lock()
				bundlesBuffer[bundle] = DownloadedBundle{BundleHash: bundle, Data: bundleData}
				bundleDownloadMu.Unlock()

				if isChannelClosed(bundleUpdates) {
					os.Exit(1)
				}
				bundleUpdates <- true
			}(bundle)
		}

		bundleDownloadWg.Wait()

		var wg sync.WaitGroup
		for i, file := range filesToDownload {
			wg.Add(1)
			go func(file ankabuffer.File, bundlesBuffer map[string]DownloadedBundle, dir string, destDir string, i int) {
				defer wg.Done()
				var fileData []byte

				if file.Chunks == nil || len(file.Chunks) == 0 { // file is not chunked
					for _, bundle := range bundlesBuffer {
						for _, chunk := range bundlesMap[bundle.BundleHash].Chunks {
							if chunk.Hash == file.Hash {
								fileData = bundle.Data[chunk.Offset : chunk.Offset+chunk.Size]
								break
							}
						}
						if fileData != nil {
							break
						}
					}
				} else { // file is chunked
					type ChunkData struct {
						Data   []byte
						Offset int64
						Size   int64
					}
					var chunksData []ChunkData
					for _, chunk := range file.Chunks { // all chunks of the file
						for _, bundle := range bundlesBuffer { // search in downloaded bundles for the chunk
							foundChunk := false
							for _, bundleChunk := range bundlesMap[bundle.BundleHash].Chunks { // each chunk of the searched bundle could be a chunk of the file
								if bundleChunk.Hash == chunk.Hash {
									foundChunk = true
									if len(bundle.Data) < int(bundleChunk.Offset+bundleChunk.Size) {
										err := fmt.Errorf("bundle data is too small. Bundle offset/size: %d/%d, BundleData length: %d, BundleHash: %s, BundleChunkHash: %s", bundleChunk.Offset, bundleChunk.Size, len(bundle.Data), bundle.BundleHash, bundleChunk.Hash)
										log.Fatal(err)
									}

									chunksData = append(chunksData, ChunkData{Data: bundle.Data[bundleChunk.Offset : bundleChunk.Offset+bundleChunk.Size], Offset: chunk.Offset, Size: chunk.Size})
								}
							}
							if foundChunk {
								break
							}
						}
					}
					sort.Slice(chunksData, func(i, j int) bool {
						return chunksData[i].Offset < chunksData[j].Offset
					})
					//if len(chunksData) > 1 {
					//	log.Println("Chunks data", chunksData[0].Offset, chunksData[len(chunksData)-1].Offset)
					//}
					for _, chunk := range chunksData {
						fileData = append(fileData, chunk.Data...)
					}
				}

				if len(fileData) == 0 {
					err := fmt.Errorf("file data is empty %s", file.Hash)
					log.Fatal(err)
				}

				offlineFilePath := filepath.Join(destDir, toDownload[i].FriendlyName)

				// anonymous function to safely defer closing file
				func() {
					path := filepath.Dir(offlineFilePath)
					_ = os.MkdirAll(path, os.ModePerm)

					fp, err := os.Create(offlineFilePath)
					if err != nil {
						log.Fatal(err)
					}
					defer func() {
						err := fp.Close()
						if err != nil {
							log.Fatal(err)
						}
					}()
					_, err = fp.Write(fileData)
					if err != nil {
						log.Fatal(err)
						return
					}
				}()

				//log.Infof("%s ✅", filepath.Base(file.Name))

				if unpack {
					Unpack(offlineFilePath, dir, destDir, title, indent, muteSpinner, silent)
					err := os.Remove(offlineFilePath)
					if err != nil {
						log.Fatal(err)
					}
				}

			}(file, bundlesBuffer, dir, destDir, i)
		}

		wg.Wait()

		if !isChannelClosed(bundleUpdates) {
			bundleUpdates <- true
		}

		feedbackWg.Wait()

	}

	return nil
}
