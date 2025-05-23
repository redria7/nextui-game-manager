package ui

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type CollectionListScreen struct {
	SearchFilter string
}

func InitCollectionList(searchFilter string) CollectionListScreen {
	return CollectionListScreen{
		SearchFilter: searchFilter,
	}
}

func (c CollectionListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.CollectionsList
}

func (c CollectionListScreen) Draw() (collection interface{}, exitCode int, e error) {
	title := "Collections"
	fmt.Println(title)

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(common.CollectionDirectory, false)
	if err != nil {
		// TODO display ui error
		common.LogStandardFatal("Error loading fetching Collection directories", err)
	}

	if fb.Items == nil || len(fb.Items) == 0 {
		return models.Collection{}, 404, nil
	}

	itemList := fb.Items

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if c.SearchFilter != "" {
		title = "[Search: \"" + c.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = utils.FilterList(itemList, c.SearchFilter)
	}

	var itemEntries shared.Items
	var collections []models.Collection
	collectionsMap := make(map[string]models.Collection)

	for _, item := range itemList {
		collection := models.Collection{
			DisplayName:    item.DisplayName,
			CollectionFile: item.Path,
		}
		collections = append(collections, collection)
		collectionsMap[item.DisplayName] = collection
		itemEntries = append(itemEntries, shared.Item{
			DisplayName: item.DisplayName,
		})
	}

	if len(itemList) == 0 {
		return models.Collection{}, 404, nil
	}

	// TODO show list

	return nil, 0, nil
}
