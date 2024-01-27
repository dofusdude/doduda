package main

import (
	"encoding/json"
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

func Map(dir string, indent string, persistenceDir string, release string, headless bool) {

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
	updatesChan <- "Load persistence"
	err := mapping.LoadPersistedElements(persistenceDir, release)
	if err != nil {
		log.Fatal(err)
	}

	var gameData *mapping.JSONGameData
	var languageData map[string]mapping.LangDict

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	updatesChan <- "Game data"
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
		updatesChan <- "Items mapping"
	} else {
		updatesChan <- "Items " + ui.HelpStyle("mapping")
	}
	mappedItems := mapping.MapItems(gameData, &languageData)
	mappedItemPath := filepath.Join(dir, "data", "MAPPED_ITEMS.json")
	marshalSave(mappedItems, mappedItemPath, indent)

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	if headless {
		updatesChan <- "Mounts mapping"
	} else {
		updatesChan <- "Mounts " + ui.HelpStyle("mapping")
	}
	mappedMounts := mapping.MapMounts(gameData, &languageData)
	mappedMountsPath := filepath.Join(dir, "data", "MAPPED_MOUNTS.json")
	marshalSave(mappedMounts, mappedMountsPath, indent)

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	if headless {
		updatesChan <- "Almanax mapping"
	} else {
		updatesChan <- "Almanax " + ui.HelpStyle("mapping")
	}
	mappedAlmanax := mapping.MapAlmanax(gameData, &languageData)
	mappedAlmanaxPath := filepath.Join(dir, "data", "MAPPED_ALMANAX.json")
	marshalSave(mappedAlmanax, mappedAlmanaxPath, indent)

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	updatesChan <- "Sets " + ui.HelpStyle("mapping")
	mappedSets := mapping.MapSets(gameData, &languageData)
	mappedSetsPath := filepath.Join(dir, "data", "MAPPED_SETS.json")
	marshalSave(mappedSets, mappedSetsPath, indent)

	if isChannelClosed(updatesChan) {
		os.Exit(1)
	}
	if headless {
		updatesChan <- "Recipes mapping"
	} else {
		updatesChan <- "Recipes " + ui.HelpStyle("mapping")
	}
	mappedRecipes := mapping.MapRecipes(gameData)
	mappedRecipesPath := filepath.Join(dir, "data", "MAPPED_RECIPES.json")
	marshalSave(mappedRecipes, mappedRecipesPath, indent)

	if persistenceDir != "" {
		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		updatesChan <- "Persist"
		err := mapping.PersistElements(filepath.Join(persistenceDir, fmt.Sprintf("elements.%s.json", release)), filepath.Join(persistenceDir, fmt.Sprintf("item_types.%s.json", release)))
		if err != nil {
			log.Fatal(err)
		}
	}

	close(updatesChan)
	spinnerWg.Wait()
}
