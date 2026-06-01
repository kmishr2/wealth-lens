package finance

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
)

type AllocationResult struct {
	AssetAllocations      []AssetAllocation      `json:"asset_allocations"`
	AssetClassAllocations []AssetClassAllocation `json:"asset_class_allocations"`
	CashAllocations       []CashAllocation       `json:"cash_allocations"`
	CurrencyTotals        []CurrencyValue        `json:"currency_totals"`
	MissingPrices         []MissingPrice         `json:"missing_prices"`
	Definition            MetricDefinition       `json:"definition"`
	IsComplete            bool                   `json:"is_complete"`
	AllocationScope       string                 `json:"allocation_scope"`
}

type AssetAllocation struct {
	AssetID     string          `json:"asset_id"`
	AssetSymbol string          `json:"asset_symbol"`
	AssetName   string          `json:"asset_name"`
	AssetClass  string          `json:"asset_class"`
	Currency    string          `json:"currency"`
	MarketValue decimal.Decimal `json:"market_value"`
	Percentage  decimal.Decimal `json:"percentage"`
}

type AssetClassAllocation struct {
	AssetClass  string          `json:"asset_class"`
	Currency    string          `json:"currency"`
	MarketValue decimal.Decimal `json:"market_value"`
	Percentage  decimal.Decimal `json:"percentage"`
}

type CashAllocation struct {
	Currency   string          `json:"currency"`
	Amount     decimal.Decimal `json:"amount"`
	Percentage decimal.Decimal `json:"percentage"`
}

func AllocationDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Portfolio Allocation",
		Formula: "Allocation percentage = component value / total portfolio value in the same currency x 100.",
		Assumptions: []string{
			"Portfolio values are grouped by currency because no foreign exchange conversion is performed.",
			"Asset values come from explicit prices and ledger-derived quantities.",
			"Cash allocation uses ledger-derived cash balances.",
			"Missing asset prices make the allocation incomplete and are reported explicitly.",
		},
		RequiredInputs: []string{
			"asset market values",
			"cash balances",
			"total value per currency",
			"asset class",
			"currency",
		},
		Explanation: "Allocation shows each asset, asset class, and cash balance as a percentage of the total valued portfolio in the same currency. It does not estimate missing prices or convert currencies.",
	}
}

func CalculateAllocation(valuation PortfolioValuationResult) (AllocationResult, error) {
	totalByCurrency := make(map[string]decimal.Decimal, len(valuation.TotalValues))
	for _, total := range valuation.TotalValues {
		if !total.Amount.GreaterThan(decimal.Zero) {
			return AllocationResult{}, fmt.Errorf("total value for currency %s must be greater than zero to calculate allocation", total.Currency)
		}
		totalByCurrency[total.Currency] = total.Amount
	}

	assetAllocations := make([]AssetAllocation, 0, len(valuation.AssetValues))
	assetClassTotals := make(map[string]map[string]decimal.Decimal)
	for _, asset := range valuation.AssetValues {
		total, ok := totalByCurrency[asset.Currency]
		if !ok {
			return AllocationResult{}, fmt.Errorf("missing total value for currency %s", asset.Currency)
		}

		assetAllocations = append(assetAllocations, AssetAllocation{
			AssetID:     asset.AssetID,
			AssetSymbol: asset.AssetSymbol,
			AssetName:   asset.AssetName,
			AssetClass:  asset.AssetClass,
			Currency:    asset.Currency,
			MarketValue: asset.MarketValue,
			Percentage:  percentage(asset.MarketValue, total),
		})
		addClassTotal(assetClassTotals, asset.AssetClass, asset.Currency, asset.MarketValue)
	}
	sort.Slice(assetAllocations, func(i, j int) bool {
		if assetAllocations[i].Currency == assetAllocations[j].Currency {
			if assetAllocations[i].AssetSymbol == assetAllocations[j].AssetSymbol {
				return assetAllocations[i].AssetID < assetAllocations[j].AssetID
			}
			return assetAllocations[i].AssetSymbol < assetAllocations[j].AssetSymbol
		}
		return assetAllocations[i].Currency < assetAllocations[j].Currency
	})

	cashAllocations := make([]CashAllocation, 0, len(valuation.CashValues))
	for _, cash := range valuation.CashValues {
		total, ok := totalByCurrency[cash.Currency]
		if !ok {
			return AllocationResult{}, fmt.Errorf("missing total value for currency %s", cash.Currency)
		}
		cashAllocations = append(cashAllocations, CashAllocation{
			Currency:   cash.Currency,
			Amount:     cash.Amount,
			Percentage: percentage(cash.Amount, total),
		})
		addClassTotal(assetClassTotals, "cash", cash.Currency, cash.Amount)
	}
	sort.Slice(cashAllocations, func(i, j int) bool {
		return cashAllocations[i].Currency < cashAllocations[j].Currency
	})

	assetClassAllocations := make([]AssetClassAllocation, 0)
	for assetClass, totalsByCurrency := range assetClassTotals {
		for currency, value := range totalsByCurrency {
			total, ok := totalByCurrency[currency]
			if !ok {
				return AllocationResult{}, fmt.Errorf("missing total value for currency %s", currency)
			}
			assetClassAllocations = append(assetClassAllocations, AssetClassAllocation{
				AssetClass:  assetClass,
				Currency:    currency,
				MarketValue: value,
				Percentage:  percentage(value, total),
			})
		}
	}
	sort.Slice(assetClassAllocations, func(i, j int) bool {
		if assetClassAllocations[i].Currency == assetClassAllocations[j].Currency {
			return assetClassAllocations[i].AssetClass < assetClassAllocations[j].AssetClass
		}
		return assetClassAllocations[i].Currency < assetClassAllocations[j].Currency
	})

	currencyTotals := append([]CurrencyValue(nil), valuation.TotalValues...)
	sort.Slice(currencyTotals, func(i, j int) bool {
		return currencyTotals[i].Currency < currencyTotals[j].Currency
	})

	return AllocationResult{
		AssetAllocations:      assetAllocations,
		AssetClassAllocations: assetClassAllocations,
		CashAllocations:       cashAllocations,
		CurrencyTotals:        currencyTotals,
		MissingPrices:         append([]MissingPrice(nil), valuation.MissingPrices...),
		Definition:            AllocationDefinition(),
		IsComplete:            valuation.IsFullyValued,
		AllocationScope:       "Allocation percentages are calculated separately per currency; no currency conversion is applied.",
	}, nil
}

func addClassTotal(totals map[string]map[string]decimal.Decimal, assetClass string, currency string, value decimal.Decimal) {
	if _, ok := totals[assetClass]; !ok {
		totals[assetClass] = make(map[string]decimal.Decimal)
	}
	totals[assetClass][currency] = totals[assetClass][currency].Add(value)
}

func percentage(value decimal.Decimal, total decimal.Decimal) decimal.Decimal {
	return value.Div(total).Mul(decimal.NewFromInt(100))
}
