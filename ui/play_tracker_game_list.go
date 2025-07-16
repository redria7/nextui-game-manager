package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
)

type PlayTrackerGamesListScreen struct {
	Console         string
	SearchFilter	string
}

func InitPlayTrackerGamesListScreen(console string, searchFilter string) PlayTrackerGamesListScreen {
	return PlayTrackerGamesListScreen{
		Console:              console,
		SearchFilter:         searchFilter,
	}
}

func (ptgls PlayTrackerGamesListScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayTrackerGameList
}

func (ptgls PlayTrackerGamesListScreen) Draw() (item interface{}, exitCode int, e error) {
	gamePlayMap, consoleMap, _ := state.GetPlayMaps()

	title := fmt.Sprintf("%.1fH : %s", float64(consoleMap[ptgls.Console])/3600.0, ptgls.Console)

	gamesList := gamePlayMap[ptgls.Console]

	if ptgls.SearchFilter != "" {
		title = "[Search: \"" + ptgls.SearchFilter + "\"]"
		gamesList = utils.FilterPlayList(gamesList, ptgls.SearchFilter)
	}

	var menuItems []gaba.MenuItem
	collectionMap := state.GetCollectionMap()
	for _, gamePlayAggregate := range gamesList {
		playHours := min(999, float64(gamePlayAggregate.PlayTimeTotal)/3600.0)
		romHomeStatus := utils.FindRomHomeFromAggregate(gamePlayAggregate)
		collections := collectionMap[gamePlayAggregate.Name]
		collectionString := ""
		for _, collection := range collections {
			collectionString = collectionString + string(collection.DisplayName[0])
		}
		if collectionString != "" {
			collectionString = "[" + collectionString + "] "
		}
		gameItem := gaba.MenuItem{
			Text:     fmt.Sprintf("%.1fH %s %s: %s", playHours, romHomeStatus, collectionString, gamePlayAggregate.Name),
			Selected: false,
			Focused:  false,
			Metadata: gamePlayAggregate,
		}
		menuItems = append(menuItems, gameItem)
	}

	options := gaba.DefaultListOptions(title, menuItems)

	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	options.SmallTitle = true
	options.EmptyMessage = "No Play Records Found"
	options.EnableAction = true
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Search"},
		// {ButtonName: "Menu", HelpText: "Help"},
		{ButtonName: "A", HelpText: "Details"},
	}

	// options.EnableHelp = true
	// options.HelpTitle = "Play Records Docs"
	// options.HelpText = []string{
	// 	"Hours played displays rounded up",
	// }

	selection, err := gaba.List(options)
	if err != nil {
		return nil, -1, err
	}

	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		query, err := gaba.Keyboard("")

		if err != nil {
			return nil, 1, err
		}

		if query.IsSome() {
			return query.Unwrap(), 4, nil
		}

		return nil, 4, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		game := selection.Unwrap().SelectedItem.Metadata.(models.PlayTrackingAggregate)
		return game, 0, nil
	}

	return nil, 2, nil
}
