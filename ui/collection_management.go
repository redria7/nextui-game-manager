package ui

import (
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type CollectionManagementScreen struct {
	Collection   models.Collection
	SearchFilter string
}

func InitCollectionManagement(collection models.Collection, searchFilter string) CollectionManagementScreen {
	return CollectionManagementScreen{
		Collection:   collection,
		SearchFilter: searchFilter,
	}
}

func (c CollectionManagementScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionManagement
}

func (c CollectionManagementScreen) Draw() (collection models.ScreenReturn, exitCode int, e error) {
	title := fmt.Sprintf("Collection: %s", c.Collection.DisplayName)

	var err error
	c.Collection, err = utils.ReadCollection(c.Collection)
	if err != nil {
		return shared.Item{}, 1, err
	}

	collectionItemsMap := make(map[string]shared.Item)

	for _, game := range c.Collection.Games {
		collectionItemsMap[game.DisplayName] = game
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "REMOVE GAME", "--confirm-button", "X", "--action-button", "Y")

	if len(c.Collection.Games) == 0 {
		return shared.Item{}, 404, err
	}

	selection, err := commonUI.DisplayList(c.Collection.Games, title, "OPTIONS", extraArgs...)
	if err != nil {
		return shared.Item{}, 1, err
	}

	if selection.ExitCode == 4 {
		return shared.Item{}, 4, nil
	}

	selectedItem := collectionItemsMap[selection.SelectedValue]

	return selectedItem, selection.ExitCode, nil
}
