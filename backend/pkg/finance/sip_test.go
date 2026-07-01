package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateSIPProjectionEndOfMonthContributions(t *testing.T) {
	result, err := CalculateSIPProjection(SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1000), MonthlyContribution: decimal.NewFromInt(100),
		AnnualReturnPercentage: decimal.NewFromInt(12), Months: 2})
	if err != nil {
		t.Fatalf("CalculateSIPProjection returned error: %v", err)
	}
	if !result.ProjectedNominalValue.Equal(decimal.RequireFromString("1221.1")) {
		t.Fatalf("nominal = %s, want 1221.1", result.ProjectedNominalValue)
	}
	if !result.TotalContributions.Equal(decimal.NewFromInt(1200)) || !result.NominalInvestmentGrowth.Equal(decimal.RequireFromString("21.1")) {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Schedule) != 2 || result.Definition.Formula == "" {
		t.Fatalf("result metadata = %+v", result)
	}
}

func TestCalculateSIPProjectionInflationAdjustedValue(t *testing.T) {
	result, err := CalculateSIPProjection(SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1200), AnnualReturnPercentage: decimal.Zero,
		AnnualInflationPercentage: decimal.NewFromInt(12), Months: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !result.ProjectedRealValue.Equal(decimal.RequireFromString("1188.1188118812")) {
		t.Fatalf("real value = %s", result.ProjectedRealValue)
	}
}

func TestCalculateSIPProjectionAllowsNegativeScenarioReturn(t *testing.T) {
	result, err := CalculateSIPProjection(SIPProjectionInput{InitialInvestment: decimal.NewFromInt(100), AnnualReturnPercentage: decimal.NewFromInt(-12), Months: 1})
	if err != nil || !result.ProjectedNominalValue.Equal(decimal.NewFromInt(99)) {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
}

func TestCalculateSIPProjectionRejectsInvalidInputs(t *testing.T) {
	tests := []SIPProjectionInput{
		{},
		{InitialInvestment: decimal.NewFromInt(-1), Months: 1},
		{InitialInvestment: decimal.NewFromInt(1), AnnualReturnPercentage: decimal.NewFromInt(-100), Months: 1},
		{InitialInvestment: decimal.NewFromInt(1), AnnualInflationPercentage: decimal.NewFromInt(-1), Months: 1},
		{InitialInvestment: decimal.NewFromInt(1), Months: 1201},
	}
	for _, input := range tests {
		if _, err := CalculateSIPProjection(input); err == nil || strings.TrimSpace(err.Error()) == "" {
			t.Fatalf("input %+v returned error %v", input, err)
		}
	}
}
