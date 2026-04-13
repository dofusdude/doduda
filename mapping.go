package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"charm.land/log/v2"
	"github.com/dofusdude/doduda/ui"
	mapping "github.com/dofusdude/dodumap"
)

func marshalSave(data any, path string, indent string) {
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
	var areasJson any
	err = json.Unmarshal(file, &areasJson)
	if err != nil {
		return 0, err
	}

	if _, ok := areasJson.(map[string]any); ok {
		return 3, nil
	} else if _, ok := areasJson.([]any); ok {
		return 2, nil
	}

	return 0, errors.New("Could not detect major version of raw data")
}

// normalizeUnityRIDTypes copies JSON files from dir to dstDir, rewriting any
// "rid" numeric values to strings required by the mapping package. The source
// files in dir are left untouched.
func normalizeUnityRIDTypes(dir string, dstDir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if strings.HasPrefix(entry.Name(), "MAPPED_") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		decoder := json.NewDecoder(bytes.NewReader(content))
		decoder.UseNumber()

		var data any
		if err := decoder.Decode(&data); err != nil {
			return fmt.Errorf("decode %s: %w", entry.Name(), err)
		}

		var output []byte
		if normalizeRIDValue(data) {
			output, err = json.Marshal(data)
			if err != nil {
				return fmt.Errorf("marshal %s: %w", entry.Name(), err)
			}
		} else {
			output = content
		}

		dstPath := filepath.Join(dstDir, entry.Name())
		if err := os.WriteFile(dstPath, output, os.ModePerm); err != nil {
			return fmt.Errorf("write %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func normalizeRIDValue(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		changed := false
		for key, child := range v {
			if key == "rid" {
				if normalized, ok := normalizeRIDScalar(child); ok {
					v[key] = normalized
					changed = true
					child = normalized
				}
			}
			if normalizeRIDValue(child) {
				changed = true
			}
		}
		return changed
	case []any:
		changed := false
		for i := range v {
			if normalizeRIDValue(v[i]) {
				changed = true
			}
		}
		return changed
	default:
		return false
	}
}

func normalizeRIDScalar(value any) (string, bool) {
	switch rid := value.(type) {
	case string:
		return "", false
	case json.Number:
		return rid.String(), true
	case int:
		return strconv.FormatInt(int64(rid), 10), true
	case int8:
		return strconv.FormatInt(int64(rid), 10), true
	case int16:
		return strconv.FormatInt(int64(rid), 10), true
	case int32:
		return strconv.FormatInt(int64(rid), 10), true
	case int64:
		return strconv.FormatInt(rid, 10), true
	case uint:
		return strconv.FormatUint(uint64(rid), 10), true
	case uint8:
		return strconv.FormatUint(uint64(rid), 10), true
	case uint16:
		return strconv.FormatUint(uint64(rid), 10), true
	case uint32:
		return strconv.FormatUint(uint64(rid), 10), true
	case uint64:
		return strconv.FormatUint(rid, 10), true
	case float32:
		floatValue := float64(rid)
		if math.Trunc(floatValue) == floatValue {
			return strconv.FormatInt(int64(floatValue), 10), true
		}
		return strconv.FormatFloat(floatValue, 'f', -1, 32), true
	case float64:
		if math.Trunc(rid) == rid {
			return strconv.FormatInt(int64(rid), 10), true
		}
		return strconv.FormatFloat(rid, 'f', -1, 64), true
	default:
		return "", false
	}
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

	switch majorVersion {
	case 2:
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
	case 3:
		var gameData *mapping.JSONGameDataUnity
		var languageData map[string]mapping.LangDictUnity

		tmpDir, err := os.MkdirTemp("", "doduda-unity-*")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		if err := normalizeUnityRIDTypes(dir, tmpDir); err != nil {
			log.Fatal(err)
		}

		gameData = mapping.ParseRawDataUnity(tmpDir)

		if isChannelClosed(updatesChan) {
			os.Exit(1)
		}
		updatesChan <- "Languages"
		languageData = mapping.ParseRawLanguagesUnity(tmpDir)
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
	default:
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
