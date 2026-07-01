package finance

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const sipPrecision int32 = 10

type SIPProjectionInput struct {
	InitialInvestment         decimal.Decimal
	MonthlyContribution       decimal.Decimal
	AnnualReturnPercentage    decimal.Decimal
	AnnualInflationPercentage decimal.Decimal
	Months                    int
}

type SIPProjectionPoint struct {
	Month                 int             `json:"month"`
	TotalContributions    decimal.Decimal `json:"total_contributions"`
	ProjectedNominalValue decimal.Decimal `json:"projected_nominal_value"`
	ProjectedRealValue    decimal.Decimal `json:"projected_real_value"`
}

type SIPProjectionResult struct {
	InitialInvestment         decimal.Decimal      `json:"initial_investment"`
	MonthlyContribution       decimal.Decimal      `json:"monthly_contribution"`
	AnnualReturnPercentage    decimal.Decimal      `json:"annual_return_percentage"`
	AnnualInflationPercentage decimal.Decimal      `json:"annual_inflation_percentage"`
	Months                    int                  `json:"months"`
	TotalContributions        decimal.Decimal      `json:"total_contributions"`
	ProjectedNominalValue     decimal.Decimal      `json:"projected_nominal_value"`
	ProjectedRealValue        decimal.Decimal      `json:"projected_real_value"`
	NominalInvestmentGrowth   decimal.Decimal      `json:"nominal_investment_growth"`
	Schedule                  []SIPProjectionPoint `json:"schedule"`
	Definition                MetricDefinition     `json:"definition"`
}

func SIPProjectionDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "SIP Projection",
		Formula: "Monthly rate = annual percentage / 12 / 100. Each month: nominal balance = prior balance x (1 + monthly return) + end-of-month contribution. Real value = nominal balance / (1 + monthly inflation rate)^elapsed months.",
		Assumptions: []string{
			"The annual return and inflation rates are explicit user inputs and remain constant for the simulation.",
			"Contributions occur at the end of each month.",
			"Return and inflation compound monthly using nominal annual rate divided by 12.",
			"Taxes, fees, missed contributions, and currency conversion are excluded.",
			"Projected values are scenarios, not predictions or advice.",
		},
		RequiredInputs: []string{"initial investment", "monthly contribution", "annual return percentage", "annual inflation percentage", "number of months"},
		Explanation:    "This deterministic scenario applies the same disclosed monthly rates and contribution amount for every simulated month.",
	}
}

func CalculateSIPProjection(input SIPProjectionInput) (SIPProjectionResult, error) {
	if input.InitialInvestment.IsNegative() {
		return SIPProjectionResult{}, fmt.Errorf("initial investment must not be negative")
	}
	if input.MonthlyContribution.IsNegative() {
		return SIPProjectionResult{}, fmt.Errorf("monthly contribution must not be negative")
	}
	if input.InitialInvestment.IsZero() && input.MonthlyContribution.IsZero() {
		return SIPProjectionResult{}, fmt.Errorf("initial investment or monthly contribution must be greater than zero")
	}
	if input.AnnualReturnPercentage.LessThanOrEqual(decimal.NewFromInt(-100)) {
		return SIPProjectionResult{}, fmt.Errorf("annual return percentage must be greater than -100")
	}
	if input.AnnualInflationPercentage.IsNegative() {
		return SIPProjectionResult{}, fmt.Errorf("annual inflation percentage must not be negative")
	}
	if input.Months <= 0 || input.Months > 1200 {
		return SIPProjectionResult{}, fmt.Errorf("months must be between 1 and 1200")
	}

	monthlyReturn := input.AnnualReturnPercentage.Div(decimal.NewFromInt(1200))
	monthlyInflation := input.AnnualInflationPercentage.Div(decimal.NewFromInt(1200))
	nominal := input.InitialInvestment
	inflationFactor := decimal.NewFromInt(1)
	totalContributions := input.InitialInvestment
	schedule := make([]SIPProjectionPoint, 0, input.Months)
	for month := 1; month <= input.Months; month++ {
		nominal = nominal.Mul(decimal.NewFromInt(1).Add(monthlyReturn)).Add(input.MonthlyContribution)
		inflationFactor = inflationFactor.Mul(decimal.NewFromInt(1).Add(monthlyInflation))
		totalContributions = totalContributions.Add(input.MonthlyContribution)
		schedule = append(schedule, SIPProjectionPoint{Month: month, TotalContributions: totalContributions.Round(sipPrecision),
			ProjectedNominalValue: nominal.Round(sipPrecision), ProjectedRealValue: nominal.Div(inflationFactor).Round(sipPrecision)})
	}
	nominal = nominal.Round(sipPrecision)
	return SIPProjectionResult{InitialInvestment: input.InitialInvestment, MonthlyContribution: input.MonthlyContribution,
		AnnualReturnPercentage: input.AnnualReturnPercentage, AnnualInflationPercentage: input.AnnualInflationPercentage,
		Months: input.Months, TotalContributions: totalContributions.Round(sipPrecision), ProjectedNominalValue: nominal,
		ProjectedRealValue: nominal.Div(inflationFactor).Round(sipPrecision), NominalInvestmentGrowth: nominal.Sub(totalContributions).Round(sipPrecision),
		Schedule: schedule, Definition: SIPProjectionDefinition()}, nil
}
