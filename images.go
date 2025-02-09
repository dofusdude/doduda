package main

import (
	"errors"
	"fmt"
	"image"
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

func removeNumberSuffix(path string, f os.FileInfo, ending string) string {
	var pathNoNum string
	if strings.Contains(f.Name(), "_#") {
		splits := strings.Split(f.Name(), "_")
		joined_except_last := strings.Join(splits[:len(splits)-1], "_")
		pathNoNum = filepath.Join(filepath.Dir(path), joined_except_last+ending)
	} else {
		pathNoNum = path
	}

	return pathNoNum
}

func cleanImages(dir string, resolution *int) error {
	type fileStuff struct {
		path string
		img  image.Config
		info os.FileInfo
	}

	files := make(map[string][]fileStuff) // buffer for files

	// populate files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".png" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		img, err := png.DecodeConfig(file)
		if err != nil {
			return nil // probably not a png, don't care
		}

		// hard filter for resolution
		if resolution != nil {
			if img.Width != *resolution || img.Height != *resolution {
				err = os.Remove(path)
				if err != nil {
					return err
				}
				return nil
			}
		}

		cleanedName := removeNumberSuffix(path, info, ".png")

		if _, ok := files[cleanedName]; !ok {
			files[cleanedName] = make([]fileStuff, 0)
		}

		files[cleanedName] = append(files[cleanedName], fileStuff{path: path, img: img, info: info})

		return nil
	})

	if err != nil {
		return err
	}

	// mark inferior resolution files for deletion
	for _, f := range files {
		higestResIdx := 0
		for i, other := range f {
			if other.img.Width*other.img.Height > f[higestResIdx].img.Width*f[higestResIdx].img.Height {
				higestResIdx = i
			}
		}

		for i, other := range f {
			if i != higestResIdx {
				err = os.Remove(other.path)
				if err != nil {
					return err
				}
			}
		}
	}

	// rename rest to final using their resolution and remove the #number
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".png" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		img, err := png.DecodeConfig(file)
		if err != nil {
			return err
		}

		var finalResStr string
		if resolution == nil {
			var resStr string
			if img.Width == img.Height {
				resStr = strconv.Itoa(img.Width)
			} else {
				resStr = fmt.Sprintf("%dx%d", img.Width, img.Height)
			}
			finalResStr = resStr
		} else {
			finalResStr = fmt.Sprintf("%d", *resolution)
		}

		cleaned := removeNumberSuffix(path, info, ".png")
		sdPath := strings.Replace(cleaned, ".png", fmt.Sprintf("-%s.png", finalResStr), 1)
		err = os.Rename(path, sdPath)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func rename_or_rmfirst(from string, to string) error {
	if _, err := os.Stat(to); err == nil {
		err = os.RemoveAll(to)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(filepath.Dir(to)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(to), os.ModePerm)
		if err != nil {
			return err
		}
	}

	err := os.Rename(from, to)
	if err != nil {
		return err
	}

	return nil
}

