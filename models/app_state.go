package models

import (
	"go.uber.org/zap/zapcore"
)

type AppState struct {
	Config *Config

	MenuPositionList []MenuPositionPointer

	GamePlayMap 	map[string][]PlayTrackingAggregate
	ConsolePlayMap 	map[string]int
	TotalPlay 		int
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)

	return nil
}

type MenuPositionPointer struct {
	SelectedIndex		int
	SelectedPosition    int
}
