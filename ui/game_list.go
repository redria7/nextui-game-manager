package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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

func (gl GameList) Draw() (item models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := gl.RomDirectory.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(gl.RomDirectory.Path)
	if err != nil {
		logger.Info("Unable to fetch ROM directory! Continuing without them",
			zap.String("rom_directory", gl.RomDirectory.Path),
			zap.Error(err))
		return shared.Item{}, 1, err
	}

	var roms shared.Items
	var displayNameToFilename = make(map[string]string)

	for _, entry := range fb.Items {
		roms = append(roms, shared.Item{
			DisplayName:          entry.DisplayName,
			Filename:             entry.Filename,
			IsDirectory:          !entry.IsMultiDiscDirectory && entry.IsDirectory,
			IsMultiDiscDirectory: entry.IsMultiDiscDirectory,
			Path:                 entry.Path,
		})

		displayNameToFilename[entry.DisplayName] = entry.Filename
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--confirm-text", "SELECT")

	if gl.SearchFilter != "" {
		title = "[Search: \"" + gl.SearchFilter + "\"]"
		extraArgs = append(extraArgs, "--cancel-text", "CLEAR SEARCH")
		roms = utils.FilterList(roms, gl.SearchFilter)
	}

	if len(roms) == 0 {
		return shared.Item{}, 404, nil
	}

	var directoryEntries shared.Items
	var itemEntries shared.Items
	displayNameToItem := make(map[string]shared.Item)

	for _, item := range roms {
		if strings.HasPrefix(item.Filename, ".") { // Skip hidden files
			continue
		}

		itemName := strings.TrimSuffix(item.Filename, filepath.Ext(item.Filename))

		if item.IsDirectory {
			itemName = "/" + itemName
			directoryEntries = append(directoryEntries, shared.Item{
				DisplayName: itemName,
			})
			displayNameToItem[itemName] = item
			continue
		}

		itemEntries = append(itemEntries, item)
		displayNameToItem[itemName] = item
	}

	allEntries := append(directoryEntries, itemEntries...)

	selection, err := cui.DisplayList(allEntries, title, "SEARCH", extraArgs...)
	if err != nil {
		return shared.Item{}, 1, err
	}

	selectedItem := displayNameToItem[selection.SelectedValue]

	return selectedItem, selection.ExitCode, nil
}
