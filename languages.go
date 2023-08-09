package main

import (
	"path/filepath"

	"github.com/dofusdude/ankabuffer"
)

func DownloadLanguageFiles(hashJson *ankabuffer.Manifest, lang string, dir string, indent string, headless bool) error {
	var langFile HashFile
	langFile.Filename = "data/i18n/i18n_" + lang + ".d2i"
	langFile.FriendlyName = lang + ".d2i"
	destPath := filepath.Join(dir, "data", "languages")
	if err := DownloadUnpackFiles(lang, hashJson, "lang_"+lang, []HashFile{langFile}, dir, destPath, true, indent, headless, false); err != nil {
		return err
	}
	return nil
}

func DownloadLanguages(hashJson *ankabuffer.Manifest, dir string, indent string, headless bool) error {
	langs := []string{"fr", "en", "es", "de", "it", "pt"}

	for _, lang := range langs {
		err := DownloadLanguageFiles(hashJson, lang, dir, indent, headless)
		if err != nil {
			return err
		}
	}
	return nil
}
