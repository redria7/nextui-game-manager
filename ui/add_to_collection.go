package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"slices"
)

type AddToCollectionScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitAddToCollectionScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) AddToCollectionScreen {
	return AddToCollectionScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (a AddToCollectionScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.AddToCollection
}

func (a AddToCollectionScreen) Draw() (collection models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(common.CollectionDirectory, false)
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to fetch Collection directories! Quitting!", "3")
		common.LogStandardFatal("Error loading fetching Collection directories", err)
	}

	if len(fb.Items) == 0 {
		return shared.ListSelection{ExitCode: 404}, 404, nil
	}

	var collections []models.Collection
	collectionsMap := make(map[string]models.Collection)

	for _, item := range fb.Items {
		collection := models.Collection{
			DisplayName:    item.DisplayName,
			CollectionFile: item.Path,
		}

		var err error
		collection, err = utils.ReadCollection(collection)

		if err != nil {
			logger.Error("Error reading collection", zap.Error(err))
		} else if !slices.ContainsFunc(collection.Games, func(element shared.Item) bool {
			return element.DisplayName == a.Game.DisplayName
		}) {
			collections = append(collections, collection)
			collectionsMap[item.DisplayName] = collection
		}
	}

	if len(collections) == 0 {
		return shared.ListSelection{ExitCode: 404}, 404, nil
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "ADD TO", "--confirm-button", "X", "--action-button", "Y")

	var itemEntries shared.Items

	for _, collection := range collections {
		itemEntries = append(itemEntries, shared.Item{
			DisplayName: collection.DisplayName,
		})
	}

	selection, err := commonUI.DisplayList(itemEntries, a.Game.DisplayName, "NEW COLLECTION", extraArgs...)
	if err != nil {
		return shared.ListSelection{ExitCode: 1}, 1, err
	}

	return collectionsMap[selection.SelectedValue], selection.ExitCode, nil
}
