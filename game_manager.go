package main

import (
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/ui"
	"nextui-game-manager/utils"
	"os"
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
}

func cleanup() {
	common.CloseLogger()
}

func main() {
	defer cleanup()
	logger := common.GetLoggerInstance()

	logger.Info("Starting Game Manager")

	var screen models.Screen
	screen = ui.InitMainMenu()

	for {
		res, code, _ := screen.Draw() // TODO figure out error handling

		switch screen.Name() {
		case models.ScreenNames.MainMenu:
			switch code {
			case 0:
				if res.Value() == "Collections" {
					screen = ui.InitCollectionList("")
					continue
				}

				platform := res.(shared.RomDirectory)

				screen = ui.InitGamesList(platform, "")
			case 1, 2:
				os.Exit(0)
			}

		case models.ScreenNames.CollectionsList:
			switch code {
			case 0:
				// screen = ui.InitCollectionManagement()
			case 4:
				// screen = ui.InitCollectionOptions()
			case 1, 2:
				screen = ui.InitMainMenu()
			}

		case models.ScreenNames.CollectionOptions:

		case models.ScreenNames.CollectionManagement:

		case models.ScreenNames.GamesList:
			switch code {
			case 0:
				selection := res.(shared.Item)
				if selection.IsDirectory {
					screen = ui.InitGamesListWithPreviousDirectory(shared.RomDirectory{
						DisplayName: selection.DisplayName,
						Tag:         selection.Tag,
						Path:        selection.Path,
					}, screen.(ui.GameList).RomDirectory, "")
				} else {
					screen = ui.InitActionsScreen(selection, screen.(ui.GameList).RomDirectory,
						screen.(ui.GameList).SearchFilter)
				}
			case 2:
				if screen.(ui.GameList).PreviousRomDirectory.Path != "" {
					screen = ui.InitGamesList(screen.(ui.GameList).PreviousRomDirectory, "")
				} else if screen.(ui.GameList).SearchFilter != "" {
					ui.InitGamesList(screen.(ui.Search).RomDirectory, "")
				} else {
					screen = ui.InitMainMenu()
				}
			case 4:
				screen = ui.InitSearch(screen.(ui.GameList).RomDirectory)
			case 404:
				if screen.(ui.GameList).SearchFilter != "" {
					ui.ShowMessage("No results found for \""+screen.(ui.GameList).SearchFilter+"\"", "3")
					screen = ui.InitSearch(screen.(ui.GameList).RomDirectory)
				} else {
					ui.ShowMessage("This system contains no items", "3")
					screen = ui.InitMainMenu()
				}
			default:
				screen = ui.InitMainMenu()
			}

		case models.ScreenNames.SearchBox:
			searchFilter := ""
			switch code {
			case 0:
				searchFilter = res.(models.WrappedString).Contents
			case 1, 2, 3:
				searchFilter = ""
			}

			screen = ui.InitGamesList(screen.(ui.Search).RomDirectory, searchFilter)

		case models.ScreenNames.Actions:
			switch code {
			case 0:
				switch models.ActionMap[res.(shared.ListSelection).SelectedValue] {
				case models.Actions.DownloadArt:
					screen = ui.InitDownloadArtScreen(screen.(ui.ActionsScreen).Game,
						screen.(ui.ActionsScreen).RomDirectory,
						screen.(ui.ActionsScreen).PreviousRomDirectory,
						screen.(ui.ActionsScreen).SearchFilter,
						state.GetAppState().Config.ArtDownloadType)
				case models.Actions.RenameRom:
					screen = ui.InitRenameRomScreen(screen.(ui.ActionsScreen).Game,
						screen.(ui.ActionsScreen).RomDirectory,
						screen.(ui.ActionsScreen).PreviousRomDirectory,
						screen.(ui.ActionsScreen).SearchFilter)
				default:
					screen = ui.InitConfirmScreen(screen.(ui.ActionsScreen).Game,
						screen.(ui.ActionsScreen).RomDirectory,
						screen.(ui.ActionsScreen).PreviousRomDirectory,
						screen.(ui.ActionsScreen).SearchFilter,
						models.ActionMap[res.(shared.ListSelection).SelectedValue])
				}
			default:
				screen = ui.InitGamesList(screen.(ui.ActionsScreen).RomDirectory,
					screen.(ui.ActionsScreen).SearchFilter)
			}

		case models.ScreenNames.RenameRom:
			rrs := screen.(ui.RenameRomScreen)
			newName := res.(models.WrappedString).Contents
			var err error
			newFilename := ""
			switch code {
			case 0:
				newFilename, err = utils.RenameRom(rrs.Game.Filename, newName, rrs.RomDirectory)
				if err != nil {
					ui.ShowMessage("Unable to rename ROM!", "3")
					screen = ui.InitActionsScreenWithPreviousDirectory(rrs.Game, rrs.RomDirectory,
						rrs.PreviousRomDirectory, rrs.SearchFilter)
					continue
				}
			}
			screen = ui.InitActionsScreenWithPreviousDirectory(shared.Item{DisplayName: newName, Filename: newFilename},
				rrs.RomDirectory, rrs.PreviousRomDirectory, rrs.SearchFilter)

		case models.ScreenNames.DownloadArt:
			switch code {
			default:
				screen = ui.InitActionsScreenWithPreviousDirectory(screen.(ui.DownloadArtScreen).Game,
					screen.(ui.DownloadArtScreen).RomDirectory,
					screen.(ui.DownloadArtScreen).PreviousRomDirectory,
					screen.(ui.DownloadArtScreen).SearchFilter)
			}

		case models.ScreenNames.Confirm:
			confirmScreen := screen.(ui.ConfirmScreen)
			switch code {
			case 0:
				switch confirmScreen.Action {
				case models.Actions.DeleteArt:
					utils.DeleteArt(confirmScreen.Game.Filename, confirmScreen.RomDirectory)
					screen = ui.InitActionsScreen(confirmScreen.Game, confirmScreen.RomDirectory,
						confirmScreen.SearchFilter)
				case models.Actions.ClearGameTracker:
					utils.ClearGameTracker(confirmScreen.Game.DisplayName, confirmScreen.RomDirectory)
					screen = ui.InitActionsScreen(confirmScreen.Game, confirmScreen.RomDirectory,
						confirmScreen.SearchFilter)
				case models.Actions.ClearSaveStates:
					utils.ClearSaveStates()
					screen = ui.InitActionsScreen(confirmScreen.Game, confirmScreen.RomDirectory,
						confirmScreen.SearchFilter)
				case models.Actions.ArchiveRom:
					utils.ArchiveRom(confirmScreen.Game.Filename, confirmScreen.RomDirectory)
					screen = ui.InitGamesListWithPreviousDirectory(confirmScreen.RomDirectory,
						confirmScreen.PreviousRomDirectory, confirmScreen.SearchFilter)
				case models.Actions.DeleteRom:
					utils.DeleteRom(confirmScreen.Game.Filename, confirmScreen.RomDirectory)
					screen = ui.InitGamesListWithPreviousDirectory(confirmScreen.RomDirectory,
						confirmScreen.PreviousRomDirectory, confirmScreen.SearchFilter)
				case models.Actions.Nuke:
					utils.Nuke(confirmScreen.Game.Filename, confirmScreen.RomDirectory)
					screen = ui.InitGamesListWithPreviousDirectory(confirmScreen.RomDirectory,
						confirmScreen.PreviousRomDirectory, confirmScreen.SearchFilter)
				case models.Actions.CollectionDelete:
					common.DeleteFile(confirmScreen.Game.Path)
					screen = ui.InitCollectionList(confirmScreen.SearchFilter)
				default:
					screen = ui.InitActionsScreenWithPreviousDirectory(confirmScreen.Game,
						confirmScreen.RomDirectory,
						confirmScreen.PreviousRomDirectory,
						confirmScreen.SearchFilter)
				}
			default:
				screen = ui.InitActionsScreenWithPreviousDirectory(confirmScreen.Game,
					confirmScreen.RomDirectory,
					confirmScreen.PreviousRomDirectory,
					confirmScreen.SearchFilter)
			}
		}

	}

}
