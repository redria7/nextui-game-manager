package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
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

	fb := filebrowser.NewFileBrowser(common.GetLoggerInstance())
	err := fb.CWD(utils.GetCollectionDirectory(), false)
	if err != nil {
		gaba.ProcessMessage("Unable to Load Collections!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(time.Second * 2)
			return nil, nil
		})
		return nil, 404, nil
	}

	if fb.Items == nil || len(fb.Items) == 0 {
		return nil, 404, nil
	}

	itemList := fb.Items

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if c.SearchFilter != "" {
		title = "[Search: \"" + c.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		itemList = utils.FilterList(itemList, c.SearchFilter)
	}

	slices.SortFunc(itemList, func(a, b shared.Item) int {
		return strings.Compare(a.DisplayName, b.DisplayName)
	})

	var menuItems []gaba.MenuItem
	for _, item := range itemList {
		col := models.Collection{DisplayName: item.DisplayName, CollectionFile: item.Path}
		col, err = utils.ReadCollection(col)

		if err != nil {
			gaba.ProcessMessage("Unable to Load Collections!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
				time.Sleep(time.Second * 2)
				return nil, nil
			})
			return nil, -1, err
		}

		collection := gaba.MenuItem{
			Text:     item.DisplayName,
			Selected: false,
			Focused:  false,
			Metadata: col,
		}
		menuItems = append(menuItems, collection)
	}

	if len(itemList) == 0 {
		title = "No Collections Found"
	}

	options := gaba.DefaultListOptions(title, menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		return selection.Unwrap().SelectedItem.Metadata.(models.Collection), 0, nil
	}

	return nil, 2, nil
}
