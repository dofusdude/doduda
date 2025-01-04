package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

type VersionFile struct {
	Main   string `json:"main"`
	Dofus3 string `json:"dofus3"`
	Beta   string `json:"beta"`
}

func VersionChanged(dir string, gameVersion string, versionFilePath string, customBodyPath string, volatile bool, initialHook *bool) (bool, string, string, error) { // changed?, old version, new version, error
	var versionFile VersionFile

	if !volatile {
		if _, err := os.Stat(versionFilePath); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(versionFilePath), 0755)
			if err != nil {
				return false, "", "", err
			}
			versionFile = VersionFile{
				Main: "",
				Beta: "",
			}
		} else {
			versions, err := os.ReadFile(versionFilePath)
			if err != nil {
				return false, "", "", err
			}

			err = json.Unmarshal(versions, &versionFile)
			if err != nil {
				return false, "", "", err
			}
		}
	} else {
		versionFile.Beta = "-"
		versionFile.Main = "-"
	}

	serverVersion := GetLatestLauncherVersion(gameVersion)
	serverVersion = serverVersion[4:] // removing updater version

	var versionChanged bool
	switch gameVersion {
	case "main":
		versionChanged = versionFile.Main != serverVersion
	case "dofus3":
		versionChanged = versionFile.Dofus3 != serverVersion
	case "beta":
		versionChanged = versionFile.Beta != serverVersion
	}

	if *initialHook {
		*initialHook = false
		return true, "-", serverVersion, nil
	}

	if versionChanged {
		var oldVersion string

		switch gameVersion {
		case "main":
			oldVersion = versionFile.Main
			versionFile.Main = serverVersion
		case "dofus3":
			oldVersion = versionFile.Dofus3
			versionFile.Dofus3 = serverVersion
		case "beta":
			oldVersion = versionFile.Beta
			versionFile.Beta = serverVersion
		}

		if !volatile {
			versionsCurrent, err := json.Marshal(versionFile)
			if err != nil {
				return false, oldVersion, serverVersion, err
			}

			err = os.WriteFile(versionFilePath, versionsCurrent, 0644)
			if err != nil {
				return false, oldVersion, serverVersion, err
			}
		}

		if oldVersion == "" { // first time, just save the file
			return false, oldVersion, serverVersion, nil
		}

		return true, oldVersion, serverVersion, nil
	}

	return false, serverVersion, serverVersion, nil
}

func watchdogTick(endTimer chan bool, dir string, gameRelease string, versionFilePath string, customBodyPath string, volatile bool, initialHook *bool, hook string, authHeader string, deadlyHook bool) {
	changed, oldVersion, newVersion, err := VersionChanged(dir, gameRelease, versionFilePath, customBodyPath, volatile, initialHook)

	if err != nil {
		log.Error(err)
	} else {
		if changed {
			message := fmt.Sprintf("🎉 Dofus %s version %s available!", gameRelease, newVersion)
			log.Info(message)

			var isJson bool
			var body []byte
			if customBodyPath == "" {
				isJson = true
				jsonBody := map[string]string{
					"message":     message,
					"old_version": oldVersion,
					"new_version": newVersion,
					"release":     gameRelease,
				}

				body, err = json.Marshal(jsonBody)
				if err != nil {
					log.Error(err)
				}
			} else {
				if filepath.Ext(customBodyPath) == ".json" {
					isJson = true
				}

				body, err = os.ReadFile(customBodyPath)
				if err != nil {
					log.Error(err)
				}

				body = []byte(os.Expand(string(body), func(key string) string {
					switch key {
					case "oldVersion":
						return oldVersion
					case "newVersion":
						return newVersion
					case "release":
						return gameRelease
					default:
						return ""
					}
				}))
			}

			req, err := http.NewRequest("POST", hook, bytes.NewBuffer(body))
			if err != nil {
				log.Error(err)
			}

			if isJson {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", "text/plain")
			}

			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Error(err)
			} else {
				err = resp.Body.Close()
				if err != nil {
					log.Error(err)
				} else {
					if deadlyHook {
						endTimer <- true
					}
				}
			}
		}
	}
}
