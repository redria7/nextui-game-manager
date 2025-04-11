package models

import (
	"qlova.tech/sum"
)

type Action struct {
	DeleteRom,
	RenameRom,
	DownloadArt,
	ReplaceArt,
	DeleteArt,
	ClearGameTracker,
	Nuke sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Delete ROM":         Actions.DeleteRom,
	"Rename ROM":         Actions.RenameRom,
	"Download Art":       Actions.DownloadArt,
	"Replace Art":        Actions.ReplaceArt,
	"Delete Art":         Actions.DeleteArt,
	"Clear Game Tracker": Actions.ClearGameTracker,
	"Nuke All":           Actions.Nuke,
}

var ActionKeys = []string{
	"Rename ROM",
	"Clear Game Tracker",
	"Delete ROM",
	"Nuke All",
}

var ActionNames = map[sum.Int[Action]]string{}

func init() {
	for name, action := range ActionMap {
		ActionNames[action] = name
	}
}
