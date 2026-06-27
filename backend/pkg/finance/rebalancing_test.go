package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateRebalancingCalculatesDriftAndAdjustments(t *testing.T) {
	result, err := CalculateRebalancing(RebalancingInput{
		CurrentAllocation: AllocationResult{
			AssetClassAllocations: []AssetClassAllocation{
				rebalancingClass("equity", "USD", "700", "70"),
				rebalancingClass("bond", "USD", "200", "20"),
				rebalancingClass("cash", "USD", "100", "10"),
			},
			CurrencyTotals: []CurrencyValue{
				{Currency: "USD", Amount: decimal.RequireFromString("1000")},
			},
			IsComplete: true,
		},
		Targets: []AllocationTarget{
			rebalancingTarget("equity", "USD", "60"),
			rebalancingTarget("bond", "USD", "30"),
			rebalancingTarget("cash", "USD", "10"),
		},
		DriftTolerancePercentage: decimal.RequireFromString("2"),
	})
	if err != nil {
		t.Fatalf("CalculateRebalancing returned error: %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("items length = %d, want 3", len(result.Items))
	}
	assertRebalancingItem(t, result.Items[0], "bond", "-10", "100", "increase", true)
	assertRebalancingItem(t, result.Items[1], "cash", "0", "0", "none", false)
	assertRebalancingItem(t, result.Items[2], "equity", "10", "-100", "decrease", true)
	if result.Definition.Name == "" || result.DriftDefinition.Name == "" || result.RebalancingScope == "" {
		t.Fatalf("metadata is incomplete: %+v", result)
	}
}

func TestCalculateRebalancingIncludesMissingCurrentAndTargetClasses(t *testing.T) {
	result, err := CalculateRebalancing(RebalancingInput{
		CurrentAllocation: AllocationResult{
			AssetClassAllocations: []AssetClassAllocation{
				rebalancingClass("equity", "USD", "800", "80"),
				rebalancingClass("alternative", "USD", "200", "20"),
			},
			CurrencyTotals: []CurrencyValue{
				{Currency: "USD", Amount: decimal.RequireFromString("1000")},
			},
			IsComplete: true,
		},
		Targets: []AllocationTarget{
			rebalancingTarget("equity", "USD", "70"),
			rebalancingTarget("bond", "USD", "30"),
		},
		DriftTolerancePercentage: decimal.Zero,
	})
	if err != nil {
		t.Fatalf("CalculateRebalancing returned error: %v", err)
	}

	if len(result.Items) != 3 {
		t.Fatalf("items length = %d, want 3", len(result.Items))
	}
	assertRebalancingItem(t, result.Items[0], "alternative", "20", "-200", "decrease", true)
	assertRebalancingItem(t, result.Items[1], "bond", "-30", "300", "increase", true)
	assertRebalancingItem(t, result.Items[2], "equity", "10", "-100", "decrease", true)
}

func TestCalculateRebalancingSuppressesAdjustmentsWithinTolerance(t *testing.T) {
	result, err := CalculateRebalancing(RebalancingInput{
		CurrentAllocation: AllocationResult{
			AssetClassAllocations: []AssetClassAllocation{
				rebalancingClass("equity", "USD", "605", "60.5"),
				rebalancingClass("bond", "USD", "395", "39.5"),
			},
			CurrencyTotals: []CurrencyValue{
				{Currency: "USD", Amount: decimal.RequireFromString("1000")},
			},
			IsComplete: true,
		},
		Targets: []AllocationTarget{
			rebalancingTarget("equity", "USD", "60"),
			rebalancingTarget("bond", "USD", "40"),
		},
		DriftTolerancePercentage: decimal.RequireFromString("1"),
	})
	if err != nil {
		t.Fatalf("CalculateRebalancing returned error: %v", err)
	}

	for _, item := range result.Items {
		if item.IsOutsideTolerance || item.Action != "none" || !item.SuggestedAdjustment.IsZero() {
			t.Fatalf("item = %+v, want no adjustment", item)
		}
	}
}

func TestCalculateRebalancingRejectsInvalidInputs(t *testing.T) {
	validAllocation := AllocationResult{
		AssetClassAllocations: []AssetClassAllocation{
			rebalancingClass("equity", "USD", "1000", "100"),
		},
		CurrencyTotals: []CurrencyValue{
			{Currency: "USD", Amount: decimal.RequireFromString("1000")},
		},
		IsComplete: true,
	}

	tests := []struct {
		name        string
		input       RebalancingInput
		wantMessage string
	}{
		{
			name: "incomplete current allocation",
			input: RebalancingInput{
				CurrentAllocation: AllocationResult{IsComplete: false},
				Targets:           []AllocationTarget{rebalancingTarget("equity", "USD", "100")},
			},
			wantMessage: "complete current allocation",
		},
		{
			name: "negative tolerance",
			input: RebalancingInput{
				CurrentAllocation:        validAllocation,
				Targets:                  []AllocationTarget{rebalancingTarget("equity", "USD", "100")},
				DriftTolerancePercentage: decimal.RequireFromString("-1"),
			},
			wantMessage: "must not be negative",
		},
		{
			name: "targets do not sum to 100",
			input: RebalancingInput{
				CurrentAllocation: validAllocation,
				Targets:           []AllocationTarget{rebalancingTarget("equity", "USD", "90")},
			},
			wantMessage: "sum to 100",
		},
		{
			name: "duplicate target",
			input: RebalancingInput{
				CurrentAllocation: validAllocation,
				Targets: []AllocationTarget{
					rebalancingTarget("equity", "USD", "50"),
					rebalancingTarget("equity", "USD", "50"),
				},
			},
			wantMessage: "duplicate allocation target",
		},
		{
			name: "target currency absent",
			input: RebalancingInput{
				CurrentAllocation: validAllocation,
				Targets:           []AllocationTarget{rebalancingTarget("equity", "INR", "100")},
			},
			wantMessage: "no portfolio total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateRebalancing(tt.input)
			if err == nil {
				t.Fatal("CalculateRebalancing returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func rebalancingClass(assetClass string, currency string, marketValue string, percentage string) AssetClassAllocation {
	return AssetClassAllocation{
		AssetClass:  assetClass,
		Currency:    currency,
		MarketValue: decimal.RequireFromString(marketValue),
		Percentage:  decimal.RequireFromString(percentage),
	}
}

func rebalancingTarget(assetClass string, currency string, percentage string) AllocationTarget {
	return AllocationTarget{
		AssetClass:       assetClass,
		Currency:         currency,
		TargetPercentage: decimal.RequireFromString(percentage),
	}
}

func assertRebalancingItem(t *testing.T, item RebalancingItem, assetClass string, drift string, adjustment string, action string, outsideTolerance bool) {
	t.Helper()

	if item.AssetClass != assetClass {
		t.Fatalf("asset class = %q, want %q", item.AssetClass, assetClass)
	}
	if !item.DriftPercentage.Equal(decimal.RequireFromString(drift)) {
		t.Fatalf("drift = %s, want %s", item.DriftPercentage, drift)
	}
	if !item.SuggestedAdjustment.Equal(decimal.RequireFromString(adjustment)) {
		t.Fatalf("adjustment = %s, want %s", item.SuggestedAdjustment, adjustment)
	}
	if item.Action != action || item.IsOutsideTolerance != outsideTolerance {
		t.Fatalf("action/outside = %s/%v, want %s/%v", item.Action, item.IsOutsideTolerance, action, outsideTolerance)
	}
}
