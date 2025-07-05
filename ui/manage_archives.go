package ui

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/utils"
	"path/filepath"
	"qlova.tech/sum"
	"strings"
)

type ManageArchivesScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	PreviousRomDirectory shared.RomDirectory
	SearchFilter         string
}

func InitManageArchivesScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) ManageArchivesScreen {
	return ManageArchivesScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (mas ManageArchivesScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageArchives
}

func (mas ManageArchivesScreen) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := "Archiving " + mas.Game.DisplayName
	
	archiveFolders, err := utils.GetArchiveFileList()
	if err != nil {
		logger.Info("Unable to fetch archive directories! Failing backwards",
			zap.String("rom_directory", mas.RomDirectory.Path),
			zap.String("rom_file", mas.Game.Filename),
			zap.Error(err))
		return nil, -1, err
	}
	var archiveFolderEntries []gabagool.MenuItem
	for _, item := range archiveFolders {
		archiveFolderEntries = append(archiveFolderEntries, gabagool.MenuItem{
			Text:               item,
			Selected:           false,
			Focused:            false,
			Metadata:           item,
			NotMultiSelectable: true,
		})
	}

	options := gabagool.DefaultListOptions(title, archiveFolderEntries)
	options.SmallTitle = true
	options.EmptyMessage = "No Archive Folders Found"
	options.EnableAction = true
	options.EnableMultiSelect = false
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Create Archive Folder"},
		{ButtonName: "Menu", HelpText: "Help"},
	}

	options.EnableHelp = true
	options.HelpTitle = "Archive Management"
	options.HelpText = []string{
		"• X: Create New Archive Folder",
		"• A: Archive ROM into selected folder",
		"• B: Exit",
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		return selection.Unwrap().SelectedItem.Text, 0, nil
	} 
	
	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	}

	return nil, 2, nil
}
