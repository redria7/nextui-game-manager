package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
	"strings"
)

type CreateArchiveScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitCreateArchiveScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) CreateArchiveScreen {
	return CreateArchiveScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (cas CreateArchiveScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CreateArchive
}

func (cas CreateArchiveScreen) Draw() (value interface{}, exitCode int, e error) {
	query, err := gabagool.Keyboard("")
	if err != nil {
		return nil, -1, err
	}

	if query.IsSome() {
		newArchive := query.Unwrap()
		if strings.HasPrefix(newArchive, ".") {
			return newArchive, 0, nil
		}
		return "." + newArchive, 0, nil
	}

	return nil, 2, nil
}
