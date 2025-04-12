package models

import (
	"go.uber.org/zap/zapcore"
)

type Config struct {
	LogLevel        string `yaml:"log_level"`
	IBackedUpMyShit bool   `yaml:"i_backed_up_my_shit"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("log_level", c.LogLevel)

	return nil
}
