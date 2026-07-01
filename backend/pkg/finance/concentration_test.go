package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateConcentrationByCurrency(t *testing.T) {
	result, err := CalculateConcentration(AllocationResult{IsComplete: true, AssetAllocations: []AssetAllocation{
		{AssetID: "b", AssetSymbol: "B", Currency: "INR", MarketValue: decimal.NewFromInt(25)},
		{AssetID: "a", AssetSymbol: "A", Currency: "INR", MarketValue: decimal.NewFromInt(75)},
		{AssetID: "c", AssetSymbol: "C", Currency: "USD", MarketValue: decimal.NewFromInt(50)},
	}})
	if err != nil {
		t.Fatalf("CalculateConcentration returned error: %v", err)
	}
	if len(result.Currencies) != 2 || result.Currencies[0].Currency != "INR" {
		t.Fatalf("currencies = %+v", result.Currencies)
	}
	inr := result.Currencies[0]
	if !inr.HerfindahlHirschmanIndex.Equal(decimal.NewFromInt(6250)) || !inr.EffectiveAssetCount.Equal(decimal.RequireFromString("1.6")) {
		t.Fatalf("INR metric = %+v", inr)
	}
	if inr.LargestAssetID != "a" || !inr.LargestAssetPercentage.Equal(decimal.NewFromInt(75)) {
		t.Fatalf("INR largest asset = %+v", inr)
	}
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Scope == "" {
		t.Fatalf("metadata incomplete: %+v", result)
	}
}

func TestCalculateConcentrationEqualWeights(t *testing.T) {
	result, err := CalculateConcentration(AllocationResult{IsComplete: true, AssetAllocations: []AssetAllocation{
		{AssetID: "a", Currency: "INR", MarketValue: decimal.NewFromInt(1)},
		{AssetID: "b", Currency: "INR", MarketValue: decimal.NewFromInt(1)},
		{AssetID: "c", Currency: "INR", MarketValue: decimal.NewFromInt(1)},
		{AssetID: "d", Currency: "INR", MarketValue: decimal.NewFromInt(1)},
	}})
	if err != nil {
		t.Fatalf("CalculateConcentration returned error: %v", err)
	}
	metric := result.Currencies[0]
	if !metric.HerfindahlHirschmanIndex.Equal(decimal.NewFromInt(2500)) || !metric.EffectiveAssetCount.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("metric = %+v", metric)
	}
}

func TestCalculateConcentrationRejectsIncompleteOrEmptyAllocation(t *testing.T) {
	_, err := CalculateConcentration(AllocationResult{IsComplete: false})
	if err == nil || !strings.Contains(err.Error(), "complete") {
		t.Fatalf("error = %v, want completeness error", err)
	}
	_, err = CalculateConcentration(AllocationResult{IsComplete: true})
	if err == nil || !strings.Contains(err.Error(), "positive asset") {
		t.Fatalf("error = %v, want positive asset error", err)
	}
}

func TestCalculateConcentrationRejectsNegativeMarketValue(t *testing.T) {
	_, err := CalculateConcentration(AllocationResult{IsComplete: true, AssetAllocations: []AssetAllocation{
		{AssetID: "a", Currency: "INR", MarketValue: decimal.NewFromInt(-1)},
	}})
	if err == nil || !strings.Contains(err.Error(), "must not be negative") {
		t.Fatalf("error = %v, want negative value error", err)
	}
}
