package benchmarks

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const benchmarkDateLayout = "2006-01-02"

type BenchmarkCreateRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

type BenchmarkObservationCreateRequest struct {
	ObservationDate string           `json:"observation_date"`
	Value           *decimal.Decimal `json:"value"`
	Source          string           `json:"source"`
	Note            string           `json:"note"`
}

type BenchmarkResponse struct {
	ID              uuid.UUID  `json:"id"`
	Code            string     `json:"code"`
	Name            string     `json:"name"`
	Currency        string     `json:"currency"`
	Source          string     `json:"source"`
	Description     string     `json:"description"`
	CreatedByUserID *uuid.UUID `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

type BenchmarkObservationResponse struct {
	ID              uuid.UUID       `json:"id"`
	BenchmarkID     uuid.UUID       `json:"benchmark_id"`
	ObservationDate string          `json:"observation_date"`
	Value           decimal.Decimal `json:"value"`
	Source          string          `json:"source"`
	Note            string          `json:"note"`
	CreatedByUserID *uuid.UUID      `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
}

type BenchmarkComparisonResponse struct {
	PortfolioID          uuid.UUID                `json:"portfolio_id"`
	BenchmarkID          uuid.UUID                `json:"benchmark_id"`
	BenchmarkCode        string                   `json:"benchmark_code"`
	BenchmarkName        string                   `json:"benchmark_name"`
	Currency             string                   `json:"currency"`
	StartDate            string                   `json:"start_date"`
	EndDate              string                   `json:"end_date"`
	PortfolioStartValue  decimal.Decimal          `json:"portfolio_start_value"`
	PortfolioEndValue    decimal.Decimal          `json:"portfolio_end_value"`
	BenchmarkStartValue  decimal.Decimal          `json:"benchmark_start_value"`
	BenchmarkEndValue    decimal.Decimal          `json:"benchmark_end_value"`
	PortfolioTotalReturn decimal.Decimal          `json:"portfolio_total_return"`
	BenchmarkTotalReturn decimal.Decimal          `json:"benchmark_total_return"`
	PortfolioCAGR        decimal.Decimal          `json:"portfolio_cagr"`
	BenchmarkCAGR        decimal.Decimal          `json:"benchmark_cagr"`
	ExcessTotalReturn    decimal.Decimal          `json:"excess_total_return"`
	ExcessCAGR           decimal.Decimal          `json:"excess_cagr"`
	ComparisonScope      string                   `json:"comparison_scope"`
	ComparisonMetadata   finance.MetricDefinition `json:"comparison_metadata"`
}

func ToBenchmarkResponse(benchmark Benchmark) BenchmarkResponse {
	return BenchmarkResponse{
		ID:              benchmark.ID,
		Code:            benchmark.Code,
		Name:            benchmark.Name,
		Currency:        benchmark.Currency,
		Source:          benchmark.Source,
		Description:     benchmark.Description,
		CreatedByUserID: benchmark.CreatedByUserID,
		CreatedAt:       benchmark.CreatedAt,
	}
}

func ToBenchmarkObservationResponse(observation BenchmarkObservation) BenchmarkObservationResponse {
	return BenchmarkObservationResponse{
		ID:              observation.ID,
		BenchmarkID:     observation.BenchmarkID,
		ObservationDate: observation.ObservationDate.UTC().Format(benchmarkDateLayout),
		Value:           observation.Value,
		Source:          observation.Source,
		Note:            observation.Note,
		CreatedByUserID: observation.CreatedByUserID,
		CreatedAt:       observation.CreatedAt,
	}
}
