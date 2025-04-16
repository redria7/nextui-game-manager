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
	Nuke,

	CollectionRename,
	CollectionDelete sum.Int[Action]
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

	"Rename Collection": Actions.CollectionRename,
	"Delete Collection": Actions.CollectionDelete,
}

var ActionMessages = map[sum.Int[Action]]string{
	Actions.RenameRom:        "Rename",
	Actions.DownloadArt:      "Download Art for",
	Actions.DeleteArt:        "Delete Art for",
	Actions.ClearGameTracker: "Clear Game Tracker for",
	Actions.ClearSaveStates:  "Clear Save States",
	Actions.ArchiveRom:       "Archive",
	Actions.DeleteRom:        "Delete",
	Actions.Nuke:             "Nuke (Deletes ROM, Art and Game Tracker)",
	Actions.CollectionDelete: "Delete",
	Actions.CollectionRename: "Rename",
}

var ActionKeys = []string{
	"Rename ROM",
	//"Clear Save States",
	"Archive ROM",
	"Delete ROM",
	"Nuclear Option",
}

var CollectionActionKeys = []string{
	"Rename Collection",
	"Delete Collection",
}

var ActionNames = map[sum.Int[Action]]string{}

func init() {
	for name, action := range ActionMap {
		ActionNames[action] = name
	}
}
