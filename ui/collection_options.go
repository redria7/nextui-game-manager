package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
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

func (c CollectionOptionsScreen) Draw() (screenReturn interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var actions []gabagool.MenuItem
	for _, action := range models.CollectionActionKeys {
		actions = append(actions, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: action,
		})
	}

	options := gabagool.DefaultListOptions(fmt.Sprintf("%s Options", c.Collection.DisplayName), actions)
	options.EnableAction = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	result, err := gabagool.List(options)

	if err != nil {
		return nil, -1, err
	}

	if result.IsSome() && !result.Unwrap().ActionTriggered && result.Unwrap().SelectedIndex != -1 {
		action := models.ActionMap[result.Unwrap().SelectedItem.Metadata.(string)]

		switch action {
		case models.Actions.CollectionRename:
			newName, err := gabagool.Keyboard(c.Collection.DisplayName)
			if err != nil {
				return nil, -1, err
			}

			if newName.IsSome() {
				updatedCol, err := utils.RenameCollection(c.Collection, newName.Unwrap())
				if err != nil {
					logger.Error("failed to rename collection", zap.Error(err))
					return nil, -1, err
				}

				return updatedCol, 4, nil
			}

		case models.Actions.CollectionDelete:
			res, _ := gabagool.ConfirmationMessage(fmt.Sprintf("Are you sure you want to delete the collection\n%s?", c.Collection.DisplayName), []gabagool.FooterHelpItem{
				{ButtonName: "B", HelpText: "Cancel"},
				{ButtonName: "X", HelpText: "Delete"},
			}, gabagool.MessageOptions{
				ImagePath:     "",
				ConfirmButton: gabagool.ButtonX,
			})

			if res.IsSome() && !res.Unwrap().Cancelled {
				utils.DeleteCollection(c.Collection)
				return nil, 3, nil
			}

		}

		return c.Collection, 2, nil
	}

	return c.Collection, 2, nil
}
