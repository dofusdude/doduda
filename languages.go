package main

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/ankabuffer"
)

func DownloadLanguageFiles(hashJson *ankabuffer.Manifest, lang string, dir string, indent string) error {
	var langFile HashFile
	langFile.Filename = "data/i18n/i18n_" + lang + ".d2i"
	langFile.FriendlyName = lang + ".d2i"
	destPath := filepath.Join(dir, "data", "languages")
	if err := DownloadUnpackFiles(hashJson, "lang_"+lang, []HashFile{langFile}, dir, destPath, true, indent); err != nil {
		return err
	}
	return nil
}

func DownloadLanguages(hashJson *ankabuffer.Manifest, dir string, indent string) error {
	log.Info("Downloading languages...")
	langs := []string{"fr", "en", "es", "de", "it", "pt"}

	fail := make(chan error)
	for _, lang := range langs {
		go func(lang string, fail chan error, dir string, indent string) {
			fail <- DownloadLanguageFiles(hashJson, lang, dir, indent)
		}(lang, fail, dir, indent)
	}

	var someFail error
	for _, lang := range langs {
		if err := <-fail; err != nil {
			someFail = err
		}
		log.Info(fmt.Sprintf("%s ✅", lang))
	}

	log.Info(".. done languages.")

	return someFail
}
