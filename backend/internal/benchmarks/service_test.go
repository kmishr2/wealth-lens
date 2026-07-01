package benchmarks

import (
	"errors"
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

type fakePortfolioRepo struct {
	portfolio *portfolios.Portfolio
	err       error
}

func (f *fakePortfolioRepo) GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	return f.portfolio, f.err
}

type fakeBenchmarkRepo struct {
	benchmark    *Benchmark
	observations map[string]*BenchmarkObservation
	created      *Benchmark
}

func (f *fakeBenchmarkRepo) Create(benchmark *Benchmark) error {
	f.created = benchmark
	return nil
}

func (f *fakeBenchmarkRepo) List(common.Pagination) ([]Benchmark, error) {
	if f.benchmark == nil {
		return []Benchmark{}, nil
	}
	return []Benchmark{*f.benchmark}, nil
}

func (f *fakeBenchmarkRepo) GetByID(uuid.UUID) (*Benchmark, error) {
	if f.benchmark == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.benchmark, nil
}

func (f *fakeBenchmarkRepo) CreateObservation(observation *BenchmarkObservation) error {
	return nil
}

func (f *fakeBenchmarkRepo) ListObservations(uuid.UUID, common.Pagination) ([]BenchmarkObservation, error) {
	return []BenchmarkObservation{}, nil
}

func (f *fakeBenchmarkRepo) GetObservationByDate(_ uuid.UUID, observationDate time.Time) (*BenchmarkObservation, error) {
	observation := f.observations[observationDate.UTC().Format(benchmarkDateLayout)]
	if observation == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return observation, nil
}

type fakeSnapshotRepo struct {
	snapshots map[string]*snapshots.PortfolioSnapshot
}

func (f *fakeSnapshotRepo) GetByPortfolioDatePeriod(_ uuid.UUID, snapshotDate time.Time, _ string) (*snapshots.PortfolioSnapshot, error) {
	snapshot := f.snapshots[snapshotDate.UTC().Format(benchmarkDateLayout)]
	if snapshot == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return snapshot, nil
}

func TestComparePortfolioCalculatesBenchmarkExcessReturn(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	benchmarkID := uuid.New()

	service := NewService(
		&fakeBenchmarkRepo{
			benchmark: &Benchmark{ID: benchmarkID, Code: "NIFTY50", Name: "Nifty 50", Currency: "INR", Source: "manual"},
			observations: map[string]*BenchmarkObservation{
				"2026-01-01": benchmarkObservation(benchmarkID, "2026-01-01", "20000"),
				"2026-02-01": benchmarkObservation(benchmarkID, "2026-02-01", "21000"),
			},
		},
		&fakePortfolioRepo{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotRepo{snapshots: map[string]*snapshots.PortfolioSnapshot{
			"2026-01-01": benchmarkSnapshot(t, userID, portfolioID, "2026-01-01", "INR", "100000"),
			"2026-02-01": benchmarkSnapshot(t, userID, portfolioID, "2026-02-01", "INR", "112000"),
		}},
	)

	response, err := service.ComparePortfolio(userID, portfolioID, benchmarkID, "2026-01-01", "2026-02-01", "")
	if err != nil {
		t.Fatalf("ComparePortfolio returned error: %v", err)
	}

	if response.Currency != "INR" {
		t.Fatalf("currency = %s, want INR", response.Currency)
	}
	if !response.PortfolioTotalReturn.Equal(decimal.RequireFromString("12")) {
		t.Fatalf("portfolio total return = %s, want 12", response.PortfolioTotalReturn)
	}
	if !response.BenchmarkTotalReturn.Equal(decimal.RequireFromString("5")) {
		t.Fatalf("benchmark total return = %s, want 5", response.BenchmarkTotalReturn)
	}
	if !response.ExcessTotalReturn.Equal(decimal.RequireFromString("7")) {
		t.Fatalf("excess total return = %s, want 7", response.ExcessTotalReturn)
	}
	if response.ComparisonMetadata.Name != "Benchmark Return Comparison" {
		t.Fatalf("metadata = %+v", response.ComparisonMetadata)
	}
}

func TestComparePortfolioRejectsCurrencyMismatch(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	benchmarkID := uuid.New()
	service := NewService(
		&fakeBenchmarkRepo{benchmark: &Benchmark{ID: benchmarkID, Code: "SPX", Name: "S&P 500", Currency: "USD", Source: "manual"}},
		&fakePortfolioRepo{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotRepo{},
	)

	_, err := service.ComparePortfolio(userID, portfolioID, benchmarkID, "2026-01-01", "2026-02-01", "INR")
	if err == nil {
		t.Fatal("ComparePortfolio returned nil error")
	}
	var appErr *common.AppError
	if !errors.As(err, &appErr) || appErr.Message != "Benchmark currency must match requested comparison currency" {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateNormalizesBenchmarkCodeAndCurrency(t *testing.T) {
	userID := uuid.New()
	repo := &fakeBenchmarkRepo{}
	service := NewService(repo, &fakePortfolioRepo{}, &fakeSnapshotRepo{})

	response, err := service.Create(userID, BenchmarkCreateRequest{
		Code:     " nifty50 ",
		Name:     "Nifty 50",
		Currency: "inr",
		Source:   "NSE manual export",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if response.Code != "NIFTY50" || response.Currency != "INR" {
		t.Fatalf("response = %+v", response)
	}
	if repo.created == nil || repo.created.CreatedByUserID == nil || *repo.created.CreatedByUserID != userID {
		t.Fatalf("created benchmark = %+v", repo.created)
	}
}

func benchmarkObservation(benchmarkID uuid.UUID, date string, value string) *BenchmarkObservation {
	return &BenchmarkObservation{
		ID:              uuid.New(),
		BenchmarkID:     benchmarkID,
		ObservationDate: benchmarkDate(date),
		Value:           decimal.RequireFromString(value),
		Source:          "manual",
	}
}

func benchmarkSnapshot(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, date string, currency string, value string) *snapshots.PortfolioSnapshot {
	t.Helper()

	totalValues, err := snapshots.NewJSONB([]finance.CurrencyValue{{Currency: currency, Amount: decimal.RequireFromString(value)}})
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

	return &snapshots.PortfolioSnapshot{
		ID:                    uuid.New(),
		PortfolioID:           portfolioID,
		SnapshotDate:          benchmarkDate(date),
		SnapshotPeriod:        snapshots.SnapshotPeriodDaily,
		TotalValues:           totalValues,
		AssetAllocations:      emptyAssetAllocations,
		AssetClassAllocations: emptyAssetClassAllocations,
		CashAllocations:       emptyCashAllocations,
		MissingPrices:         emptyMissingPrices,
		IsFullyValued:         true,
		ValuationScope:        "test",
		AllocationScope:       "test",
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       userID,
	}
}

func benchmarkDate(date string) time.Time {
	parsed, err := time.Parse(benchmarkDateLayout, date)
	if err != nil {
		panic(err)
	}
	return parsed
}
