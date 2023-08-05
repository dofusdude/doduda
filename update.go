package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/ankabuffer"
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

func DownloadMountsImages(mounts *mapping.JSONGameData, hashJson *ankabuffer.Manifest, worker int, dir string, indent string) {
	arr := Values(mounts.Mounts)
	workerSlices := PartitionSlice(arr, worker)

	wg := sync.WaitGroup{}
	for _, workerSlice := range workerSlices {
		wg.Add(1)
		go func(workerSlice []mapping.JSONGameMount, dir string, indent string) {
			defer wg.Done()
			DownloadMountImageWorker(hashJson, "main", dir, workerSlice, indent)
		}(workerSlice, dir, indent)
	}
	wg.Wait()
}

func DownloadMountImageWorker(manifest *ankabuffer.Manifest, fragment string, dir string, workerSlice []mapping.JSONGameMount, indent string) {
	wg := sync.WaitGroup{}

	for _, mount := range workerSlice {
		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer wg.Done()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.png", mountId)
			image.FriendlyName = fmt.Sprintf("%d.png", mountId)
			outPath := filepath.Join(dir, "data", "img", "mount")
			_ = DownloadUnpackFiles(manifest, fragment, []HashFile{image}, dir, outPath, true, indent)
		}(mount.Id, &wg, dir)

		//  Missing bundle for content/gfx/mounts/162.swf
		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer wg.Done()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.swf", mountId)
			image.FriendlyName = fmt.Sprintf("%d.swf", mountId)
			outPath := filepath.Join(dir, "data", "vector", "mount")
			_ = DownloadUnpackFiles(manifest, fragment, []HashFile{image}, dir, outPath, false, indent)
		}(mount.Id, &wg, dir)
	}

	wg.Wait()
}

func GetLatestLauncherVersion(beta bool) string {
	versionResponse, err := http.Get("https://cytrus.cdn.ankama.com/cytrus.json")
	if err != nil {
		log.Fatal(err)
	}

	versionBody, err := io.ReadAll(versionResponse.Body)
	if err != nil {
		log.Fatal(err)
	}

	var versionJson map[string]interface{}
	err = json.Unmarshal(versionBody, &versionJson)
	if err != nil {
		log.Fatal(err)
	}

	games := versionJson["games"].(map[string]interface{})
	dofus := games["dofus"].(map[string]interface{})
	platform := dofus["platforms"].(map[string]interface{})
	windows := platform["windows"].(map[string]interface{})

	if beta {
		return windows["beta"].(string)
	} else {
		return windows["main"].(string)
	}
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
	os.MkdirAll(fmt.Sprintf("%s/data/tmp/vector", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/data/img/item", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/data/img/mount", dir), os.ModePerm)

	os.MkdirAll(fmt.Sprintf("%s/data/vector/item", dir), os.ModePerm)
	os.MkdirAll(fmt.Sprintf("%s/data/vector/mount", dir), os.ModePerm)

	os.MkdirAll(fmt.Sprintf("%s/data/languages", dir), os.ModePerm)

	err := touchFileIfNotExists(fmt.Sprintf("%s/data/img/index.html", dir))
	if err != nil {
		log.Fatal(err)
	}
	err = touchFileIfNotExists("data/img/item/index.html")
	if err != nil {
		log.Fatal(err)
	}
	err = touchFileIfNotExists("data/img/mount/index.html")
	if err != nil {
		log.Fatal(err)
	}
}

func GetReleaseManifest(version string, beta bool, dir string) (ankabuffer.Manifest, error) {
	log.Info("Downloading release manifest...")
	var gameVersionType string
	if beta {
		gameVersionType = "beta"
	} else {
		gameVersionType = "main"
	}
	gameHashesUrl := fmt.Sprintf("https://cytrus.cdn.ankama.com/dofus/releases/%s/windows/%s.manifest", gameVersionType, version)
	hashResponse, err := http.Get(gameHashesUrl)
	if err != nil {
		log.Fatal(err)
		return ankabuffer.Manifest{}, err
	}

	hashBody, err := io.ReadAll(hashResponse.Body)
	if err != nil {
		log.Fatal(err)
		return ankabuffer.Manifest{}, err
	}

	fileHashes := *ankabuffer.ParseManifest(hashBody)

	log.Info("... release manifest downloaded")

	return fileHashes, nil
}

