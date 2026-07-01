package goals

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakeGoalStore struct {
	goal            *Goal
	monthlySnapshot *MonthlyGoalSnapshot
	createCalls     int
}

func (f *fakeGoalStore) Create(goal *Goal) error { f.goal = goal; return nil }
func (f *fakeGoalStore) ListByPortfolio(uuid.UUID, common.Pagination) ([]Goal, error) {
	return nil, nil
}
func (f *fakeGoalStore) ListActiveByPortfolio(uuid.UUID) ([]Goal, error) {
	if f.goal == nil {
		return nil, nil
	}
	return []Goal{*f.goal}, nil
}
func (f *fakeGoalStore) GetInPortfolio(uuid.UUID, uuid.UUID) (*Goal, error) {
	if f.goal == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.goal, nil
}
func (f *fakeGoalStore) Update(goal *Goal) error { f.goal = goal; return nil }
func (f *fakeGoalStore) Delete(*Goal) error      { return nil }
func (f *fakeGoalStore) CreateMonthlySnapshot(snapshot *MonthlyGoalSnapshot) error {
	f.createCalls++
	f.monthlySnapshot = snapshot
	return nil
}
func (f *fakeGoalStore) GetMonthlySnapshotByGoalMonth(uuid.UUID, time.Time) (*MonthlyGoalSnapshot, error) {
	if f.monthlySnapshot == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.monthlySnapshot, nil
}
func (f *fakeGoalStore) ListMonthlySnapshots(uuid.UUID, common.Pagination) ([]MonthlyGoalSnapshot, error) {
	return nil, nil
}

type fakeGoalPortfolioReader struct{ portfolio *portfolios.Portfolio }

func (f *fakeGoalPortfolioReader) GetOwned(uuid.UUID, uuid.UUID) (*portfolios.Portfolio, error) {
	if f.portfolio == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.portfolio, nil
}

type fakeGoalSnapshotReader struct{ snapshot *snapshots.PortfolioSnapshot }

func (f *fakeGoalSnapshotReader) GetByPortfolioDatePeriod(uuid.UUID, time.Time, string) (*snapshots.PortfolioSnapshot, error) {
	if f.snapshot == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.snapshot, nil
}

func TestCreateMonthlySnapshotCalculatesAndIsIdempotent(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	goalID := uuid.New()
	repo := &fakeGoalStore{goal: &Goal{
		ID: goalID, PortfolioID: portfolioID, TargetAmount: decimal.RequireFromString("1000000"),
		Currency: "INR", TargetDate: mustGoalDate("2026-12-31"), Status: StatusActive,
	}}
	service := NewService(
		repo,
		&fakeGoalPortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeGoalSnapshotReader{snapshot: validDailySnapshot(t, portfolioID, "2026-01-31", "INR", "250000")},
	)

	first, err := service.CreateMonthlySnapshot(userID, portfolioID, goalID, "2026-01-31")
	if err != nil {
		t.Fatalf("CreateMonthlySnapshot returned error: %v", err)
	}
	if !first.ProgressPercentage.Equal(decimal.RequireFromString("25")) || first.MonthsRemaining != 11 {
		t.Fatalf("snapshot = %+v", first)
	}
	if first.GoalProgressMetadata.Name != "Goal Progress" {
		t.Fatalf("metadata = %+v", first.GoalProgressMetadata)
	}

	second, err := service.CreateMonthlySnapshot(userID, portfolioID, goalID, "2026-01-31")
	if err != nil {
		t.Fatalf("idempotent CreateMonthlySnapshot returned error: %v", err)
	}
	if second.GoalID != goalID || repo.createCalls != 1 {
		t.Fatalf("second = %+v, create calls = %d", second, repo.createCalls)
	}
}

func TestCreateMonthlySnapshotRequiresExactDailySnapshot(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	goalID := uuid.New()
	service := NewService(
		&fakeGoalStore{goal: &Goal{ID: goalID, PortfolioID: portfolioID, TargetAmount: decimal.NewFromInt(100), Currency: "INR", TargetDate: mustGoalDate("2026-12-31")}},
		&fakeGoalPortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeGoalSnapshotReader{},
	)

	_, err := service.CreateMonthlySnapshot(userID, portfolioID, goalID, "2026-01-31")
	if err == nil {
		t.Fatal("CreateMonthlySnapshot error = nil, want missing daily snapshot error")
	}
}

func TestCreateMonthlySnapshotRequiresGoalCurrency(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	goalID := uuid.New()
	service := NewService(
		&fakeGoalStore{goal: &Goal{ID: goalID, PortfolioID: portfolioID, TargetAmount: decimal.NewFromInt(100), Currency: "INR", TargetDate: mustGoalDate("2026-12-31")}},
		&fakeGoalPortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeGoalSnapshotReader{snapshot: validDailySnapshot(t, portfolioID, "2026-01-31", "USD", "100")},
	)

	_, err := service.CreateMonthlySnapshot(userID, portfolioID, goalID, "2026-01-31")
	if err == nil {
		t.Fatal("CreateMonthlySnapshot error = nil, want missing currency error")
	}
}

func validDailySnapshot(t *testing.T, portfolioID uuid.UUID, date, currency, amount string) *snapshots.PortfolioSnapshot {
	t.Helper()
	totals, err := snapshots.NewJSONB([]finance.CurrencyValue{{Currency: currency, Amount: decimal.RequireFromString(amount)}})
	if err != nil {
		t.Fatal(err)
	}
	empty, _ := snapshots.NewJSONB([]any{})
	metadata, _ := snapshots.NewJSONB(finance.MetricDefinition{})
	return &snapshots.PortfolioSnapshot{
		PortfolioID: portfolioID, SnapshotDate: mustGoalDate(date), SnapshotPeriod: snapshots.SnapshotPeriodDaily,
		TotalValues: totals, AssetAllocations: empty, AssetClassAllocations: empty, CashAllocations: empty,
		MissingPrices: empty, ValuationMetadata: metadata, AllocationMetadata: metadata, HoldingsMetadata: metadata,
	}
}

func mustGoalDate(raw string) time.Time {
	date, err := time.Parse(goalDateLayout, raw)
	if err != nil {
		panic(err)
	}
	return date
}
