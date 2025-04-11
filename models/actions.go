package models

import "qlova.tech/sum"

type Actions struct {
	DeleteRom,
	DownloadArt,
	DeleteArt,
	ClearGameTracker,
	Nuke sum.Int[Screen]
}

var ActionMap = map[string]sum.Int[Screen]{
	"Delete Rom":         Actions{}.DeleteRom,
	"Download Art":       Actions{}.DownloadArt,
	"Delete Art":         Actions{}.DeleteArt,
	"Clear Game Tracker": Actions{}.ClearGameTracker,
	"Nuke All":           Actions{}.Nuke,
}

var ActionKeys = []string{
	"Download Art",
	"Delete Art",
	"Clear Game Tracker",
	"Delete Rom",
	"Nuke All",
}
