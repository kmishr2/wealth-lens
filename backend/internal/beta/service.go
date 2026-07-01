package beta

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/benchmarks"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

type portfolioReader interface {
	GetOwned(uuid.UUID, uuid.UUID) (*portfolios.Portfolio, error)
}
type benchmarkReader interface {
	GetByID(uuid.UUID) (*benchmarks.Benchmark, error)
	ListObservationsByDateRange(uuid.UUID, time.Time, time.Time) ([]benchmarks.BenchmarkObservation, error)
}
type snapshotReader interface {
	ListByPortfolioDateRange(uuid.UUID, time.Time, time.Time, string) ([]snapshots.PortfolioSnapshot, error)
}
type cashFlowReader interface {
	ListExternalCashFlows(uuid.UUID, time.Time, time.Time) ([]transactions.ExternalCashFlowRecord, error)
}

type Service struct {
	portfolios portfolioReader
	benchmarks benchmarkReader
	snapshots  snapshotReader
	cashFlows  cashFlowReader
}

func NewService(portfolios portfolioReader, benchmarks benchmarkReader, snapshots snapshotReader, cashFlows cashFlowReader) *Service {
	return &Service{portfolios: portfolios, benchmarks: benchmarks, snapshots: snapshots, cashFlows: cashFlows}
}

func (s *Service) Get(userID, portfolioID, benchmarkID uuid.UUID, rawStart, rawEnd, rawCurrency string) (Response, error) {
	if _, err := s.portfolios.GetOwned(userID, portfolioID); errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, common.NotFound("Portfolio not found")
	} else if err != nil {
		return Response{}, err
	}
	benchmark, err := s.benchmarks.GetByID(benchmarkID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, common.NotFound("Benchmark not found")
	}
	if err != nil {
		return Response{}, err
	}
	start, end, err := parseDateRange(rawStart, rawEnd)
	if err != nil {
		return Response{}, err
	}
	currency := common.NormalizeCurrency(rawCurrency)
	if currency == "" {
		currency = benchmark.Currency
	}
	if !common.ValidateCurrency(currency) || currency != benchmark.Currency {
		return Response{}, common.BadRequest("Requested currency must match the benchmark currency")
	}

	snapshotRecords, err := s.snapshots.ListByPortfolioDateRange(portfolioID, start, end, snapshots.SnapshotPeriodDaily)
	if err != nil {
		return Response{}, err
	}
	snapshotResponses, err := snapshots.ToResponses(snapshotRecords)
	if err != nil {
		return Response{}, err
	}
	observations, err := s.benchmarks.ListObservationsByDateRange(benchmarkID, start, end)
	if err != nil {
		return Response{}, err
	}
	benchmarkByDate := make(map[string]decimal.Decimal, len(observations))
	for _, observation := range observations {
		benchmarkByDate[observation.ObservationDate.UTC().Format(dateLayout)] = observation.Value
	}

	aligned := make([]finance.BetaObservation, 0)
	for _, snapshot := range snapshotResponses {
		if !snapshot.IsFullyValued {
			continue
		}
		benchmarkValue, ok := benchmarkByDate[snapshot.SnapshotDate]
		if !ok {
			continue
		}
		portfolioValue, ok := totalForCurrency(snapshot.TotalValues, currency)
		if !ok || !portfolioValue.GreaterThan(decimal.Zero) {
			continue
		}
		date, _ := time.Parse(dateLayout, snapshot.SnapshotDate)
		aligned = append(aligned, finance.BetaObservation{Date: date, PortfolioValue: portfolioValue, BenchmarkValue: benchmarkValue})
	}
	if len(aligned) < 3 {
		return Response{}, common.BadRequest("Beta requires at least three exact aligned portfolio and benchmark observations")
	}
	sort.Slice(aligned, func(i, j int) bool { return aligned[i].Date.Before(aligned[j].Date) })
	flows, err := s.cashFlows.ListExternalCashFlows(portfolioID, endOfDay(aligned[0].Date), endOfDay(aligned[len(aligned)-1].Date))
	if err != nil {
		return Response{}, err
	}
	for index := 1; index < len(aligned); index++ {
		aligned[index].NetExternalCashFlow = cashFlowForInterval(flows, currency, endOfDay(aligned[index-1].Date), endOfDay(aligned[index].Date))
	}
	result, err := finance.CalculateBeta(aligned)
	if err != nil {
		return Response{}, common.BadRequest(fmt.Sprintf("Beta calculation error: %s", err.Error()))
	}
	return Response{PortfolioID: portfolioID, BenchmarkID: benchmarkID, BenchmarkCode: benchmark.Code, Currency: currency,
		StartDate: start.Format(dateLayout), EndDate: end.Format(dateLayout), AlignedCount: len(aligned), PairedReturnCount: result.PairedReturnCount,
		Beta: result.Beta, Scope: "Beta uses exact aligned immutable daily portfolio snapshots and benchmark observations, with ledger-derived external cash flows removed and no interpolation or currency conversion.", MetricMetadata: result.Definition}, nil
}

func parseDateRange(rawStart, rawEnd string) (time.Time, time.Time, error) {
	start, err := time.Parse(dateLayout, rawStart)
	if err != nil {
		return time.Time{}, time.Time{}, common.BadRequest("Start date must use YYYY-MM-DD format")
	}
	end, err := time.Parse(dateLayout, rawEnd)
	if err != nil {
		return time.Time{}, time.Time{}, common.BadRequest("End date must use YYYY-MM-DD format")
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, common.BadRequest("End date must be after start date")
	}
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if end.After(today) {
		return time.Time{}, time.Time{}, common.BadRequest("End date cannot be in the future")
	}
	return start, end, nil
}

func totalForCurrency(values []finance.CurrencyValue, currency string) (decimal.Decimal, bool) {
	for _, value := range values {
		if value.Currency == currency {
			return value.Amount, true
		}
	}
	return decimal.Zero, false
}

func cashFlowForInterval(records []transactions.ExternalCashFlowRecord, currency string, startAfter, endAt time.Time) decimal.Decimal {
	total := decimal.Zero
	for _, record := range records {
		if record.Currency == currency && record.OccurredAt.After(startAfter) && !record.OccurredAt.After(endAt) {
			total = total.Add(record.Amount)
		}
	}
	return total
}

func endOfDay(date time.Time) time.Time {
	return date.UTC().Truncate(24 * time.Hour).Add(24*time.Hour - time.Nanosecond)
}
