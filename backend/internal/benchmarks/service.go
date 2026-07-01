package benchmarks

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const defaultBenchmarkObservationSource = "manual"

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type benchmarkStore interface {
	Create(benchmark *Benchmark) error
	List(pagination common.Pagination) ([]Benchmark, error)
	GetByID(benchmarkID uuid.UUID) (*Benchmark, error)
	CreateObservation(observation *BenchmarkObservation) error
	ListObservations(benchmarkID uuid.UUID, pagination common.Pagination) ([]BenchmarkObservation, error)
	GetObservationByDate(benchmarkID uuid.UUID, observationDate time.Time) (*BenchmarkObservation, error)
}

type snapshotReader interface {
	GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*snapshots.PortfolioSnapshot, error)
}

type Service struct {
	repo          benchmarkStore
	portfolioRepo portfolioReader
	snapshotRepo  snapshotReader
}

func NewService(repo benchmarkStore, portfolioRepo portfolioReader, snapshotRepo snapshotReader) *Service {
	return &Service{
		repo:          repo,
		portfolioRepo: portfolioRepo,
		snapshotRepo:  snapshotRepo,
	}
}

func (s *Service) Create(userID uuid.UUID, req BenchmarkCreateRequest) (BenchmarkResponse, error) {
	benchmark, err := buildBenchmark(userID, req)
	if err != nil {
		return BenchmarkResponse{}, err
	}

	if err := s.repo.Create(benchmark); err != nil {
		if common.IsUniqueViolation(err) {
			return BenchmarkResponse{}, common.Conflict("Benchmark code already exists")
		}
		return BenchmarkResponse{}, err
	}

	return ToBenchmarkResponse(*benchmark), nil
}

func (s *Service) List(pagination common.Pagination) ([]BenchmarkResponse, error) {
	benchmarks, err := s.repo.List(pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]BenchmarkResponse, 0, len(benchmarks))
	for _, benchmark := range benchmarks {
		responses = append(responses, ToBenchmarkResponse(benchmark))
	}
	return responses, nil
}

func (s *Service) CreateObservation(userID uuid.UUID, benchmarkID uuid.UUID, req BenchmarkObservationCreateRequest) (BenchmarkObservationResponse, error) {
	if _, err := s.getBenchmark(benchmarkID); err != nil {
		return BenchmarkObservationResponse{}, err
	}

	observation, err := buildObservation(userID, benchmarkID, req)
	if err != nil {
		return BenchmarkObservationResponse{}, err
	}

	if err := s.repo.CreateObservation(observation); err != nil {
		if common.IsUniqueViolation(err) {
			return BenchmarkObservationResponse{}, common.Conflict("Benchmark observation already exists for this date")
		}
		return BenchmarkObservationResponse{}, err
	}

	return ToBenchmarkObservationResponse(*observation), nil
}

func (s *Service) ListObservations(benchmarkID uuid.UUID, pagination common.Pagination) ([]BenchmarkObservationResponse, error) {
	if _, err := s.getBenchmark(benchmarkID); err != nil {
		return nil, err
	}

	observations, err := s.repo.ListObservations(benchmarkID, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]BenchmarkObservationResponse, 0, len(observations))
	for _, observation := range observations {
		responses = append(responses, ToBenchmarkObservationResponse(observation))
	}
	return responses, nil
}

func (s *Service) ComparePortfolio(userID uuid.UUID, portfolioID uuid.UUID, benchmarkID uuid.UUID, rawStartDate string, rawEndDate string, rawCurrency string) (BenchmarkComparisonResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	benchmark, err := s.getBenchmark(benchmarkID)
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	startDate, endDate, err := parseBenchmarkDateRange(rawStartDate, rawEndDate)
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	currency := common.NormalizeCurrency(rawCurrency)
	if currency == "" {
		currency = benchmark.Currency
	}
	if !common.ValidateCurrency(currency) {
		return BenchmarkComparisonResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	if currency != benchmark.Currency {
		return BenchmarkComparisonResponse{}, common.BadRequest("Benchmark currency must match requested comparison currency")
	}

	startSnapshot, err := s.getSnapshot(portfolioID, startDate, "Start snapshot not found")
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}
	endSnapshot, err := s.getSnapshot(portfolioID, endDate, "End snapshot not found")
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	startPortfolioValue, endPortfolioValue, err := portfolioValuesForCurrency(*startSnapshot, *endSnapshot, currency)
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	startObservation, err := s.getObservation(benchmark.ID, startDate, "Start benchmark observation not found")
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}
	endObservation, err := s.getObservation(benchmark.ID, endDate, "End benchmark observation not found")
	if err != nil {
		return BenchmarkComparisonResponse{}, err
	}

	result, err := finance.CalculateBenchmarkComparison(finance.BenchmarkComparisonInput{
		PortfolioBeginningValue: startPortfolioValue,
		PortfolioEndingValue:    endPortfolioValue,
		BenchmarkBeginningValue: startObservation.Value,
		BenchmarkEndingValue:    endObservation.Value,
		StartDate:               startDate,
		EndDate:                 endDate,
	})
	if err != nil {
		return BenchmarkComparisonResponse{}, common.BadRequest(fmt.Sprintf("Benchmark comparison error: %s", err.Error()))
	}

	return BenchmarkComparisonResponse{
		PortfolioID:          portfolioID,
		BenchmarkID:          benchmark.ID,
		BenchmarkCode:        benchmark.Code,
		BenchmarkName:        benchmark.Name,
		Currency:             currency,
		StartDate:            dateString(startDate),
		EndDate:              dateString(endDate),
		PortfolioStartValue:  startPortfolioValue,
		PortfolioEndValue:    endPortfolioValue,
		BenchmarkStartValue:  startObservation.Value,
		BenchmarkEndValue:    endObservation.Value,
		PortfolioTotalReturn: result.PortfolioTotalReturn,
		BenchmarkTotalReturn: result.BenchmarkTotalReturn,
		PortfolioCAGR:        result.PortfolioCAGR,
		BenchmarkCAGR:        result.BenchmarkCAGR,
		ExcessTotalReturn:    result.ExcessTotalReturn,
		ExcessCAGR:           result.ExcessCAGR,
		ComparisonScope:      "Benchmark comparison uses immutable daily portfolio snapshots and exact benchmark observations for the same dates and currency. No default benchmark, price interpolation, currency conversion, or prediction is applied.",
		ComparisonMetadata:   result.Definition,
	}, nil
}

