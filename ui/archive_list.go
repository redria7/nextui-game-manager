package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"slices"
	"strings"
	"time"
)

type ArchiveListScreen struct {}

func InitArchiveListScreen() ArchiveListScreen {
	return ArchiveListScreen{}
}

func (als ArchiveListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveList
}

// Lists available archive folders
func (als ArchiveListScreen) Draw() (item interface{}, exitCode int, e error) {
	title := "Archives"
	
	archiveFolders, err := utils.GetArchiveFileListBasic()
	if err != nil {
		gaba.ProcessMessage("Unable to Load Archives!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(time.Second * 2)
			return nil, nil
		})
		return nil, 404, nil
	}

	if archiveFolders == nil || len(archiveFolders) == 0 {
		return nil, 404, nil
	}

	var menuItems []gaba.MenuItem
	for _, archiveFolder := range archiveFolders {
		archive := gaba.MenuItem{
			Text:     archiveFolder,
			Selected: false,
			Focused:  false,
			Metadata: archiveFolder,
		}
		menuItems = append(menuItems, archive)
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
		archive := selection.Unwrap().SelectedItem.Metadata.(string)
		archiveDirectory := shared.RomDirectory{
			DisplayName: archive,
			Path: 		 utils.GetArchiveRoot(archive),
		}
		return archiveDirectory, 0, nil
	}

	return nil, 2, nil
}
