package main

import (
	"log"
	"stud_hub/internal"
	"stud_hub/internal/config"

	"github.com/spf13/pflag"
)

// @title Stud Hub API
// @version 1.0
// @description HTTP API for Stud Hub backend.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
