package finance

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type GoalProgressInput struct {
	CurrentValue decimal.Decimal
	TargetValue  decimal.Decimal
	SnapshotDate time.Time
	TargetDate   time.Time
}

type GoalProgressResult struct {
	ProgressPercentage          decimal.Decimal  `json:"progress_percentage"`
	RemainingAmount             decimal.Decimal  `json:"remaining_amount"`
	MonthsRemaining             int              `json:"months_remaining"`
	RequiredMonthlyContribution decimal.Decimal  `json:"required_monthly_contribution"`
	IsTargetReached             bool             `json:"is_target_reached"`
	Definition                  MetricDefinition `json:"definition"`
}

func GoalProgressDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Goal Progress",
		Formula: "Progress percentage = current value / target value x 100. Remaining amount = max(target value - current value, 0). Months remaining = max(whole calendar month boundaries from snapshot date to target date, rounded up when there are remaining days, 0). Required monthly contribution = remaining amount / months remaining when months remain; otherwise 0.",
		Assumptions: []string{
			"Current value and target value use the same currency.",
			"Target value must be greater than zero.",
			"Required monthly contribution is zero on or after the target date because no contribution periods remain.",
			"No investment return, inflation, taxes, fees, or currency conversion is assumed.",
		},
		RequiredInputs: []string{
			"current value",
			"target value",
			"snapshot date",
			"target date",
		},
		Explanation: "Goal progress is a deterministic gap calculation from observed portfolio value to an explicit target. It is not a forecast or recommendation.",
	}
}

func CalculateGoalProgress(input GoalProgressInput) (GoalProgressResult, error) {
	if input.CurrentValue.IsNegative() {
		return GoalProgressResult{}, fmt.Errorf("current value must not be negative")
	}
	if !input.TargetValue.GreaterThan(decimal.Zero) {
		return GoalProgressResult{}, fmt.Errorf("target value must be greater than zero")
	}
	if input.SnapshotDate.IsZero() {
		return GoalProgressResult{}, fmt.Errorf("snapshot date is required")
	}
	if input.TargetDate.IsZero() {
		return GoalProgressResult{}, fmt.Errorf("target date is required")
	}
	remaining := input.TargetValue.Sub(input.CurrentValue)
	if remaining.IsNegative() {
		remaining = decimal.Zero
	}

	monthsRemaining := calendarMonthsRemaining(input.SnapshotDate, input.TargetDate)
	progress := input.CurrentValue.Div(input.TargetValue).Mul(decimal.NewFromInt(100)).Round(performanceResultDecimal)
	requiredMonthlyContribution := decimal.Zero
	if monthsRemaining > 0 {
		requiredMonthlyContribution = remaining.Div(decimal.NewFromInt(int64(monthsRemaining))).Round(performanceResultDecimal)
	}

	return GoalProgressResult{
		ProgressPercentage:          progress,
		RemainingAmount:             remaining,
		MonthsRemaining:             monthsRemaining,
		RequiredMonthlyContribution: requiredMonthlyContribution,
		IsTargetReached:             remaining.IsZero(),
		Definition:                  GoalProgressDefinition(),
	}, nil
}

func calendarMonthsRemaining(snapshotDate time.Time, targetDate time.Time) int {
	snapshot := utcDate(snapshotDate)
	target := utcDate(targetDate)

	months := (target.Year()-snapshot.Year())*12 + int(target.Month()-snapshot.Month())
	if target.Day() > snapshot.Day() {
		months++
	}
	if months < 0 {
		return 0
	}
	return months
}
