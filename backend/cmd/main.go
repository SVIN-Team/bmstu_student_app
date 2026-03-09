package main

import (
	"log"
	"stud_hub/internal"
	"stud_hub/internal/config"

	"github.com/spf13/pflag"
)

func main() {
	configPath := pflag.StringP("config", "c", "config.yaml", "Path to application config file")
	pflag.Parse()

	if *configPath == "" {
		log.Fatalf("No config path passed")
	}

	cfg, err := config.LoadApplicationConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	internal.Run(cfg)
}
