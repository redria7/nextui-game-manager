package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"time"
)

type GlobalActionsScreen struct {
}

func InitGlobalActionsScreen() GlobalActionsScreen {
	return GlobalActionsScreen{}
}

func (gas GlobalActionsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.GlobalActions
}

func (gas GlobalActionsScreen) Draw() (value interface{}, exitCode int, e error) {
	globalActions := models.GlobalActionKeys

	var globalActionEntries []gabagool.MenuItem
	for _, action := range globalActions {
		globalActionEntries = append(globalActionEntries, gabagool.MenuItem{
			Text:     action,
			Selected: false,
			Focused:  false,
			Metadata: models.GlobalActionMap[action],
		})
	}

	options := gabagool.DefaultListOptions("Global Actions", globalActionEntries)
	options.FooterHelpItems = []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gabagool.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)

		if selection.Unwrap().SelectedItem.Metadata == models.Actions.GlobalDownloadArt {
			if utils.ConfirmAction("Depending on the size of your collection\nthis process may take a while.\n\nContinue?") {
				noArt, err := utils.FindRomsWithoutArt()
				if err != nil {
					utils.ShowTimedMessage("Failed to scan for missing art", time.Second*2)
					return nil, -1, err
				}

				platformCount := len(noArt)
				missingArtCount := 0

				for _, missing := range noArt {
					missingArtCount += len(missing)
				}

				if missingArtCount == 0 {
					utils.ShowTimedMessage("All your games have art!\nGo play something!", time.Second*2)
					return nil, 0, nil
				}

				var artPaths []string

				gabagool.ProcessMessage(fmt.Sprintf("Searching for art...\n%d Platforms | %d Games Total", platformCount, missingArtCount), gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
					for romDir, games := range noArt {
						for _, game := range games {
							if artPath := utils.FindArt(romDir, game, state.GetAppState().Config.ArtDownloadType); artPath != "" {
								artPaths = append(artPaths, artPath)
							}
						}
					}
					return nil, nil
				})

				if len(artPaths) == 0 {
					utils.ShowTimedMessage("No art found!", time.Second*2)
					return
				} else {
					message := fmt.Sprintf("Art found for %d/%d games!", len(artPaths), missingArtCount)
					utils.ShowTimedMessage(message, time.Second*2)
				}
			}
		}

		return selection.Unwrap().SelectedItem.Text, 0, nil
	}

	return nil, 2, nil
}
