package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateGoalProgress(t *testing.T) {
	result, err := CalculateGoalProgress(GoalProgressInput{
		CurrentValue: decimal.RequireFromString("250000"),
		TargetValue:  decimal.RequireFromString("1000000"),
		SnapshotDate: goalDate("2026-01-31"),
		TargetDate:   goalDate("2026-12-31"),
	})
	if err != nil {
		t.Fatalf("CalculateGoalProgress returned error: %v", err)
	}

	if !result.ProgressPercentage.Equal(decimal.RequireFromString("25")) {
		t.Fatalf("progress = %s, want 25", result.ProgressPercentage)
	}
	if !result.RemainingAmount.Equal(decimal.RequireFromString("750000")) {
		t.Fatalf("remaining = %s, want 750000", result.RemainingAmount)
	}
	if result.MonthsRemaining != 11 {
		t.Fatalf("months remaining = %d, want 11", result.MonthsRemaining)
	}
	if !result.RequiredMonthlyContribution.Equal(decimal.RequireFromString("68181.8181818182")) {
		t.Fatalf("required monthly = %s", result.RequiredMonthlyContribution)
	}
	if result.IsTargetReached {
		t.Fatal("IsTargetReached = true, want false")
	}
	if result.Definition.Name != "Goal Progress" {
		t.Fatalf("definition = %+v", result.Definition)
	}
}

func TestCalculateGoalProgressHandlesReachedTarget(t *testing.T) {
	result, err := CalculateGoalProgress(GoalProgressInput{
		CurrentValue: decimal.RequireFromString("120"),
		TargetValue:  decimal.RequireFromString("100"),
		SnapshotDate: goalDate("2026-01-31"),
		TargetDate:   goalDate("2026-02-28"),
	})
	if err != nil {
		t.Fatalf("CalculateGoalProgress returned error: %v", err)
	}

	if !result.RemainingAmount.IsZero() || !result.RequiredMonthlyContribution.IsZero() || !result.IsTargetReached {
		t.Fatalf("result = %+v, want reached target with zero remaining", result)
	}
}

func TestCalculateGoalProgressRoundsUpPartialMonths(t *testing.T) {
	result, err := CalculateGoalProgress(GoalProgressInput{
		CurrentValue: decimal.RequireFromString("0"),
		TargetValue:  decimal.RequireFromString("100"),
		SnapshotDate: goalDate("2026-01-31"),
		TargetDate:   goalDate("2026-03-01"),
	})
	if err != nil {
		t.Fatalf("CalculateGoalProgress returned error: %v", err)
	}
	if result.MonthsRemaining != 2 {
		t.Fatalf("months remaining = %d, want 2", result.MonthsRemaining)
	}
}

func TestCalculateGoalProgressRejectsInvalidInputs(t *testing.T) {
	_, err := CalculateGoalProgress(GoalProgressInput{
		CurrentValue: decimal.Zero,
		TargetValue:  decimal.Zero,
		SnapshotDate: goalDate("2026-01-31"),
		TargetDate:   goalDate("2026-12-31"),
	})
	if err == nil || !strings.Contains(err.Error(), "target value") {
		t.Fatalf("error = %v, want target value error", err)
	}

	result, err := CalculateGoalProgress(GoalProgressInput{
		CurrentValue: decimal.Zero,
		TargetValue:  decimal.RequireFromString("100"),
		SnapshotDate: goalDate("2026-01-31"),
		TargetDate:   goalDate("2026-01-31"),
	})
	if err != nil {
		t.Fatalf("CalculateGoalProgress returned error at target date: %v", err)
	}
	if result.MonthsRemaining != 0 || !result.RequiredMonthlyContribution.IsZero() || !result.RemainingAmount.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("result = %+v, want preserved shortfall with no contribution periods", result)
	}
}

func goalDate(date string) time.Time {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return parsed
}
