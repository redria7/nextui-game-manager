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

type GameList struct {
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitGamesList(romDirectory shared.RomDirectory, searchFilter string) GameList {
	return InitGamesListWithPreviousDirectory(romDirectory, shared.RomDirectory{}, searchFilter)
}

func InitGamesListWithPreviousDirectory(romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, searchFilter string) GameList {
	return GameList{
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (gl GameList) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GamesList
}

func (gl GameList) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := gl.RomDirectory.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(gl.RomDirectory.Path, false)
	if err != nil {
		logger.Info("Unable to fetch ROM directory! Continuing without them",
			zap.String("rom_directory", gl.RomDirectory.Path),
			zap.Error(err))
		return shared.Item{}, 1, err
	}

	var roms shared.Items

	for _, entry := range fb.Items {
		roms = append(roms, shared.Item{
			DisplayName:          entry.DisplayName,
			Filename:             entry.Filename,
			IsDirectory:          !entry.IsMultiDiscDirectory && entry.IsDirectory,
			IsMultiDiscDirectory: entry.IsMultiDiscDirectory,
			Path:                 entry.Path,
		})
	}

	if gl.SearchFilter != "" {
		title = "[Search: \"" + gl.SearchFilter + "\"]"
		roms = utils.FilterList(roms, gl.SearchFilter)
	}

	if len(roms) == 0 {
		return shared.Item{}, 404, nil
	}

	var directoryEntries []gabagool.MenuItem
	var itemEntries []gabagool.MenuItem

	for _, item := range roms {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if item.IsDirectory {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, gabagool.MenuItem{
				Text:               itemName,
				Selected:           false,
				Focused:            false,
				Metadata:           item,
				NotMultiSelectable: true,
			})
		} else {
			itemEntries = append(itemEntries, gabagool.MenuItem{
				Text:     itemName,
				Selected: false,
				Focused:  false,
				Metadata: item,
			})
		}
	}

	allEntries := append(directoryEntries, itemEntries...)

	options := gabagool.DefaultListOptions(title, allEntries)
	options.EnableAction = true
	options.EnableMultiSelect = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Select", HelpText: "Multi"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		return selection.Unwrap().SelectedItem.Metadata.(shared.Item), 0, nil
	}

	return nil, 2, nil
}
