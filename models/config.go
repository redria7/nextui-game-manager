package models

import (
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap/zapcore"
	"qlova.tech/sum"
)

type Config struct {
	LogLevel        string                          `yaml:"log_level"`
	ArtDownloadType sum.Int[shared.ArtDownloadType] `yaml:"art_download_type"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("log_level", c.LogLevel)

	return nil
}
