package allocations

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
)

type PortfolioAllocationResponse struct {
	PortfolioID           uuid.UUID                      `json:"portfolio_id"`
	AssetAllocations      []finance.AssetAllocation      `json:"asset_allocations"`
	AssetClassAllocations []finance.AssetClassAllocation `json:"asset_class_allocations"`
	CashAllocations       []finance.CashAllocation       `json:"cash_allocations"`
	CurrencyTotals        []finance.CurrencyValue        `json:"currency_totals"`
	MissingPrices         []finance.MissingPrice         `json:"missing_prices"`
	IsComplete            bool                           `json:"is_complete"`
	AllocationScope       string                         `json:"allocation_scope"`
	MetricMetadata        finance.MetricDefinition       `json:"metric_metadata"`
	ValuationMetadata     finance.MetricDefinition       `json:"valuation_metadata"`
	HoldingsMetadata      finance.MetricDefinition       `json:"holdings_metadata"`
}

func ToResponse(portfolioID uuid.UUID, holdings finance.HoldingsResult, valuation finance.PortfolioValuationResult, allocation finance.AllocationResult) PortfolioAllocationResponse {
	return PortfolioAllocationResponse{
		PortfolioID:           portfolioID,
		AssetAllocations:      allocation.AssetAllocations,
		AssetClassAllocations: allocation.AssetClassAllocations,
		CashAllocations:       allocation.CashAllocations,
		CurrencyTotals:        allocation.CurrencyTotals,
		MissingPrices:         allocation.MissingPrices,
		IsComplete:            allocation.IsComplete,
		AllocationScope:       allocation.AllocationScope,
		MetricMetadata:        allocation.Definition,
		ValuationMetadata:     valuation.Definition,
		HoldingsMetadata:      holdings.Definition,
	}
}
