package main

import (
	"errors"
	"strconv"

	"github.com/dofusdude/ankabuffer"
)

func DownloadAchievements(hashJson *ankabuffer.Manifest, bin int, version int, dir string, indent string, headless bool) error {
	outPath := dir

	if version == 3 {
		fileNames := []HashFile{
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementcategoriesroot.asset.bundle", FriendlyName: "achievement_categories.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementobjectivesroot.asset.bundle", FriendlyName: "achievement_objectives.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementrewardsroot.asset.bundle", FriendlyName: "achievement_rewards.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementsroot.asset.bundle", FriendlyName: "achievements.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementprogressstepsroot.asset.bundle", FriendlyName: "achievement_progress_steps.asset.bundle"},
			{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_achievementprogressroot.asset.bundle", FriendlyName: "achievement_progress.asset.bundle"},
		}

		err := DownloadUnpackFiles("Achievements", bin, hashJson, "data", fileNames, dir, outPath, true, indent, headless, false, true)
		return err
	} else {
		return errors.New("unsupported version: " + strconv.Itoa(version))
	}
}
