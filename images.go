package main

import (
	"errors"
	"fmt"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func cleanImages(dir string, resolution int) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("Error accessing file", "err", err)
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".png" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Error("Error opening file", "err", err)
			return err
		}
		defer file.Close()

		img, err := png.DecodeConfig(file)
		if err != nil {
			log.Errorf("Error decoding image, skipping %s\n", path)
			return nil
		}

		if img.Width != resolution || img.Height != resolution {
			err = os.Remove(path)
			if err != nil {
				log.Errorf("Error removing file: %s\n", path)
				return err
			}
			return nil
		}

		if strings.Contains(info.Name(), "_") {
			oldPath := path
			path = filepath.Join(filepath.Dir(path), strings.Split(info.Name(), "_")[0]+".png")

			err = os.Rename(oldPath, path)
			if err != nil {
				log.Error("Renaming file failed", "err", err)
				return err
			}
		}

		sdPath := strings.Replace(path, ".png", fmt.Sprintf("-%d.png", resolution), 1)
		err = os.Rename(path, sdPath)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, bin int, version int, dir string, headless bool) error {
	inPath := filepath.Join(dir, "tmp")
	outPath := filepath.Join(dir, "img", "item")

	if version == 2 {
		fileNames := []HashFile{
			{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
			{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
			{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
			{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
			{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
		}

		if err := DownloadUnpackFiles("Item Bitmaps", bin, hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
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

		inPath = filepath.Join(dir, "tmp", "vector")
		outPath = filepath.Join(dir, "vector", "item")
		if err := DownloadUnpackFiles("Item Vectors", bin, hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
			return err
		}

		unpackD2pFolder("Item Vectors", inPath, outPath, headless)

		return nil
	} else if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_assets_1x.bundle", FriendlyName: "item_images_1.imagebundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_assets_2x.bundle", FriendlyName: "item_images_2.imagebundle"},
			//{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/mount_.bundle", FriendlyName: "mount_images.bundle"},
		}

		err := PullImages([]string{"stelzo/assetstudio-cli:latest"}, false, headless)
		if err != nil {
			return err
		}

		err = DownloadUnpackFiles("Images", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false)
		if err != nil {
			return err
		}

		feedbacks := make(chan string)

		var feedbackWg sync.WaitGroup
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Images", feedbacks, false, headless)
		}()

		defer func() {
			close(feedbacks)
			feedbackWg.Wait()
		}()

		feedbacks <- "cleaning"

		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "items", "1x"), filepath.Join(outPath, "1x"))
		if err != nil {
			return err
		}

		err = os.Rename(filepath.Join(outPath, "Assets", "BuiltAssets", "items", "2x"), filepath.Join(outPath, "2x"))
		if err != nil {
			return err
		}

		err = os.RemoveAll(filepath.Join(outPath, "Assets"))
		if err != nil {
			return err
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			if cleanImages(filepath.Join(outPath, "1x"), 64) != nil {
				log.Error("Error cleaning images", "err", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if cleanImages(filepath.Join(outPath, "2x"), 128) != nil {
				log.Error("Error cleaning images", "err", err)
			}
		}()

		wg.Wait()

		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
