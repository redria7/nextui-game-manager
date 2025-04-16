package main

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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
	fb := filebrowser.NewFileBrowser(logger)

	var romDirectories []shared.RomDirectory

	err = fb.CWD(common.CollectionDirectory)
	if err != nil {
		logger.Info("Unable to fetch Collection directories! Continuing without them", zap.Error(err))
	}

	if len(fb.Items) > 0 {
		romDirectories = append(romDirectories, shared.RomDirectory{
			DisplayName: "Collections",
			Tag:         "Collections",
			Path:        common.CollectionDirectory,
		})
	}

	err = fb.CWD(common.RomDirectory)
	if err != nil {
		_, _ = commonUI.ShowMessage("Unable to fetch ROM directories! Quitting!", "3")
		common.LogStandardFatal("Error loading fetching ROM directories", err)
	}

	romDirectoryMap := make(map[string]shared.RomDirectory)

	for _, item := range fb.Items {
		romDirectory := shared.RomDirectory{
			DisplayName: item.DisplayName,
			Tag:         item.Tag,
			Path:        item.Path,
		}
		romDirectories = append(romDirectories, romDirectory)
		romDirectoryMap[item.DisplayName] = romDirectory
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
	logger := common.GetLoggerInstance()

	for {
		appState := state.GetAppState()

		selection, err := ui.ScreenFuncs[appState.CurrentScreen]()
		if err != nil {
			logger.Error("Error loading screen")
		}

		// Hacky way to handle bad input on deep sleep
		if strings.Contains(selection.Value, "SetRawBrightness") ||
			strings.Contains(selection.Value, "nSetRawVolume") {
			continue
		}

		switch appState.CurrentScreen {
		case ui.Screens.MainMenu:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.Loading)
				selection := strings.TrimSpace(selection.Value)

				if selection == "Collections" {
					ui.SetScreen(ui.Screens.CollectionsList)
					continue
				}

				directory := appState.RomDirectoryMap[selection]

				state.SetSection(models.Section{
					Name:           directory.DisplayName,
					LocalDirectory: directory.Path,
				})

			case 1, 2:
				os.Exit(0)
			}

		case ui.Screens.Loading:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.GamesList)
			case 1:
				ui.ShowMessage("Error listing ROMs", "3")
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.GamesList:
			switch selection.ExitCode {
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

		case ui.Screens.CollectionsList:
			switch selection.ExitCode {
			case 0:
				selection := strings.TrimSpace(selection.Value)
				collection := appState.CollectionDirectoryMap[selection]

				state.SetSection(models.Section{
					Name:               collection.DisplayName,
					CollectionFilePath: collection.Path,
				})
				ui.SetScreen(ui.Screens.CollectionManagement)
			case 2:
				if appState.SearchFilter != "" {
					state.SetSearchFilter("")
				} else {
					ui.SetScreen(ui.Screens.MainMenu)
				}
			default:
				ui.SetScreen(ui.Screens.MainMenu)
			}

		case ui.Screens.CollectionManagement:
			switch selection.ExitCode {
			case 0:
				ui.SetScreen(ui.Screens.Actions)
			case 4:
				selection := strings.TrimSpace(selection.Value)
				collection := appState.CollectionDirectoryMap[selection]

				state.SetSection(models.Section{
					Name:               collection.DisplayName,
					CollectionFilePath: collection.Path,
				})

				ui.SetScreen(ui.Screens.CollectionOptions)
			default:
				ui.SetScreen(ui.Screens.CollectionsList)
			}

		case ui.Screens.CollectionOptions:
			switch selection.ExitCode {
			case 0:
				action := strings.TrimSpace(selection.Value)

				state.SetSelectedAction(action)

				switch models.ActionMap[action] {
				case models.Actions.CollectionRename:
					ui.SetScreen(ui.Screens.RenameCollection)
				case models.Actions.CollectionDelete:
					ui.SetScreen(ui.Screens.Confirm)
				}
			default:
				ui.SetScreen(ui.Screens.CollectionManagement)

			}

		case ui.Screens.SearchBox:
			switch selection.ExitCode {
			case 0:
				state.SetSearchFilter(strings.TrimSpace(selection.Value))
			case 1, 2, 3:
				state.SetSearchFilter("")
			}

			ui.SetScreen(ui.Screens.GamesList)

		case ui.Screens.Actions:
			switch selection.ExitCode {
			case 0:
				{
					state.SetSelectedAction(strings.TrimSpace(selection.Value))

					switch appState.SelectedAction {
					case models.Actions.DownloadArt:
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
			switch selection.ExitCode {
			case 0:
				switch appState.SelectedAction {
				case models.Actions.DeleteArt:
					utils.DeleteArt()
					ui.SetScreen(ui.Screens.Actions)
				case models.Actions.ClearGameTracker:
					utils.ClearGameTracker()
					ui.SetScreen(ui.Screens.Actions)
				case models.Actions.ClearSaveStates:
					utils.ClearSaveStates()
					ui.SetScreen(ui.Screens.Actions)
				case models.Actions.ArchiveRom:
					utils.ArchiveRom()
					ui.SetScreen(ui.Screens.GamesList)
				case models.Actions.DeleteRom:
					utils.DeleteRom()
					ui.SetScreen(ui.Screens.GamesList)
				case models.Actions.Nuke:
					utils.Nuke()
					ui.SetScreen(ui.Screens.GamesList)
				default:
					ui.SetScreen(ui.Screens.Actions)
				}
			default:
				ui.SetScreen(ui.Screens.Actions)
			}

		case ui.Screens.RenameRom:
			switch selection.ExitCode {
			case 0:
				utils.RenameRom(selection.Value)
			}
			ui.SetScreen(ui.Screens.Actions)

		case ui.Screens.DownloadArt:
			switch selection.ExitCode {
			case 0:
				logger.Debug("Showing Art Download", zap.String("last_saved_art_path", state.GetAppState().LastSavedArtPath))

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
