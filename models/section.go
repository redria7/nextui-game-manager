package models

import "go.uber.org/zap/zapcore"

type Section struct {
	Name           string `yaml:"section_name"`
	LocalDirectory string `yaml:"local_directory"`
}

type Sections []Section

func (s Sections) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, section := range s {
		_ = enc.AppendObject(section)
	}

	return nil
}

func (s Section) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", s.Name)
	enc.AddString("local_directory", s.LocalDirectory)

	return nil
}
