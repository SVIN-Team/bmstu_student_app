package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ApplicationConfig struct {
	LoggerConfig `yaml:"logger"`
}

func LoadApplicationConfig(path string) (*ApplicationConfig, error) {
	cfg := new(ApplicationConfig)
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true) // строгий режим

	err = dec.Decode(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
