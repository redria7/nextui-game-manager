package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type AppState struct {
	Config        *Config
	CurrentScreen sum.Int[Screen]

	RomDirectories  []shared.RomDirectory
	RomDirectoryMap map[string]shared.RomDirectory

	CurrentSection                  Section
	CurrentItemsList                shared.Items
	CurrentItemListWithExtensionMap map[string]string
	SearchFilter                    string
	SelectedFile                    string
	SelectedFileHasArt              bool
	SelectedAction                  sum.Int[Action]

	LastSavedArtPath string
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)
	enc.AddString("current_screen", a.CurrentScreen.String())
	_ = enc.AddObject("current_section", a.CurrentSection)
	_ = enc.AddArray("current_items_list", a.CurrentItemsList)
	enc.AddString("search_filter", a.SearchFilter)
	enc.AddString("selected_file", a.SelectedFile)
	enc.AddBool("selected_file_has_art", a.SelectedFileHasArt)
	enc.AddString("last_saved_art_path", a.LastSavedArtPath)

	return nil
}
