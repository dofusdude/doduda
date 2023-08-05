package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/ankabuffer"
	"github.com/emirpasic/gods/maps/treebidimap"
	gutils "github.com/emirpasic/gods/utils"
)

type PersistentStringKeysMap struct {
	Entries *treebidimap.Map `json:"entries"`
	NextId  int              `json:"next_id"`
}

var (
	PersistedElements PersistentStringKeysMap
	PersistedTypes    PersistentStringKeysMap
)

func LoadPersistedElements(dir string) error {
	var element_path string
	var item_type_path string

	element_path = filepath.Join(dir, "persistent", "elements.json")
	item_type_path = filepath.Join(dir, "persistent", "item_types.json")

	data, err := os.ReadFile(element_path)
	if err != nil {
		return err
	}

	var elements []string
	err = json.Unmarshal(data, &elements)
	if err != nil {
		fmt.Println(err)
	}

	PersistedElements = PersistentStringKeysMap{
		Entries: treebidimap.NewWith(gutils.IntComparator, gutils.StringComparator),
		NextId:  0,
	}

	for _, entry := range elements {
		PersistedElements.Entries.Put(PersistedElements.NextId, entry)
		PersistedElements.NextId++
	}

	data, err = os.ReadFile(item_type_path)
	if err != nil {
		return err
	}

	var types []string
	err = json.Unmarshal(data, &types)
	if err != nil {
		fmt.Println(err)
	}

	PersistedTypes = PersistentStringKeysMap{
		Entries: treebidimap.NewWith(gutils.IntComparator, gutils.StringComparator),
		NextId:  0,
	}

	for _, entry := range types {
		PersistedTypes.Entries.Put(PersistedTypes.NextId, entry)
		PersistedTypes.NextId++
	}

	return nil
}

