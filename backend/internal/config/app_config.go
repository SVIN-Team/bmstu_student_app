package config

import (
	"context"
	"os"
	"stud_hub/util/logger"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAccessSecretKey  = "default_access_secret_key"
	DefaultRefreshSecretKey = "default_refresh_secret_key"
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

	cfg.AuthConfig.AccessSecretKey = os.Getenv("ACCESS_SECRET_KEY")
	cfg.AuthConfig.RefreshSecretKey = os.Getenv("REFRESH_SECRET_KEY")

	if cfg.AuthConfig.AccessSecretKey == "" {
		cfg.AuthConfig.AccessSecretKey = DefaultAccessSecretKey
	}

	if cfg.AuthConfig.RefreshSecretKey == "" {
		cfg.AuthConfig.RefreshSecretKey = DefaultRefreshSecretKey
	}

	return cfg, nil
}
