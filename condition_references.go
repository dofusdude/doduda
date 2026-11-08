package main

import (
	"encoding/json"
	"fmt"
	mapping "github.com/dofusdude/dodumap"
	"os"
	"path/filepath"
)

// The downloader supplies entity catalogs; interpretation belongs to dodumap.
func loadConditionReferences(dir string, data *mapping.JSONGameDataUnity, languages map[string]mapping.LangDictUnity) error {
	data.ConditionReferences = map[string]map[int]mapping.ConditionReference{}
	for _, source := range []struct{ file, class, code string }{
		{"quests.json", "QuestData", "Qa"}, {"spell_states.json", "SpellStateData", "HS"},
		{"achievements.json", "AchievementData", "OA"}, {"spells.json", "SpellData", "PT"},
		{"jobs.json", "JobData", "PJ"}, {"areas.json", "AreaData", "Po"}, {"subareas.json", "SubAreaData", "PB"},
		{"alterations.json", "AlterationData", "HA"}, {"server_game_types.json", "ServerGameTypeData", "SC"},
		{"achievement_objectives.json", "AchievementObjectiveData", "OO"}, {"emoticons.json", "EmoticonData", "PE"},
		{"smiley_packs.json", "SmileyPackData", "Os"}, {"havenbag_themes.json", "HavenbagThemeData", "OH"},
		{"alignment_sides.json", "AlignmentSideData", "Ps"}, {"alliance_rights.json", "AllianceRightData", "Ox"},
		{"monsters.json", "MonsterData", "EM"}, {"item_sets.json", "ItemSetData", "OS"},
		{"world_events.json", "WorldEventData", "WE"},
		{"almanax.json", "AlmanaxCalendarData", "Ad"},
	} {
		records, err := readUnityRecords[rawAchievement](dir, source.file, source.class)
		if err != nil {
			return err
		}
		data.ConditionReferences[source.code] = map[int]mapping.ConditionReference{}
		for _, record := range records {
			data.ConditionReferences[source.code][record.ID] = mapping.ConditionReference{AnkamaID: record.ID, Name: localizedText(record.NameID, languages)}
		}
	}
	data.ConditionReferences["Qf"], data.ConditionReferences["Qc"] = data.ConditionReferences["Qa"], data.ConditionReferences["Qa"]
	for _, source := range []struct{ file, code string }{{"breeds.json", "PG"}, {"titles.json", "Ot"}, {"maps_information.json", "Pm"}, {"quest_steps.json", "Qs"}, {"quest_objectives.json", "Qo"}} {
		body, err := os.ReadFile(filepath.Join(dir, source.file))
		if err != nil {
			return err
		}
		var asset unityAsset
		if err = json.Unmarshal(body, &asset); err != nil {
			return err
		}
		data.ConditionReferences[source.code] = map[int]mapping.ConditionReference{}
		for _, entry := range asset.References.RefIDs {
			var record struct {
				ID          int         `json:"id"`
				NameID      flexibleInt `json:"nameId"`
				ShortNameID flexibleInt `json:"shortNameId"`
				MaleNameID  flexibleInt `json:"nameMaleId"`
				X           int         `json:"posX"`
				Y           int         `json:"posY"`
				QuestID     int         `json:"questId"`
				StepID      int         `json:"stepId"`
				TypeID      int         `json:"typeId"`
			}
			if err = json.Unmarshal(entry.Data, &record); err != nil {
				return err
			}
			if record.ID <= 0 {
				continue
			}
			nameID := record.NameID
			if source.code == "PG" {
				nameID = record.ShortNameID
			}
			if source.code == "Ot" {
				nameID = record.MaleNameID
			}
			name := localizedText(nameID, languages)
			if source.code == "Pm" {
				for lang := range languages {
					name[lang] = fmt.Sprintf("Map [%d, %d]", record.X, record.Y)
				}
			}
			if source.code == "Qo" {
				for lang := range languages {
					name[lang] = fmt.Sprintf("Quest objective %d (step %d)", record.ID, record.StepID)
				}
			}
			data.ConditionReferences[source.code][record.ID] = mapping.ConditionReference{AnkamaID: record.ID, Name: name}
		}
	}
	return nil
}
