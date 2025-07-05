package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
	"strings"
)

type CreateArchiveScreen struct {
	RomDirectory shared.RomDirectory
}

func InitCreateArchiveScreen(romDirectory shared.RomDirectory) CreateArchiveScreen {
	return CreateArchiveScreen{
		RomDirectory: romDirectory,
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
