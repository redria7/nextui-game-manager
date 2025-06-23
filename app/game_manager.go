package main

import (
	"fmt"
	_ "github.com/UncleJunVIP/certifiable"
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
				selections := res.(shared.Items)

				if len(selections) == 0 {
					break
				} else if len(selections) == 1 {
					selection := selections[0]
					if selection.IsDirectory {
						screen = ui.InitGamesListWithPreviousDirectory(shared.RomDirectory{
							DisplayName: selection.DisplayName,
							Tag:         selection.Tag,
							Path:        selection.Path,
						}, screen.(ui.GameList).RomDirectory, "")
					} else {
						screen = ui.InitActionsScreen(selection,
							screen.(ui.GameList).RomDirectory,
							screen.(ui.GameList).PreviousRomDirectory,
							screen.(ui.GameList).SearchFilter)
					}
				} else {
					screen = ui.InitBulkOptionsScreen(selections, screen.(ui.GameList).RomDirectory,
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
					screen = ui.InitAddToCollectionScreen([]shared.Item{as.Game}, as.RomDirectory, as.PreviousRomDirectory, as.SearchFilter)
				case models.Actions.ClearGameTracker:
					result, err := gaba.ConfirmationMessage(fmt.Sprintf("Clear %s from Game Tracker?", as.Game.DisplayName), []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "I Changed My Mind"},
						{ButtonName: "A", HelpText: "Do It!"},
					},
						gaba.MessageOptions{})
					if err != nil || result.IsSome() {
						res := utils.ClearGameTracker(as.Game.Filename, as.RomDirectory)
						if res {
							gaba.ProcessMessage("Game Tracker data cleared!",
								gaba.ProcessMessageOptions{}, func() (interface{}, error) {
									time.Sleep(3 * time.Second)
									return nil, nil
								})
						} else {
							gaba.ProcessMessage("Unable to clear Game Tracker data!",
								gaba.ProcessMessageOptions{}, func() (interface{}, error) {
									time.Sleep(3 * time.Second)
									return nil, nil
								})
						}
					}
				case models.Actions.ArchiveRom:
					result, err := gaba.ConfirmationMessage(fmt.Sprintf("Archive %s?", as.Game.DisplayName), []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "I Changed My Mind"},
						{ButtonName: "A", HelpText: "Yes"},
					},
						gaba.MessageOptions{})
					if err != nil || result.IsSome() {
						err = utils.ArchiveRom(as.Game, as.RomDirectory)
						if err != nil {
							gaba.ProcessMessage(fmt.Sprintf("Unable to archive %s!", as.Game.DisplayName), gaba.ProcessMessageOptions{}, func() (interface{}, error) {
								time.Sleep(3 * time.Second)
								return nil, nil
							})
						} else {
							screen = ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory,
								as.SearchFilter)
						}
					}
				case models.Actions.DeleteRom:
					result, err := gaba.ConfirmationMessage(fmt.Sprintf("Delete %s?", as.Game.DisplayName), []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "I Changed My Mind"},
						{ButtonName: "A", HelpText: "Yes"},
					},
						gaba.MessageOptions{})
					if err != nil || result.IsSome() {
						utils.DeleteRom(as.Game, as.RomDirectory)
						screen = ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory,
							as.SearchFilter)
					}
				case models.Actions.Nuke:
					result, err := gaba.ConfirmationMessage(fmt.Sprintf("Nuke %s?", as.Game.DisplayName), []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "I Changed My Mind"},
						{ButtonName: "A", HelpText: "Yes"},
					},
						gaba.MessageOptions{})
					if err != nil || result.IsSome() {
						utils.Nuke(as.Game, as.RomDirectory)
						screen = ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory,
							as.SearchFilter)
					}
				default:
				}
			default:
				screen = ui.InitGamesListWithPreviousDirectory(as.RomDirectory, as.PreviousRomDirectory,
					as.SearchFilter)
			}

		case models.ScreenNames.BulkActions:
			ba := screen.(ui.BulkOptionsScreen)
			switch code {
			case 0:
				switch models.ActionMap[res.(string)] {
				case models.Actions.DownloadArt:
					var artPaths []string

					gaba.ProcessMessage(fmt.Sprintf("Downloading art for %d games...", len(ba.Games)),
						gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
							for _, game := range ba.Games {
								artPath := utils.FindArt(ba.RomDirectory, game, state.GetAppState().Config.ArtDownloadType)

								if artPath != "" {
									artPaths = append(artPaths, artPath)
								}
							}
							return nil, nil
						})

					if len(artPaths) == 0 {
						gaba.ProcessMessage("No art found!",
							gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
								time.Sleep(time.Millisecond * 2000)
								return nil, nil
							})
					} else if len(ba.Games) > 1 {
						gaba.ProcessMessage(fmt.Sprintf("Art found for %d/%d games!", len(artPaths), len(ba.Games)),
							gaba.ProcessMessageOptions{ShowThemeBackground: true}, func() (interface{}, error) {
								time.Sleep(time.Millisecond * 2000)
								return nil, nil
							})
					}
				case models.Actions.DeleteArt:
					confirm, _ := gaba.ConfirmationMessage("Delete art for the selected games?", []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "Cancel"},
						{ButtonName: "X", HelpText: "Remove"},
					}, gaba.MessageOptions{
						ImagePath:     "",
						ConfirmButton: gaba.ButtonX,
					})

					if confirm.IsSome() && !confirm.Unwrap().Cancelled {
						for _, game := range ba.Games {
							utils.DeleteArt(game.Filename, ba.RomDirectory)
						}
					}
				case models.Actions.CollectionAdd:
					screen = ui.InitAddToCollectionScreen(ba.Games, ba.RomDirectory, ba.PreviousRomDirectory, ba.SearchFilter)
				case models.Actions.ArchiveRom:
					confirm, _ := gaba.ConfirmationMessage("Archive the selected games?", []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "Cancel"},
						{ButtonName: "X", HelpText: "Remove"},
					}, gaba.MessageOptions{
						ImagePath:     "",
						ConfirmButton: gaba.ButtonX,
					})

					if confirm.IsSome() && !confirm.Unwrap().Cancelled {
						for _, game := range ba.Games {
							utils.ArchiveRom(game, ba.RomDirectory)
						}
					}
				case models.Actions.DeleteRom:
					confirm, _ := gaba.ConfirmationMessage("Delete the selected games?", []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "Cancel"},
						{ButtonName: "X", HelpText: "Remove"},
					}, gaba.MessageOptions{
						ImagePath:     "",
						ConfirmButton: gaba.ButtonX,
					})

					if confirm.IsSome() && !confirm.Unwrap().Cancelled {
						for _, game := range ba.Games {
							utils.DeleteRom(game, ba.RomDirectory)
						}
					}
				case models.Actions.Nuke:
					confirm, _ := gaba.ConfirmationMessage("Nuke the selected games?", []gaba.FooterHelpItem{
						{ButtonName: "B", HelpText: "Cancel"},
						{ButtonName: "X", HelpText: "Remove"},
					}, gaba.MessageOptions{
						ImagePath:     "",
						ConfirmButton: gaba.ButtonX,
					})

					if confirm.IsSome() && !confirm.Unwrap().Cancelled {
						for _, game := range ba.Games {
							utils.Nuke(game, ba.RomDirectory)
						}
					}
				}
			default:
				screen = ui.InitGamesListWithPreviousDirectory(ba.RomDirectory, ba.PreviousRomDirectory,
					ba.SearchFilter)
			}

		case models.ScreenNames.AddToCollection:
			atc := screen.(ui.AddToCollectionScreen)
			switch code {
			case 0:
				screen = ui.InitAddToCollectionScreen(atc.Games,
					atc.RomDirectory,
					atc.PreviousRomDirectory,
					atc.SearchFilter)
			case 4, 404:
				screen = ui.InitCreateCollectionScreen(atc.Games, atc.RomDirectory,
					atc.PreviousRomDirectory, atc.SearchFilter)
			case 2:
				gameCount := len(atc.Games)
				if gameCount == 1 {
					screen = ui.InitActionsScreen(atc.Games[0],
						atc.RomDirectory,
						atc.PreviousRomDirectory,
						atc.SearchFilter)
				} else {
					screen = ui.InitBulkOptionsScreen(atc.Games,
						atc.RomDirectory,
						atc.PreviousRomDirectory,
						atc.SearchFilter)
				}

			}

		case models.ScreenNames.CollectionCreate:
			cc := screen.(ui.CreateCollectionScreen)
			screen = ui.InitAddToCollectionScreen(cc.Games, cc.RomDirectory, cc.PreviousRomDirectory, cc.SearchFilter)

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
