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
	CollectionAdd,

	GlobalDownloadArt sum.Int[Action]
}

var Actions = sum.Int[Action]{}.Sum()

var ActionMap = map[string]sum.Int[Action]{
	"Rename ROM":         Actions.RenameRom,
	"Download Art":       Actions.DownloadArt,
	"Delete Art":         Actions.DeleteArt,
	"Clear Game Tracker": Actions.ClearGameTracker,
	"Archive ROM":        Actions.ArchiveRom,
	"Rename Archive":     Actions.ArchiveRename,
	"Delete Archive":     Actions.ArchiveDelete,
	"Delete ROM":         Actions.DeleteRom,
	"Nuclear Option":     Actions.Nuke,

	"Rename Collection": Actions.CollectionRename,
	"Delete Collection": Actions.CollectionDelete,
	"Add to Collection": Actions.CollectionAdd,
}

var GlobalActionMap = map[string]sum.Int[Action]{
	"Download Missing Art": Actions.GlobalDownloadArt,
}

var ActionKeys = []string{
	"Rename ROM",
	"Add to Collection",
	//"Clear Save States",
	"Archive ROM",
	"Delete ROM",
	//"Nuclear Option",
}

var GlobalActionKeys = []string{
	"Download Missing Art",
}

var BulkActionKeys = []string{
	"Add to Collection",
	"Download Art",
	"Delete Art",
	//"Clear Game Tracker",
	"Archive ROM",
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
