//go:build integration

package transactions_test

import (
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func TestCreateManyRollsBackEntireImport(t *testing.T) {
	databaseURL := os.Getenv("BACKEND_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BACKEND_TEST_DATABASE_URL is required")
	}
	assertIsolatedTestDatabase(t, databaseURL)
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	db = db.Begin()
	if db.Error != nil {
		t.Fatalf("begin fixture transaction: %v", db.Error)
	}
	defer db.Rollback()
	seedIntegrityFixture(t, db)
	validID := uuid.New()
	amount := decimal.NewFromInt(100)
	makeTransaction := func(id uuid.UUID, accountID uuid.UUID) *transactions.Transaction {
		return &transactions.Transaction{
			ID:              id,
			PortfolioID:     uuid.MustParse("10000000-0000-0000-0000-000000000011"),
			AccountID:       accountID,
			TransactionType: transactions.TransactionTypeDeposit,
			OccurredAt:      time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC),
			CreatedByUserID: uuid.MustParse("10000000-0000-0000-0000-000000000001"),
			Entries:         []transactions.TransactionEntry{{EntryKind: transactions.EntryKindCash, Amount: &amount, Currency: "USD"}},
		}
	}
	repo := transactions.NewRepository(db)
	err = repo.CreateMany([]*transactions.Transaction{
		makeTransaction(validID, uuid.MustParse("10000000-0000-0000-0000-000000000021")),
		makeTransaction(uuid.New(), uuid.New()),
	})
	if err == nil {
		t.Fatal("CreateMany succeeded, want foreign-key failure")
	}
	var count int64
	if err := db.Table("transactions").Where("id = ?", validID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("first imported transaction count = %d, want rollback", count)
	}
}

func TestLedgerRelationalIntegrity(t *testing.T) {
	databaseURL := os.Getenv("BACKEND_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BACKEND_TEST_DATABASE_URL is required")
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

	seedIntegrityFixture(t, db)

	expectConstraintViolation(t, db, `
		INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000101',
			 '10000000-0000-0000-0000-000000000011',
			 '10000000-0000-0000-0000-000000000022',
			 'deposit', '2026-01-10T10:00:00Z',
			 '10000000-0000-0000-0000-000000000001')
	`, "cross-portfolio account")

	expectConstraintViolation(t, db, `
		INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000102',
			 '10000000-0000-0000-0000-000000000011',
			 '10000000-0000-0000-0000-000000000021',
			 'deposit', '2026-01-10T10:00:00Z',
			 '10000000-0000-0000-0000-000000000002')
	`, "non-owner creator")

	expectConstraintViolation(t, db, `
		INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, reverses_transaction_id, created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000103',
			 '10000000-0000-0000-0000-000000000012',
			 '10000000-0000-0000-0000-000000000022',
			 'reversal', '2026-01-11T10:00:00Z',
			 '10000000-0000-0000-0000-000000000100',
			 '10000000-0000-0000-0000-000000000002')
	`, "cross-portfolio reversal")

	expectConstraintViolation(t, db, `
		INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, corrects_transaction_id, created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000104',
			 '10000000-0000-0000-0000-000000000012',
			 '10000000-0000-0000-0000-000000000022',
			 'deposit', '2026-01-11T10:00:00Z',
			 '10000000-0000-0000-0000-000000000100',
			 '10000000-0000-0000-0000-000000000002')
	`, "cross-portfolio correction")

	expectConstraintViolation(t, db, `
		INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000105',
			 '10000000-0000-0000-0000-000000000011',
			 '10000000-0000-0000-0000-000000000021',
			 'reversal', '2026-01-11T10:00:00Z',
			 '10000000-0000-0000-0000-000000000001')
	`, "unlinked reversal")

	expectConstraintViolation(t, db, `
		INSERT INTO portfolio_snapshots
			(portfolio_id, snapshot_date, snapshot_period, total_values,
			 asset_allocations, asset_class_allocations, cash_allocations,
			 missing_prices, is_fully_valued, valuation_scope, allocation_scope,
			 valuation_metadata, allocation_metadata, holdings_metadata,
			 created_by_user_id)
		VALUES
			('10000000-0000-0000-0000-000000000011', '2026-01-10', 'daily',
			 '[]', '[]', '[]', '[]', '[]', true, 'test', 'test', '{}', '{}', '{}',
			 '10000000-0000-0000-0000-000000000002')
	`, "snapshot created by non-owner")
}

func assertIsolatedTestDatabase(t *testing.T, databaseURL string) {
	t.Helper()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse BACKEND_TEST_DATABASE_URL: %v", err)
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	if !strings.HasPrefix(name, "wealth_lens_test_") {
		t.Fatalf("refusing to run against non-test database %q", name)
	}
}

func seedIntegrityFixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone)
		 VALUES ('10000000-0000-0000-0000-000000000001', 'integrity-a@example.com', 'not-used', 'Integrity A', 'USD', 'UTC')`,
		`INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone)
		 VALUES ('10000000-0000-0000-0000-000000000002', 'integrity-b@example.com', 'not-used', 'Integrity B', 'USD', 'UTC')`,
		`INSERT INTO portfolios (id, user_id, name, base_currency, deleted_at)
		 VALUES ('10000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000001', 'Integrity A', 'USD', now())`,
		`INSERT INTO portfolios (id, user_id, name, base_currency, deleted_at)
		 VALUES ('10000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000002', 'Integrity B', 'USD', now())`,
		`INSERT INTO accounts (id, portfolio_id, name, account_type, currency)
		 VALUES ('10000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000011', 'Account A', 'brokerage', 'USD')`,
		`INSERT INTO accounts (id, portfolio_id, name, account_type, currency)
		 VALUES ('10000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000012', 'Account B', 'brokerage', 'USD')`,
		`INSERT INTO transactions
			(id, portfolio_id, account_id, transaction_type, occurred_at, created_by_user_id)
		 VALUES
			('10000000-0000-0000-0000-000000000100',
			 '10000000-0000-0000-0000-000000000011',
			 '10000000-0000-0000-0000-000000000021',
			 'deposit', '2026-01-10T09:00:00Z',
			 '10000000-0000-0000-0000-000000000001')`,
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed integrity fixture: %v", err)
	}
}

func expectConstraintViolation(t *testing.T, db *gorm.DB, statement string, name string) {
	t.Helper()
	if err := db.Exec(statement).Error; err == nil {
		t.Fatalf("%s insert succeeded, want constraint violation", name)
	}
}
