//go:build integration

package snapshots_test

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	testUserID      = "00000000-0000-0000-0000-000000000001"
	testPortfolioID = "00000000-0000-0000-0000-000000000002"
	testAccountID   = "00000000-0000-0000-0000-000000000003"
	testAssetID     = "00000000-0000-0000-0000-000000000004"
	testSnapshotDay = "2026-01-15"
)

func TestDailySnapshotJobIntegration(t *testing.T) {
	databaseURL := os.Getenv("SNAPSHOT_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SNAPSHOT_TEST_DATABASE_URL is required")
	}
	assertIsolatedTestDatabase(t, databaseURL)

	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get SQL database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	seedSnapshotFixture(t, db)

	portfolioRepo := portfolios.NewRepository(db)
	holdingsRepo := holdings.NewRepository(db)
	priceRepo := prices.NewRepository(db)
	snapshotRepo := snapshots.NewRepository(db)
	snapshotService := snapshots.NewService(snapshotRepo, holdingsRepo, priceRepo, portfolioRepo)
	job := snapshots.NewDailyJob(portfolioRepo, snapshotService, 10)

	first := runSuccessfulJob(t, job)
	second := runSuccessfulJob(t, job)
	if first.Processed != 1 || second.Processed != 1 {
		t.Fatalf("processed counts = %d and %d, want 1 and 1", first.Processed, second.Processed)
	}

	portfolioID := uuid.MustParse(testPortfolioID)
	snapshotDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	stored, err := snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, snapshots.SnapshotPeriodDaily)
	if err != nil {
		t.Fatalf("get stored snapshot: %v", err)
	}

	var count int64
	if err := db.Model(&snapshots.PortfolioSnapshot{}).
		Where("portfolio_id = ? AND snapshot_date = ? AND snapshot_period = ?", portfolioID, snapshotDate, snapshots.SnapshotPeriodDaily).
		Count(&count).Error; err != nil {
		t.Fatalf("count snapshots: %v", err)
	}
	if count != 1 {
		t.Fatalf("snapshot count = %d, want 1 after two job runs", count)
	}

	response, err := snapshots.ToResponse(*stored)
	if err != nil {
		t.Fatalf("decode stored snapshot: %v", err)
	}
	assertSnapshotValues(t, response)
	assertSnapshotImmutable(t, db, stored.ID)
}

func assertIsolatedTestDatabase(t *testing.T, databaseURL string) {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse SNAPSHOT_TEST_DATABASE_URL: %v", err)
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if !strings.HasPrefix(databaseName, "wealth_lens_test_") {
		t.Fatalf("refusing to run against non-test database %q", databaseName)
	}
}

func seedSnapshotFixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone)
		 VALUES ('00000000-0000-0000-0000-000000000001', 'snapshot-test@example.com', 'not-used', 'Snapshot Test', 'USD', 'UTC')`,
		`INSERT INTO portfolios (id, user_id, name, description, base_currency)
		 VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Integration Portfolio', '', 'USD')`,
		`INSERT INTO accounts (id, portfolio_id, name, account_type, institution_name, currency)
		 VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002', 'Integration Brokerage', 'brokerage', '', 'USD')`,
		`INSERT INTO assets (id, symbol, name, asset_class, currency, exchange)
		 VALUES ('00000000-0000-0000-0000-000000000004', 'TEST', 'Integration Test Asset', 'equity', 'USD', 'TEST')`,
		`INSERT INTO transactions (id, portfolio_id, account_id, transaction_type, occurred_at, description, created_by_user_id)
		 VALUES ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003', 'deposit', '2026-01-14T10:00:00Z', 'Initial contribution', '00000000-0000-0000-0000-000000000001')`,
		`INSERT INTO transaction_entries (id, transaction_id, entry_kind, amount, currency)
		 VALUES ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', 'cash', 125, 'USD')`,
		`INSERT INTO transactions (id, portfolio_id, account_id, transaction_type, occurred_at, description, created_by_user_id)
		 VALUES ('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003', 'buy', '2026-01-15T10:00:00Z', 'Buy test asset', '00000000-0000-0000-0000-000000000001')`,
		`INSERT INTO transaction_entries (id, transaction_id, entry_kind, asset_id, quantity, currency)
		 VALUES ('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000102', 'asset', '00000000-0000-0000-0000-000000000004', 2, 'USD')`,
		`INSERT INTO transaction_entries (id, transaction_id, entry_kind, amount, currency)
		 VALUES ('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000102', 'cash', -100, 'USD')`,
		`INSERT INTO asset_prices (id, asset_id, price, currency, priced_at, source, note, created_by_user_id)
		 VALUES ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000004', 50, 'USD', '2026-01-15T12:00:00Z', 'integration-test', '', '00000000-0000-0000-0000-000000000001')`,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed snapshot fixture: %v", err)
	}
}

