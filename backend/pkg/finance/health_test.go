package finance

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateHealthScoreUsesConfirmedWeights(t *testing.T) {
	result, err := CalculateHealthScore(HealthScoreInput{
		Currency: "inr", LargestAssetPercentage: decimal.NewFromInt(20), HoldingCount: 8,
		MaximumAbsoluteDriftPercentage: decimal.NewFromInt(5), AnnualizedVolatilityPercentage: decimal.NewFromInt(12),
		VolatilityThresholdPercentage: decimal.NewFromInt(15), MaximumDrawdownPercentage: decimal.NewFromInt(-10),
		DrawdownThresholdPercentage: decimal.NewFromInt(-20), DataQuality: DataQualityComplete,
	})
	if err != nil {
		t.Fatalf("CalculateHealthScore returned error: %v", err)
	}
	if result.Score != 100 || result.Maximum != 100 || result.Currency != "INR" {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Components) != 5 || result.Definition.Formula == "" {
		t.Fatalf("result metadata = %+v", result)
	}
}

func TestCalculateHealthScoreUsesLowerDiversificationBand(t *testing.T) {
	input := validHealthInput()
	input.LargestAssetPercentage = decimal.NewFromInt(30)
	input.HoldingCount = 3
	result, err := CalculateHealthScore(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Components[0].Points != 10 {
		t.Fatalf("diversification points = %d, want 10", result.Components[0].Points)
	}
}

func TestHealthScoreBandBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*HealthScoreInput)
		component int
		want      int
	}{
		{"drift 10", func(i *HealthScoreInput) { i.MaximumAbsoluteDriftPercentage = decimal.NewFromInt(10) }, 1, 20},
		{"drift over 30", func(i *HealthScoreInput) { i.MaximumAbsoluteDriftPercentage = decimal.RequireFromString("30.1") }, 1, 0},
		{"volatility 125", func(i *HealthScoreInput) { i.AnnualizedVolatilityPercentage = decimal.RequireFromString("18.75") }, 2, 10},
		{"volatility over 150", func(i *HealthScoreInput) { i.AnnualizedVolatilityPercentage = decimal.RequireFromString("22.51") }, 2, 0},
		{"drawdown 125", func(i *HealthScoreInput) { i.MaximumDrawdownPercentage = decimal.NewFromInt(-25) }, 3, 8},
		{"drawdown over 150", func(i *HealthScoreInput) { i.MaximumDrawdownPercentage = decimal.RequireFromString("-30.1") }, 3, 0},
		{"minor quality", func(i *HealthScoreInput) { i.DataQuality = DataQualityMinor }, 4, 10},
		{"partial quality", func(i *HealthScoreInput) { i.DataQuality = DataQualityPartial }, 4, 5},
		{"major quality", func(i *HealthScoreInput) { i.DataQuality = DataQualityMajor }, 4, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validHealthInput()
			tt.mutate(&input)
			result, err := CalculateHealthScore(input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Components[tt.component].Points != tt.want {
				t.Fatalf("points = %d, want %d", result.Components[tt.component].Points, tt.want)
			}
		})
	}
}

func TestCalculateHealthScoreRejectsInvalidThresholds(t *testing.T) {
	input := validHealthInput()
	input.VolatilityThresholdPercentage = decimal.Zero
	if _, err := CalculateHealthScore(input); err == nil {
		t.Fatal("expected volatility threshold error")
	}
	input = validHealthInput()
	input.DrawdownThresholdPercentage = decimal.Zero
	if _, err := CalculateHealthScore(input); err == nil {
		t.Fatal("expected drawdown threshold error")
	}
	input = validHealthInput()
	input.DataQuality = "unknown"
	if _, err := CalculateHealthScore(input); err == nil {
		t.Fatal("expected data quality error")
	}
}

func validHealthInput() HealthScoreInput {
	return HealthScoreInput{Currency: "INR", LargestAssetPercentage: decimal.NewFromInt(20), HoldingCount: 8,
		MaximumAbsoluteDriftPercentage: decimal.NewFromInt(5), AnnualizedVolatilityPercentage: decimal.NewFromInt(10),
		VolatilityThresholdPercentage: decimal.NewFromInt(15), MaximumDrawdownPercentage: decimal.NewFromInt(-10),
		DrawdownThresholdPercentage: decimal.NewFromInt(-20), DataQuality: DataQualityComplete}
}

func TestDefaultHealthProfiles(t *testing.T) {
	moderate, err := DefaultHealthProfile("")
	if err != nil {
		t.Fatal(err)
	}
	if moderate.Name != RiskProfileModerate || !moderate.VolatilityThresholdPercentage.Equal(decimal.NewFromInt(15)) || !moderate.DrawdownThresholdPercentage.Equal(decimal.NewFromInt(-20)) {
		t.Fatalf("moderate profile = %+v", moderate)
	}
	if _, err := DefaultHealthProfile("unknown"); err == nil {
		t.Fatal("expected unknown profile error")
	}
}

func TestCalculateMaximumRiskCategoryDrift(t *testing.T) {
	allocation := AllocationResult{IsComplete: true,
		AssetAllocations: []AssetAllocation{
			{Currency: "INR", RiskCategory: "equity", Percentage: decimal.NewFromInt(50)},
			{Currency: "INR", RiskCategory: "debt", Percentage: decimal.NewFromInt(30)},
		},
		CashAllocations: []CashAllocation{{Currency: "INR", Percentage: decimal.NewFromInt(20)}},
	}
	profile, _ := DefaultHealthProfile(RiskProfileModerate)
	drift, unclassified, err := CalculateMaximumRiskCategoryDrift(allocation, "INR", profile.Targets)
	if err != nil {
		t.Fatal(err)
	}
	if !drift.Equal(decimal.NewFromInt(15)) || unclassified {
		t.Fatalf("drift = %s, unclassified = %t", drift, unclassified)
	}

	allocation.AssetAllocations = append(allocation.AssetAllocations, AssetAllocation{Currency: "INR", Percentage: decimal.NewFromInt(5)})
	_, unclassified, err = CalculateMaximumRiskCategoryDrift(allocation, "INR", profile.Targets)
	if err != nil || !unclassified {
		t.Fatalf("err = %v, unclassified = %t", err, unclassified)
	}
}
