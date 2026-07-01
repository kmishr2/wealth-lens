package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/goals"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
)

func main() {
	os.Exit(run())
}

func run() int {
	snapshotMonthEnd := flag.String("date", "", "UTC month-end date in YYYY-MM-DD format (required)")
	flag.Parse()
	if *snapshotMonthEnd == "" {
		fmt.Fprintln(os.Stderr, "snapshot month-end date is required; use -date YYYY-MM-DD")
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
	goalRepo := goals.NewRepository(db)
	snapshotRepo := snapshots.NewRepository(db)
	goalService := goals.NewService(goalRepo, portfolioRepo, snapshotRepo)
	job := goals.NewMonthlySnapshotJob(portfolioRepo, goalService, goals.DefaultPortfolioBatchSize)

	result, err := job.Run(*snapshotMonthEnd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run monthly goal snapshots: %v\n", err)
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
