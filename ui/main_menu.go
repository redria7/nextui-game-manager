package ui

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	cui "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
	"go.uber.org/zap"
	"nextui-game-manager/models"
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

func (m MainMenu) Draw() (romDirectory models.ScreenReturn, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	fb := filebrowser.NewFileBrowser(logger)

	romDirectoryMap := make(map[string]shared.RomDirectory)
	var romDirectories shared.RomDirectories

	err := fb.CWD(common.CollectionDirectory, false)
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
	}

	err = fb.CWD(common.RomDirectory, false)
	if err != nil {
		_, _ = cui.ShowMessage("Unable to fetch ROM directories! Quitting!", "3")
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
		}
	}

	var extraArgs []string
	extraArgs = append(extraArgs, "--cancel-text", "QUIT")

	selection, err := cui.DisplayList(romDirectories, "Game Manager", "", extraArgs...)
	if err != nil {
		return shared.RomDirectory{}, 1, err
	}

	if selection.ExitCode == 0 {
		return romDirectoryMap[selection.SelectedValue], selection.ExitCode, nil
	}

	return shared.RomDirectory{}, selection.ExitCode, nil
}
