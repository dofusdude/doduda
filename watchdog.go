package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func VersionChanged(dir string, gameVersion string, versionFilePath string, customBodyPath string, persistence bool, initialHook *bool) (bool, string, string, error) { // changed?, old version, new version, error
	var versionFile VersionFile

	if persistence {
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

	isBeta := gameVersion == "beta"
	serverVersion := GetLatestLauncherVersion(isBeta)
	serverVersion = serverVersion[4:] // removing updater version

	var versionChanged bool
	if isBeta {
		versionChanged = versionFile.Beta != serverVersion
	} else {
		versionChanged = versionFile.Main != serverVersion
	}

	if *initialHook {
		*initialHook = false
		return true, "-", serverVersion, nil
	}

	if versionChanged {
		var oldVersion string
		if isBeta {
			oldVersion = versionFile.Beta
			versionFile.Beta = serverVersion
		} else {
			oldVersion = versionFile.Main
			versionFile.Main = serverVersion
		}

		if persistence {
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

func SpawnWatchdog(endTimer chan bool, dir string, gameRelease string, hook string, token string, versionFilePath string, customBodyPath string, persistence bool, initialHook *bool, deadlyHook bool, interval uint32) *time.Ticker {
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	go func() {
		for range ticker.C {
			changed, oldVersion, newVersion, err := VersionChanged(dir, gameRelease, versionFilePath, customBodyPath, persistence, initialHook)

			if err != nil {
				log.Error(err)
			} else {
				if changed {
					log.Info(fmt.Sprintf("🎉 New %s version available: %s", gameRelease, newVersion))

					var isJson bool
					var body []byte
					if customBodyPath == "" {
						isJson = true
						jsonBody := map[string]string{
							"message":     "🎉 New Dofus version available!",
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

					if token != "" {
						req.Header.Set("Authorization", "Bearer "+token)
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
	}()

	return ticker
}
