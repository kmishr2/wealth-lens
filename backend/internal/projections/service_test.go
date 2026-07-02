package projections

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakePortfolioReader struct{ err error }

func (f *fakePortfolioReader) GetOwned(userID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &portfolios.Portfolio{ID: portfolioID, UserID: userID}, nil
}

func TestCalculateSIPUsesExplicitInputs(t *testing.T) {
	annualReturn := decimal.NewFromInt(12)
	initial := decimal.NewFromInt(1000)
	monthly := decimal.NewFromInt(100)
	response, err := NewService(&fakePortfolioReader{}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{
		Currency: "inr", InitialInvestment: &initial, MonthlyContribution: &monthly, AnnualReturnPercentage: &annualReturn, Months: 2,
	})
	if err != nil {
		t.Fatalf("CalculateSIP returned error: %v", err)
	}
	if response.Currency != "INR" || !response.ProjectedNominalValue.Equal(decimal.RequireFromString("1221.1")) || response.Scope == "" {
		t.Fatalf("response = %+v", response)
	}
}

func TestCalculateSIPRequiresReturnAssumption(t *testing.T) {
	if _, err := NewService(&fakePortfolioReader{}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{Currency: "INR", Months: 12}); err == nil {
		t.Fatal("CalculateSIP error = nil, want missing return error")
	}
}

func TestCalculateSIPRejectsUnownedPortfolio(t *testing.T) {
	annualReturn := decimal.NewFromInt(10)
	if _, err := NewService(&fakePortfolioReader{err: gorm.ErrRecordNotFound}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{Currency: "INR", AnnualReturnPercentage: &annualReturn, Months: 12}); err == nil {
		t.Fatal("CalculateSIP error = nil, want portfolio error")
	}
}

func TestCompareWhatIfReturnsDifferencesFromFirstScenario(t *testing.T) {
	annualReturn := decimal.NewFromInt(12)
	initial := decimal.NewFromInt(1000)
	baselineContribution := decimal.NewFromInt(100)
	higherContribution := decimal.NewFromInt(200)
	response, err := NewService(&fakePortfolioReader{}).CompareWhatIf(uuid.New(), uuid.New(), WhatIfRequest{
		Currency: "INR", Scenarios: []WhatIfScenarioRequest{
			{Name: "Baseline", Input: SIPScenarioInput{InitialInvestment: &initial, MonthlyContribution: &baselineContribution, AnnualReturnPercentage: &annualReturn, Months: 2}},
			{Name: "Higher", Input: SIPScenarioInput{InitialInvestment: &initial, MonthlyContribution: &higherContribution, AnnualReturnPercentage: &annualReturn, Months: 2}},
		},
	})
	if err != nil {
		t.Fatalf("CompareWhatIf returned error: %v", err)
	}
	if response.BaselineName != "Baseline" || len(response.Scenarios) != 2 || response.Currency != "INR" {
		t.Fatalf("response = %+v", response)
	}
	if !response.Scenarios[1].NominalDifferenceFromBaseline.Equal(decimal.NewFromInt(201)) || response.Scope == "" {
		t.Fatalf("comparison = %+v", response)
	}
}

func TestCompareWhatIfRejectsMissingScenarioReturn(t *testing.T) {
	initial := decimal.NewFromInt(1)
	_, err := NewService(&fakePortfolioReader{}).CompareWhatIf(uuid.New(), uuid.New(), WhatIfRequest{Currency: "INR", Scenarios: []WhatIfScenarioRequest{
		{Name: "one", Input: SIPScenarioInput{InitialInvestment: &initial, Months: 12}},
		{Name: "two", Input: SIPScenarioInput{InitialInvestment: &initial, Months: 12}},
	}})
	if err == nil {
		t.Fatal("CompareWhatIf error = nil, want missing return error")
	}
}
