package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

type VersionFile struct {
	Main string `json:"main"`
	Beta string `json:"beta"`
}

func IsVersionChanged(dir string, isBeta bool) (bool, error) {
	versionFilePath := filepath.Join(dir, "data", "version", "version.json")

	versions, err := os.ReadFile(versionFilePath)
	if err != nil {
		return false, err
	}

	var versionFile VersionFile
	err = json.Unmarshal(versions, &versionFile)
	if err != nil {
		return false, err
	}

	serverVersion := GetLatestLauncherVersion(isBeta)

	var versionChanged bool
	if isBeta {
		versionChanged = versionFile.Beta != serverVersion
	} else {
		versionChanged = versionFile.Main != serverVersion
	}

	if versionChanged {
		if isBeta {
			versionFile.Beta = serverVersion
		} else {
			versionFile.Main = serverVersion
		}

		// marshal the version file
		versions, err = json.Marshal(versionFile)
		if err != nil {
			return false, err
		}

		// write the version file
		err = os.WriteFile(versionFilePath, versions, 0644)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func SpawnWatchdog(dir string, isBeta bool, hook string) {
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			changed, err := IsVersionChanged(dir, isBeta)
			if err != nil {
				log.Error(err)
			} else {
				if changed {
					log.Info("🎉 New version available!")

					resp, err := http.Get(hook)
					if err != nil {
						log.Error(err)
					} else {
						resp.Body.Close()
						if err != nil {
							log.Error(err)
						}
					}
				}
			}
		}
	}()
}