func download_unpack_clean_dedup_multires(errorChan chan error, topic string, bin int, hashJson *ankabuffer.Manifest, dir string, outPath string, fileNames []HashFile, semaphore chan struct{}, feedbacks chan string, innerTopicPlural string, resolutionMap map[string]*int, ressubdirs ...string) {
	err := DownloadUnpackFiles(topic+" 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", true, true)
	if err != nil {
		errorChan <- err
	}

	<-semaphore

	innerWg := sync.WaitGroup{}
	if len(ressubdirs) == 0 {
		ressubdirs = append(ressubdirs, "") // hacky :)
	}

	for _, ressubdir := range ressubdirs {
		for res, resolution := range resolutionMap {
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				var fromPath string
				var toPath string
				if ressubdir == "" {
					fromPath = filepath.Join("Assets", "BuiltAssets", innerTopicPlural)
					toPath = filepath.Join(outPath, res)
				} else {
					fromPath = filepath.Join("Assets", "BuiltAssets", innerTopicPlural, ressubdir)
					toPath = filepath.Join(outPath, ressubdir, res)
				}
				from := filepath.Join(outPath, fromPath, res)
				if err := rename_or_rmfirst(from, toPath); err != nil {
					errorChan <- err
				}

				var cleanPath string
				if ressubdir == "" {
					cleanPath = filepath.Join(outPath, res)
				} else {
					cleanPath = filepath.Join(outPath, ressubdir, res)
				}

				if cleanImages(cleanPath, resolution) != nil {
					errorChan <- err
				}
			}()
		}
	}

	innerWg.Wait()
	err = os.RemoveAll(filepath.Join(outPath, "Assets"))
	if err != nil {
		errorChan <- err
	}
	feedbacks <- "✅ " + topic
}

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, bin int, maxConcurrentDownloads int, version int, dir string, ignore []string, headless bool) error {
	if version == 2 {
		fileNames := []HashFile{
			{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
			{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
			{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
			{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
			{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
		}
		inPath := filepath.Join(dir, "tmp")

		if !ignoresRegex(ignore, "images-items") {
			if err := DownloadUnpackFiles("Item Bitmaps", bin, hashJson, "main", fileNames, dir, inPath, false, "", headless, false); err != nil {
				return err
			}
			outPath := filepath.Join(dir, "img", "item")
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
		}

		return nil
	} else if version == 3 {

		// -- Worldmaps --
		if !ignoresRegex(ignore, "images-worldmaps") {
			outPath := filepath.Join(dir, "img", "worldmap")
			fileNames := []HashFile{
				{Filename: "REGEX:Dofus_Data/StreamingAssets/Content/Picto/Worldmaps/worldmap_assets__*", FriendlyName: "worldmap_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Worldmap 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			// not cleaning since names are not unique enough without #number
		}

		// -- UI Ornaments --
		if !ignoresRegex(ignore, "images-ui-ornaments") {
			outPath := filepath.Join(dir, "img", "ui", "ornaments")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/ornament_assets_all.bundle", FriendlyName: "ornament_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("UI Ornaments 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- UI Documents --
		if !ignoresRegex(ignore, "images-ui-documents") {
			outPath := filepath.Join(dir, "img", "ui", "documents")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/document_assets_all.bundle", FriendlyName: "document_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("UI Documents 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- UI Guidebook --
		if !ignoresRegex(ignore, "images-ui-guidebook") {
			outPath := filepath.Join(dir, "img", "ui", "guidebook")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/guidebook_assets_all.bundle", FriendlyName: "guidebook_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("UI Guidebook 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- UI house --
		if !ignoresRegex(ignore, "images-ui-house") {
			outPath := filepath.Join(dir, "img", "ui", "house")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/house_assets_all.bundle", FriendlyName: "house_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("UI House 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- UI illustrations --
		if !ignoresRegex(ignore, "images-ui-illustrations") {
			outPath := filepath.Join(dir, "img", "ui", "illustration")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/illus_assets_all.bundle", FriendlyName: "illustration_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("UI Illustration 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}
		}

		// -- Suggestion --
		if !ignoresRegex(ignore, "images-misc-suggestions") {
			outPath := filepath.Join(dir, "img", "misc", "suggestion")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/suggestion_assets_all.bundle", FriendlyName: "suggestion_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Suggestion 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- Icons --
		if !ignoresRegex(ignore, "images-misc-icons") {
			outPath := filepath.Join(dir, "img", "misc", "icon")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/icon_assets_all.bundle", FriendlyName: "icon_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Icon 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- Flag --
		if !ignoresRegex(ignore, "images-misc-flags") {
			outPath := filepath.Join(dir, "img", "misc", "flag")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/flag_assets_all.bundle", FriendlyName: "flag_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Flag 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- Guildrank --
		if !ignoresRegex(ignore, "images-misc-guildranks") {
			outPath := filepath.Join(dir, "img", "misc", "guildrank")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/guildrank_assets_all.bundle", FriendlyName: "guildrank_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Guildrank 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}

		}

		// -- Arena Categories --
		if !ignoresRegex(ignore, "images-misc-arena") {
			outPath := filepath.Join(dir, "img", "misc", "arena")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/arena_assets_all.bundle", FriendlyName: "arena_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Arena 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- Achievement Categories --
		if !ignoresRegex(ignore, "images-achievement_categories") {
			outPath := filepath.Join(dir, "img", "achievement_category")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/achievementcategory_assets_all.bundle", FriendlyName: "achievement_category_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Achievement Category 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- Achievements --
		if !ignoresRegex(ignore, "images-achievements") {
			outPath := filepath.Join(dir, "img", "achievement")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/achievement_assets_all.bundle", FriendlyName: "achievement_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Achievement 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}

			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		// -- spell states --
		if !ignoresRegex(ignore, "images-spell_states") {
			outPath := filepath.Join(dir, "img", "spell_state")
			fileNames := []HashFile{
				{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Spells/spellstate_assets_all.bundle", FriendlyName: "spell_state_images.imagebundle"},
			}

			if err := DownloadUnpackFiles("Spell State 🖼️", bin, hashJson, "picto", fileNames, dir, outPath, true, "", headless, false); err != nil {
				return err
			}
			if err := cleanImages(outPath, nil); err != nil {
				return err
			}
		}

		const totalDownloads = 14 // just to buffer, must be at least the number of go routines started below

		feedbacks := make(chan string)

		var feedbackWg sync.WaitGroup
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Multiscale 🖼️", feedbacks, false, headless)
		}()

		defer func() {
			close(feedbacks)
			feedbackWg.Wait()
		}()

		feedbacks <- "⬇️ in parallel"

		semaphore := make(chan struct{}, maxConcurrentDownloads)
		errorChan := make(chan error, totalDownloads)
		var wg sync.WaitGroup

		// -- items --
		if !ignoresRegex(ignore, "images-items") {
			wg.Add(1) // 1
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_assets_1x.bundle", FriendlyName: "item_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_assets_2x.bundle", FriendlyName: "item_images_2.imagebundle"},
				}

				lowRes := 64
				highRes := 128
				download_unpack_clean_dedup_multires(errorChan, "Items", bin, hashJson, dir, filepath.Join(dir, "img", "item"), fileNames, semaphore, feedbacks, "items", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- mounts --
		if !ignoresRegex(ignore, "images-mounts") {
			wg.Add(1) // 2
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/mount_assets_.bundle", FriendlyName: "mount_images.imagebundle"},
				}

				lowRes := 64
				highRes := 256
				download_unpack_clean_dedup_multires(errorChan, "Mounts", bin, hashJson, dir, filepath.Join(dir, "img", "mount"), fileNames, semaphore, feedbacks, "mounts", map[string]*int{"small": &lowRes, "big": &highRes})
			}()
		}

		// -- emotes --
		if !ignoresRegex(ignore, "images-emotes") {
			wg.Add(1) // 3
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emote_assets_1x.bundle", FriendlyName: "emote_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emote_assets_2x.bundle", FriendlyName: "emote_images_2.imagebundle"},
				}

				lowRes := 32
				highRes := 64
				download_unpack_clean_dedup_multires(errorChan, "Emotes", bin, hashJson, dir, filepath.Join(dir, "img", "emote"), fileNames, semaphore, feedbacks, "emotes", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- classes --
		if !ignoresRegex(ignore, "images-class_heads") {
			wg.Add(1) // 4
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/class_assets_1x.bundle", FriendlyName: "class_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/class_assets_2x.bundle", FriendlyName: "class_images_2.imagebundle"},
				}

				lowRes := 32
				highRes := 64
				download_unpack_clean_dedup_multires(errorChan, "Class Heads", bin, hashJson, dir, filepath.Join(dir, "img", "class_head"), fileNames, semaphore, feedbacks, filepath.Join("classes", "heads", "small"), map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- alignment --
		if !ignoresRegex(ignore, "images-alignment") {
			wg.Add(1) // 5
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/alignment_assets_1x.bundle", FriendlyName: "alignment_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/alignment_assets_2x.bundle", FriendlyName: "alignment_images_2.imagebundle"},
				}

				download_unpack_clean_dedup_multires(errorChan, "Alignment", bin, hashJson, dir, filepath.Join(dir, "img", "alignment"), fileNames, semaphore, feedbacks, "alignments", map[string]*int{"1x": nil, "2x": nil})
			}()
		}

		// -- challenge --
		if !ignoresRegex(ignore, "images-challenges") {
			wg.Add(1) // 6
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/challenge_assets_1x.bundle", FriendlyName: "challenge_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/challenge_assets_2x.bundle", FriendlyName: "challenge_images_2.imagebundle"},
				}

				lowRes := 32
				highRes := 64
				download_unpack_clean_dedup_multires(errorChan, "Challenges", bin, hashJson, dir, filepath.Join(dir, "img", "challenge"), fileNames, semaphore, feedbacks, "challenges", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- companion --
		if !ignoresRegex(ignore, "images-companions") {
			wg.Add(1) // 7
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/companion_assets_1x.bundle", FriendlyName: "companion_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/companion_assets_2x.bundle", FriendlyName: "companion_images_2.imagebundle"},
				}

				lowRes := 84
				highRes := 168
				download_unpack_clean_dedup_multires(errorChan, "Companions", bin, hashJson, dir, filepath.Join(dir, "img", "companion"), fileNames, semaphore, feedbacks, "companions", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- cosmetic --
		if !ignoresRegex(ignore, "images-cosmetics") {
			wg.Add(1) // 8
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/cosmetic_assets_1x.bundle", FriendlyName: "cosmetic_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/cosmetic_assets_2x.bundle", FriendlyName: "cosmetic_images_2.imagebundle"},
				}

				lowRes := 64
				highRes := 128
				download_unpack_clean_dedup_multires(errorChan, "Cosmetics", bin, hashJson, dir, filepath.Join(dir, "img", "cosmetic"), fileNames, semaphore, feedbacks, "cosmetics", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- smileys --
		if !ignoresRegex(ignore, "images-smileys") {
			wg.Add(1) // 9
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/smiley_assets_1x.bundle", FriendlyName: "smiley_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/smiley_assets_2x.bundle", FriendlyName: "smiley_images_2.imagebundle"},
				}

				lowRes := 32
				highRes := 64
				download_unpack_clean_dedup_multires(errorChan, "Smileys", bin, hashJson, dir, filepath.Join(dir, "img", "smiley"), fileNames, semaphore, feedbacks, "smilies", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- jobs --
		if !ignoresRegex(ignore, "images-jobs") {
			wg.Add(1) // 10
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/job_assets_1x.bundle", FriendlyName: "job_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/job_assets_2x.bundle", FriendlyName: "job_images_2.imagebundle"},
				}

				lowRes := 32
				highRes := 64
				download_unpack_clean_dedup_multires(errorChan, "Jobs", bin, hashJson, dir, filepath.Join(dir, "img", "job"), fileNames, semaphore, feedbacks, "jobs", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- emblem --
		if !ignoresRegex(ignore, "images-emblems") {
			wg.Add(1) // 11
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emblem_assets_1x.bundle", FriendlyName: "emblem_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/emblem_assets_2x.bundle", FriendlyName: "emblem_images_2.imagebundle"},
				}

				lowRes := 64
				highRes := 128
				download_unpack_clean_dedup_multires(errorChan, "Emblems", bin, hashJson, dir, filepath.Join(dir, "img", "emblem"), fileNames, semaphore, feedbacks, filepath.Join("emblems", "big"), map[string]*int{"1x": &lowRes, "2x": &highRes}, "backcontent", "outlinealliance", "outlineguild", "up")
			}()
		}

		// -- monsters --
		if !ignoresRegex(ignore, "images-monsters") {
			wg.Add(1) // 12
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Monsters/monster_assets_1x.bundle", FriendlyName: "monster_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Monsters/monster_assets_2x.bundle", FriendlyName: "monster_images_2.imagebundle"},
				}

				lowRes := 64
				highRes := 128
				download_unpack_clean_dedup_multires(errorChan, "Monsters", bin, hashJson, dir, filepath.Join(dir, "img", "monster"), fileNames, semaphore, feedbacks, "monsters", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- spells --
		if !ignoresRegex(ignore, "images-spells") {
			wg.Add(1) // 13
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Spells/spell_assets_1x.bundle", FriendlyName: "spell_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Spells/spell_assets_2x.bundle", FriendlyName: "spell_images_2.imagebundle"},
				}

				lowRes := 48
				highRes := 96
				download_unpack_clean_dedup_multires(errorChan, "Spells", bin, hashJson, dir, filepath.Join(dir, "img", "spell"), fileNames, semaphore, feedbacks, "spells", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		// -- preset (statistics) --
		if !ignoresRegex(ignore, "images-statistics") {
			wg.Add(1) // 14
			go func() {
				defer wg.Done()
				semaphore <- struct{}{}
				fileNames := []HashFile{
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/preset_assets_1x.bundle", FriendlyName: "preset_images_1.imagebundle"},
					{Filename: "Dofus_Data/StreamingAssets/Content/Picto/UI/preset_assets_2x.bundle", FriendlyName: "preset_images_2.imagebundle"},
				}

				lowRes := 48
				highRes := 96
				download_unpack_clean_dedup_multires(errorChan, "Statistics", bin, hashJson, dir, filepath.Join(dir, "img", "statistics"), fileNames, semaphore, feedbacks, "presets", map[string]*int{"1x": &lowRes, "2x": &highRes})
			}()
		}

		go func() {
			wg.Wait()
			close(errorChan)
		}()

		for err := range errorChan {
			if err != nil {
				return err
			}
		}

		wg.Wait()

		return nil
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
