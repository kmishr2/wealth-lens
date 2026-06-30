//go:build integration

package prices_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/shopspring/decimal"
)

func TestAutomatedPriceInsertIsIdempotent(t *testing.T) {
	databaseURL := os.Getenv("BACKEND_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BACKEND_TEST_DATABASE_URL is required")
	}
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}

	assetID := uuid.New()
	if err := db.Exec(`
		INSERT INTO assets (id, symbol, name, asset_class, currency, exchange)
		VALUES (?, ?, 'Ingestion Test', 'fund', 'INR', 'AMFI')
	`, assetID, "TEST"+strings.ToUpper(assetID.String()[:8])).Error; err != nil {
		t.Fatalf("insert asset: %v", err)
	}

	marketDate := time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC)
	newPrice := func() *prices.AssetPrice {
		return &prices.AssetPrice{
			AssetID: assetID, Price: decimal.RequireFromString("123.45"), Currency: "INR",
			PricedAt: marketDate.Add(18*time.Hour + 29*time.Minute), MarketDate: &marketDate,
			Source: "amfi", Note: "integration test",
		}
	}
	repo := prices.NewRepository(db)
	inserted, err := repo.CreateAutomated(newPrice())
	if err != nil || !inserted {
		t.Fatalf("first insert: inserted=%t err=%v", inserted, err)
	}
	inserted, err = repo.CreateAutomated(newPrice())
	if err != nil || inserted {
		t.Fatalf("second insert: inserted=%t err=%v", inserted, err)
	}

	var count int64
	if err := db.Table("asset_prices").Where("asset_id = ? AND market_date = ? AND source = 'amfi'", assetID, marketDate).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("stored prices = %d, want 1", count)
	}
}
