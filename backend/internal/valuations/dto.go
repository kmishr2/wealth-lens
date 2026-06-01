package valuations

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
)

type PortfolioValuationResponse struct {
	PortfolioID    uuid.UUID                `json:"portfolio_id"`
	AssetValues    []finance.AssetValuation `json:"asset_values"`
	CashValues     []finance.CashBalance    `json:"cash_values"`
	TotalValues    []finance.CurrencyValue  `json:"total_values"`
	MissingPrices  []finance.MissingPrice   `json:"missing_prices"`
	IsFullyValued  bool                     `json:"is_fully_valued"`
	ValuationScope string                   `json:"valuation_scope"`
	MetricMetadata finance.MetricDefinition `json:"metric_metadata"`
	HoldingsMeta   finance.MetricDefinition `json:"holdings_metadata"`
}

func ToResponse(portfolioID uuid.UUID, holdings finance.HoldingsResult, valuation finance.PortfolioValuationResult) PortfolioValuationResponse {
	return PortfolioValuationResponse{
		PortfolioID:    portfolioID,
		AssetValues:    valuation.AssetValues,
		CashValues:     valuation.CashValues,
		TotalValues:    valuation.TotalValues,
		MissingPrices:  valuation.MissingPrices,
		IsFullyValued:  valuation.IsFullyValued,
		ValuationScope: valuation.ValuationScope,
		MetricMetadata: valuation.Definition,
		HoldingsMeta:   holdings.Definition,
	}
}
