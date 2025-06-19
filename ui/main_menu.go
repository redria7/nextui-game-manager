package ui

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type MainMenu struct {
}

func InitMainMenu() MainMenu {
	return MainMenu{}
}

func (m MainMenu) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.MainMenu
}

func (m MainMenu) Draw() (romDirectory interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	fb := filebrowser.NewFileBrowser(logger)

	var menuItems []gaba.MenuItem
	romDirectoryMap := make(map[string]shared.RomDirectory)
	var romDirectories shared.RomDirectories

	err := fb.CWD(utils.GetCollectionDirectory(), false)
	if err != nil {
		logger.Info("Unable to fetch Collection directories! Continuing without them", zap.Error(err))
	}

	if len(fb.Items) > 0 {
		collections := shared.RomDirectory{
			DisplayName: "Collections",
			Tag:         "Collections",
			Path:        common.CollectionDirectory,
		}
		romDirectories = append(romDirectories, collections)
		romDirectoryMap["Collections"] = collections

		menuItems = append(menuItems, gaba.MenuItem{
			Text:     "Collections",
			Selected: false,
			Focused:  false,
			Metadata: collections,
		})
	}

	err = fb.CWD(utils.GetRomDirectory(), state.GetAppState().Config.HideEmpty)
	if err != nil {
		gaba.ConfirmationMessage("Unable to fetch ROM directories!", []gaba.FooterHelpItem{
			{ButtonName: "B", HelpText: "Quit"},
		}, gaba.MessageOptions{})
		common.LogStandardFatal("Error loading fetching ROM directories", err)
	}

	for _, item := range fb.Items {
		if item.IsDirectory {
			romDirectory := shared.RomDirectory{
				DisplayName: item.DisplayName,
				Tag:         item.Tag,
				Path:        item.Path,
			}
			romDirectories = append(romDirectories, romDirectory)
			romDirectoryMap[item.DisplayName] = romDirectory

			menuItems = append(menuItems, gaba.MenuItem{
				Text:     romDirectory.DisplayName,
				Selected: false,
				Focused:  false,
				Metadata: romDirectory,
			})
		}
	}

	options := gaba.DefaultListOptions("Game Manager", menuItems)
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "X", HelpText: "Settings"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		return selection.Unwrap().SelectedItem.Metadata.(shared.RomDirectory), 0, nil
	}

	return nil, 2, nil
}
