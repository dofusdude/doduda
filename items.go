package main

import (
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/dofusdude/ankabuffer"
)

func DownloadItems(hashJson *ankabuffer.Manifest, dir string, pythonPath string) error {
	log.Info("Downloading items...")
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
		{Filename: "data/common/AlmanaxCalendars.d2o", FriendlyName: "almanax.d2o"},
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
	}

	outPath := filepath.Join(dir, "data")
	err := DownloadUnpackFiles(hashJson, "main", fileNames, dir, outPath, true, pythonPath)

	log.Info("... downloaded items")
	return err
}