func runSuccessfulJob(t *testing.T, job *snapshots.DailyJob) snapshots.DailyJobResult {
	t.Helper()
	result, err := job.Run(testSnapshotDay)
	if err != nil {
		t.Fatalf("run daily snapshot job: %v", err)
	}
	if result.Succeeded != 1 || result.Failed != 0 || len(result.Failures) != 0 {
		t.Fatalf("job result = %+v, want one success and no failures", result)
	}
	return result
}

func assertSnapshotValues(t *testing.T, response snapshots.PortfolioSnapshotResponse) {
	t.Helper()
	if response.PortfolioID != uuid.MustParse(testPortfolioID) || response.SnapshotDate != testSnapshotDay {
		t.Fatalf("snapshot identity = %s/%s, want %s/%s", response.PortfolioID, response.SnapshotDate, testPortfolioID, testSnapshotDay)
	}
	if !response.IsFullyValued || len(response.MissingPrices) != 0 {
		t.Fatalf("valuation completeness = %t, missing prices = %+v", response.IsFullyValued, response.MissingPrices)
	}
	if len(response.TotalValues) != 1 || response.TotalValues[0].Currency != "USD" || !response.TotalValues[0].Amount.Equal(decimal.NewFromInt(125)) {
		t.Fatalf("total values = %+v, want USD 125", response.TotalValues)
	}
	if len(response.AssetAllocations) != 1 || !response.AssetAllocations[0].Percentage.Equal(decimal.NewFromInt(80)) {
		t.Fatalf("asset allocations = %+v, want 80 percent", response.AssetAllocations)
	}
	if len(response.CashAllocations) != 1 || !response.CashAllocations[0].Percentage.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("cash allocations = %+v, want 20 percent", response.CashAllocations)
	}
	if response.HoldingsMetadata.Formula == "" || response.ValuationMetadata.Formula == "" || response.AllocationMetadata.Formula == "" {
		t.Fatalf("metric metadata is incomplete: %+v", response)
	}
}

func assertSnapshotImmutable(t *testing.T, db *gorm.DB, snapshotID uuid.UUID) {
	t.Helper()
	if err := db.Exec("UPDATE portfolio_snapshots SET valuation_scope = 'mutated' WHERE id = ?", snapshotID).Error; err == nil {
		t.Fatal("snapshot update succeeded, want immutability trigger error")
	}
	if err := db.Exec("DELETE FROM portfolio_snapshots WHERE id = ?", snapshotID).Error; err == nil {
		t.Fatal("snapshot delete succeeded, want immutability trigger error")
	}

	var count int64
	if err := db.Model(&snapshots.PortfolioSnapshot{}).Where("id = ?", snapshotID).Count(&count).Error; err != nil {
		t.Fatalf("count immutable snapshot: %v", err)
	}
	if count != 1 {
		t.Fatalf("snapshot count after mutation attempts = %d, want 1", count)
	}
}
