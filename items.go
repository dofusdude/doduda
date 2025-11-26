package main

import (
	"errors"
	"strconv"

	"github.com/dofusdude/ankabuffer"
)

func DownloadItems(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	outPath := dir

	if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemsdataroot.asset.bundle", FriendlyName: "items.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemtypesdataroot.asset.bundle", FriendlyName: "item_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemsetsdataroot.asset.bundle", FriendlyName: "item_sets.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_effectsdataroot.asset.bundle", FriendlyName: "effects.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_bonusesdataroot.asset.bundle", FriendlyName: "bonuses.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_recipesdataroot.asset.bundle", FriendlyName: "recipes.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellsdataroot.asset.bundle", FriendlyName: "spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spelltypesdataroot.asset.bundle", FriendlyName: "spell_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breedsdataroot.asset.bundle", FriendlyName: "breeds.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breedrolesdataroot.asset.bundle", FriendlyName: "breed_roles.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characteristicsdataroot.asset.bundle", FriendlyName: "characteristics.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characterxpmappingsdataroot.asset.bundle", FriendlyName: "char_xp_mappings.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mountsdataroot.asset.bundle", FriendlyName: "mounts.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_monsterracesdataroot.asset.bundle", FriendlyName: "monster_races.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_monstersdataroot.asset.bundle", FriendlyName: "monsters.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companioncharacteristicsdataroot.asset.bundle", FriendlyName: "companion_chars.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companionspellsdataroot.asset.bundle", FriendlyName: "companion_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_companionsdataroot.asset.bundle", FriendlyName: "companions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_areasdataroot.asset.bundle", FriendlyName: "areas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_mountfamiliesdataroot.asset.bundle", FriendlyName: "mount_family.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_servergametypesdataroot.asset.bundle", FriendlyName: "server_game_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_characteristiccategoriesdataroot.asset.bundle", FriendlyName: "chars_categories.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_creaturebonestypesdataroot.asset.bundle", FriendlyName: "creature_bone_types.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_creaturebonesoverridesdataroot.asset.bundle", FriendlyName: "creature_bone_overrides.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_evolutiveeffectsdataroot.asset.bundle", FriendlyName: "evol_effects.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_bonusescriterionsdataroot.asset.bundle", FriendlyName: "bonus_criterions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_titlesdataroot.asset.bundle", FriendlyName: "titles.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_dungeonsdataroot.asset.bundle", FriendlyName: "dungeons.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellpairsdataroot.asset.bundle", FriendlyName: "spell_pairs.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellstatesdataroot.asset.bundle", FriendlyName: "spell_states.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellvariantsdataroot.asset.bundle", FriendlyName: "spell_variants.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spelllevelsdataroot.asset.bundle", FriendlyName: "spell_levels.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_custommodebreedspellsdataroot.asset.bundle", FriendlyName: "custom_breed_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_forgettablespellsdataroot.asset.bundle", FriendlyName: "forgettable_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellbombsdataroot.asset.bundle", FriendlyName: "bomb_spells.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellconversionsdataroot.asset.bundle", FriendlyName: "spell_conversions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_spellscriptsdataroot.asset.bundle", FriendlyName: "spell_scripts.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_subareasdataroot.asset.bundle", FriendlyName: "subareas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_superareasdataroot.asset.bundle", FriendlyName: "superareas.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breachworldmapcoordinatesdataroot.asset.bundle", FriendlyName: "breach_worldmap_coordinates.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_breachworldmapsectorsdataroot.asset.bundle", FriendlyName: "breach_worldmap_sectors.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcsdataroot.asset.bundle", FriendlyName: "npcs.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcactionsdataroot.asset.bundle", FriendlyName: "npc_actions.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcdialogskindataroot.asset.bundle", FriendlyName: "npc_dialogskin.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_npcmessagesdataroot.asset.bundle", FriendlyName: "npc_messages.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_infomessagesdataroot.asset.bundle", FriendlyName: "info_messages.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_calendareventsdataroot.asset.bundle", FriendlyName: "event_calendar.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_randomdropgroupsdataroot.asset.bundle", FriendlyName: "random_drop_groups.asset.bundle"},
		}

		err := DownloadUnpackFiles("Items", bin, hashJson, "data", fileNames, dir, outPath, true, indent, headless, false)
		if err != nil {
			return err
		}

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

		err := DownloadUnpackFiles("Items", bin, hashJson, "main", fileNames, dir, outPath, true, indent, headless, false)

		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
