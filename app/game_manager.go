package main

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"log"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/ui"
	"nextui-game-manager/utils"
	"os"
	"time"
)

func init() {
	gaba.InitSDL(gaba.GabagoolOptions{
		WindowTitle:    "Game Manager",
		ShowBackground: true,
	})

	common.SetLogLevel("ERROR")
	common.InitIncludes()

	config, err := state.LoadConfig()
	if err != nil {
		config = &models.Config{
			ArtDownloadType: shared.ArtDownloadTypeFromString["BOX_ART"],
			HideEmpty:       false,
			LogLevel:        "ERROR",
		}

		err := utils.SaveConfig(config)
		if err != nil {
			log.Fatal("Unable to save config", zap.Error(err))
		}
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
			case 4:
				screen = ui.InitSettingsScreen()
			}

		case models.ScreenNames.Settings:
			screen = ui.InitMainMenu()

		case models.ScreenNames.CollectionsList:
			switch code {
			case 0:
				col := res.(models.Collection)
				screen = ui.InitCollectionManagement(col)
			default:
				screen = ui.InitMainMenu()
			}

		case models.ScreenNames.CollectionManagement:
			switch code {
			case 0:
				collection := res.(models.Collection)
				screen = ui.InitCollectionManagement(collection)
			case 4:
				screen = ui.InitCollectionOptions(screen.(ui.CollectionManagement).Collection,
					screen.(ui.CollectionManagement).SearchFilter)
			case -1, 2:
				screen = ui.InitCollectionList(screen.(ui.CollectionManagement).SearchFilter)
			}

		case models.ScreenNames.CollectionOptions:
			switch code {
			case 0, 3:
				collectionOptions := screen.(ui.CollectionOptionsScreen)
				screen = ui.InitCollectionList(collectionOptions.SearchFilter)
			case 2:
				screen = ui.InitCollectionManagement(screen.(ui.CollectionOptionsScreen).Collection)
			case 4:
				updatedCollection := res.(models.Collection)
				screen = ui.InitCollectionOptions(updatedCollection, screen.(ui.CollectionOptionsScreen).SearchFilter)
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
					screen = ui.InitActionsScreen(res.(shared.Item),
						screen.(ui.GameList).RomDirectory,
						screen.(ui.GameList).PreviousRomDirectory,
						screen.(ui.GameList).SearchFilter)
				}
			case 2:
				if screen.(ui.GameList).PreviousRomDirectory.Path != "" {
					screen = ui.InitGamesList(screen.(ui.GameList).PreviousRomDirectory, "")
				} else if screen.(ui.GameList).SearchFilter != "" {
					screen = ui.InitGamesList(screen.(ui.GameList).RomDirectory, "")
				} else {
					screen = ui.InitMainMenu()
				}
			case 4:
				screen = ui.InitSearch(screen.(ui.GameList).RomDirectory)
			case 404:
				if screen.(ui.GameList).SearchFilter != "" {
					gaba.ProcessMessage(fmt.Sprintf("No results found for %s!", screen.(ui.GameList).SearchFilter), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
						time.Sleep(1250 * time.Millisecond)
						return nil, nil
					})
					screen = ui.InitSearch(screen.(ui.GameList).RomDirectory)
				} else {
					gaba.ProcessMessage(fmt.Sprintf("%s is empty!", screen.(ui.GameList).RomDirectory.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
						time.Sleep(3 * time.Second)
						return nil, nil
					})
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
				switch models.ActionMap[res.(string)] {
				case models.Actions.DownloadArt:
					screen = ui.InitDownloadArtScreen(as.Game,
						as.RomDirectory,
						as.PreviousRomDirectory,
						as.SearchFilter,
						state.GetAppState().Config.ArtDownloadType)
				case models.Actions.DeleteArt:
					existingArtFilename, err := utils.FindExistingArt(as.Game.Filename, as.RomDirectory)
					if err != nil {
						logger.Error("failed to find existing arts", zap.Error(err))
						gaba.ProcessMessage("Unable to delete art!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
							time.Sleep(3 * time.Second)
							return nil, nil
						})
						break
					}

					result, err := gaba.ConfirmationMessage("Delete this beautiful art?", []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "I Changed My Mind"},
						{ButtonName: "A", HelpText: "Trash It!"},
					},
						gaba.MessageOptions{
							ImagePath: existingArtFilename,
						})
					if err != nil || result.IsSome() {
						common.DeleteFile(existingArtFilename)
						screen = ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
					}
				case models.Actions.RenameRom:
					newName, err := gaba.Keyboard(as.Game.DisplayName)
					if err != nil {
						gaba.ProcessMessage("Unable to rename ROM!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
							time.Sleep(3 * time.Second)
							return nil, nil
						})
						break
					}
					if newName.IsSome() {
						path, err := utils.RenameRom(as.Game, newName.Unwrap(), as.RomDirectory)
						if err != nil {
							gaba.ProcessMessage("Unable to rename ROM!", gaba.ProcessMessageOptions{}, func() (interface{}, error) {
								time.Sleep(3 * time.Second)
								return nil, nil
							})
						} else {
							as.Game.DisplayName = newName.Unwrap()
							as.Game.Filename = path
							screen = ui.InitActionsScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
						}
					}

				case models.Actions.CollectionAdd:
					screen = ui.InitAddToCollectionScreen(as.Game, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)

				default:
				}
			default:
				screen = ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory,
					as.SearchFilter)
			}

		case models.ScreenNames.AddToCollection:
			atc := screen.(ui.AddToCollectionScreen)
			switch code {
			case 0:
				screen = ui.InitAddToCollectionScreen(atc.Game,
					atc.RomDirectory,
					atc.PreviousRomDirectory,
					atc.SearchFilter)
			case 404:
				screen = ui.InitCreateCollectionScreen(atc.Game, atc.RomDirectory, atc.PreviousRomDirectory, atc.SearchFilter)
			default:
				screen = ui.InitActionsScreen(atc.Game,
					atc.RomDirectory,
					atc.PreviousRomDirectory,
					atc.SearchFilter)

			}

		case models.ScreenNames.CollectionCreate:
			cc := screen.(ui.CreateCollectionScreen)
			screen = ui.InitAddToCollectionScreen(cc.Game, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)

		case models.ScreenNames.DownloadArt:
			switch code {
			default:
				screen = ui.InitActionsScreen(screen.(ui.DownloadArtScreen).Game,
					screen.(ui.DownloadArtScreen).RomDirectory,
					screen.(ui.DownloadArtScreen).PreviousRomDirectory,
					screen.(ui.DownloadArtScreen).SearchFilter)
			}

		}

	}

}
