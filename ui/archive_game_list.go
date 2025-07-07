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

type ArchiveGamesListScreen struct {
	Archive 			 string
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
}

func InitArchiveGamesList(archive string, romDirectory shared.RomDirectory, searchFilter string) ArchiveGamesListScreen {
	return InitArchiveGamesListScreenWithPreviousDirectory(archive, romDirectory, shared.RomDirectory{}, searchFilter)
}

func InitArchiveGamesListScreenWithPreviousDirectory(archive string, romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, searchFilter string) ArchiveGamesListScreen {
	return ArchiveGamesListScreen{
		Archive:			  archive,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
	}
}

func (agl ArchiveGamesListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ArchiveGamesListScreen
}

// Lists ROMs in the current archive/directory path and allows for restoration
func (agl ArchiveGamesListScreen) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	title := Archive + " : " + agl.RomDirectory.DisplayName

	fb := filebrowser.NewFileBrowser(logger)

	err := fb.CWD(agl.RomDirectory.Path, false)
	if err != nil {
		logger.Info("Unable to fetch ROM directory! Continuing without them",
			zap.String("rom_directory", agl.RomDirectory.Path),
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

	if agl.SearchFilter != "" {
		title = "[Search: \"" + agl.SearchFilter + "\"]"
		roms = utils.FilterList(roms, agl.SearchFilter)
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
	options.SmallTitle = true
	options.EmptyMessage = "No ROMs Found"
	options.EnableAction = true
	options.EnableMultiSelect = true
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		{ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Restore"},
	}

	options.EnableHelp = true
	options.HelpTitle = "Archive ROMs List Controls"
	options.HelpText = []string{
		"• X: Open Options",
		"• Select: Toggle Multi-Select",
		"• Start: Confirm Multi-Selection",
	}

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		query, err := gabagool.Keyboard("")
		
		if err != nil {
			return nil, 1, err
		}

		if query.IsSome() {
			return query.Unwrap(), 4, nil
		}

		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		var selectedItems shared.Items
		rawSelection := selection.Unwrap().SelectedItems
		
		confirmMessage := fmt.Sprintf("Restore %s from archive %s?", rawSelection[0].DisplayName, agl.Archive)
		successMessage := fmt.Sprintf("Restored %s from archive %s!", rawSelection[0].DisplayName, agl.Archive)
		if len(rawSelection) > 1 {
			confirmMessage = fmt.Sprintf("Restore %d from archive %s?", len(rawSelection), agl.Archive)
			successMessage = fmt.Sprintf("Restored %d games from archive %s!", len(rawSelection), agl.Archive)
		} else {
			if rawSelection[0].IsDirectory {
				newRomDirectory := shared.RomDirectory{
					DisplayName: rawSelection[0].DisplayName,
					Tag:         rawSelection[0].Tag,
					Path:        rawSelection[0].Path,
				}
				return newRomDirectory, 0, nil
			}
		}

		if !confirmAction(message) {
			return nil, 404, nil
		}

		for _, item := range rawSelection {
			err := utils.RestoreRom(item, agl.RomDirectory, agl.Archive)
			if err != nil {
				gaba.ProcessMessage(fmt.Sprintf("Unable to restore %s!", item.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
					time.Sleep(time.Second * 2)
					return nil, nil
				})
				return nil, 0, err
			}
		}

		gaba.ProcessMessage(successMessage, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
			time.Sleep(time.Second * 2)
			return nil, nil
		})

		return nil, 0, nil
	}

	return nil, 2, nil
}