func PersistElements(elementPath string, itemTypePath string) error {
	elements := make([]string, PersistedElements.NextId)
	it := PersistedElements.Entries.Iterator()
	for it.Next() {
		elements[it.Key().(int)] = it.Value().(string)
	}

	elementsJson, err := json.MarshalIndent(elements, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(elementPath, elementsJson, 0644)
	if err != nil {
		return err
	}

	types := make([]string, PersistedTypes.NextId)
	it = PersistedTypes.Entries.Iterator()
	for it.Next() {
		types[it.Key().(int)] = it.Value().(string)
	}

	typesJson, err := json.MarshalIndent(types, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(itemTypePath, typesJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Parse(dir string, indentFlag bool) {
	log.Info("Parsing...")

	var indent string
	if indentFlag {
		indent = "    "
	} else {
		indent = ""
	}

	startParsing := time.Now()
	gameData := ParseRawData(dir)
	languageData := ParseRawLanguages(dir)
	log.Infof("... %.2fs", time.Since(startParsing).Seconds())

	startMapping := time.Now()
	err := LoadPersistedElements(dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Mapping...")
	mappedItems := MapItems(gameData, &languageData)
	mappedItemPath := filepath.Join(dir, "data", "MAPPED_ITEMS.json")
	out, err := os.Create(mappedItemPath)
	if err != nil {
		fmt.Println(err)
	}
	defer out.Close()

	outBytes, err := json.MarshalIndent(mappedItems, "", indent)
	if err != nil {
		fmt.Println(err)
		return
	}

	out.Write(outBytes)
	log.Infof("📁 %s", mappedItemPath)

	mappedMounts := MapMounts(gameData, &languageData)
	mappedMountsPath := filepath.Join(dir, "data", "MAPPED_MOUNTS.json")
	out, err = os.Create(mappedMountsPath)
	if err != nil {
		fmt.Println(err)
	}
	defer out.Close()

	outBytes, err = json.MarshalIndent(mappedMounts, "", indent)
	if err != nil {
		fmt.Println(err)
		return
	}

	out.Write(outBytes)
	log.Infof("📁 %s", mappedMountsPath)

	mappedSets := MapSets(gameData, &languageData)
	mappedSetsPath := filepath.Join(dir, "data", "MAPPED_SETS.json")
	outSets, err := os.Create(mappedSetsPath)
	if err != nil {
		fmt.Println(err)
	}
	defer outSets.Close()

	outSetsBytes, err := json.MarshalIndent(mappedSets, "", indent)
	if err != nil {
		fmt.Println(err)
		return
	}

	outSets.Write(outSetsBytes)
	log.Infof("📁 %s", mappedSetsPath)

	mappedRecipes := MapRecipes(gameData)
	var outRecipes *os.File
	mappedRecipesPath := filepath.Join(dir, "data", "MAPPED_RECIPES.json")
	outRecipes, err = os.Create(mappedRecipesPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outRecipes.Close()

	var outRecipeBytes []byte
	outRecipeBytes, err = json.MarshalIndent(mappedRecipes, "", indent)
	if err != nil {
		log.Fatal(err)
	}

	outRecipes.Write(outRecipeBytes)
	log.Infof("📁 %s", mappedRecipesPath)

	err = PersistElements(filepath.Join(dir, "persistent", "elements.json"), filepath.Join(dir, "persistent", "item_types.json"))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("... %.2fs", time.Since(startMapping).Seconds())
	mappedSets = nil
	mappedItems = nil
}

func DownloadMountImageWorker(manifest *ankabuffer.Manifest, fragment string, dir string, workerSlice []JSONGameMount) {
	wg := sync.WaitGroup{}

	for _, mount := range workerSlice {
		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer wg.Done()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.png", mountId)
			image.FriendlyName = fmt.Sprintf("%d.png", mountId)
			outPath := filepath.Join(dir, "data/img/mount")
			_ = DownloadUnpackFiles(manifest, fragment, []HashFile{image}, dir, outPath, true)
		}(mount.Id, &wg, dir)

		//  Missing bundle for content/gfx/mounts/162.swf
		wg.Add(1)
		go func(mountId int, wg *sync.WaitGroup, dir string) {
			defer wg.Done()
			var image HashFile
			image.Filename = fmt.Sprintf("content/gfx/mounts/%d.swf", mountId)
			image.FriendlyName = fmt.Sprintf("%d.swf", mountId)
			outPath := filepath.Join(dir, "data/vector/mount")
			_ = DownloadUnpackFiles(manifest, fragment, []HashFile{image}, dir, outPath, false)
		}(mount.Id, &wg, dir)
	}

	wg.Wait()
}

func PartitionSlice[T any](items []T, parts int) (chunks [][]T) {
	var divided [][]T

	chunkSize := (len(items) + parts - 1) / parts

	for i := 0; i < len(items); i += chunkSize {
		end := i + chunkSize

		if end > len(items) {
			end = len(items)
		}

		divided = append(divided, items[i:end])
	}

	return divided
}

// https://stackoverflow.com/questions/13422578/in-go-how-to-get-a-slice-of-values-from-a-map
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

func DownloadMountsImages(mounts *JSONGameData, hashJson *ankabuffer.Manifest, worker int, dir string) {
	arr := Values(mounts.Mounts)
	workerSlices := PartitionSlice(arr, worker)

	wg := sync.WaitGroup{}
	for _, workerSlice := range workerSlices {
		wg.Add(1)
		go func(workerSlice []JSONGameMount, dir string) {
			defer wg.Done()
			DownloadMountImageWorker(hashJson, "main", dir, workerSlice)
		}(workerSlice, dir)
	}
	wg.Wait()
}

func isActiveEffect(name map[string]string) bool {
	regex := regexp.MustCompile(`^\(.*\)$`)
	if regex.Match([]byte(name["en"])) {
		return true
	}
	if strings.Contains(name["de"], "(Ziel)") {
		return true
	}
	return false
}

func ParseEffects(data *JSONGameData, allEffects [][]JSONGameItemPossibleEffect, langs *map[string]LangDict) [][]MappedMultilangEffect {
	var mappedAllEffects [][]MappedMultilangEffect
	for _, effects := range allEffects {
		var mappedEffects []MappedMultilangEffect
		for _, effect := range effects {

			var mappedEffect MappedMultilangEffect
			currentEffect := data.effects[effect.EffectId]

			numIsSpell := false
			if strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zauberspruchs #1") || strings.Contains((*langs)["de"].Texts[currentEffect.DescriptionId], "Zaubers #1") {
				numIsSpell = true
			}

			mappedEffect.Type = make(map[string]string)
			mappedEffect.Templated = make(map[string]string)
			var minMaxRemove int
			for _, lang := range Languages {
				var diceNum int
				var diceSide int
				var value int

				diceNum = effect.MinimumValue

				diceSide = effect.MaximumValue

				value = effect.Value

				effectName := (*langs)[lang].Texts[currentEffect.DescriptionId]
				if lang == "de" {
					effectName = strings.ReplaceAll(effectName, "{~ps}{~zs}", "") // german has error in template
				}

				if effectName == "#1" { // is spell description from dicenum 1
					effectName = "-special spell-"
					mappedEffect.Min = 0
					mappedEffect.Max = 0
					mappedEffect.Type[lang] = effectName
					mappedEffect.Templated[lang] = (*langs)[lang].Texts[data.spells[diceNum].DescriptionId]
					mappedEffect.IsMeta = true
				} else {
					templatedName := effectName
					templatedName, minMaxRemove = NumSpellFormatter(templatedName, lang, data, langs, &diceNum, &diceSide, &value, currentEffect.DescriptionId, numIsSpell, currentEffect.UseDice)
					if templatedName == "" { // found effect that should be discarded for now
						break
					}
					templatedName = SingularPluralFormatter(templatedName, effect.MinimumValue, lang)

					effectName = DeleteDamageFormatter(effectName)
					effectName = SingularPluralFormatter(effectName, effect.MinimumValue, lang)

					mappedEffect.Min = diceNum
					mappedEffect.Max = diceSide
					mappedEffect.Type[lang] = effectName
					mappedEffect.Templated[lang] = templatedName
					mappedEffect.IsMeta = false
				}

				if lang == "en" && mappedEffect.Type[lang] == "" {
					break
				}
			}

			if mappedEffect.Type["en"] == "()" || mappedEffect.Type["en"] == "" {
				continue
			}

			mappedEffect.Active = isActiveEffect(mappedEffect.Type)
			searchTypeEn := mappedEffect.Type["en"]
			if mappedEffect.Active {
				searchTypeEn += " (Active)"
			}
			key, foundKey := PersistedElements.Entries.GetKey(searchTypeEn)
			if foundKey {
				mappedEffect.ElementId = key.(int)
			} else {
				mappedEffect.ElementId = PersistedElements.NextId
				PersistedElements.Entries.Put(PersistedElements.NextId, searchTypeEn)
				PersistedElements.NextId++
			}

			mappedEffect.MinMaxIrrelevant = minMaxRemove

			mappedEffects = append(mappedEffects, mappedEffect)
		}
		if len(mappedEffects) > 0 {
			mappedAllEffects = append(mappedAllEffects, mappedEffects)
		}
	}
	if len(mappedAllEffects) == 0 {
		return nil
	}
	return mappedAllEffects
}

func ParseCondition(condition string, langs *map[string]LangDict, data *JSONGameData) []MappedMultilangCondition {
	if condition == "" || (!strings.Contains(condition, "&") && !strings.Contains(condition, "<") && !strings.Contains(condition, ">")) {
		return nil
	}

	condition = strings.ReplaceAll(condition, "\n", "")

	lower := strings.ToLower(condition)

	var outs []MappedMultilangCondition

	var parts []string
	if strings.Contains(lower, "&") {
		parts = strings.Split(lower, "&")
	} else {
		parts = []string{lower}
	}

	operators := []string{"<", ">", "=", "!"}

	for _, part := range parts {
		var out MappedMultilangCondition
		out.Templated = make(map[string]string)

		foundCond := false
		for _, operator := range operators { // try every known operator against it
			if strings.Contains(part, operator) {
				var outTmp MappedMultilangCondition
				outTmp.Templated = make(map[string]string)
				foundConditionElement := ConditionWithOperator(part, operator, langs, &out, data)
				if foundConditionElement {
					foundCond = true
				}
			}
		}

		if foundCond {
			outs = append(outs, out)
		}
	}

	if len(outs) == 0 {
		return nil
	}

	return outs
}

type HasId interface {
	GetID() int
}

func CleanJSON(jsonStr string) string {
	jsonStr = strings.ReplaceAll(jsonStr, "NaN", "null")
	jsonStr = strings.ReplaceAll(jsonStr, "\"null\"", "null")
	jsonStr = strings.ReplaceAll(jsonStr, " ", " ")
	return jsonStr
}

func ParseRawDataPart[T HasId](fileSource string, result chan map[int]T, dir string) {
	file, err := os.ReadFile(filepath.Join(dir, "data", fileSource))
	if err != nil {
		fmt.Print(err)
	}
	fileStr := CleanJSON(string(file))
	var fileJson []T
	err = json.Unmarshal([]byte(fileStr), &fileJson)
	if err != nil {
		fmt.Println(err)
	}
	items := make(map[int]T)
	for _, item := range fileJson {
		items[item.GetID()] = item
	}
	result <- items
}

func ParseRawData(dir string) *JSONGameData {
	var data JSONGameData
	itemChan := make(chan map[int]JSONGameItem)
	itemTypeChan := make(chan map[int]JSONGameItemType)
	itemSetsChan := make(chan map[int]JSONGameSet)
	itemEffectsChan := make(chan map[int]JSONGameEffect)
	itemBonusesChan := make(chan map[int]JSONGameBonus)
	itemRecipesChang := make(chan map[int]JSONGameRecipe)
	spellsChan := make(chan map[int]JSONGameSpell)
	spellTypesChan := make(chan map[int]JSONGameSpellType)
	areasChan := make(chan map[int]JSONGameArea)
	mountsChan := make(chan map[int]JSONGameMount)
	breedsChan := make(chan map[int]JSONGameBreed)
	mountFamilyChan := make(chan map[int]JSONGameMountFamily)
	npcsChan := make(chan map[int]JSONGameNPC)

	go func() {
		ParseRawDataPart("npcs.json", npcsChan, dir)
	}()
	go func() {
		ParseRawDataPart("mount_family.json", mountFamilyChan, dir)
	}()
	go func() {
		ParseRawDataPart("breeds.json", breedsChan, dir)
	}()
	go func() {
		ParseRawDataPart("mounts.json", mountsChan, dir)
	}()
	go func() {
		ParseRawDataPart("areas.json", areasChan, dir)
	}()
	go func() {
		ParseRawDataPart("spell_types.json", spellTypesChan, dir)
	}()
	go func() {
		ParseRawDataPart("spells.json", spellsChan, dir)
	}()
	go func() {
		ParseRawDataPart("recipes.json", itemRecipesChang, dir)
	}()
	go func() {
		ParseRawDataPart("items.json", itemChan, dir)
	}()
	go func() {
		ParseRawDataPart("item_types.json", itemTypeChan, dir)
	}()
	go func() {
		ParseRawDataPart("item_sets.json", itemSetsChan, dir)
	}()
	go func() {
		ParseRawDataPart("bonuses.json", itemBonusesChan, dir)
	}()
	go func() {
		ParseRawDataPart("effects.json", itemEffectsChan, dir)
	}()

	data.Items = <-itemChan
	close(itemChan)

	data.bonuses = <-itemBonusesChan
	close(itemBonusesChan)

	data.effects = <-itemEffectsChan
	close(itemEffectsChan)

	data.ItemTypes = <-itemTypeChan
	close(itemTypeChan)

	data.Sets = <-itemSetsChan
	close(itemSetsChan)

	data.Recipes = <-itemRecipesChang
	close(itemRecipesChang)

	data.spells = <-spellsChan
	close(spellsChan)

	data.spellTypes = <-spellTypesChan
	close(spellTypesChan)

	data.areas = <-areasChan
	close(areasChan)

	data.Mounts = <-mountsChan
	close(mountsChan)

	data.classes = <-breedsChan
	close(breedsChan)

	data.MountFamilys = <-mountFamilyChan
	close(mountFamilyChan)

	data.npcs = <-npcsChan
	close(npcsChan)

	return &data
}

func ParseLangDict(langCode string, dir string) LangDict {
	var err error

	dataPath := filepath.Join(dir, "data", "languages")
	var data LangDict
	data.IdText = make(map[int]int)
	data.Texts = make(map[int]string)
	data.NameText = make(map[string]int)

	langFile, err := os.ReadFile(filepath.Join(dataPath, langCode+".json"))
	if err != nil {
		fmt.Print(err)
	}

	langFileStr := CleanJSON(string(langFile))
	var langJson JSONLangDict
	err = json.Unmarshal([]byte(langFileStr), &langJson)
	if err != nil {
		fmt.Println(err)
	}

	for key, value := range langJson.IdText {
		keyParsed, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println(err)
		}
		data.IdText[keyParsed] = value
	}

	for key, value := range langJson.Texts {
		keyParsed, err := strconv.Atoi(key)
		if err != nil {
			fmt.Println(err)
		}
		data.Texts[keyParsed] = value
	}
	data.NameText = langJson.NameText
	return data
}

func ParseRawLanguages(dir string) map[string]LangDict {
	data := make(map[string]LangDict)
	for _, lang := range Languages {
		data[lang] = ParseLangDict(lang, dir)
	}
	return data
}
