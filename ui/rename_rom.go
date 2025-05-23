package ui

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type RenameRomScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitRenameRomScreen(game shared.Item, romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, searchFilter string) RenameRomScreen {
	return RenameRomScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (r RenameRomScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.RenameRom
}

func (r RenameRomScreen) Draw() (value interface{}, exitCode int, e error) {
	// TODO Keyboard

	//newFilename := strings.ReplaceAll(outValue, "\n", "")

	return "", 0, nil
}
