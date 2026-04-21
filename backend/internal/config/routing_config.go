package config

type RoutingConfig struct {
	Port    uint   `yaml:"port"`
	GinMode string `yaml:"gin_mode"`
}
