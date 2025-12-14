package main

import (
	"errors"
	"strconv"

	"github.com/dofusdude/ankabuffer"
)

func DownloadQuests(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	outPath := dir

	if version == 2 {
		fileNames := []HashFile{
			{Filename: "data/common/Quests.d2o", FriendlyName: "quests.d2o"},
			{Filename: "data/common/QuestSteps.d2o", FriendlyName: "quest_steps.d2o"},
			{Filename: "data/common/QuestStepRewards.d2o", FriendlyName: "quest_step_rewards.d2o"},
			{Filename: "data/common/QuestObjectives.d2o", FriendlyName: "quest_objectives.d2o"},
			{Filename: "data/common/QuestCategory.d2o", FriendlyName: "quest_categories.d2o"},
			{Filename: "data/common/AlmanaxCalendars.d2o", FriendlyName: "almanax.d2o"},
		}

		err := DownloadUnpackFiles("Quests", bin, hashJson, "main", fileNames, dir, outPath, true, indent, headless, false)
		return err
	} else if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_questsdataroot.asset.bundle", FriendlyName: "quests.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_queststepsdataroot.asset.bundle", FriendlyName: "quest_steps.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_queststeprewardsdataroot.asset.bundle", FriendlyName: "quest_step_rewards.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_questobjectivesdataroot.asset.bundle", FriendlyName: "quest_objectives.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_questcategoriesdataroot.asset.bundle", FriendlyName: "quest_categories.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_almanaxcalendarsdataroot.asset.bundle", FriendlyName: "almanax.asset.bundle"},
		}

		err := DownloadUnpackFiles("Quests", bin, hashJson, "data", fileNames, dir, outPath, true, indent, headless, false)
		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
