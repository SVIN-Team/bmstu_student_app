package config

import "time"

type AuthConfig struct {
	AccessSecretKey  string
	RefreshSecretKey string
	AccessLifeTime   time.Duration `yaml:"access_life_time"`
	RefreshLifeTime  time.Duration `yaml:"refresh_life_time"`
}
