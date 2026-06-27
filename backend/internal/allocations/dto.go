package allocations

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
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

type RebalancingRequest struct {
	Targets                  []finance.AllocationTarget `json:"targets"`
	DriftTolerancePercentage decimal.Decimal            `json:"drift_tolerance_percentage"`
}

type PortfolioRebalancingResponse struct {
	PortfolioID              uuid.UUID                 `json:"portfolio_id"`
	Items                    []finance.RebalancingItem `json:"items"`
	DriftTolerancePercentage decimal.Decimal           `json:"drift_tolerance_percentage"`
	RebalancingScope         string                    `json:"rebalancing_scope"`
	MetricMetadata           finance.MetricDefinition  `json:"metric_metadata"`
	DriftMetadata            finance.MetricDefinition  `json:"drift_metadata"`
	AllocationMetadata       finance.MetricDefinition  `json:"allocation_metadata"`
	ValuationMetadata        finance.MetricDefinition  `json:"valuation_metadata"`
	HoldingsMetadata         finance.MetricDefinition  `json:"holdings_metadata"`
}

func ToRebalancingResponse(portfolioID uuid.UUID, holdings finance.HoldingsResult, valuation finance.PortfolioValuationResult, allocation finance.AllocationResult, rebalancing finance.RebalancingResult) PortfolioRebalancingResponse {
	return PortfolioRebalancingResponse{
		PortfolioID:              portfolioID,
		Items:                    rebalancing.Items,
		DriftTolerancePercentage: rebalancing.DriftTolerancePercentage,
		RebalancingScope:         rebalancing.RebalancingScope,
		MetricMetadata:           rebalancing.Definition,
		DriftMetadata:            rebalancing.DriftDefinition,
		AllocationMetadata:       allocation.Definition,
		ValuationMetadata:        valuation.Definition,
		HoldingsMetadata:         holdings.Definition,
	}
}
