package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type ActionsScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitActionsScreen(game shared.Item, romDirectory shared.RomDirectory, searchFilter string) ActionsScreen {
	return InitActionsScreenWithPreviousDirectory(game, romDirectory, shared.RomDirectory{}, searchFilter)
}

func InitActionsScreenWithPreviousDirectory(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string) ActionsScreen {
	return ActionsScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (a ActionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Actions
}

func (a ActionsScreen) Draw() (action models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	existingArtFilename, err := utils.FindExistingArt(a.Game.Filename, a.RomDirectory)
	if err != nil {
		logger.Error("failed to find existing arts", zap.Error(err))
	}

	hasGameTrackerData := utils.HasGameTrackerData(a.Game.Filename, a.RomDirectory)

	actions := models.ActionKeys

	if existingArtFilename == "" {
		actions = utils.InsertIntoSlice(actions, 1, "Download Art")
	} else {
		actions = utils.InsertIntoSlice(actions, 1, "Delete Art")
	}

	if hasGameTrackerData {
		actions = utils.InsertIntoSlice(actions, 2, "Clear Game Tracker")
	}

	var actionEntries shared.Items
	for _, action := range actions {
		actionEntries = append(actionEntries, shared.Item{DisplayName: action})
	}

	res, err := cui.DisplayList(actionEntries, a.Game.DisplayName, "")
	if err != nil {
		return shared.Item{}, 1, err
	}

	return shared.ListSelection{SelectedValue: res.SelectedValue}, res.ExitCode, nil
}
