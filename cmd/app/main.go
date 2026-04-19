package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lminimum/LiteDock/config"
	"github.com/lminimum/LiteDock/internal/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
