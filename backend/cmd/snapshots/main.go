package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
)

func main() {
	os.Exit(run())
}

func run() int {
	snapshotDate := flag.String("date", "", "UTC snapshot date in YYYY-MM-DD format (required)")
	flag.Parse()
	if *snapshotDate == "" {
		fmt.Fprintln(os.Stderr, "snapshot date is required; use -date YYYY-MM-DD")
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

	portfolioRepo := portfolios.NewRepository(db)
	holdingsRepo := holdings.NewRepository(db)
	priceRepo := prices.NewRepository(db)
	snapshotRepo := snapshots.NewRepository(db)
	snapshotService := snapshots.NewService(snapshotRepo, holdingsRepo, priceRepo, portfolioRepo)
	job := snapshots.NewDailyJob(portfolioRepo, snapshotService, snapshots.DefaultPortfolioBatchSize)

	result, err := job.Run(*snapshotDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run daily snapshots: %v\n", err)
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
