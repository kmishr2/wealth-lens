package main

import (
	"log"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/server"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	api := server.New(cfg, db)
	if err := api.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run api: %v", err)
	}
}
