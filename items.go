package main

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/dofusdude/ankabuffer"
)

func DownloadItems(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	outPath := dir

	if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemsroot.asset.bundle", FriendlyName: "items.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemtypesroot.asset.bundle", FriendlyName: "item_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemsetsroot.asset.bundle", FriendlyName: "item_sets.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_effectsroot.asset.bundle", FriendlyName: "effects.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_bonusesroot.asset.bundle", FriendlyName: "bonuses.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_recipesroot.asset.bundle", FriendlyName: "recipes.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellsroot.asset.bundle", FriendlyName: "spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spelltypesroot.asset.bundle", FriendlyName: "spell_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breedsroot.asset.bundle", FriendlyName: "breeds.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breedrolesroot.asset.bundle", FriendlyName: "breed_roles.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characteristicsroot.asset.bundle", FriendlyName: "characteristics.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characterxpmappingsroot.asset.bundle", FriendlyName: "char_xp_mappings.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mountsroot.asset.bundle", FriendlyName: "mounts.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_monsterracesroot.asset.bundle", FriendlyName: "monster_races.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_monstersroot.asset.bundle", FriendlyName: "monsters.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companioncharacteristicsroot.asset.bundle", FriendlyName: "companion_chars.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companionspellsroot.asset.bundle", FriendlyName: "companion_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companionsroot.asset.bundle", FriendlyName: "companions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_areasroot.asset.bundle", FriendlyName: "areas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mountfamilyroot.asset.bundle", FriendlyName: "mount_family.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_servergametypesroot.asset.bundle", FriendlyName: "server_game_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characteristiccategoriesroot.asset.bundle", FriendlyName: "chars_categories.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_creaturebonestypesroot.asset.bundle", FriendlyName: "creature_bone_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_creaturebonesoverridesroot.asset.bundle", FriendlyName: "creature_bone_overrides.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_evolutiveeffectsroot.asset.bundle", FriendlyName: "evol_effects.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_bonusescriterionsroot.asset.bundle", FriendlyName: "bonus_criterions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_titlesroot.asset.bundle", FriendlyName: "titles.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_dungeonsroot.asset.bundle", FriendlyName: "dungeons.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellpairsroot.asset.bundle", FriendlyName: "spell_pairs.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellstatesroot.asset.bundle", FriendlyName: "spell_states.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellvariantsroot.asset.bundle", FriendlyName: "spell_variants.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spelllevelsroot.asset.bundle", FriendlyName: "spell_levels.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_custommodebreedspellsroot.asset.bundle", FriendlyName: "custom_breed_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_forgettablespellsroot.asset.bundle", FriendlyName: "forgettable_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellbombsroot.asset.bundle", FriendlyName: "bomb_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellconversionsroot.asset.bundle", FriendlyName: "spell_conversions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellscriptsroot.asset.bundle", FriendlyName: "spell_scripts.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_subareasroot.asset.bundle", FriendlyName: "subareas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_superareasroot.asset.bundle", FriendlyName: "superareas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breachworldmapcoordinatesroot.asset.bundle", FriendlyName: "breach_worldmap_coordinates.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breachworldmapsectorsroot.asset.bundle", FriendlyName: "breach_worldmap_sectors.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcsroot.asset.bundle", FriendlyName: "npcs.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcactionsroot.asset.bundle", FriendlyName: "npc_actions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcdialogskinroot.asset.bundle", FriendlyName: "npc_dialogskin.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcmessagesroot.asset.bundle", FriendlyName: "npc_messages.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_infomessagesroot.asset.bundle", FriendlyName: "info_messages.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_calendareventsroot.asset.bundle", FriendlyName: "event_calendar.asset.bundle"},
		}

		err := DownloadUnpackFiles("Items", bin, hashJson, "data", fileNames, dir, outPath, true, indent, headless, false, true)
		if err != nil {
			return err
		}

		fileNames = []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/aa/StandaloneWindows64/worldassets_assets_all.bundle", FriendlyName: "worldassets.imagebundle"},
		}
		err = DownloadUnpackFiles("Worldgraph", bin, hashJson, "win32_x64", fileNames, dir, outPath, true, indent, headless, false, true)
		if err != nil {
			return err
		}
		worldgraph_source := filepath.Join(outPath, "WorldData", "world-graph.json")
		worldgraph_target := filepath.Join(outPath, "world_graph.json")
		os.Rename(worldgraph_source, worldgraph_target)

		return nil
	} else if version == 2 {
		fileNames := []HashFile{
			{Filename: "data/common/Items.d2o", FriendlyName: "items.d2o"},
			{Filename: "data/common/ItemTypes.d2o", FriendlyName: "item_types.d2o"},
			{Filename: "data/common/ItemSets.d2o", FriendlyName: "item_sets.d2o"},
			{Filename: "data/common/Effects.d2o", FriendlyName: "effects.d2o"},
			{Filename: "data/common/Bonuses.d2o", FriendlyName: "bonuses.d2o"},
			{Filename: "data/common/Recipes.d2o", FriendlyName: "recipes.d2o"},
			{Filename: "data/common/Spells.d2o", FriendlyName: "spells.d2o"},
			{Filename: "data/common/SpellTypes.d2o", FriendlyName: "spell_types.d2o"},
			{Filename: "data/common/Breeds.d2o", FriendlyName: "breeds.d2o"},
			{Filename: "data/common/Mounts.d2o", FriendlyName: "mounts.d2o"},
			{Filename: "data/common/Idols.d2o", FriendlyName: "idols.d2o"},
			{Filename: "data/common/MonsterRaces.d2o", FriendlyName: "monster_races.d2o"},
			{Filename: "data/common/Monsters.d2o", FriendlyName: "monsters.d2o"},
			{Filename: "data/common/CompanionCharacteristics.d2o", FriendlyName: "companion_chars.d2o"},
			{Filename: "data/common/CompanionSpells.d2o", FriendlyName: "companion_spells.d2o"},
			{Filename: "data/common/Companions.d2o", FriendlyName: "companions.d2o"},
			{Filename: "data/common/Areas.d2o", FriendlyName: "areas.d2o"},
			{Filename: "data/common/MountFamily.d2o", FriendlyName: "mount_family.d2o"},
			{Filename: "data/common/Npcs.d2o", FriendlyName: "npcs.d2o"},
			{Filename: "data/common/ServerGameTypes.d2o", FriendlyName: "server_game_types.d2o"},
			{Filename: "data/common/CharacteristicCategories.d2o", FriendlyName: "chars_categories.d2o"},
			{Filename: "data/common/CreatureBonesTypes.d2o", FriendlyName: "creature_bone_types.d2o"},
			{Filename: "data/common/CreatureBonesOverrides.d2o", FriendlyName: "create_bone_overrides.d2o"},
			{Filename: "data/common/EvolutiveEffects.d2o", FriendlyName: "evol_effects.d2o"},
			{Filename: "data/common/BonusesCriterions.d2o", FriendlyName: "bonus_criterions.d2o"},
			{Filename: "data/common/Titles.d2o", FriendlyName: "titles.d2o"},
		}

		err := DownloadUnpackFiles("Items", bin, hashJson, "main", fileNames, dir, outPath, true, indent, headless, false, true)

		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}

func DownloadMaps(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	outPath := dir

	if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mappositionsroot.asset.bundle", FriendlyName: "map_positions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mapcoordinatesroot.asset.bundle", FriendlyName: "map_coordinates.asset.bundle"},
		}

		err := DownloadUnpackFiles("Map Metadata", bin, hashJson, "data", fileNames, dir, outPath, true, indent, headless, false, true)
		if err != nil {
			return err
		}

		fileNames = []HashFile{
			{Filename: `REGEX:Dofus_Data/StreamingAssets/Content/Map/Data/mapdata_assets_world_\d+\.bundle`, FriendlyName: "map.mapbundle"},
		}

		mapOutPath := path.Join(outPath, "map")
		err = DownloadUnpackFiles("Map", bin, hashJson, "map_common", fileNames, dir, mapOutPath, true, indent, headless, false, false)
		if err != nil {
			return err
		}

		err = os.RemoveAll(path.Join(outPath, "Dofus_Data"))

		return err

	} else {
		// This might throw an error on every Dofus 2 download without an ignore but I think it is currently impossible to load Dofus 2 data.
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
