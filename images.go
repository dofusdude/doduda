package main

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/dofusdude/ankabuffer"
	"github.com/dofusdude/doduda/unpack"
)

func unpackD2pFolder(inPath string, outPath string) {
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
		}

	}
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, dir string) error {
	fileNames := []HashFile{
		{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
		{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
		{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
		{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
		{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
	}

	inPath := filepath.Join(dir, "data", "tmp")
	outPath := filepath.Join(dir, "data", "img", "item")
	if err := DownloadUnpackFiles(hashJson, "main", fileNames, dir, inPath, false, ""); err != nil {
		return err
	}

	unpackD2pFolder(inPath, outPath)

	fileNames = []HashFile{
		{Filename: "content/gfx/items/vector0.d2p", FriendlyName: "vector_0.d2p"},
		{Filename: "content/gfx/items/vector0_1.d2p", FriendlyName: "vector_1.d2p"},
		{Filename: "content/gfx/items/vector1.d2p", FriendlyName: "vector_2.d2p"},
		{Filename: "content/gfx/items/vector1_1.d2p", FriendlyName: "vector_3.d2p"},
		{Filename: "content/gfx/items/vector1_2.d2p", FriendlyName: "vector_4.d2p"},
	}

	inPath = filepath.Join(dir, "data", "tmp", "vector")
	outPath = filepath.Join(dir, "data", "vector", "item")
	if err := DownloadUnpackFiles(hashJson, "main", fileNames, dir, inPath, false, ""); err != nil {
		return err
	}

	unpackD2pFolder(inPath, outPath)

	return nil
}
