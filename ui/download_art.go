package ui

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"nextui-game-manager/models"
	"qlova.tech/sum"
)

type DownloadArtScreen struct {
	Game                 shared.Item
	RomDirectory         shared.RomDirectory
	SearchFilter         string
	PreviousRomDirectory shared.RomDirectory
	DownloadType         sum.Int[shared.ArtDownloadType]
}

func InitDownloadArtScreen(game shared.Item, romDirectory shared.RomDirectory,
	previousRomDirectory shared.RomDirectory,
	searchFilter string, downloadType sum.Int[shared.ArtDownloadType]) DownloadArtScreen {
	return DownloadArtScreen{
		Game:                 game,
		RomDirectory:         romDirectory,
		PreviousRomDirectory: previousRomDirectory,
		SearchFilter:         searchFilter,
		DownloadType:         downloadType,
	}
}

func (da DownloadArtScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadArt
}

func (da DownloadArtScreen) Draw() (value interface{}, exitCode int, e error) {
	artPath := "" // TODO use art fetcher

	if artPath == "" {
		// TODO say no art found
		return shared.Item{}, 404, nil
	}

	// TODO use this art?

	return nil, 0, nil

}
