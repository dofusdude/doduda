package main

import (
	"path/filepath"

	"github.com/dofusdude/ankabuffer"
)

func DownloadQuests(hashJson *ankabuffer.Manifest, dir string, indent string, headless bool) error {
	fileNames := []HashFile{
		{Filename: "data/common/Quests.d2o", FriendlyName: "quests.d2o"},
		{Filename: "data/common/QuestSteps.d2o", FriendlyName: "quest_steps.d2o"},
		{Filename: "data/common/QuestStepRewards.d2o", FriendlyName: "quest_step_rewards.d2o"},
		{Filename: "data/common/QuestObjectives.d2o", FriendlyName: "quest_objectives.d2o"},
		{Filename: "data/common/QuestCategory.d2o", FriendlyName: "quest_categories.d2o"},
		{Filename: "data/common/AlmanaxCalendars.d2o", FriendlyName: "almanax.d2o"},
	}

	outPath := filepath.Join(dir, "data")

	err := DownloadUnpackFiles("Quests", hashJson, "main", fileNames, dir, outPath, true, indent, headless, false)

	return err
}
