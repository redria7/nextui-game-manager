package ui

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type Search struct {
	RomDirectory shared.RomDirectory
}

func InitSearch(romDirectory shared.RomDirectory) Search {
	return Search{
		RomDirectory: romDirectory,
	}
}

func (s Search) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.SearchBox
}

func (s Search) Draw() (value interface{}, exitCode int, e error) {

	return "KEY BOARD OUTPUT", 0, nil
}
