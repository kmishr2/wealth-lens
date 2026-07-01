package beta

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

type Response struct {
	PortfolioID       uuid.UUID                `json:"portfolio_id"`
	BenchmarkID       uuid.UUID                `json:"benchmark_id"`
	BenchmarkCode     string                   `json:"benchmark_code"`
	Currency          string                   `json:"currency"`
	StartDate         string                   `json:"start_date"`
	EndDate           string                   `json:"end_date"`
	AlignedCount      int                      `json:"aligned_observation_count"`
	PairedReturnCount int                      `json:"paired_return_count"`
	Beta              decimal.Decimal          `json:"beta"`
	Scope             string                   `json:"scope"`
	MetricMetadata    finance.MetricDefinition `json:"metric_metadata"`
}
