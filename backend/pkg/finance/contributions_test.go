package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateContributionAnalysisSeparatesGrowth(t *testing.T) {
	result, err := CalculateContributionAnalysis(ContributionAnalysisInput{
		BeginningValue: decimal.NewFromInt(1000), EndingValue: decimal.NewFromInt(1400),
		StartDate: contributionDate("2026-01-01"), EndDate: contributionDate("2026-03-31"),
		CashFlows: []ExternalContribution{
			{Date: contributionDate("2026-01-15"), Amount: decimal.NewFromInt(300)},
			{Date: contributionDate("2026-01-20"), Amount: decimal.NewFromInt(-50)},
			{Date: contributionDate("2026-03-01"), Amount: decimal.NewFromInt(100)},
		},
	})
	if err != nil {
		t.Fatalf("CalculateContributionAnalysis returned error: %v", err)
	}
	if !result.Contributions.Equal(decimal.NewFromInt(400)) || !result.Withdrawals.Equal(decimal.NewFromInt(50)) ||
		!result.NetContributions.Equal(decimal.NewFromInt(350)) || !result.InvestmentGrowth.Equal(decimal.NewFromInt(50)) {
		t.Fatalf("result = %+v", result)
	}
	if result.EventCount != 3 || len(result.MonthlyBuckets) != 2 || result.MonthlyBuckets[0].Month != "2026-01" {
		t.Fatalf("buckets = %+v", result.MonthlyBuckets)
	}
	if result.Definition.Formula == "" {
		t.Fatalf("definition = %+v", result.Definition)
	}
}

func TestCalculateContributionAnalysisHandlesNoCashFlows(t *testing.T) {
	result, err := CalculateContributionAnalysis(ContributionAnalysisInput{BeginningValue: decimal.NewFromInt(100), EndingValue: decimal.NewFromInt(110),
		StartDate: contributionDate("2026-01-01"), EndDate: contributionDate("2026-01-31")})
	if err != nil || !result.InvestmentGrowth.Equal(decimal.NewFromInt(10)) || result.EventCount != 0 {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
}

func TestCalculateContributionAnalysisRejectsOutOfRangeFlow(t *testing.T) {
	_, err := CalculateContributionAnalysis(ContributionAnalysisInput{StartDate: contributionDate("2026-01-01"), EndDate: contributionDate("2026-01-31"),
		CashFlows: []ExternalContribution{{Date: contributionDate("2026-01-01"), Amount: decimal.NewFromInt(1)}}})
	if err == nil || !strings.Contains(err.Error(), "cash flow") {
		t.Fatalf("error = %v", err)
	}
}

func contributionDate(raw string) time.Time {
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		panic(err)
	}
	return date
}
