package finance

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type ExternalContribution struct {
	Date   time.Time
	Amount decimal.Decimal
}

type ContributionAnalysisInput struct {
	BeginningValue decimal.Decimal
	EndingValue    decimal.Decimal
	StartDate      time.Time
	EndDate        time.Time
	CashFlows      []ExternalContribution
}

type MonthlyContributionBucket struct {
	Month            string          `json:"month"`
	Contributions    decimal.Decimal `json:"contributions"`
	Withdrawals      decimal.Decimal `json:"withdrawals"`
	NetContributions decimal.Decimal `json:"net_contributions"`
	EventCount       int             `json:"event_count"`
}

type ContributionAnalysisResult struct {
	BeginningValue   decimal.Decimal             `json:"beginning_value"`
	EndingValue      decimal.Decimal             `json:"ending_value"`
	Contributions    decimal.Decimal             `json:"contributions"`
	Withdrawals      decimal.Decimal             `json:"withdrawals"`
	NetContributions decimal.Decimal             `json:"net_contributions"`
	InvestmentGrowth decimal.Decimal             `json:"investment_growth"`
	EventCount       int                         `json:"event_count"`
	MonthlyBuckets   []MonthlyContributionBucket `json:"monthly_buckets"`
	Definition       MetricDefinition            `json:"definition"`
}

func ContributionAnalysisDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Contribution Analysis",
		Formula: "Contributions = sum(positive external cash flows). Withdrawals = absolute sum(negative external cash flows). Net contributions = contributions - withdrawals. Investment growth = ending portfolio value - beginning portfolio value - net contributions.",
		Assumptions: []string{
			"Beginning and ending values are exact immutable daily snapshot values in the same currency.",
			"Only ledger deposit and withdrawal cash flows after the start snapshot and through the end snapshot are external flows.",
			"Monthly buckets use UTC calendar months.",
			"No currency conversion, interpolation, return forecast, or attribution by asset is performed.",
		},
		RequiredInputs: []string{"beginning portfolio value", "ending portfolio value", "start date", "end date", "dated external cash flows"},
		Explanation:    "Contribution analysis separates observed portfolio value change into net money added or removed and the remaining investment growth component.",
	}
}

func CalculateContributionAnalysis(input ContributionAnalysisInput) (ContributionAnalysisResult, error) {
	if input.BeginningValue.IsNegative() || input.EndingValue.IsNegative() {
		return ContributionAnalysisResult{}, fmt.Errorf("beginning and ending values must not be negative")
	}
	if input.StartDate.IsZero() || input.EndDate.IsZero() || !input.EndDate.After(input.StartDate) {
		return ContributionAnalysisResult{}, fmt.Errorf("end date must be after start date")
	}
	contributions := decimal.Zero
	withdrawals := decimal.Zero
	buckets := make(map[string]*MonthlyContributionBucket)
	eventCount := 0
	for index, flow := range input.CashFlows {
		if flow.Date.IsZero() || !flow.Date.After(input.StartDate) || flow.Date.After(endOfFinanceDay(input.EndDate)) {
			return ContributionAnalysisResult{}, fmt.Errorf("cash flow %d must occur after start date and on or before end date", index)
		}
		if flow.Amount.IsZero() {
			continue
		}
		month := flow.Date.UTC().Format("2006-01")
		bucket, ok := buckets[month]
		if !ok {
			bucket = &MonthlyContributionBucket{Month: month}
			buckets[month] = bucket
		}
		if flow.Amount.IsPositive() {
			contributions = contributions.Add(flow.Amount)
			bucket.Contributions = bucket.Contributions.Add(flow.Amount)
		} else {
			amount := flow.Amount.Abs()
			withdrawals = withdrawals.Add(amount)
			bucket.Withdrawals = bucket.Withdrawals.Add(amount)
		}
		bucket.NetContributions = bucket.Contributions.Sub(bucket.Withdrawals)
		bucket.EventCount++
		eventCount++
	}
	orderedBuckets := make([]MonthlyContributionBucket, 0, len(buckets))
	for _, bucket := range buckets {
		orderedBuckets = append(orderedBuckets, *bucket)
	}
	sort.Slice(orderedBuckets, func(i, j int) bool { return orderedBuckets[i].Month < orderedBuckets[j].Month })
	net := contributions.Sub(withdrawals)
	return ContributionAnalysisResult{BeginningValue: input.BeginningValue, EndingValue: input.EndingValue,
		Contributions: contributions, Withdrawals: withdrawals, NetContributions: net,
		InvestmentGrowth: input.EndingValue.Sub(input.BeginningValue).Sub(net), EventCount: eventCount,
		MonthlyBuckets: orderedBuckets, Definition: ContributionAnalysisDefinition()}, nil
}

func endOfFinanceDay(date time.Time) time.Time {
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, 1).Add(-time.Nanosecond)
}
