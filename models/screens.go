package models

import "qlova.tech/sum"

type Screen struct {
	MainMenu,
	Loading,
	GamesList,
	SearchBox,
	Actions,
	Confirm,
	DownloadArt sum.Int[Screen]
}
