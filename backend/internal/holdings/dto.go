package holdings

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

type HoldingsResponse struct {
	PortfolioID    uuid.UUID                `json:"portfolio_id"`
	AssetHoldings  []AssetHoldingResponse   `json:"asset_holdings"`
	CashBalances   []CashBalanceResponse    `json:"cash_balances"`
	MetricMetadata finance.MetricDefinition `json:"metric_metadata"`
}

type AssetHoldingResponse struct {
	AssetID     string          `json:"asset_id"`
	AssetSymbol string          `json:"asset_symbol"`
	AssetName   string          `json:"asset_name"`
	AssetClass  string          `json:"asset_class"`
	Quantity    decimal.Decimal `json:"quantity"`
	Currency    string          `json:"currency"`
}

type CashBalanceResponse struct {
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
}

func ToResponse(portfolioID uuid.UUID, result finance.HoldingsResult) HoldingsResponse {
	assetHoldings := make([]AssetHoldingResponse, 0, len(result.AssetHoldings))
	for _, holding := range result.AssetHoldings {
		assetHoldings = append(assetHoldings, AssetHoldingResponse{
			AssetID:     holding.AssetID,
			AssetSymbol: holding.AssetSymbol,
			AssetName:   holding.AssetName,
			AssetClass:  holding.AssetClass,
			Quantity:    holding.Quantity,
			Currency:    holding.Currency,
		})
	}

	cashBalances := make([]CashBalanceResponse, 0, len(result.CashBalances))
	for _, balance := range result.CashBalances {
		cashBalances = append(cashBalances, CashBalanceResponse{
			Currency: balance.Currency,
			Amount:   balance.Amount,
		})
	}

	return HoldingsResponse{
		PortfolioID:    portfolioID,
		AssetHoldings:  assetHoldings,
		CashBalances:   cashBalances,
		MetricMetadata: result.Definition,
	}
}
