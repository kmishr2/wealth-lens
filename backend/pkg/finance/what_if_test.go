package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCompareSIPScenariosUsesFirstAsBaseline(t *testing.T) {
	result, err := CompareSIPScenarios([]NamedSIPScenario{
		{Name: "Current", Input: SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1000), MonthlyContribution: decimal.NewFromInt(100), AnnualReturnPercentage: decimal.NewFromInt(12), Months: 2}},
		{Name: "Higher contribution", Input: SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1000), MonthlyContribution: decimal.NewFromInt(200), AnnualReturnPercentage: decimal.NewFromInt(12), Months: 2}},
	})
	if err != nil {
		t.Fatalf("CompareSIPScenarios returned error: %v", err)
	}
	if result.BaselineName != "Current" || result.Months != 2 || len(result.Scenarios) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if !result.Scenarios[0].NominalDifferenceFromBaseline.IsZero() {
		t.Fatalf("baseline difference = %s", result.Scenarios[0].NominalDifferenceFromBaseline)
	}
	if !result.Scenarios[1].NominalDifferenceFromBaseline.Equal(decimal.RequireFromString("201")) ||
		!result.Scenarios[1].ContributionDifferenceFromBaseline.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("comparison = %+v", result.Scenarios[1])
	}
	if result.Definition.Formula == "" {
		t.Fatalf("definition = %+v", result.Definition)
	}
}

func TestCompareSIPScenariosRejectsInvalidSets(t *testing.T) {
	valid := SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1), Months: 12}
	tests := [][]NamedSIPScenario{
		{{Name: "only", Input: valid}},
		{{Name: "same", Input: valid}, {Name: " Same ", Input: valid}},
		{{Name: "one", Input: valid}, {Name: "two", Input: SIPProjectionInput{InitialInvestment: decimal.NewFromInt(1), Months: 24}}},
	}
	for _, scenarios := range tests {
		if _, err := CompareSIPScenarios(scenarios); err == nil || strings.TrimSpace(err.Error()) == "" {
			t.Fatalf("scenarios %+v returned %v", scenarios, err)
		}
	}
}
