package main

import (
	"fmt"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	commonUI "github.com/UncleJunVIP/nextui-pak-shared-functions/ui"
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

	if _, err := os.Stat(common.CollectionDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(common.CollectionDirectory, 0755)
		if err != nil {
			ui.ShowMessage("Unable to create collection directory! Quitting!", "3")
			logger.Fatal("Unable to create collection directory", zap.Error(err))
		}
	}

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
				if res.(shared.RomDirectory).DisplayName == "Collections" {
					screen = ui.InitCollectionList("")
					continue
				}

				platform := res.(shared.RomDirectory)

				screen = ui.InitGamesList(platform, "")
			case 1, 2:
				os.Exit(0)
			}

		case models.ScreenNames.CollectionsList:
			collection := res.(models.Collection)
			switch code {
			case 0:
				screen = ui.InitCollectionManagement(collection, screen.(ui.CollectionListScreen).SearchFilter)
			case 4:
			// TODO add search box
			case 404:
				ui.ShowMessage("No collections found!", "3")
				screen = ui.InitMainMenu()
			case 1, 2:
				screen = ui.InitMainMenu()
			}

		case models.ScreenNames.CollectionManagement:
			collection := screen.(ui.CollectionManagementScreen).Collection
			switch code {
			case 0:
				updatedCollection, err := utils.RemoveCollectionGame(collection, res.(shared.Item).DisplayName)
				if err != nil {
					ui.ShowMessage("Unable to remove game from collection!", "3")
					screen = ui.InitCollectionManagement(collection, screen.(ui.CollectionManagementScreen).SearchFilter)
				} else {
					screen = ui.InitCollectionManagement(updatedCollection, screen.(ui.CollectionManagementScreen).SearchFilter)
				}
			case 4:
				screen = ui.InitCollectionOptions(collection, screen.(ui.CollectionManagementScreen).SearchFilter)
			case 404:
				ui.ShowMessage("No games found in collection", "3")
				screen = ui.InitCollectionList(screen.(ui.CollectionManagementScreen).SearchFilter)
			default:
				screen = ui.InitCollectionList(screen.(ui.CollectionManagementScreen).SearchFilter)
			}

		case models.ScreenNames.CollectionOptions:
			collectionOptions := screen.(ui.CollectionOptionsScreen)
			switch code {
			case 0:
				action := models.ActionMap[res.(shared.ListSelection).SelectedValue]

				switch action {
				case models.Actions.CollectionRename:
					screen = ui.InitRenameCollectionScreen(collectionOptions.Collection)
				case models.Actions.CollectionDelete:
					message := fmt.Sprintf("Delete %s?", collectionOptions.Collection.DisplayName)

					code, err := commonUI.ShowMessageWithOptions(message, "0",
						"--confirm-text", "DO IT!",
						"--confirm-show", "true",
						"--confirm-button", "X",
						"--cancel-show", "true",
						"--cancel-text", "CHANGED MY MIND",
					)

					if err != nil {
						logger := common.GetLoggerInstance()
						logger.Info("Oh no", zap.Error(err))
					}

					switch code {
					case 0:
						utils.DeleteCollection(collectionOptions.Collection)
						screen = ui.InitCollectionList(collectionOptions.SearchFilter)
					default:
						screen = ui.InitCollectionOptions(collectionOptions.Collection, collectionOptions.SearchFilter)
					}
				}
			case 2:
				screen = ui.InitCollectionManagement(collectionOptions.Collection, collectionOptions.SearchFilter)
			}

		case models.ScreenNames.CollectionRename:
			rcs := screen.(ui.RenameCollectionScreen)
			newName := res.(models.WrappedString).Contents
			var err error
			switch code {
			case 0:
				err = utils.RenameCollection(rcs.Collection, newName)
				if err != nil {
					ui.ShowMessage("Unable to rename ROM!", "3")
				} else {
					rcs.Collection.DisplayName = newName
				}
			}

			screen = ui.InitCollectionOptions(rcs.Collection, "")

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
			as := screen.(ui.ActionsScreen)
			switch code {
			case 0:
				switch models.ActionMap[res.(shared.ListSelection).SelectedValue] {
				case models.Actions.DownloadArt:
					screen = ui.InitDownloadArtScreen(as.Game,
						as.RomDirectory,
						as.PreviousRomDirectory,
						as.SearchFilter,
						state.GetAppState().Config.ArtDownloadType)
				case models.Actions.RenameRom:
					screen = ui.InitRenameRomScreen(as.Game,
						as.RomDirectory,
						as.PreviousRomDirectory,
						as.SearchFilter)
				case models.Actions.CollectionAdd:
					screen = ui.InitAddToCollectionScreen(as.Game,
						as.RomDirectory,
						as.PreviousRomDirectory,
						as.SearchFilter)
				default:
					screen = ui.InitConfirmScreen(as.Game,
						as.RomDirectory,
						as.PreviousRomDirectory,
						as.SearchFilter,
						models.ActionMap[res.(shared.ListSelection).SelectedValue])
				}
			default:
				screen = ui.InitGamesList(as.RomDirectory,
					as.SearchFilter)
			}

		case models.ScreenNames.AddToCollection:
			atc := screen.(ui.AddToCollectionScreen)
			switch code {
			case 0:
				_, err := utils.AddCollectionGame(res.(models.Collection), atc.Game)
				if err != nil {
					ui.ShowMessage("Unable to add game to collection!", "2")
				} else {
					ui.ShowMessage("Added to collection!", "2")
					continue
				}
			case 4:
				screen = ui.InitCreateCollectionScreen(atc.Game, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
				continue
			case 404:
				screen = ui.InitCreateCollectionScreen(atc.Game, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
				continue
			}

			screen = ui.InitActionsScreenWithPreviousDirectory(atc.Game,
				atc.RomDirectory,
				atc.PreviousRomDirectory,
				atc.SearchFilter)

		case models.ScreenNames.CollectionCreate:
			cc := screen.(ui.CreateCollectionScreen)
			switch code {
			case 0:
				ui.ShowMessage("Created collection & added game!", "2")
				screen = ui.InitAddToCollectionScreen(cc.Game, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)
			case 2:
				screen = ui.InitActionsScreenWithPreviousDirectory(cc.Game, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)
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
				} else {
					screen = ui.InitActionsScreenWithPreviousDirectory(shared.Item{DisplayName: newName, Filename: newFilename},
						rrs.RomDirectory, rrs.PreviousRomDirectory, rrs.SearchFilter)
					continue
				}
			}

			screen = ui.InitActionsScreenWithPreviousDirectory(rrs.Game, rrs.RomDirectory,
				rrs.PreviousRomDirectory, rrs.SearchFilter)

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
					utils.ClearGameTracker(confirmScreen.Game.Filename, confirmScreen.RomDirectory)
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
