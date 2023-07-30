package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/ankabuffer"
)

var Languages = []string{"de", "en", "es", "fr", "it", "pt"}

type HashFile struct {
	Hash         string
	Filename     string
	FriendlyName string
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

func Download(beta bool, dir string, pythonPath string, manifest string, mountsWorker int, ignore []string) error {
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

		// check if manifest is a file
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

		marshalledBytes, _ := json.MarshalIndent(ankaManifest, "", "  ")
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
		go func(manifest *ankabuffer.Manifest, dir string, pythonPath string) {
			defer waitGrp.Done()
			if err := DownloadLanguages(manifest, dir, pythonPath); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir, pythonPath)
	}

	if !contains(ignore, "images") {
		waitGrp.Add(1)
		go func(manifest *ankabuffer.Manifest, dir string, pythonPath string) {
			defer waitGrp.Done()
			if err := DownloadImagesLauncher(manifest, dir, pythonPath); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir, pythonPath)
	}

	if !contains(ignore, "items") {
		waitGrp.Add(1)
		go func(manifest *ankabuffer.Manifest, dir string, pythonPath string) {
			defer waitGrp.Done()
			if err := DownloadItems(manifest, dir, pythonPath); err != nil {
				log.Fatal(err)
			}
		}(&ankaManifest, dir, pythonPath)
	}

	waitGrp.Wait()

	if !contains(ignore, "mountsimages") {
		log.Info("Parsing for missing mount images...")
		gamedata := ParseRawData(dir)
		log.Info("Downloading mount images...")
		DownloadMountsImages(gamedata, &ankaManifest, mountsWorker, dir, pythonPath)
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

func Unpack(file string, dir string, destDir string, pythonPath string) {
	var err error

	suffix := filepath.Ext(file)[1:]

	if suffix == "png" || suffix == "jpg" || suffix == "jpeg" {
		return // no need to unpack images files
	}

	absConvertCmd := fmt.Sprintf("%s/PyDofus/%s_unpack.py", dir, suffix)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Fatal(err)
	}

	fileNoExt := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	absOutPath := filepath.Join(destDir, fileNoExt+".json")
	fileButJson := strings.Replace(file, filepath.Ext(file), ".json", 1)

	log.Infof("📖 %s -> %s", file, absOutPath)

	err = exec.Command(pythonPath, absConvertCmd, file).Run()
	if err != nil {
		log.Fatalf("Unpacking failed: %s %s %s with Error %v", pythonPath, absConvertCmd, file, err)
	}

	err = os.Rename(fileButJson, absOutPath)
	if err != nil {
		log.Fatal(err)
	}
}

func DownloadUnpackFiles(manifest *ankabuffer.Manifest, fragment string, toDownload []HashFile, dir string, destDir string, unpack bool, pythonPath string) error {
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
			log.Infof("%s ✅ -> 📂 %s", file.Name, offlineFilePath)

			if unpack {
				Unpack(offlineFilePath, dir, destDir, pythonPath)
			}
		}(file, bundlesBuffer, dir, destDir, i)
	}

	wg.Wait()
	return nil
}
