package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type AppState struct {
	Config        *Config
	CurrentScreen sum.Int[Screen]

	RomDirectories []shared.RomDirectory

	CurrentSection   Section
	CurrentItemsList shared.Items
	SearchFilter     string
	SelectedFile     string
	SelectedAction   string

	LastSavedArtPath string
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)
	enc.AddString("current_screen", a.CurrentScreen.String())
	_ = enc.AddObject("current_section", a.CurrentSection)
	_ = enc.AddArray("current_items_list", a.CurrentItemsList)
	enc.AddString("search_filter", a.SearchFilter)
	enc.AddString("selected_file", a.SelectedFile)
	enc.AddString("selected_action", a.SelectedAction)
	enc.AddString("last_saved_art_path", a.LastSavedArtPath)

	return nil
}
