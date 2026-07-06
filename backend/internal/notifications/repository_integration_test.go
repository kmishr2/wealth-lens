//go:build integration

package notifications_test

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/notifications"
)

func TestGoalNoticesRespectOwnershipAndReachedSnapshot(t *testing.T) {
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

	userID, otherUserID := uuid.New(), uuid.New()
	portfolioID, otherPortfolioID := uuid.New(), uuid.New()
	goalID, otherGoalID := uuid.New(), uuid.New()
	for _, fixture := range []struct {
		userID, portfolioID, goalID    uuid.UUID
		email, portfolioName, goalName string
	}{
		{userID, portfolioID, goalID, "notices-owner@example.com", "Owner portfolio", "Education"},
		{otherUserID, otherPortfolioID, otherGoalID, "notices-other@example.com", "Other portfolio", "Other goal"},
	} {
		if err := db.Exec(`INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone) VALUES (?, ?, 'hash', 'Notice Test', 'INR', 'Asia/Kolkata')`, fixture.userID, fixture.email).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO portfolios (id, user_id, name, base_currency) VALUES (?, ?, ?, 'INR')`, fixture.portfolioID, fixture.userID, fixture.portfolioName).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`INSERT INTO goals (id, portfolio_id, name, target_amount, currency, target_date, status, created_by_user_id) VALUES (?, ?, ?, 100000, 'INR', '2026-07-20', 'active', ?)`, fixture.goalID, fixture.portfolioID, fixture.goalName, fixture.userID).Error; err != nil {
			t.Fatal(err)
		}
	}
	accountID, assetID, transactionID := uuid.New(), uuid.New(), uuid.New()
	if err := db.Exec(`INSERT INTO accounts (id, portfolio_id, name, account_type, currency) VALUES (?, ?, 'Brokerage', 'brokerage', 'INR')`, accountID, portfolioID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO assets (id, symbol, name, asset_class, currency, exchange, is_active) VALUES (?, 'TEST', 'Test Equity', 'equity', 'INR', 'NSE', true)`, assetID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO transactions (id, portfolio_id, account_id, transaction_type, occurred_at, created_by_user_id) VALUES (?, ?, ?, 'buy', '2026-06-01', ?)`, transactionID, portfolioID, accountID, userID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO transaction_entries (transaction_id, entry_kind, asset_id, quantity, currency) VALUES (?, 'asset', ?, 1, 'INR')`, transactionID, assetID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO asset_prices (asset_id, price, currency, priced_at, source, created_by_user_id) VALUES (?, 100, 'INR', '2026-06-29', 'test', ?)`, assetID, userID).Error; err != nil {
		t.Fatal(err)
	}

	service := notifications.NewService(notifications.NewRepository(db))
	asOfDate := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	result, err := service.List(userID, asOfDate)
	if err != nil || len(result) != 2 || findNotice(result, "goal_target_date") == nil || findNotice(result, "asset_price_stale") == nil {
		t.Fatalf("owner notices before snapshot = %+v, err = %v", result, err)
	}

	if err := db.Exec(`
		INSERT INTO monthly_goal_snapshots
			(portfolio_id, goal_id, snapshot_month_end, current_value, target_value, currency,
			 progress_percentage, remaining_amount, months_remaining, required_monthly_contribution,
			 is_target_reached, goal_progress_metadata, created_by_user_id)
		VALUES (?, ?, '2026-06-30', 100000, 100000, 'INR', 100, 0, 1, 0, true, '{}', ?)
	`, portfolioID, goalID, userID).Error; err != nil {
		t.Fatal(err)
	}
	result, err = service.List(userID, asOfDate)
	if err != nil || len(result) != 1 || result[0].Kind != "asset_price_stale" || result[0].EntityID != assetID {
		t.Fatalf("owner notices after reached snapshot = %+v, err = %v", result, err)
	}
}

func findNotice(notices []notifications.Response, kind string) *notifications.Response {
	for index := range notices {
		if notices[index].Kind == kind {
			return &notices[index]
		}
	}
	return nil
}
