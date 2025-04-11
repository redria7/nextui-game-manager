package main

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	sharedModels "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/ui"
	"nextui-game-manager/utils"
	"os"
	"strings"
)

func init() {
	common.SetLogLevel("ERROR")
	common.ConfigureEnvironment()

	config, err := state.LoadConfig()
	if err != nil {
		ui.ShowMessage("Unable to parse config.yml! Quitting!", "3")
		common.LogStandardFatal("Error loading config", err)
	}

	common.SetLogLevel(config.LogLevel)

	logger := common.GetLoggerInstance()

	logger.Debug("Config Loaded",
		zap.Object("config", config))

	state.SetConfig(config)

	appState := state.GetAppState()

	romDirectories, err := common.FetchRomDirectories()
	if err != nil {
		logger.Error("Issue fetching rom directories", zap.Error(err))
	}

	romDirectoryMap := make(map[string]sharedModels.RomDirectory)

	for _, dir := range romDirectories {
		romDirectoryMap[dir.DisplayName] = dir
	}

	appState.RomDirectories = romDirectories
	appState.RomDirectoryMap = romDirectoryMap
	appState.CurrentScreen = ui.Screens.MainMenu

	state.UpdateAppState(appState)
}

func cleanup() {
	common.CloseLogger()
}

func main() {
	defer cleanup()

	for {
		appState := state.GetAppState()

		selection := ui.ScreenFuncs[appState.CurrentScreen]()

		// Hacky way to handle bad input on deep sleep
		if strings.Contains(selection.Value, "SetRawBrightness") ||
			strings.Contains(selection.Value, "nSetRawVolume") {
			continue
		}

		switch appState.CurrentScreen {
		case ui.Screens.MainMenu:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.Loading)
				selection := strings.TrimSpace(selection.Value)

				directory := appState.RomDirectoryMap[selection]

				state.SetSection(models.Section{
					Name:           directory.DisplayName,
					LocalDirectory: directory.Path,
				})

			case 1, 2:
				os.Exit(0)
			}

		case ui.Screens.Loading:
			switch selection.Code {
			case 0:
				ui.SetScreen(ui.Screens.GamesList)
			case 1:
				ui.ShowMessage("Error listing ROMs", "3")
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.GamesList:
			switch selection.Code {
			case 0:
				state.SetSelectedFile(strings.TrimSpace(selection.Value))

				ui.SetScreen(ui.Screens.Actions)
			case 2:
				if appState.SearchFilter != "" {
					state.SetSearchFilter("")
				} else {
					ui.SetScreen(ui.Screens.MainMenu)
				}
			case 4:
				ui.SetScreen(ui.Screens.SearchBox)
			case 404:
				if appState.SearchFilter != "" {
					ui.ShowMessage("No results found for \""+appState.SearchFilter+"\"", "3")
					state.SetSearchFilter("")
					ui.SetScreen(ui.Screens.SearchBox)
				} else {
					ui.ShowMessage("This system contains no items", "3")
					ui.SetScreen(ui.Screens.MainMenu)
				}
			}

		case ui.Screens.SearchBox:
			switch selection.Code {
			case 0:
				state.SetSearchFilter(strings.TrimSpace(selection.Value))
			case 1, 2, 3:
				state.SetSearchFilter("")
			}

			ui.SetScreen(ui.Screens.GamesList)

		case ui.Screens.Actions:
			switch selection.Code {
			case 0:
				{
					state.SetSelectedAction(strings.TrimSpace(selection.Value))

					switch appState.SelectedAction {
					case models.Actions.DownloadArt:
						ui.SetScreen(ui.Screens.DownloadArt)
					case models.Actions.ReplaceArt:
						ui.SetScreen(ui.Screens.DownloadArt)
					case models.Actions.RenameRom:
						ui.SetScreen(ui.Screens.RenameRom)
					default:
						ui.SetScreen(ui.Screens.Confirm)
					}
				}
			default:
				ui.SetScreen(ui.Screens.GamesList)
			}

		case ui.Screens.Confirm:
			switch selection.Code {
			case 0:
				switch appState.SelectedAction {
				case models.Actions.DeleteArt:
					ui.SetScreen(ui.Screens.Actions)
				case models.Actions.ClearGameTracker:
					utils.ClearGameTracker()
					ui.SetScreen(ui.Screens.Actions)
				case models.Actions.DeleteRom:
					utils.DeleteRom()
					ui.SetScreen(ui.Screens.Loading)
				case models.Actions.Nuke:
					utils.Nuke()
					ui.SetScreen(ui.Screens.Loading)
				default:
					ui.SetScreen(ui.Screens.Actions)
				}
			default:
				ui.SetScreen(ui.Screens.Actions)
			}

		case ui.Screens.RenameRom:
			switch selection.Code {
			case 0:
				utils.RenameRom(selection.Value)
			}
			ui.SetScreen(ui.Screens.Actions)

		case ui.Screens.DownloadArt:
			switch selection.Code {
			case 0:
				code := ui.ShowMessageWithOptions("　　　　　　　　　　　　　　　　　　　　　　　　　", "0",
					"--background-image", state.GetAppState().LastSavedArtPath,
					"--confirm-text", "Use",
					"--confirm-show", "true",
					"--action-button", "X",
					"--action-text", "I'll Find My Own",
					"--action-show", "true",
					"--message-alignment", "bottom")

				if code == 2 || code == 4 {
					common.DeleteFile(state.GetAppState().LastSavedArtPath)
				}
			case 1:
				ui.ShowMessage("Could not find art :(", "3")
			}
			ui.SetScreen(ui.Screens.Actions)
		}

	}

}
