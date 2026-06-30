package main

import (
	"log"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/server"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	api, err := server.New(cfg, db)
	if err != nil {
		log.Fatalf("configure server: %v", err)
	}
	if err := api.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run api: %v", err)
	}
}
