package performance

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakePortfolioReader struct {
	userID      uuid.UUID
	portfolioID uuid.UUID
	portfolio   *portfolios.Portfolio
	err         error
}

func (f *fakePortfolioReader) GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	f.userID = userID
	f.portfolioID = portfolioID
	return f.portfolio, f.err
}

type fakeSnapshotReader struct {
	snapshots map[string]snapshots.PortfolioSnapshot
	calls     []snapshotCall
	err       error
}

type snapshotCall struct {
	portfolioID uuid.UUID
	date        time.Time
	period      string
}

func (f *fakeSnapshotReader) GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*snapshots.PortfolioSnapshot, error) {
	f.calls = append(f.calls, snapshotCall{portfolioID: portfolioID, date: snapshotDate, period: snapshotPeriod})
	if f.err != nil {
		return nil, f.err
	}
	snapshot, ok := f.snapshots[dateString(snapshotDate)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &snapshot, nil
}

type fakeExternalCashFlowReader struct {
	portfolioID uuid.UUID
	startAfter  time.Time
	endAt       time.Time
	records     []transactions.ExternalCashFlowRecord
	err         error
}

func (f *fakeExternalCashFlowReader) ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error) {
	f.portfolioID = portfolioID
	f.startAfter = startAfter
	f.endAt = endAt
	return f.records, f.err
}

