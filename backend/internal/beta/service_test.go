package beta

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/benchmarks"
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

type fakeBenchmarkReader struct {
	benchmark    *benchmarks.Benchmark
	observations []benchmarks.BenchmarkObservation
}

func (f *fakeBenchmarkReader) GetByID(uuid.UUID) (*benchmarks.Benchmark, error) {
	if f.benchmark == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.benchmark, nil
}
func (f *fakeBenchmarkReader) ListObservationsByDateRange(uuid.UUID, time.Time, time.Time) ([]benchmarks.BenchmarkObservation, error) {
	return f.observations, nil
}

type fakeSnapshotReader struct{ records []snapshots.PortfolioSnapshot }

func (f *fakeSnapshotReader) ListByPortfolioDateRange(uuid.UUID, time.Time, time.Time, string) ([]snapshots.PortfolioSnapshot, error) {
	return f.records, nil
}

type fakeCashFlowReader struct {
	records []transactions.ExternalCashFlowRecord
}

func (f *fakeCashFlowReader) ListExternalCashFlows(uuid.UUID, time.Time, time.Time) ([]transactions.ExternalCashFlowRecord, error) {
	return f.records, nil
}

func TestGetCalculatesCashFlowAdjustedBetaFromExactDates(t *testing.T) {
	benchmarkID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(
		&fakePortfolioReader{},
		&fakeBenchmarkReader{benchmark: &benchmarks.Benchmark{ID: benchmarkID, Code: "NIFTY50", Currency: "INR"}, observations: []benchmarks.BenchmarkObservation{
			{ObservationDate: mustDate("2026-01-01"), Value: decimal.NewFromInt(100)},
			{ObservationDate: mustDate("2026-01-02"), Value: decimal.NewFromInt(110)},
			{ObservationDate: mustDate("2026-01-03"), Value: decimal.RequireFromString("104.5")},
		}},
		&fakeSnapshotReader{records: []snapshots.PortfolioSnapshot{
			validSnapshot(t, portfolioID, "2026-01-01", "100"),
			validSnapshot(t, portfolioID, "2026-01-02", "120"),
			validSnapshot(t, portfolioID, "2026-01-03", "96"),
		}},
		&fakeCashFlowReader{records: []transactions.ExternalCashFlowRecord{{OccurredAt: mustDate("2026-01-02").Add(12 * time.Hour), Currency: "INR", Amount: decimal.NewFromInt(10)}}},
	)

	response, err := service.Get(uuid.New(), portfolioID, benchmarkID, "2026-01-01", "2026-01-03", "INR")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !response.Beta.Equal(decimal.NewFromInt(2)) || response.AlignedCount != 3 || response.PairedReturnCount != 2 {
		t.Fatalf("response = %+v", response)
	}
	if response.MetricMetadata.Name != "Portfolio Beta" || response.Scope == "" {
		t.Fatalf("metadata = %+v", response)
	}
}

func TestGetRejectsInsufficientExactOverlap(t *testing.T) {
	benchmarkID := uuid.New()
	portfolioID := uuid.New()
	service := NewService(&fakePortfolioReader{}, &fakeBenchmarkReader{benchmark: &benchmarks.Benchmark{ID: benchmarkID, Currency: "INR"}, observations: []benchmarks.BenchmarkObservation{
		{ObservationDate: mustDate("2026-01-01"), Value: decimal.NewFromInt(100)},
	}}, &fakeSnapshotReader{records: []snapshots.PortfolioSnapshot{validSnapshot(t, portfolioID, "2026-01-01", "100")}}, &fakeCashFlowReader{})
	if _, err := service.Get(uuid.New(), portfolioID, benchmarkID, "2026-01-01", "2026-01-03", "INR"); err == nil {
		t.Fatal("Get error = nil, want insufficient overlap error")
	}
}

func TestGetReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(&fakePortfolioReader{err: gorm.ErrRecordNotFound}, &fakeBenchmarkReader{}, &fakeSnapshotReader{}, &fakeCashFlowReader{})
	if _, err := service.Get(uuid.New(), uuid.New(), uuid.New(), "2026-01-01", "2026-01-03", "INR"); err == nil {
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
