package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/doduda/ui"
	mapping "github.com/dofusdude/dodumap"
)

func marshalSave(data interface{}, path string, indent string) {
	out, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	var outBytes []byte
	if indent != "" {
		outBytes, err = json.MarshalIndent(data, "", indent)
	} else {
		outBytes, err = json.Marshal(data)
	}
	if err != nil {
		log.Fatal(err)
	}

	out.Write(outBytes)
}

func detectRawDataMajorVersion(dir string) (int, error) {
	file, err := os.ReadFile(filepath.Join(dir, "areas.json"))
	if err != nil {
		fmt.Print(err)
	}
	var areasJson interface{}
	err = json.Unmarshal(file, &areasJson)
	if err != nil {
		return 0, err
	}

	if _, ok := areasJson.(map[string]interface{}); ok {
		return 3, nil
	} else if _, ok := areasJson.([]interface{}); ok {
		return 2, nil
	}

	return 0, errors.New("Could not detect major version of raw data")
}

func Map(dir string, indent string, persistenceDir string, release string, headless bool) {
	majorVersion, err := detectRawDataMajorVersion(dir)
	if err != nil {
		log.Fatal(err)
	}

	updatesChan := make(chan string)
	spinnerWg := sync.WaitGroup{}
	spinnerWg.Add(1)
	go func() {
		defer spinnerWg.Done()
		ui.Spinner("", updatesChan, false, headless)
	}()

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	updatesChan <- "⬇️ Persistence"
	err = mapping.LoadPersistedElements(persistenceDir, release, majorVersion)
	if err != nil {
		log.Fatal(err)
	}

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	updatesChan <- "Game data"

	if majorVersion == 2 {
		var gameData *mapping.JSONGameData
		var languageData map[string]mapping.LangDict

		gameData = mapping.ParseRawData(dir)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		updatesChan <- "Languages"
		languageData = mapping.ParseRawLanguages(dir)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Items 🧠"
		} else {
			updatesChan <- "Items " + ui.HelpStyle("mapping")
		}
		mappedItems := mapping.MapItems(gameData, &languageData)
		mappedItemPath := filepath.Join(dir, "MAPPED_ITEMS.json")
		marshalSave(mappedItems, mappedItemPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Mounts 🧠"
		} else {
			updatesChan <- "Mounts " + ui.HelpStyle("mapping")
		}
		mappedMounts := mapping.MapMounts(gameData, &languageData)
		mappedMountsPath := filepath.Join(dir, "MAPPED_MOUNTS.json")
		marshalSave(mappedMounts, mappedMountsPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Almanax 🧠"
		} else {
			updatesChan <- "Almanax " + ui.HelpStyle("mapping")
		}
		mappedAlmanax := mapping.MapAlmanax(gameData, &languageData)
		mappedAlmanaxPath := filepath.Join(dir, "MAPPED_ALMANAX.json")
		marshalSave(mappedAlmanax, mappedAlmanaxPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Sets 🧠"
		} else {
			updatesChan <- "Sets " + ui.HelpStyle("mapping")
		}
		mappedSets := mapping.MapSets(gameData, &languageData)
		mappedSetsPath := filepath.Join(dir, "MAPPED_SETS.json")
		marshalSave(mappedSets, mappedSetsPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Recipes 🧠"
		} else {
			updatesChan <- "Recipes " + ui.HelpStyle("mapping")
		}
		mappedRecipes := mapping.MapRecipes(gameData)
		mappedRecipesPath := filepath.Join(dir, "MAPPED_RECIPES.json")
		marshalSave(mappedRecipes, mappedRecipesPath, indent)
	} else if majorVersion == 3 {
		var gameData *mapping.JSONGameDataUnity
		var languageData map[string]mapping.LangDictUnity

		gameData = mapping.ParseRawDataUnity(dir)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		updatesChan <- "Languages"
		languageData = mapping.ParseRawLanguagesUnity(dir)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}

		if headless {
			updatesChan <- "Items 🧠"
		} else {
			updatesChan <- "Items " + ui.HelpStyle("mapping")
		}
		mappedItems := mapping.MapItemsUnity(gameData, &languageData)
		mappedItemPath := filepath.Join(dir, "MAPPED_ITEMS.json")
		marshalSave(mappedItems, mappedItemPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Mounts 🧠"
		} else {
			updatesChan <- "Mounts " + ui.HelpStyle("mapping")
		}
		mappedMounts := mapping.MapMountsUnity(gameData, &languageData)
		mappedMountsPath := filepath.Join(dir, "MAPPED_MOUNTS.json")
		marshalSave(mappedMounts, mappedMountsPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Almanax 🧠"
		} else {
			updatesChan <- "Almanax " + ui.HelpStyle("mapping")
		}
		mappedAlmanax := mapping.MapAlmanaxUnity(gameData, &languageData)
		mappedAlmanaxPath := filepath.Join(dir, "MAPPED_ALMANAX.json")
		marshalSave(mappedAlmanax, mappedAlmanaxPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Sets 🧠"
		} else {
			updatesChan <- "Sets " + ui.HelpStyle("mapping")
		}
		mappedSets := mapping.MapSetsUnity(gameData, &languageData)
		mappedSetsPath := filepath.Join(dir, "MAPPED_SETS.json")
		marshalSave(mappedSets, mappedSetsPath, indent)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		if headless {
			updatesChan <- "Recipes 🧠"
		} else {
			updatesChan <- "Recipes " + ui.HelpStyle("mapping")
		}
		mappedRecipes := mapping.MapRecipesUnity(gameData)
		mappedRecipesPath := filepath.Join(dir, "MAPPED_RECIPES.json")
		marshalSave(mappedRecipes, mappedRecipesPath, indent)
	} else {
		log.Fatal("Unsupported major version of raw data")
	}

	if persistenceDir != "" {
		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		updatesChan <- "💾"
		dofus3prefix := ""
		if majorVersion == 3 {
			dofus3prefix = ".dofus3"
		}

		releasePersist := release
		if releasePersist == "dofus3" {
			releasePersist = "main"
		}
		err := mapping.PersistElements(filepath.Join(persistenceDir, fmt.Sprintf("elements%s.%s.json", dofus3prefix, releasePersist)), filepath.Join(persistenceDir, fmt.Sprintf("item_types%s.%s.json", dofus3prefix, releasePersist)))
		if err != nil {
			log.Fatal(err)
		}
	}

	close(updatesChan)
	spinnerWg.Wait()
}
