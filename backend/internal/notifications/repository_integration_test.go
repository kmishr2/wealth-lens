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

	service := notifications.NewService(notifications.NewRepository(db))
	asOfDate := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	result, err := service.List(userID, asOfDate)
	if err != nil || len(result) != 1 || result[0].EntityID != goalID || result[0].DataAsOfDate != nil {
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
	if err != nil || len(result) != 0 {
		t.Fatalf("owner notices after reached snapshot = %+v, err = %v", result, err)
	}
}