func (s *Service) getBenchmark(benchmarkID uuid.UUID) (*Benchmark, error) {
	benchmark, err := s.repo.GetByID(benchmarkID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Benchmark not found")
	}
	if err != nil {
		return nil, err
	}
	return benchmark, nil
}

func (s *Service) getOwnedPortfolio(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Portfolio not found")
	}
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (s *Service) getSnapshot(portfolioID uuid.UUID, snapshotDate time.Time, notFoundMessage string) (*snapshots.PortfolioSnapshot, error) {
	snapshot, err := s.snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, snapshots.SnapshotPeriodDaily)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound(notFoundMessage)
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *Service) getObservation(benchmarkID uuid.UUID, observationDate time.Time, notFoundMessage string) (*BenchmarkObservation, error) {
	observation, err := s.repo.GetObservationByDate(benchmarkID, observationDate)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound(notFoundMessage)
	}
	if err != nil {
		return nil, err
	}
	return observation, nil
}

func buildBenchmark(userID uuid.UUID, req BenchmarkCreateRequest) (*Benchmark, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, common.BadRequest("Benchmark code is required")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, common.BadRequest("Benchmark name is required")
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		return nil, common.BadRequest("Benchmark source is required")
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return nil, common.BadRequest("Currency must be a three-letter uppercase code")
	}

	return &Benchmark{
		Code:            code,
		Name:            name,
		Currency:        currency,
		Source:          source,
		Description:     strings.TrimSpace(req.Description),
		CreatedByUserID: &userID,
	}, nil
}

func buildObservation(userID uuid.UUID, benchmarkID uuid.UUID, req BenchmarkObservationCreateRequest) (*BenchmarkObservation, error) {
	observationDate, err := parseBenchmarkDate(req.ObservationDate, "Observation date")
	if err != nil {
		return nil, err
	}
	if req.Value == nil || !req.Value.GreaterThan(decimal.Zero) {
		return nil, common.BadRequest("Benchmark value must be greater than zero")
	}
	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = defaultBenchmarkObservationSource
	}

	return &BenchmarkObservation{
		BenchmarkID:     benchmarkID,
		ObservationDate: observationDate,
		Value:           req.Value.Round(10),
		Source:          source,
		Note:            strings.TrimSpace(req.Note),
		CreatedByUserID: &userID,
	}, nil
}

func portfolioValuesForCurrency(startSnapshot snapshots.PortfolioSnapshot, endSnapshot snapshots.PortfolioSnapshot, currency string) (decimal.Decimal, decimal.Decimal, error) {
	startResponse, err := snapshots.ToResponse(startSnapshot)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	endResponse, err := snapshots.ToResponse(endSnapshot)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	startTotals := totalsByCurrency(startResponse.TotalValues)
	endTotals := totalsByCurrency(endResponse.TotalValues)
	startValue, startOK := startTotals[currency]
	endValue, endOK := endTotals[currency]
	if !startOK || !endOK {
		return decimal.Zero, decimal.Zero, common.BadRequest("Portfolio snapshots do not contain the requested currency")
	}
	if !startValue.GreaterThan(decimal.Zero) || !endValue.GreaterThan(decimal.Zero) {
		return decimal.Zero, decimal.Zero, common.BadRequest("Portfolio comparison values must be greater than zero")
	}

	return startValue, endValue, nil
}

func totalsByCurrency(values []finance.CurrencyValue) map[string]decimal.Decimal {
	totals := make(map[string]decimal.Decimal, len(values))
	for _, value := range values {
		totals[value.Currency] = value.Amount
	}
	return totals
}

func parseBenchmarkDateRange(rawStartDate string, rawEndDate string) (time.Time, time.Time, error) {
	startDate, err := parseBenchmarkDate(rawStartDate, "Start date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := parseBenchmarkDate(rawEndDate, "End date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !endDate.After(startDate) {
		return time.Time{}, time.Time{}, common.BadRequest("End date must be after start date")
	}
	return startDate, endDate, nil
}

func parseBenchmarkDate(raw string, label string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s is required", label))
	}

	date, err := time.Parse(benchmarkDateLayout, raw)
	if err != nil {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s must use YYYY-MM-DD format", label))
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s cannot be in the future", label))
	}
	return date, nil
}

func dateString(date time.Time) string {
	return date.UTC().Format(benchmarkDateLayout)
}
