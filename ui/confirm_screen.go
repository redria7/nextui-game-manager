package ui

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type ConfirmScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	Action               sum.Int[models.Action]
	PreviousRomDirectory shared.RomDirectory
}

func InitConfirmScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string, action sum.Int[models.Action]) ConfirmScreen {
	return ConfirmScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
		Action:               action,
	}
}

func (c ConfirmScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Confirm
}

func (c ConfirmScreen) Draw() (unused models.ScreenReturn, exitCode int, e error) {
	actionMessage := models.ActionMessages[c.Action]

	target := c.Game.DisplayName
	if c.Action == models.Actions.CollectionDelete {
		target = c.RomDirectory.DisplayName
	}

	message := fmt.Sprintf("%s %s?", actionMessage, target)

	code, err := commonUI.ShowMessageWithOptions(message, "0",
		"--confirm-text", "DO IT!",
		"--confirm-show", "true",
		"--confirm-button", "X",
		"--cancel-show", "true",
		"--cancel-text", "CHANGED MY MIND",
	)

	if err != nil {
		logger := common.GetLoggerInstance()
		logger.Info("Oh no", zap.Error(err))
		return shared.Item{}, 1, err
	}

	return shared.Item{}, code, nil
}
