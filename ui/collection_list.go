package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
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

func (c CollectionListScreen) Draw() (collection models.ScreenReturn, exitCode int, e error) {
	title := "Collections"

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(common.CollectionDirectory)
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to fetch Collection directories! Quitting!", "3")
		common.LogStandardFatal("Error loading fetching Collection directories", err)
	}

	if len(fb.Items) == 0 {
		return shared.ListSelection{ExitCode: 404}, 404, nil
	}

	itemList := fb.Items
	var collectionDirectory []shared.RomDirectory
	collectionDirectoryMap := make(map[string]shared.RomDirectory)

	for _, item := range fb.Items {
		romDirectory := shared.RomDirectory{
			DisplayName: item.DisplayName,
			Tag:         item.Tag,
			Path:        item.Path,
		}
		collectionDirectory = append(collectionDirectory, romDirectory)
		collectionDirectoryMap[item.DisplayName] = romDirectory
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if c.SearchFilter != "" {
		title = "[Search: \"" + c.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = utils.FilterList(itemList, c.SearchFilter)
	}

	if len(itemList) == 0 {
		return shared.ListSelection{ExitCode: 404}, 404, nil
	}

	var itemEntries shared.Items

	for _, item := range itemList {
		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))
		itemEntries = append(itemEntries, shared.Item{
			DisplayName: itemName,
		})
	}

	selection, err := commonUI.DisplayList(itemEntries, title, "", extraArgs...)
	if err != nil {
		return shared.ListSelection{ExitCode: 1}, 1, err
	}

	return collectionDirectoryMap[selection.SelectedValue], selection.ExitCode, nil
}
