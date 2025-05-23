package models

import "qlova.tech/sum"

type ScreenName struct {
	MainMenu,
	GamesList,
	SearchBox,
	Actions,
	Confirm,
	RenameRom,
	DownloadArt,

	CollectionsList,
	CollectionOptions,
	CollectionManagement,
	CollectionCreate sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value interface{}, exitCode int, e error)
}
