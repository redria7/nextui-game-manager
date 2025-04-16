package models

import "qlova.tech/sum"

type Screen struct {
	MainMenu,
	Loading,
	GamesList,
	SearchBox,
	Actions,
	AddToCollection,
	Confirm,
	RenameRom,
	DownloadArt,

	CollectionsList,
	CollectionOptions,
	RenameCollection,
	CollectionManagement sum.Int[Screen]
}
