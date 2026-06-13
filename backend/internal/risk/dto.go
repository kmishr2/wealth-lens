package risk

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const riskDateLayout = "2006-01-02"

type PortfolioRiskResponse struct {
	PortfolioID        uuid.UUID                `json:"portfolio_id"`
	StartDate          string                   `json:"start_date"`
	EndDate            string                   `json:"end_date"`
	PeriodsPerYear     decimal.Decimal          `json:"periods_per_year"`
	CurrencyRisk       []CurrencyRiskResponse   `json:"currency_risk"`
	RiskScope          string                   `json:"risk_scope"`
	VolatilityMetadata finance.MetricDefinition `json:"volatility_metadata"`
	DrawdownMetadata   finance.MetricDefinition `json:"drawdown_metadata"`
}

type CurrencyRiskResponse struct {
	Currency             string          `json:"currency"`
	ObservationCount     int             `json:"observation_count"`
	PeriodicReturnCount  int             `json:"periodic_return_count"`
	AnnualizedVolatility decimal.Decimal `json:"annualized_volatility"`
	MaximumDrawdown      decimal.Decimal `json:"maximum_drawdown"`
	PeakDate             string          `json:"peak_date"`
	TroughDate           string          `json:"trough_date"`
}

func formatRiskDate(date time.Time) string {
	return date.UTC().Format(riskDateLayout)
}
