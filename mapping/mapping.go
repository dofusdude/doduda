package mapping

import (
	"fmt"
	"log"
)

var Languages = []string{"de", "en", "es", "fr", "it", "pt"}

func MapSets(data *JSONGameData, langs *map[string]LangDict) []MappedMultilangSet {
	var mappedSets []MappedMultilangSet
	for _, set := range data.Sets {
		var mappedSet MappedMultilangSet
		mappedSet.AnkamaId = set.Id
		mappedSet.ItemIds = set.ItemIds
		mappedSet.Effects = ParseEffects(data, set.Effects, langs)

		highestLevel := 0
		for _, item := range set.ItemIds {
			if data.Items[item].Level > highestLevel {
				highestLevel = data.Items[item].Level
			}
		}
		mappedSet.Level = highestLevel

		mappedSet.Name = make(map[string]string)
		for _, lang := range Languages {
			mappedSet.Name[lang] = (*langs)[lang].Texts[set.NameId]
		}

		mappedSets = append(mappedSets, mappedSet)
	}

	if len(mappedSets) == 0 {
		return nil
	}

	return mappedSets
}

func MapRecipes(data *JSONGameData) []MappedMultilangRecipe {
	var mappedRecipes []MappedMultilangRecipe

	for _, recipe := range data.Recipes {
		ingredientCount := len(recipe.IngredientIds)
		var mappedRecipe MappedMultilangRecipe
		mappedRecipe.ResultId = recipe.Id
		mappedRecipe.Entries = make([]MappedMultilangRecipeEntry, ingredientCount)
		for i := 0; i < ingredientCount; i++ {
			var recipeEntry MappedMultilangRecipeEntry
			recipeEntry.ItemId = recipe.IngredientIds[i]
			recipeEntry.Quantity = recipe.Quantities[i]
			mappedRecipe.Entries[i] = recipeEntry
		}
		mappedRecipes = append(mappedRecipes, mappedRecipe)
	}

	if len(mappedRecipes) == 0 {
		return nil
	}

	return mappedRecipes
}

func MapMounts(data *JSONGameData, langs *map[string]LangDict) []MappedMultilangMount {
	var mappedMounts []MappedMultilangMount
	for _, mount := range data.Mounts {
		var mappedMount MappedMultilangMount
		mappedMount.AnkamaId = mount.Id
		mappedMount.FamilyId = mount.FamilyId
		mappedMount.Name = make(map[string]string)
		mappedMount.FamilyName = make(map[string]string)

		for _, lang := range Languages {
			mappedMount.Name[lang] = (*langs)[lang].Texts[mount.NameId]
			mappedMount.FamilyName[lang] = (*langs)[lang].Texts[data.MountFamilys[mount.FamilyId].NameId]
		}

		allEffectResult := ParseEffects(data, [][]JSONGameItemPossibleEffect{mount.Effects}, langs)
		if len(allEffectResult) > 0 {
			mappedMount.Effects = allEffectResult[0]
		}

		mappedMounts = append(mappedMounts, mappedMount)
	}

	if len(mappedMounts) == 0 {
		return nil
	}

	return mappedMounts
}

func MapItems(data *JSONGameData, langs *map[string]LangDict) []MappedMultilangItem {
	var filteredItems []JSONGameItem

	for key, value := range data.Items {
		if langs == nil || data == nil {
			log.Fatal("langs is nil")
		}
		frName := (*langs)["fr"].Texts[value.NameId]
		itemType := data.ItemTypes[value.TypeId]
		category := itemType.CategoryId
		deTypeName := (*langs)["de"].Texts[data.ItemTypes[value.TypeId].NameId]
		if frName == "" || category == 4 || deTypeName == "Hauptquesten" {
			continue // skip unnamed and hidden items
		}
		filteredItems = append(filteredItems, data.Items[key])
	}

	mappedItems := make([]MappedMultilangItem, len(filteredItems))
	for idx, item := range filteredItems {
		mappedItems[idx].AnkamaId = item.Id
		mappedItems[idx].Level = item.Level
		mappedItems[idx].Pods = item.Pods
		mappedItems[idx].Image = fmt.Sprintf("https://static.ankama.com/dofus/www/game/items/200/%d.png", item.IconId)
		mappedItems[idx].Name = make(map[string]string, len(Languages))
		mappedItems[idx].Description = make(map[string]string, len(Languages))
		mappedItems[idx].Type.Name = make(map[string]string, len(Languages))
		mappedItems[idx].IconId = item.IconId

		for _, lang := range Languages {
			mappedItems[idx].Name[lang] = (*langs)[lang].Texts[item.NameId]
			mappedItems[idx].Description[lang] = (*langs)[lang].Texts[item.DescriptionId]
			mappedItems[idx].Type.Name[lang] = (*langs)[lang].Texts[data.ItemTypes[item.TypeId].NameId]
		}

		mappedItems[idx].Type.Id = item.TypeId
		mappedItems[idx].Type.SuperTypeId = data.ItemTypes[item.TypeId].SuperTypeId
		mappedItems[idx].Type.CategoryId = data.ItemTypes[item.TypeId].CategoryId

		searchTypeEn := mappedItems[idx].Type.Name["en"]
		key, foundKey := PersistedTypes.Entries.GetKey(searchTypeEn)
		if foundKey {
			mappedItems[idx].Type.ItemTypeId = key.(int)
		} else {
			mappedItems[idx].Type.ItemTypeId = PersistedTypes.NextId
			PersistedTypes.Entries.Put(PersistedTypes.NextId, searchTypeEn)
			PersistedTypes.NextId++
		}

		mappedItems[idx].UsedInRecipes = item.RecipeIds
		allEffectResult := ParseEffects(data, [][]JSONGameItemPossibleEffect{item.PossibleEffects}, langs)
		if len(allEffectResult) > 0 {
			mappedItems[idx].Effects = allEffectResult[0]
		}
		mappedItems[idx].Range = item.Range
		mappedItems[idx].MinRange = item.MinRange
		mappedItems[idx].CriticalHitProbability = item.CriticalHitProbability
		mappedItems[idx].CriticalHitBonus = item.CriticalHitBonus
		mappedItems[idx].ApCost = item.ApCost
		mappedItems[idx].TwoHanded = item.TwoHanded
		mappedItems[idx].MaxCastPerTurn = item.MaxCastPerTurn
		mappedItems[idx].DropMonsterIds = item.DropMonsterIds
		mappedItems[idx].HasParentSet = item.ItemSetId != -1
		if mappedItems[idx].HasParentSet {
			mappedItems[idx].ParentSet.Id = item.ItemSetId
			mappedItems[idx].ParentSet.Name = make(map[string]string, len(Languages))
			for _, lang := range Languages {
				mappedItems[idx].ParentSet.Name[lang] = (*langs)[lang].Texts[data.Sets[item.ItemSetId].NameId]
			}
		}

		if len(item.Criteria) != 0 && mappedItems[idx].Type.Name["de"] != "Verwendbarer Temporis-Gegenstand" { // TODO Temporis got some weird conditions, need to play to see the items, not in normal game
			mappedItems[idx].Conditions = ParseCondition(item.Criteria, langs, data)
		}
	}

	return mappedItems
}
