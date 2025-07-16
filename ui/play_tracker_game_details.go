package ui

import (
	"fmt"
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-game-manager/models"
	"nextui-game-manager/state"
	"nextui-game-manager/utils"
	"qlova.tech/sum"
	"strconv"
	"time"
)

type PlayTrackerGameDetailsScreen struct {
	Console         		string
	SearchFilter			string
	GameAggregate			models.PlayTrackingAggregate
	Game                 	shared.Item
	RomDirectory         	shared.RomDirectory
	PreviousRomDirectory 	shared.RomDirectory
	PlayTrackerOrigin		bool
}

func InitPlayTrackerGameDetailsScreenFromPlayTracker(console string, searchFilter string, gameAggregate models.PlayTrackingAggregate) PlayTrackerGameDetailsScreen {
	return PlayTrackerGameDetailsScreen{
		Console:      		console,
		SearchFilter: 		searchFilter,
		GameAggregate: 		gameAggregate,
		PlayTrackerOrigin: 	true,
	}
}

func InitPlayTrackerGameDetailsScreenFromActions(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory, searchFilter string) PlayTrackerGameDetailsScreen {
	gamePlayMap, _, _ := state.GetPlayMaps()
	gameAggregate, console := utils.CollectGameAggregateFromGame(game, gamePlayMap)
	return PlayTrackerGameDetailsScreen{
		Console:				console,
		SearchFilter: 			searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayTrackerOrigin: 		false,
	}
}

func InitPlayTrackerGameDetailsScreenFromSelf(console string, searchFilter string, gameAggregate models.PlayTrackingAggregate, game shared.Item, 
	romDirectory shared.RomDirectory, previousRomDirectory shared.RomDirectory, playTrackerOrigin bool) PlayTrackerGameDetailsScreen {
	return PlayTrackerGameDetailsScreen{
		Console:				console,
		SearchFilter: 			searchFilter,
		GameAggregate: 			gameAggregate,
		Game:      				game,
		RomDirectory: 			romDirectory,
		PreviousRomDirectory:	previousRomDirectory,
		PlayTrackerOrigin: 		playTrackerOrigin,
	}
}

func (ptgds PlayTrackerGameDetailsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.PlayTrackerGameDetails
}

func (ptgds PlayTrackerGameDetailsScreen) Draw() (selection interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	var sections []gaba.Section

	_, consolePlayMap, totalPlay := state.GetPlayMaps()
	sections = append(sections, gaba.NewInfoSection(
		"",
		[]gaba.MetadataItem{
			{Label: "Total Play Time", Value: utils.ConvertSecondsToHumanReadable(ptgds.GameAggregate.PlayTimeTotal)},
			{Label: "  Play Sessions", Value: strconv.Itoa(ptgds.GameAggregate.PlayCountTotal)},
			{Label: "   First Played", Value: ptgds.GameAggregate.FirstPlayedTime.Format(time.UnixDate)},
			{Label: "    Last Played", Value: ptgds.GameAggregate.LastPlayedTime.Format(time.UnixDate)},
			{Label: "Average Session", Value: utils.ConvertSecondsToHumanReadable(ptgds.GameAggregate.PlayTimeTotal/ptgds.GameAggregate.PlayCountTotal)},
			{Label: "   Pct of Total", Value: fmt.Sprintf("%.2f%%", (float64(ptgds.GameAggregate.PlayTimeTotal)/float64(totalPlay))*100)},
			{Label: " Pct of Console", Value: fmt.Sprintf("%.2f%%", (float64(ptgds.GameAggregate.PlayTimeTotal)/float64(consolePlayMap[ptgds.Console]))*100)},
		},
	))

	// sections = append(sections, gaba.NewImageSection(
	// 	"Pak Repository",
	// 	"", //                                                   add a filePath string
	// 	int32(256),
	// 	int32(256),
	// 	gaba.AlignCenter,
	// ))

	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false

	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "History"},
	}

	sel, err := gaba.DetailScreen(fmt.Sprintf("%s Play Stats", ptgds.GameAggregate.Name), options, footerItems)
	if err != nil {
		logger.Error("Unable to display Play History screen", zap.Error(err))
		return nil, -1, err
	}

	if sel.IsNone() {
		return nil, 2, nil
	}

	return nil, 0, nil
}
