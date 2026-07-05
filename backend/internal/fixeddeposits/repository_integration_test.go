//go:build integration

package fixeddeposits_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/fixeddeposits"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/shopspring/decimal"
)

func TestCreatePersistsCompleteLedgerBackedBundle(t *testing.T) {
	databaseURL := os.Getenv("BACKEND_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BACKEND_TEST_DATABASE_URL is required")
	}
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	db = db.Begin()
	if db.Error != nil {
		t.Fatalf("begin fixture transaction: %v", db.Error)
	}
	defer db.Rollback()
	userID, portfolioID, accountID := uuid.New(), uuid.New(), uuid.New()
	if err := db.Exec(`
		INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone)
		VALUES (?, ?, 'hash', 'FD Test', 'INR', 'Asia/Kolkata')
	`, userID, "fd-"+userID.String()+"@example.com").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO portfolios (id, user_id, name, base_currency)
		VALUES (?, ?, 'FD Portfolio', 'INR')
	`, portfolioID, userID).Error; err != nil {
		t.Fatalf("insert portfolio: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO accounts (id, portfolio_id, name, account_type, currency)
		VALUES (?, ?, 'FD Bank', 'bank', 'INR')
	`, accountID, portfolioID).Error; err != nil {
		t.Fatalf("insert account: %v", err)
	}

	repo := fixeddeposits.NewRepository(db)
	service := fixeddeposits.NewService(
		repo,
		portfolios.NewRepository(db),
		accounts.NewRepository(db),
		prices.NewRepository(db),
	)
	principal := decimal.RequireFromString("250000")
	rate := decimal.RequireFromString("7.10")
	currentValue := decimal.RequireFromString("261500")
	response, err := service.Create(userID, portfolioID, accountID, fixeddeposits.CreateRequest{
		Name: "Integration FD", Principal: &principal, Currency: "INR",
		AnnualInterestRate: &rate, StartDate: "2025-01-01", MaturityDate: "2027-01-01",
		CurrentValue: &currentValue, CurrentValueDate: "2026-01-01",
	})
	if err != nil {
		t.Fatalf("create fixed deposit: %v", err)
	}

	for table, id := range map[string]uuid.UUID{
		"fixed_deposits": response.ID,
		"assets":         response.AssetID,
		"transactions":   response.OpeningTransactionID,
	} {
		var count int64
		if err := db.Table(table).Where("id = ?", id).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("%s count = %d, err = %v", table, count, err)
		}
	}
	var entryCount, priceCount int64
	if err := db.Table("transaction_entries").Where("transaction_id = ?", response.OpeningTransactionID).Count(&entryCount).Error; err != nil || entryCount != 2 {
		t.Fatalf("entry count = %d, err = %v", entryCount, err)
	}
	if err := db.Table("asset_prices").Where("asset_id = ?", response.AssetID).Count(&priceCount).Error; err != nil || priceCount != 1 {
		t.Fatalf("price count = %d, err = %v", priceCount, err)
	}
}
