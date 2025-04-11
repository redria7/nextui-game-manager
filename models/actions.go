package models

import (
	"qlova.tech/sum"
)

type Action struct {
	DeleteRom,
	RenameRom,
	DownloadArt,
	DeleteArt,
	ClearGameTracker,
	Nuke sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Delete ROM":         Actions.DeleteRom,
	"Rename ROM":         Actions.RenameRom,
	"Download Art":       Actions.DownloadArt,
	"Delete Art":         Actions.DeleteArt,
	"Clear Game Tracker": Actions.ClearGameTracker,
	"Nuke All":           Actions.Nuke,
}

var ActionKeys = []string{
	"Download Art",
	"Delete Art",
	"Rename ROM",
	"Clear Game Tracker",
	"Delete ROM",
	"Nuke All",
}
