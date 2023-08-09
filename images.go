package main

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/dofusdude/ankabuffer"
	"github.com/dofusdude/doduda/ui"
	"github.com/dofusdude/doduda/unpack"
)

func unpackD2pFolder(title string, inPath string, outPath string, headless bool) {
	files := []string{}
	filepath.Walk(inPath, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".d2p" {
			files = append(files, path)
		}
		return nil
	})

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		os.MkdirAll(outPath, os.ModePerm)
	}

	updateProgress := make(chan bool, len(files))
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ui.Progress("Unpack "+title, len(files), updateProgress, 0, true, headless)
	}()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		converted := unpack.NewD2P(f).GetFiles()
		for filename, specs := range converted {
			outFile := filepath.Join(outPath, filename)

			if filepath.Ext(filename) == ".swl" {
				log.Warnf("can not unpack swl file %s", filename)
			}

			f, err := os.Create(outFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_, err = f.Write(specs["binary"].([]byte))
			if err != nil {
				log.Fatal(err)
			}
			if isChannelClosed(updateProgress) {
				os.Exit(1)
			}
		}
		updateProgress <- true
	}

	wg.Wait()
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, dir string, headless bool) error {
	fileNames := []HashFile{
		{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
		{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
		{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
		{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
		{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
	}

	inPath := filepath.Join(dir, "data", "tmp")
	outPath := filepath.Join(dir, "data", "img", "item")
	if err := DownloadUnpackFiles("Item Bitmaps", hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
		return err
	}

	unpackD2pFolder("Item Bitmaps", inPath, outPath, headless)

	fileNames = []HashFile{
		{Filename: "content/gfx/items/vector0.d2p", FriendlyName: "vector_0.d2p"},
		{Filename: "content/gfx/items/vector0_1.d2p", FriendlyName: "vector_1.d2p"},
		{Filename: "content/gfx/items/vector1.d2p", FriendlyName: "vector_2.d2p"},
		{Filename: "content/gfx/items/vector1_1.d2p", FriendlyName: "vector_3.d2p"},
		{Filename: "content/gfx/items/vector1_2.d2p", FriendlyName: "vector_4.d2p"},
	}

	inPath = filepath.Join(dir, "data", "tmp", "vector")
	outPath = filepath.Join(dir, "data", "vector", "item")
	if err := DownloadUnpackFiles("Item Vectors", hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
		return err
	}

	unpackD2pFolder("Item Vectors", inPath, outPath, headless)

	return nil
}
