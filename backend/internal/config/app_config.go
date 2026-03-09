package config

import (
	"context"
	"os"
	"stud_hub/util/logger/logger"

	"gopkg.in/yaml.v3"
)

type ApplicationConfig struct {
	LoggerConfig LoggerConfig `yaml:"logger"`
}

func LoadApplicationConfig(path string) (*ApplicationConfig, error) {
	cfg := new(ApplicationConfig)
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Errorf(context.Background(), "Error while closing config file: %s", err)
		}
	}(file)

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true) // строгий режим

	err = dec.Decode(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
