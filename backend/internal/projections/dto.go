package projections

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

type SIPRequest struct {
	Currency                  string           `json:"currency"`
	InitialInvestment         *decimal.Decimal `json:"initial_investment"`
	MonthlyContribution       *decimal.Decimal `json:"monthly_contribution"`
	AnnualReturnPercentage    *decimal.Decimal `json:"annual_return_percentage"`
	AnnualInflationPercentage *decimal.Decimal `json:"annual_inflation_percentage"`
	Months                    int              `json:"months"`
}

type SIPResponse struct {
	PortfolioID uuid.UUID `json:"portfolio_id"`
	Currency    string    `json:"currency"`
	finance.SIPProjectionResult
	Scope string `json:"scope"`
}