func TestGetCalculatesCAGRAndXIRRFromSnapshotsAndExternalCashFlows(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	snapshotReader := &fakeSnapshotReader{
		snapshots: map[string]snapshots.PortfolioSnapshot{
			"2025-01-01": storedPerformanceSnapshot(t, userID, portfolioID, startDate, "USD", "1000"),
			"2026-01-01": storedPerformanceSnapshot(t, userID, portfolioID, endDate, "USD", "1300"),
		},
	}
	cashFlowReader := &fakeExternalCashFlowReader{
		records: []transactions.ExternalCashFlowRecord{
			{
				OccurredAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
				Currency:   "USD",
				Amount:     decimal.RequireFromString("100"),
			},
			{
				OccurredAt: time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC),
				Currency:   "USD",
				Amount:     decimal.RequireFromString("-50"),
			},
		},
	}
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		snapshotReader,
		cashFlowReader,
	)

	response, err := service.Get(userID, portfolioID, "2025-01-01", "2026-01-01")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if len(snapshotReader.calls) != 2 {
		t.Fatalf("snapshot calls = %d, want 2", len(snapshotReader.calls))
	}
	if cashFlowReader.portfolioID != portfolioID {
		t.Fatalf("cash flow portfolio = %s, want %s", cashFlowReader.portfolioID, portfolioID)
	}
	expectedStartAfter := time.Date(2025, 1, 1, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	expectedEndAt := time.Date(2026, 1, 1, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !cashFlowReader.startAfter.Equal(expectedStartAfter) || !cashFlowReader.endAt.Equal(expectedEndAt) {
		t.Fatalf("cash flow range = %s..%s, want %s..%s", cashFlowReader.startAfter, cashFlowReader.endAt, expectedStartAfter, expectedEndAt)
	}
	if len(response.CurrencyReturns) != 1 {
		t.Fatalf("currency returns length = %d, want 1", len(response.CurrencyReturns))
	}
	result := response.CurrencyReturns[0]
	if result.Currency != "USD" {
		t.Fatalf("currency = %s, want USD", result.Currency)
	}
	if !result.BeginningValue.Equal(decimal.RequireFromString("1000")) || !result.EndingValue.Equal(decimal.RequireFromString("1300")) {
		t.Fatalf("values = %s -> %s, want 1000 -> 1300", result.BeginningValue, result.EndingValue)
	}
	if !result.NetExternalCashFlow.Equal(decimal.RequireFromString("50")) {
		t.Fatalf("net external cash flow = %s, want 50", result.NetExternalCashFlow)
	}
	if !result.ProfitLoss.Equal(decimal.RequireFromString("250")) {
		t.Fatalf("profit loss = %s, want 250", result.ProfitLoss)
	}
	assertApprox(t, result.CAGR, decimal.RequireFromString("30"), decimal.RequireFromString("0.1"))
	assertApprox(t, result.XIRR, decimal.RequireFromString("24.05"), decimal.RequireFromString("0.1"))
	if result.CashFlowCount != 2 {
		t.Fatalf("cash flow count = %d, want 2", result.CashFlowCount)
	}
	if response.PnLMetadata.Name == "" || response.CAGRMetadata.Name == "" || response.XIRRMetadata.Name == "" || response.PerformanceScope == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestGetFiltersToCurrenciesPresentInBothSnapshots(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotReader{
			snapshots: map[string]snapshots.PortfolioSnapshot{
				"2025-01-01": storedPerformanceSnapshotWithTotals(t, userID, portfolioID, startDate, []finance.CurrencyValue{
					{Currency: "INR", Amount: decimal.RequireFromString("1000")},
					{Currency: "USD", Amount: decimal.RequireFromString("1000")},
				}),
				"2026-01-01": storedPerformanceSnapshotWithTotals(t, userID, portfolioID, endDate, []finance.CurrencyValue{
					{Currency: "USD", Amount: decimal.RequireFromString("1100")},
				}),
			},
		},
		&fakeExternalCashFlowReader{},
	)

	response, err := service.Get(userID, portfolioID, "2025-01-01", "2026-01-01")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if len(response.CurrencyReturns) != 1 || response.CurrencyReturns[0].Currency != "USD" {
		t.Fatalf("currency returns = %+v, want only USD", response.CurrencyReturns)
	}
}

func TestGetReturnsNotFoundWhenSnapshotIsMissing(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotReader{snapshots: map[string]snapshots.PortfolioSnapshot{}},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(userID, portfolioID, "2025-01-01", "2026-01-01")
	if err == nil {
		t.Fatal("Get returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusNotFound || appErr.Message != "Start snapshot not found" {
		t.Fatalf("error = %+v, want start snapshot not found", appErr)
	}
}

func TestGetRejectsInvalidDateRange(t *testing.T) {
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}},
		&fakeSnapshotReader{},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(uuid.New(), uuid.New(), "2026-01-01", "2025-01-01")
	if err == nil {
		t.Fatal("Get returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusBadRequest)
	}
}

func TestGetReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(
		&fakePortfolioReader{err: gorm.ErrRecordNotFound},
		&fakeSnapshotReader{},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(uuid.New(), uuid.New(), "2025-01-01", "2026-01-01")
	if err == nil {
		t.Fatal("Get returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
}

func storedPerformanceSnapshot(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, snapshotDate time.Time, currency string, amount string) snapshots.PortfolioSnapshot {
	t.Helper()
	return storedPerformanceSnapshotWithTotals(t, userID, portfolioID, snapshotDate, []finance.CurrencyValue{
		{Currency: currency, Amount: decimal.RequireFromString(amount)},
	})
}

func storedPerformanceSnapshotWithTotals(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, snapshotDate time.Time, totals []finance.CurrencyValue) snapshots.PortfolioSnapshot {
	t.Helper()

	totalValues, err := snapshots.NewJSONB(totals)
	if err != nil {
		t.Fatalf("total values JSON: %v", err)
	}
	emptyAssetAllocations, err := snapshots.NewJSONB([]finance.AssetAllocation{})
	if err != nil {
		t.Fatalf("asset allocations JSON: %v", err)
	}
	emptyAssetClassAllocations, err := snapshots.NewJSONB([]finance.AssetClassAllocation{})
	if err != nil {
		t.Fatalf("asset class allocations JSON: %v", err)
	}
	emptyCashAllocations, err := snapshots.NewJSONB([]finance.CashAllocation{})
	if err != nil {
		t.Fatalf("cash allocations JSON: %v", err)
	}
	emptyMissingPrices, err := snapshots.NewJSONB([]finance.MissingPrice{})
	if err != nil {
		t.Fatalf("missing prices JSON: %v", err)
	}
	valuationMetadata, err := snapshots.NewJSONB(finance.PortfolioValuationDefinition())
	if err != nil {
		t.Fatalf("valuation metadata JSON: %v", err)
	}
	allocationMetadata, err := snapshots.NewJSONB(finance.AllocationDefinition())
	if err != nil {
		t.Fatalf("allocation metadata JSON: %v", err)
	}
	holdingsMetadata, err := snapshots.NewJSONB(finance.HoldingsDefinition())
	if err != nil {
		t.Fatalf("holdings metadata JSON: %v", err)
	}

	return snapshots.PortfolioSnapshot{
		ID:                    uuid.New(),
		PortfolioID:           portfolioID,
		SnapshotDate:          snapshotDate,
		SnapshotPeriod:        snapshots.SnapshotPeriodDaily,
		TotalValues:           totalValues,
		AssetAllocations:      emptyAssetAllocations,
		AssetClassAllocations: emptyAssetClassAllocations,
		CashAllocations:       emptyCashAllocations,
		MissingPrices:         emptyMissingPrices,
		IsFullyValued:         true,
		ValuationScope:        "Totals are grouped by currency; no currency conversion is applied.",
		AllocationScope:       "Allocation percentages are calculated separately per currency; no currency conversion is applied.",
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       userID,
		CreatedAt:             time.Now().UTC(),
	}
}

func assertApprox(t *testing.T, got decimal.Decimal, want decimal.Decimal, tolerance decimal.Decimal) {
	t.Helper()

	diff := got.Sub(want).Abs()
	if diff.GreaterThan(tolerance) {
		t.Fatalf("value = %s, want %s +/- %s", got, want, tolerance)
	}
}
