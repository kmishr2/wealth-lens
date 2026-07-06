//go:build integration

package fixeddeposits_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/fixeddeposits"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/notifications"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
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

	transactionService := transactions.NewService(
		transactions.NewRepository(db), portfolios.NewRepository(db), accounts.NewRepository(db), assets.NewRepository(db),
	)
	one := decimal.NewFromInt(1)
	_, err = transactionService.Create(userID, portfolioID, transactions.TransactionCreateRequest{
		AccountID: accountID, TransactionType: transactions.TransactionTypeSell,
		OccurredAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		Entries:    []transactions.TransactionEntryRequest{{EntryKind: transactions.EntryKindAsset, AssetID: &response.AssetID, Quantity: &one, Currency: "INR"}},
	})
	if err == nil || err.Error() != "Fixed deposit assets must be managed through fixed deposit endpoints at index 0" {
		t.Fatalf("generic fixed-deposit transaction error = %v", err)
	}
	if _, err := transactionService.Reverse(userID, portfolioID, response.OpeningTransactionID, transactions.TransactionReversalRequest{
		Reason: "should be rejected", OccurredAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}); err == nil || err.Error() != "Fixed deposit ledger events cannot be reversed or corrected directly" {
		t.Fatalf("opening transaction reversal error = %v", err)
	}
	notificationService := notifications.NewService(notifications.NewRepository(db))
	reminders, err := notificationService.List(userID, time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC))
	if err != nil || len(reminders) != 1 || reminders[0].EntityID != response.ID || reminders[0].Status != "upcoming" {
		t.Fatalf("maturity reminders = %+v, err = %v", reminders, err)
	}

	proceeds := decimal.RequireFromString("263100")
	closed, err := service.Close(userID, portfolioID, accountID, response.ID, fixeddeposits.CloseRequest{
		ClosureType: "premature", ClosedAt: "2026-01-02", Proceeds: &proceeds, Note: "Integration closure",
	})
	if err != nil {
		t.Fatalf("close fixed deposit: %v", err)
	}
	if closed.Status != "closed" || closed.ClosingTransactionID == nil {
		t.Fatalf("closed response = %+v", closed)
	}
	var closureCount, closingEntryCount int64
	if err := db.Table("fixed_deposit_closures").Where("fixed_deposit_id = ?", response.ID).Count(&closureCount).Error; err != nil || closureCount != 1 {
		t.Fatalf("closure count = %d, err = %v", closureCount, err)
	}
	if err := db.Table("transaction_entries").Where("transaction_id = ?", *closed.ClosingTransactionID).Count(&closingEntryCount).Error; err != nil || closingEntryCount != 2 {
		t.Fatalf("closing entry count = %d, err = %v", closingEntryCount, err)
	}
	if _, err := transactionService.Reverse(userID, portfolioID, *closed.ClosingTransactionID, transactions.TransactionReversalRequest{
		Reason: "should be rejected", OccurredAt: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
	}); err == nil || err.Error() != "Fixed deposit ledger events cannot be reversed or corrected directly" {
		t.Fatalf("closing transaction reversal error = %v", err)
	}
	reminders, err = notificationService.List(userID, time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC))
	if err != nil || len(reminders) != 0 {
		t.Fatalf("closed-deposit reminders = %+v, err = %v", reminders, err)
	}
}
