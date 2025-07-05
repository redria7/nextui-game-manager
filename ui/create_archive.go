package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type CreateArchive struct {
	RomDirectory shared.RomDirectory
}

func InitCreateArchive(romDirectory shared.RomDirectory) Search {
	return CreateArchive{
		RomDirectory: romDirectory,
	}
}

func (cas CreateArchive) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CreateArchive
}

func (cas CreateArchive) Draw() (value interface{}, exitCode int, e error) {
	query, err := gabagool.Keyboard("")
	if err != nil {
		return nil, -1, err
	}

	if query.IsSome() {
		return query.Unwrap(), 0, nil
	}

	return nil, 2, nil
}
