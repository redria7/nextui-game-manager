package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
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

func (a ActionsScreen) Draw() (action interface{}, exitCode int, e error) {
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

	var actionEntries []gabagool.MenuItem
	for _, action := range actions {
		actionEntries = append(actionEntries, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(a.Game.DisplayName, actionEntries)
	options.SmallTitle = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		return selection.Unwrap().SelectedItem.Text, 0, nil
	}

	return nil, 2, nil
}
