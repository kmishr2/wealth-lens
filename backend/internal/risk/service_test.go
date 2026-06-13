package risk

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
	portfolio *portfolios.Portfolio
	err       error
}

func (f *fakePortfolioReader) GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	return f.portfolio, f.err
}

type fakeSnapshotRangeReader struct {
	portfolioID uuid.UUID
	startDate   time.Time
	endDate     time.Time
	period      string
	snapshots   []snapshots.PortfolioSnapshot
	err         error
}

func (f *fakeSnapshotRangeReader) ListByPortfolioDateRange(portfolioID uuid.UUID, startDate time.Time, endDate time.Time, snapshotPeriod string) ([]snapshots.PortfolioSnapshot, error) {
	f.portfolioID = portfolioID
	f.startDate = startDate
	f.endDate = endDate
	f.period = snapshotPeriod
	return f.snapshots, f.err
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

func TestGetCalculatesCashFlowAdjustedRiskFromSnapshots(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	snapshotReader := &fakeSnapshotRangeReader{
		snapshots: []snapshots.PortfolioSnapshot{
			storedRiskSnapshot(t, userID, portfolioID, "2026-01-01", "USD", "100", true),
			storedRiskSnapshot(t, userID, portfolioID, "2026-01-02", "USD", "210", true),
			storedRiskSnapshot(t, userID, portfolioID, "2026-01-03", "USD", "189", true),
		},
	}
	cashFlowReader := &fakeExternalCashFlowReader{
		records: []transactions.ExternalCashFlowRecord{
			{
				OccurredAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
				Currency:   "USD",
				Amount:     decimal.RequireFromString("100"),
			},
		},
	}
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		snapshotReader,
		cashFlowReader,
	)

	response, err := service.Get(userID, portfolioID, "2026-01-01", "2026-01-03", "252")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if snapshotReader.portfolioID != portfolioID || snapshotReader.period != snapshots.SnapshotPeriodDaily {
		t.Fatalf("snapshot query portfolio=%s period=%s", snapshotReader.portfolioID, snapshotReader.period)
	}
	expectedStartAfter := time.Date(2026, 1, 1, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	expectedEndAt := time.Date(2026, 1, 3, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !cashFlowReader.startAfter.Equal(expectedStartAfter) || !cashFlowReader.endAt.Equal(expectedEndAt) {
		t.Fatalf("cash flow range = %s..%s, want %s..%s", cashFlowReader.startAfter, cashFlowReader.endAt, expectedStartAfter, expectedEndAt)
	}
	if len(response.CurrencyRisk) != 1 {
		t.Fatalf("currency risk length = %d, want 1", len(response.CurrencyRisk))
	}
	result := response.CurrencyRisk[0]
	if result.Currency != "USD" || result.ObservationCount != 3 || result.PeriodicReturnCount != 2 {
		t.Fatalf("risk result = %+v", result)
	}
	assertRiskApprox(t, result.AnnualizedVolatility, decimal.RequireFromString("224.4994"), decimal.RequireFromString("0.001"))
	if !result.MaximumDrawdown.Equal(decimal.RequireFromString("-10")) {
		t.Fatalf("maximum drawdown = %s, want -10", result.MaximumDrawdown)
	}
	if result.PeakDate != "2026-01-02" || result.TroughDate != "2026-01-03" {
		t.Fatalf("peak/trough = %s/%s, want 2026-01-02/2026-01-03", result.PeakDate, result.TroughDate)
	}
	if response.VolatilityMetadata.Name == "" || response.DrawdownMetadata.Name == "" || response.RiskScope == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestGetRejectsIncompleteSnapshot(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotRangeReader{
			snapshots: []snapshots.PortfolioSnapshot{
				storedRiskSnapshot(t, userID, portfolioID, "2026-01-01", "USD", "100", true),
				storedRiskSnapshot(t, userID, portfolioID, "2026-01-02", "USD", "110", false),
				storedRiskSnapshot(t, userID, portfolioID, "2026-01-03", "USD", "120", true),
			},
		},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(userID, portfolioID, "2026-01-01", "2026-01-03", "252")
	if err == nil {
		t.Fatal("Get returned nil error")
	}
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestGetRequiresThreeSnapshots(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotRangeReader{
			snapshots: []snapshots.PortfolioSnapshot{
				storedRiskSnapshot(t, userID, portfolioID, "2026-01-01", "USD", "100", true),
				storedRiskSnapshot(t, userID, portfolioID, "2026-01-02", "USD", "110", true),
			},
		},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(userID, portfolioID, "2026-01-01", "2026-01-02", "252")
	if err == nil {
		t.Fatal("Get returned nil error")
	}
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestGetRequiresExplicitPeriodsPerYear(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotRangeReader{},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(userID, portfolioID, "2026-01-01", "2026-01-03", "")
	if err == nil {
		t.Fatal("Get returned nil error")
	}
	assertAppErrorStatus(t, err, http.StatusBadRequest)
}

func TestGetReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(
		&fakePortfolioReader{err: gorm.ErrRecordNotFound},
		&fakeSnapshotRangeReader{},
		&fakeExternalCashFlowReader{},
	)

	_, err := service.Get(uuid.New(), uuid.New(), "2026-01-01", "2026-01-03", "252")
	if err == nil {
		t.Fatal("Get returned nil error")
	}
	assertAppErrorStatus(t, err, http.StatusNotFound)
}

func storedRiskSnapshot(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, date string, currency string, amount string, isFullyValued bool) snapshots.PortfolioSnapshot {
	t.Helper()

	snapshotDate, err := time.Parse(riskDateLayout, date)
	if err != nil {
		t.Fatalf("snapshot date: %v", err)
	}
	totalValues, err := snapshots.NewJSONB([]finance.CurrencyValue{
		{Currency: currency, Amount: decimal.RequireFromString(amount)},
	})
	if err != nil {
		t.Fatalf("total values JSON: %v", err)
	}
	emptyAssetAllocations, _ := snapshots.NewJSONB([]finance.AssetAllocation{})
	emptyAssetClassAllocations, _ := snapshots.NewJSONB([]finance.AssetClassAllocation{})
	emptyCashAllocations, _ := snapshots.NewJSONB([]finance.CashAllocation{})
	emptyMissingPrices, _ := snapshots.NewJSONB([]finance.MissingPrice{})
	valuationMetadata, _ := snapshots.NewJSONB(finance.PortfolioValuationDefinition())
	allocationMetadata, _ := snapshots.NewJSONB(finance.AllocationDefinition())
	holdingsMetadata, _ := snapshots.NewJSONB(finance.HoldingsDefinition())

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
		IsFullyValued:         isFullyValued,
		ValuationScope:        "Totals are grouped by currency; no currency conversion is applied.",
		AllocationScope:       "Allocation percentages are calculated separately per currency; no currency conversion is applied.",
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       userID,
		CreatedAt:             time.Now().UTC(),
	}
}

func assertAppErrorStatus(t *testing.T, err error, status int) {
	t.Helper()

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != status {
		t.Fatalf("status = %d, want %d", appErr.Status, status)
	}
}

func assertRiskApprox(t *testing.T, got decimal.Decimal, want decimal.Decimal, tolerance decimal.Decimal) {
	t.Helper()

	if got.Sub(want).Abs().GreaterThan(tolerance) {
		t.Fatalf("value = %s, want %s +/- %s", got, want, tolerance)
	}
}
