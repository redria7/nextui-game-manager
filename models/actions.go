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
	ArchiveRename,
	ArchiveDelete,
	DeleteRom,
	Nuke,

	CollectionRename,
	CollectionDelete,
	CollectionAdd sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Rename ROM":         Actions.RenameRom,
	"Download Art":       Actions.DownloadArt,
	"Delete Art":         Actions.DeleteArt,
	"Clear Game Tracker": Actions.ClearGameTracker,
	"Archive ROM":        Actions.ArchiveRom,
	"Rename Archive": 	  Actions.ArchiveRename,
	"Delete Archive": 	  Actions.ArchiveDelete,
	"Delete ROM":         Actions.DeleteRom,
	"Nuclear Option":     Actions.Nuke,

	"Rename Collection": Actions.CollectionRename,
	"Delete Collection": Actions.CollectionDelete,
	"Add to Collection": Actions.CollectionAdd,
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
	Actions.ArchiveRename: "Rename",
	Actions.ArchiveDelete: "Delete",
	Actions.CollectionRename: "Rename",
	Actions.CollectionAdd:    "Add to",
}

var ActionKeys = []string{
	"Rename ROM",
	"Add to Collection",
	//"Clear Save States",
	"Archive ROM",
	"Delete ROM",
	//"Nuclear Option",
}

var BulkActionKeys = []string{
	"Add to Collection",
	"Download Art",
	"Delete Art",
	//"Clear Game Tracker",
	//"Archive ROM",
	//"Delete ROM",
	//"Nuclear Option",
}

var CollectionActionKeys = []string{
	"Rename Collection",
	"Delete Collection",
}

var ArchiveActionKeys = []string{
	"Rename Archive",
	"Delete Archive",
}

var ActionNames = map[sum.Int[Action]]string{}

func init() {
	for name, action := range ActionMap {
		ActionNames[action] = name
	}
}
