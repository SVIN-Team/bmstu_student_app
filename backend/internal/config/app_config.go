package config

import (
	"fmt"
	"log"
	"os"
	"stud_hub/util/logger/logger"

	"gopkg.in/yaml.v3"
)

type ApplicationConfig struct {
	LoggerConfig LoggerConfig `yaml:"logger"`
	AuthConfig   AuthConfig   `yaml:"auth"`
}

func LoadApplicationConfig(path string) (*ApplicationConfig, error) {
	cfg := new(ApplicationConfig)
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func(file *os.File) {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Error while closing config file: %s", cerr)
		}
	}(file)

	dec := yaml.NewDecoder(file)
	dec.KnownFields(true) // строгий режим

	err = dec.Decode(cfg)
	if err != nil {
		return cfg, err
	}

	accessSecret := os.Getenv("ACCESS_SECRET_KEY")
	if accessSecret == "" {
		return nil, fmt.Errorf("ACCESS_SECRET_KEY environment variable is not set")
	}

	refreshSecret := os.Getenv("REFRESH_SECRET_KEY")
	if refreshSecret == "" {
		return nil, fmt.Errorf("REFRESH_SECRET_KEY environment variable is not set")
	}

	cfg.AuthConfig.AccessSecretKey = accessSecret
	cfg.AuthConfig.RefreshSecretKey = refreshSecret

	return cfg, nil
}
