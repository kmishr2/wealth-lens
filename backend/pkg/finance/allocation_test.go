package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateAllocationGroupsByCurrencyAndAssetClass(t *testing.T) {
	valuation := PortfolioValuationResult{
		AssetValues: []AssetValuation{
			allocationAssetValue("asset-b", "VTI", "Vanguard Total Stock Market ETF", "equity", "USD", "150"),
			allocationAssetValue("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "USD", "50"),
		},
		CashValues: []CashBalance{
			valuationCash("USD", "50"),
			valuationCash("INR", "500"),
		},
		TotalValues: []CurrencyValue{
			valuationTotal("USD", "250"),
			valuationTotal("INR", "500"),
		},
		IsFullyValued: true,
	}

	result, err := CalculateAllocation(valuation)
	if err != nil {
		t.Fatalf("CalculateAllocation returned error: %v", err)
	}

	if !result.IsComplete {
		t.Fatal("IsComplete = false, want true")
	}
	if len(result.AssetAllocations) != 2 {
		t.Fatalf("asset allocations length = %d, want 2", len(result.AssetAllocations))
	}
	assertAssetAllocation(t, result.AssetAllocations[0], "asset-a", "BND", "USD", "20")
	assertAssetAllocation(t, result.AssetAllocations[1], "asset-b", "VTI", "USD", "60")

	if len(result.CashAllocations) != 2 {
		t.Fatalf("cash allocations length = %d, want 2", len(result.CashAllocations))
	}
	assertCashAllocation(t, result.CashAllocations[0], "INR", "100")
	assertCashAllocation(t, result.CashAllocations[1], "USD", "20")

	if len(result.AssetClassAllocations) != 4 {
		t.Fatalf("asset class allocations length = %d, want 4", len(result.AssetClassAllocations))
	}
	assertClassAllocation(t, result.AssetClassAllocations[0], "cash", "INR", "100")
	assertClassAllocation(t, result.AssetClassAllocations[1], "bond", "USD", "20")
	assertClassAllocation(t, result.AssetClassAllocations[2], "cash", "USD", "20")
	assertClassAllocation(t, result.AssetClassAllocations[3], "equity", "USD", "60")

	if result.Definition.Name == "" || result.Definition.Formula == "" || result.AllocationScope == "" {
		t.Fatalf("allocation metadata is incomplete: %+v", result)
	}
}

func TestCalculateAllocationCarriesMissingPriceCompleteness(t *testing.T) {
	valuation := PortfolioValuationResult{
		CashValues: []CashBalance{
			valuationCash("USD", "100"),
		},
		TotalValues: []CurrencyValue{
			valuationTotal("USD", "100"),
		},
		MissingPrices: []MissingPrice{
			{AssetID: "asset-a", AssetSymbol: "VTI"},
		},
		IsFullyValued: false,
	}

	result, err := CalculateAllocation(valuation)
	if err != nil {
		t.Fatalf("CalculateAllocation returned error: %v", err)
	}

	if result.IsComplete {
		t.Fatal("IsComplete = true, want false")
	}
	if len(result.MissingPrices) != 1 || result.MissingPrices[0].AssetID != "asset-a" {
		t.Fatalf("missing prices = %+v, want asset-a", result.MissingPrices)
	}
}

func TestCalculateAllocationRejectsInvalidTotals(t *testing.T) {
	tests := []struct {
		name        string
		valuation   PortfolioValuationResult
		wantMessage string
	}{
		{
			name: "zero total",
			valuation: PortfolioValuationResult{
				TotalValues: []CurrencyValue{valuationTotal("USD", "0")},
			},
			wantMessage: "greater than zero",
		},
		{
			name: "missing total",
			valuation: PortfolioValuationResult{
				AssetValues: []AssetValuation{
					allocationAssetValue("asset-a", "VTI", "Vanguard Total Stock Market ETF", "equity", "USD", "100"),
				},
				TotalValues: []CurrencyValue{valuationTotal("INR", "100")},
			},
			wantMessage: "missing total value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateAllocation(tt.valuation)
			if err == nil {
				t.Fatal("CalculateAllocation returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func allocationAssetValue(assetID string, symbol string, name string, assetClass string, currency string, marketValue string) AssetValuation {
	return AssetValuation{
		AssetID:     assetID,
		AssetSymbol: symbol,
		AssetName:   name,
		AssetClass:  assetClass,
		Currency:    currency,
		MarketValue: decimal.RequireFromString(marketValue),
	}
}

func valuationTotal(currency string, amount string) CurrencyValue {
	return CurrencyValue{
		Currency: currency,
		Amount:   decimal.RequireFromString(amount),
	}
}

func assertAssetAllocation(t *testing.T, allocation AssetAllocation, assetID string, symbol string, currency string, percentage string) {
	t.Helper()

	if allocation.AssetID != assetID {
		t.Fatalf("asset id = %q, want %q", allocation.AssetID, assetID)
	}
	if allocation.AssetSymbol != symbol {
		t.Fatalf("symbol = %q, want %q", allocation.AssetSymbol, symbol)
	}
	if allocation.Currency != currency {
		t.Fatalf("currency = %q, want %q", allocation.Currency, currency)
	}
	if !allocation.Percentage.Equal(decimal.RequireFromString(percentage)) {
		t.Fatalf("percentage = %s, want %s", allocation.Percentage, percentage)
	}
}

func assertCashAllocation(t *testing.T, allocation CashAllocation, currency string, percentage string) {
	t.Helper()

	if allocation.Currency != currency {
		t.Fatalf("currency = %q, want %q", allocation.Currency, currency)
	}
	if !allocation.Percentage.Equal(decimal.RequireFromString(percentage)) {
		t.Fatalf("percentage = %s, want %s", allocation.Percentage, percentage)
	}
}

func assertClassAllocation(t *testing.T, allocation AssetClassAllocation, assetClass string, currency string, percentage string) {
	t.Helper()

	if allocation.AssetClass != assetClass {
		t.Fatalf("asset class = %q, want %q", allocation.AssetClass, assetClass)
	}
	if allocation.Currency != currency {
		t.Fatalf("currency = %q, want %q", allocation.Currency, currency)
	}
	if !allocation.Percentage.Equal(decimal.RequireFromString(percentage)) {
		t.Fatalf("percentage = %s, want %s", allocation.Percentage, percentage)
	}
}
