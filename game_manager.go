package main

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
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
	gaba.InitSDL(gaba.GabagoolOptions{
		WindowTitle:    "Game Manager",
		ShowBackground: true,
	})

	common.SetLogLevel("ERROR")

	config, err := state.LoadConfig()
	if err != nil {
		// TODO make config
	}

	common.SetLogLevel(config.LogLevel)

	logger := common.GetLoggerInstance()

	logger.Debug("Config Loaded",
		zap.Object("config", config))

	if _, err := os.Stat(utils.GetCollectionDirectory()); os.IsNotExist(err) {
		err := os.MkdirAll(utils.GetCollectionDirectory(), 0755)
		if err != nil {
			gaba.ConfirmationMessage("Unable to create Collections directory!", []gaba.FooterHelpItem{
				{ButtonName: "B", HelpText: "Quit"},
			}, gaba.MessageOptions{})
			logger.Fatal("Unable to create collection directory", zap.Error(err))
		}
	}

	state.SetConfig(config)
}

func main() {
	defer gaba.CloseSDL()
	defer common.CloseLogger()

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

		case models.ScreenNames.CollectionManagement:

		case models.ScreenNames.CollectionOptions:
			collectionOptions := screen.(ui.CollectionOptionsScreen)
			switch code {
			case 0:
				action := models.ActionMap["replace me"]

				switch action {
				case models.Actions.CollectionRename:
				case models.Actions.CollectionDelete:
					// TODO confirmation

					switch code {
					case 0:
						utils.DeleteCollection(collectionOptions.Collection)
						screen = ui.InitCollectionList(collectionOptions.SearchFilter)
					default:
						screen = ui.InitCollectionOptions(collectionOptions.Collection, collectionOptions.SearchFilter)
					}
				}
			case 2:
			}

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
					// TODO no results message
					screen = ui.InitSearch(screen.(ui.GameList).RomDirectory)
				} else {
					// TODO new items for system message
					screen = ui.InitMainMenu()
				}
			default:
				screen = ui.InitMainMenu()
			}

		case models.ScreenNames.SearchBox:
			searchFilter := ""
			switch code {
			case 0:
				searchFilter = res.(string)
			case 1, 2, 3:
				searchFilter = ""
			}

			screen = ui.InitGamesList(screen.(ui.Search).RomDirectory, searchFilter)

		case models.ScreenNames.Actions:
			as := screen.(ui.ActionsScreen)
			switch code {
			case 0:
				switch models.ActionMap["REPLACE ME"] {
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
				default:
				}
			default:
				screen = ui.InitGamesList(as.RomDirectory,
					as.SearchFilter)
			}

		case models.ScreenNames.CollectionCreate:
			cc := screen.(ui.CreateCollectionScreen)
			switch code {
			case 0:
			case 2:
				screen = ui.InitActionsScreenWithPreviousDirectory(cc.Game, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)
			}

		case models.ScreenNames.RenameRom:
			rrs := screen.(ui.RenameRomScreen)
			newName := res.(string)
			var err error
			newFilename := ""
			switch code {
			case 0:
				newFilename, err = utils.RenameRom(rrs.Game.Filename, newName, rrs.RomDirectory)
				if err != nil {
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

		}

	}

}
