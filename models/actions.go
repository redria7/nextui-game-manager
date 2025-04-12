package models

import (
	"qlova.tech/sum"
)

type Action struct {
	RenameRom,
	DownloadArt,
	DeleteArt,
	ClearGameTracker,
	ClearSaveStates,
	ArchiveRom,
	DeleteRom,
	Nuke sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Rename ROM":         Actions.RenameRom,
	"Download Art":       Actions.DownloadArt,
	"Delete Art":         Actions.DeleteArt,
	"Clear Game Tracker": Actions.ClearGameTracker,
	"Archive ROM":        Actions.ArchiveRom,
	"Delete ROM":         Actions.DeleteRom,
	"Nuclear Option":     Actions.Nuke,
}

var ActionMessages = map[sum.Int[Action]]string{
	Actions.RenameRom:        "Rename",
	Actions.DownloadArt:      "Download Art for",
	Actions.DeleteArt:        "Delete Art for",
	Actions.ClearGameTracker: "Clear Game Tracker for",
	Actions.ClearSaveStates:  "Clear Save States",
	Actions.ArchiveRom:       "Archive",
	Actions.DeleteRom:        "Delete",
	Actions.Nuke:             "Nuke",
}

var ActionKeys = []string{
	"Rename ROM",
	//"Clear Save States",
	"Archive ROM",
	"Delete ROM",
	"Nuclear Option",
}

var ActionNames = map[sum.Int[Action]]string{}

func init() {
	for name, action := range ActionMap {
		ActionNames[action] = name
	}
}
