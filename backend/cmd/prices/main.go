package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/marketdata"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
)

func main() { os.Exit(run()) }

func run() int {
	fromFlag := flag.String("from", "", "first Indian market date in YYYY-MM-DD (default: 5 days before to)")
	toFlag := flag.String("to", "", "last Indian market date in YYYY-MM-DD (default: yesterday in Asia/Kolkata)")
	flag.Parse()

	india, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load Asia/Kolkata timezone: %v\n", err)
		return 1
	}
	now := time.Now().In(india)
	toDefault := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
	to, err := parseDate(*toFlag, toDefault)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid to date: %v\n", err)
		return 2
	}
	from, err := parseDate(*fromFlag, to.AddDate(0, 0, -5))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid from date: %v\n", err)
		return 2
	}

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect database: %v\n", err)
		return 1
	}
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get database connection: %v\n", err)
		return 1
	}
	defer sqlDB.Close()

	client := &http.Client{Timeout: 30 * time.Second}
	job := marketdata.NewJob(
		assets.NewRepository(db),
		prices.NewRepository(db),
		marketdata.NewAMFIProvider(client, cfg.AMFINAVURL),
		marketdata.NewUpstoxProvider(client, cfg.UpstoxAPIURL, cfg.UpstoxAccessToken),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	result, err := job.Run(ctx, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run market price ingestion: %v\n", err)
		return 1
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "encode result: %v\n", err)
		return 1
	}
	if result.Failed > 0 {
		return 1
	}
	return 0
}

func parseDate(raw string, fallback time.Time) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", raw)
}
