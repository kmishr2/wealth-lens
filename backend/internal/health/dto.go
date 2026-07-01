package health

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

type CurrencyConfiguration struct {
	Currency                      string                       `json:"currency"`
	Targets                       []finance.RiskCategoryTarget `json:"targets"`
	VolatilityThresholdPercentage *decimal.Decimal             `json:"volatility_threshold_percentage"`
	DrawdownThresholdPercentage   *decimal.Decimal             `json:"drawdown_threshold_percentage"`
}

type ScoreRequest struct {
	AsOfDate               string                  `json:"as_of_date"`
	RiskProfile            string                  `json:"risk_profile"`
	CurrencyConfigurations []CurrencyConfiguration `json:"currency_configurations"`
}

type ScoreResponse struct {
	PortfolioID    uuid.UUID                   `json:"portfolio_id"`
	StartDate      string                      `json:"start_date"`
	EndDate        string                      `json:"end_date"`
	RiskProfile    string                      `json:"risk_profile"`
	PeriodsPerYear decimal.Decimal             `json:"periods_per_year"`
	Scores         []finance.HealthScoreResult `json:"scores"`
	Scope          string                      `json:"scope"`
}
