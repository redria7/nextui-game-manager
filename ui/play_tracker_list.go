package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"qlova.tech/sum"
	"math"
)

type PlayTrackerListScreen struct {}

func InitPlayTrackerListScreen() PlayTrackerListScreen {
	return PlayTrackerListScreen{}
}

func (ptls PlayTrackerListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayTrackerList
}

// Lists available play tracker consoles
func (ptls PlayTrackerListScreen) Draw() (item interface{}, exitCode int, e error) {
	title := "Play Tracker Consoles"

	_, consolePlayMap, _ := state.GetPlayMaps()

	if consolePlayMap == nil || len(consolePlayMap) == 0 {
		return nil, 404, nil
	}

	var menuItems []gaba.MenuItem
	for console, playTime := range consolePlayMap {
		consoleItem := gaba.MenuItem{
			Text:     fmt.Sprintf("(%4.0fH) %s", min(9999, math.Ceil(float64(playTime)/3600.0)), console),
			Selected: false,
			Focused:  false,
			Metadata: console,
		}
		menuItems = append(menuItems, consoleItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		console := selection.Unwrap().SelectedItem.Metadata.(string)
		return console, 0, nil
	}

	return nil, 2, nil
}
