package ui

import (
	"fmt"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type CollectionOptionsScreen struct {
	Collection   models.Collection
	SearchFilter string
}

func InitCollectionOptions(collection models.Collection, searchFilter string) CollectionOptionsScreen {
	return CollectionOptionsScreen{
		Collection:   collection,
		SearchFilter: searchFilter,
	}
}

func (c CollectionOptionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionOptions
}

func (c CollectionOptionsScreen) Draw() (screenReturn models.ScreenReturn, exitCode int, e error) {
	title := fmt.Sprintf("Collection Options: %s", c.Collection.DisplayName)

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	var actions shared.Items
	for _, action := range models.CollectionActionKeys {
		actions = append(actions, shared.Item{DisplayName: action})
	}

	selection, err := commonUI.DisplayList(actions, title, "", extraArgs...)
	if err != nil {
		return c.Collection, 1, err
	}

	return shared.ListSelection{SelectedValue: selection.SelectedValue}, selection.ExitCode, nil
}
