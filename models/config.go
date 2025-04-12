package models

import (
	"go.uber.org/zap/zapcore"
)

type Config struct {
	LogLevel string `yaml:"log_level"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("log_level", c.LogLevel)

	return nil
}
