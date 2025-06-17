package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
	"time"
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
	res, err := gaba.Keyboard("")

	if err != nil {
		return nil, -1, err
	}

	if res.IsSome() {
		newCollectionName := strings.ReplaceAll(res.Unwrap(), " ", "")

		if newCollectionName == "" {
			return nil, 2, nil
		}

		utils.AddCollectionGame(models.Collection{
			DisplayName:    newCollectionName,
			CollectionFile: filepath.Join(utils.GetCollectionDirectory(), newCollectionName+".txt"),
		}, c.Game)

		gaba.ProcessMessage(fmt.Sprintf("Created %s!\nAlso added %s!", newCollectionName, c.Game.DisplayName),
			gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(time.Second * 2)
				return nil, nil
			})

		return nil, 0, nil
	}

	return nil, 2, nil
}
