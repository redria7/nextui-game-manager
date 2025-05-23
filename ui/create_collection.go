package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

type CreateCollectionScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitCreateCollectionScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) CreateCollectionScreen {
	return CreateCollectionScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (c CreateCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionCreate
}

func (c CreateCollectionScreen) Draw() (collection interface{}, exitCode int, e error) {

	newCollectionName := strings.ReplaceAll("KEYBOARD OUTPUT", "\n", "")

	utils.AddCollectionGame(models.Collection{
		DisplayName:    newCollectionName,
		CollectionFile: filepath.Join(common.CollectionDirectory, newCollectionName+".txt"),
	}, c.Game)

	return nil, 0, nil
}
