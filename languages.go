package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/dofusdude/ankabuffer"
	"github.com/dofusdude/doduda/ui"
)

// TODO remove release when native implementation for .bin files exists.
func DownloadLanguageFiles(release string, hashJson *ankabuffer.Manifest, bin int, version int, lang string, dir string, indent string, headless bool) error {
	destPath := filepath.Join(dir, "languages")

	if version == 2 {
		var langFile HashFile
		langFile.Filename = "data/i18n/i18n_" + lang + ".d2i"
		langFile.FriendlyName = lang + ".d2i"
		err := DownloadUnpackFiles(lang, bin, hashJson, "lang_"+lang, []HashFile{langFile}, dir, destPath, true, indent, headless, false)
		return err
	} else if version == 3 {
		// TODO remove gh api part when native implementation for .json files exists.
		feedbacks := make(chan string)

		var feedbackWg sync.WaitGroup
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Languages", feedbacks, false, headless)
		}()

		feedbacks <- "searching"

		ghUrl := fmt.Sprintf("https://api.github.com/repos/dofusdude/dofus3-lang-%s/releases/latest", release)
		releaseApiResponse, err := http.Get(ghUrl)
		if err != nil {
			return err
		}

		releaseApiResponseBody, err := io.ReadAll(releaseApiResponse.Body)
		if err != nil {
			return err
		}

		var v map[string]interface{}
		err = json.Unmarshal(releaseApiResponseBody, &v)
		if err != nil {
			return err
		}

		assets := v["assets"].([]interface{})
		found := false
		for _, asset := range assets {
			assetMap := asset.(map[string]interface{})
			if assetMap["name"].(string) == fmt.Sprintf("%s.i18n.json", lang) {
				found = true
				assetUrl := assetMap["browser_download_url"].(string)

				feedbacks <- "loading " + lang

				imagesResponse, err := http.Get(assetUrl)
				if err != nil {
					return err
				}

				file := filepath.Join(destPath, lang+".json")
				out, err := os.Create(file)
				if err != nil {
					return err
				}

				_, err = io.Copy(out, imagesResponse.Body)
				if err != nil {
					return err
				}
			}
		}

		if !found {
			close(feedbacks)
			feedbackWg.Wait()
			return errors.New("Could not find the specified file in the latest release")
		}

		close(feedbacks)
		feedbackWg.Wait()

		/*langFile := HashFile{
			Filename:     "Dofus_Data/StreamingAssets/Content/I18n/" + lang + ".bin",
			FriendlyName: lang + ".bin",
		}
		err := DownloadUnpackFiles(lang, bin, hashJson, "i18n", []HashFile{langFile}, dir, destPath, false, indent, headless, false)*/
		return nil
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}

func DownloadLanguages(release string, hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	var langs []string
	if version == 2 {
		langs = []string{"fr", "en", "es", "de", "it", "pt"}
	} else if version == 3 {
		langs = []string{"fr", "en", "es", "de", "pt"}
	}

	for _, lang := range langs {
		err := DownloadLanguageFiles(release, hashJson, bin, version, lang, dir, indent, headless)
		if err != nil {
			return err
		}
	}
	return nil
}
