package main

import (
	"errors"
	"path/filepath"
	"strconv"

	"github.com/dofusdude/ankabuffer"
)

func DownloadLanguageFiles(hashJson *ankabuffer.Manifest, bin int, version int, lang string, dir string, indent string, headless bool) error {
	destPath := filepath.Join(dir, "languages")

	if version == 2 {
		var langFile HashFile
		langFile.Filename = "data/i18n/i18n_" + lang + ".d2i"
		langFile.FriendlyName = lang + ".d2i"
		err := DownloadUnpackFiles(lang, bin, hashJson, "lang_"+lang, []HashFile{langFile}, dir, destPath, true, indent, headless, false)
		return err
	} else if version == 3 {
		langFile := HashFile{
			Filename:     "Dofus_Data/StreamingAssets/Content/I18n/" + lang + ".bin",
			FriendlyName: lang + ".bin",
		}
		err := DownloadUnpackFiles(lang, bin, hashJson, "i18n", []HashFile{langFile}, dir, destPath, false, indent, headless, false)
		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}

func DownloadLanguages(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	langs := []string{"fr", "en", "es", "de", "it", "pt"}

	for _, lang := range langs {
		err := DownloadLanguageFiles(hashJson, bin, version, lang, dir, indent, headless)
		if err != nil {
			return err
		}
	}
	return nil
}
