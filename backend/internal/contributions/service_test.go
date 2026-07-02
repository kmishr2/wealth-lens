package contributions

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakePortfolioReader struct{ err error }

func (f *fakePortfolioReader) GetOwned(userID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &portfolios.Portfolio{ID: portfolioID, UserID: userID}, nil
}

type fakeSnapshotReader struct {
	byDate map[string]snapshots.PortfolioSnapshot
}

func (f *fakeSnapshotReader) GetByPortfolioDatePeriod(_ uuid.UUID, date time.Time, _ string) (*snapshots.PortfolioSnapshot, error) {
	record, ok := f.byDate[date.Format(dateLayout)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &record, nil
}

type fakeCashFlowReader struct {
	records           []transactions.ExternalCashFlowRecord
	startAfter, endAt time.Time
}

func (f *fakeCashFlowReader) ListExternalCashFlows(_ uuid.UUID, startAfter, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error) {
	f.startAfter, f.endAt = startAfter, endAt
	return f.records, nil
}

func TestGetCalculatesContributionAttributionForCurrency(t *testing.T) {
	portfolioID := uuid.New()
	cashFlows := &fakeCashFlowReader{records: []transactions.ExternalCashFlowRecord{
		{OccurredAt: mustDate("2026-02-10").Add(12 * time.Hour), Currency: "INR", Amount: decimal.NewFromInt(300)},
		{OccurredAt: mustDate("2026-02-20").Add(12 * time.Hour), Currency: "INR", Amount: decimal.NewFromInt(-50)},
		{OccurredAt: mustDate("2026-02-20").Add(12 * time.Hour), Currency: "USD", Amount: decimal.NewFromInt(999)},
	}}
	service := NewService(&fakePortfolioReader{}, &fakeSnapshotReader{byDate: map[string]snapshots.PortfolioSnapshot{
		"2026-01-31": validSnapshot(t, portfolioID, "2026-01-31", "1000"),
		"2026-03-31": validSnapshot(t, portfolioID, "2026-03-31", "1400"),
	}}, cashFlows)

	response, err := service.Get(uuid.New(), portfolioID, "2026-01-31", "2026-03-31", "inr")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !response.NetContributions.Equal(decimal.NewFromInt(250)) || !response.InvestmentGrowth.Equal(decimal.NewFromInt(150)) {
		t.Fatalf("response = %+v", response)
	}
	if response.Currency != "INR" || response.EventCount != 2 || len(response.MonthlyBuckets) != 1 || response.Scope == "" {
		t.Fatalf("response = %+v", response)
	}
	if cashFlows.startAfter.Day() != 31 || cashFlows.startAfter.Hour() != 23 || cashFlows.endAt.Day() != 31 {
		t.Fatalf("cash flow range = %s to %s", cashFlows.startAfter, cashFlows.endAt)
	}
}

func TestGetRejectsMissingSnapshotCurrency(t *testing.T) {
	portfolioID := uuid.New()
	service := NewService(&fakePortfolioReader{}, &fakeSnapshotReader{byDate: map[string]snapshots.PortfolioSnapshot{
		"2026-01-31": validSnapshot(t, portfolioID, "2026-01-31", "1000"),
		"2026-03-31": validSnapshot(t, portfolioID, "2026-03-31", "1400"),
	}}, &fakeCashFlowReader{})
	if _, err := service.Get(uuid.New(), portfolioID, "2026-01-31", "2026-03-31", "USD"); err == nil {
		t.Fatal("Get error = nil, want missing currency error")
	}
}

func TestGetRejectsUnownedPortfolio(t *testing.T) {
	service := NewService(&fakePortfolioReader{err: gorm.ErrRecordNotFound}, &fakeSnapshotReader{}, &fakeCashFlowReader{})
	if _, err := service.Get(uuid.New(), uuid.New(), "2026-01-31", "2026-03-31", "INR"); err == nil {
		t.Fatal("Get error = nil, want portfolio error")
	}
}

func validSnapshot(t *testing.T, portfolioID uuid.UUID, date, value string) snapshots.PortfolioSnapshot {
	t.Helper()
	totals, _ := snapshots.NewJSONB([]finance.CurrencyValue{{Currency: "INR", Amount: decimal.RequireFromString(value)}})
	empty, _ := snapshots.NewJSONB([]any{})
	metadata, _ := snapshots.NewJSONB(finance.MetricDefinition{})
	return snapshots.PortfolioSnapshot{PortfolioID: portfolioID, SnapshotDate: mustDate(date), SnapshotPeriod: snapshots.SnapshotPeriodDaily,
		TotalValues: totals, AssetAllocations: empty, AssetClassAllocations: empty, CashAllocations: empty, MissingPrices: empty,
		IsFullyValued: true, ValuationMetadata: metadata, AllocationMetadata: metadata, HoldingsMetadata: metadata}
}

func mustDate(raw string) time.Time {
	date, err := time.Parse(dateLayout, raw)
	if err != nil {
		panic(err)
	}
	return date
}