func contains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func Download(beta bool, dir string, manifest string, mountsWorker int, ignore []string, indent string) error {
	CleanUp(dir)
	CreateDataDirectoryStructure(dir)

	var ankaManifest ankabuffer.Manifest
	manifestSearchPath := "manifest.json"

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

	if manifestPath == "" {
		version := GetLatestLauncherVersion(beta)

		var err error

		ankaManifest, err = GetReleaseManifest(version, beta, dir)
		if err != nil {
			return err
		}

		marshalledBytes, err := json.Marshal(ankaManifest)
		if err != nil {
			log.Fatal(err)
		}
		os.WriteFile(manifestSearchPath, marshalledBytes, os.ModePerm)
	} else {
		log.Infof("Using manifest file %s", manifestPath)
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
		log.Info("... manifest loaded")
	}

	var waitGrp sync.WaitGroup
	if !contains(ignore, "languages") {
		waitGrp.Add(1)
		go func(manifest *ankabuffer.Manifest, dir string) {
			defer waitGrp.Done()
			if err := DownloadLanguages(manifest, dir, indent); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir)
	}

	if !contains(ignore, "images") {
		waitGrp.Add(1)
		go func(manifest *ankabuffer.Manifest, dir string) {
			defer waitGrp.Done()
			if err := DownloadImagesLauncher(manifest, dir); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir)
	}

	if !contains(ignore, "items") {
		waitGrp.Add(1)
		go func(manifest *ankabuffer.Manifest, dir string, indent string) {
			defer waitGrp.Done()
			if err := DownloadItems(manifest, dir, indent); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir, indent)
	}

	waitGrp.Wait()

	if !contains(ignore, "mountsimages") {
		log.Info("Parsing for missing mount images...")
		gamedata := mapping.ParseRawData(dir)
		log.Info("Downloading mount images...")
		DownloadMountsImages(gamedata, &ankaManifest, mountsWorker, dir, indent)
		log.Info("... mount images downloaded")
	}

	os.RemoveAll(fmt.Sprintf("%s/data/tmp", dir))

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

func CleanUp(dir string) {
	os.RemoveAll(fmt.Sprintf("%s/data", dir))
}

func Unpack(file string, dir string, destDir string, indent string) {
	suffix := filepath.Ext(file)[1:]

	if suffix == "png" || suffix == "jpg" || suffix == "jpeg" {
		return // no need to unpack images files
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Fatal(err)
	}

	fileNoExt := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	absOutPath := filepath.Join(destDir, fileNoExt+".json")

	supportedUnpack := []string{"d2o", "d2i"}
	isSupported := false
	for _, unpackType := range supportedUnpack {
		if suffix == unpackType {
			isSupported = true
			break
		}
	}

	if !isSupported {
		log.Warnf("Unsupported file type for unpacking %s", suffix)
	}

	if suffix == "d2o" {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

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

		os.WriteFile(absOutPath, marshalledBytes, os.ModePerm)
	}

	if suffix == "d2i" {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

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

		os.WriteFile(absOutPath, marshalledBytes, os.ModePerm)
	}
}

func DownloadUnpackFiles(manifest *ankabuffer.Manifest, fragment string, toDownload []HashFile, dir string, destDir string, unpack bool, indent string) error {
	var filesToDownload []ankabuffer.File
	for i, file := range toDownload {
		if manifest.Fragments[fragment].Files[file.Filename].Name == "" {
			continue
		}
		filesToDownload = append(filesToDownload, manifest.Fragments[fragment].Files[file.Filename])
		toDownload[i].Hash = manifest.Fragments[fragment].Files[file.Filename].Hash
	}

	bundles := ankabuffer.GetNeededBundles(filesToDownload)

	if len(bundles) == 0 && len(filesToDownload) > 0 {
		for _, file := range filesToDownload {
			log.Warn("Missing bundle for", file.Name)
		}
	}

	if len(bundles) == 0 {
		log.Warn("No bundles to download")
		return nil
	}

	bundlesMap := ankabuffer.GetBundleHashMap(manifest)

	type DownloadedBundle struct {
		BundleHash string
		Data       []byte
	}

	//bundleData := make(chan DownloadedBundle, len(bundles))
	bundlesBuffer := make(map[string]DownloadedBundle)

	for _, bundle := range bundles {
		//go func(bundleHash string, data chan DownloadedBundle) {
		bundleData, err := DownloadBundle(bundle)
		if err != nil {
			return fmt.Errorf("could not download bundle %s: %s", bundle, err)
		}
		res := DownloadedBundle{BundleHash: bundle, Data: bundleData}
		bundlesBuffer[bundle] = res
	}

	var wg sync.WaitGroup
	for i, file := range filesToDownload {
		if file.Name == "" {
			continue
		}
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

			fp, err := os.Create(offlineFilePath)
			if err != nil {
				log.Fatal(err)
			}
			defer fp.Close()
			_, err = fp.Write(fileData)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Infof("%s ✅", filepath.Base(file.Name))

			if unpack {
				Unpack(offlineFilePath, dir, destDir, indent)
				err = os.Remove(offlineFilePath)
				if err != nil {
					log.Fatal(err)
				}
			}
		}(file, bundlesBuffer, dir, destDir, i)
	}

	wg.Wait()
	return nil
}
