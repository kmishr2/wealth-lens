package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateCAGRAnnualizesGrowth(t *testing.T) {
	result, err := CalculateCAGR(CAGRInput{
		BeginningValue: decimal.RequireFromString("1000"),
		EndingValue:    decimal.RequireFromString("1210"),
		StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CalculateCAGR returned error: %v", err)
	}

	assertDecimalApprox(t, result.Rate, decimal.RequireFromString("10"), decimal.RequireFromString("0.05"))
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculatePeriodPnLAdjustsForContributionsAndWithdrawals(t *testing.T) {
	result, err := CalculatePeriodPnL(PeriodPnLInput{
		BeginningValue:      decimal.RequireFromString("1000"),
		EndingValue:         decimal.RequireFromString("1300"),
		NetExternalCashFlow: decimal.RequireFromString("50"),
	})
	if err != nil {
		t.Fatalf("CalculatePeriodPnL returned error: %v", err)
	}

	if !result.Amount.Equal(decimal.RequireFromString("250")) {
		t.Fatalf("amount = %s, want 250", result.Amount)
	}
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculatePeriodPnLHandlesLossAndNetWithdrawal(t *testing.T) {
	result, err := CalculatePeriodPnL(PeriodPnLInput{
		BeginningValue:      decimal.RequireFromString("1000"),
		EndingValue:         decimal.RequireFromString("850"),
		NetExternalCashFlow: decimal.RequireFromString("-100"),
	})
	if err != nil {
		t.Fatalf("CalculatePeriodPnL returned error: %v", err)
	}

	if !result.Amount.Equal(decimal.RequireFromString("-50")) {
		t.Fatalf("amount = %s, want -50", result.Amount)
	}
}

func TestCalculatePeriodPnLRejectsNegativePortfolioValues(t *testing.T) {
	tests := []struct {
		name        string
		input       PeriodPnLInput
		wantMessage string
	}{
		{
			name: "negative beginning value",
			input: PeriodPnLInput{
				BeginningValue: decimal.RequireFromString("-1"),
				EndingValue:    decimal.RequireFromString("100"),
			},
			wantMessage: "beginning value",
		},
		{
			name: "negative ending value",
			input: PeriodPnLInput{
				BeginningValue: decimal.RequireFromString("100"),
				EndingValue:    decimal.RequireFromString("-1"),
			},
			wantMessage: "ending value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculatePeriodPnL(tt.input)
			if err == nil {
				t.Fatal("CalculatePeriodPnL returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestCalculateCAGRHandlesNegativeGrowth(t *testing.T) {
	result, err := CalculateCAGR(CAGRInput{
		BeginningValue: decimal.RequireFromString("1000"),
		EndingValue:    decimal.RequireFromString("810"),
		StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("CalculateCAGR returned error: %v", err)
	}

	assertDecimalApprox(t, result.Rate, decimal.RequireFromString("-10"), decimal.RequireFromString("0.05"))
}

func TestCalculateCAGRRejectsInvalidInputs(t *testing.T) {
	validStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validEnd := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		input       CAGRInput
		wantMessage string
	}{
		{
			name: "zero beginning value",
			input: CAGRInput{
				BeginningValue: decimal.Zero,
				EndingValue:    decimal.RequireFromString("100"),
				StartDate:      validStart,
				EndDate:        validEnd,
			},
			wantMessage: "beginning value",
		},
		{
			name: "zero ending value",
			input: CAGRInput{
				BeginningValue: decimal.RequireFromString("100"),
				EndingValue:    decimal.Zero,
				StartDate:      validStart,
				EndDate:        validEnd,
			},
			wantMessage: "ending value",
		},
		{
			name: "missing start date",
			input: CAGRInput{
				BeginningValue: decimal.RequireFromString("100"),
				EndingValue:    decimal.RequireFromString("110"),
				EndDate:        validEnd,
			},
			wantMessage: "start date",
		},
		{
			name: "end before start",
			input: CAGRInput{
				BeginningValue: decimal.RequireFromString("100"),
				EndingValue:    decimal.RequireFromString("110"),
				StartDate:      validEnd,
				EndDate:        validStart,
			},
			wantMessage: "end date must be after start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateCAGR(tt.input)
			if err == nil {
				t.Fatal("CalculateCAGR returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestCalculateBenchmarkComparisonCalculatesExcessReturn(t *testing.T) {
	result, err := CalculateBenchmarkComparison(BenchmarkComparisonInput{
		PortfolioBeginningValue: decimal.RequireFromString("100"),
		PortfolioEndingValue:    decimal.RequireFromString("125"),
		BenchmarkBeginningValue: decimal.RequireFromString("20000"),
		BenchmarkEndingValue:    decimal.RequireFromString("22000"),
		StartDate:               mustPerformanceDate("2025-01-01"),
		EndDate:                 mustPerformanceDate("2026-01-01"),
	})
	if err != nil {
		t.Fatalf("CalculateBenchmarkComparison returned error: %v", err)
	}

	if !result.PortfolioTotalReturn.Equal(decimal.RequireFromString("25")) {
		t.Fatalf("portfolio total return = %s, want 25", result.PortfolioTotalReturn)
	}
	if !result.BenchmarkTotalReturn.Equal(decimal.RequireFromString("10")) {
		t.Fatalf("benchmark total return = %s, want 10", result.BenchmarkTotalReturn)
	}
	if !result.ExcessTotalReturn.Equal(decimal.RequireFromString("15")) {
		t.Fatalf("excess total return = %s, want 15", result.ExcessTotalReturn)
	}
	if result.Definition.Name != "Benchmark Return Comparison" {
		t.Fatalf("definition = %+v", result.Definition)
	}
}

func TestCalculateBenchmarkComparisonRejectsInvalidInputs(t *testing.T) {
	_, err := CalculateBenchmarkComparison(BenchmarkComparisonInput{
		PortfolioBeginningValue: decimal.Zero,
		PortfolioEndingValue:    decimal.RequireFromString("125"),
		BenchmarkBeginningValue: decimal.RequireFromString("20000"),
		BenchmarkEndingValue:    decimal.RequireFromString("22000"),
		StartDate:               mustPerformanceDate("2025-01-01"),
		EndDate:                 mustPerformanceDate("2026-01-01"),
	})
	if err == nil {
		t.Fatal("CalculateBenchmarkComparison returned nil error")
	}
}

func TestCalculateXIRRCalculatesAnnualizedMoneyWeightedReturn(t *testing.T) {
	result, err := CalculateXIRR([]CashFlow{
		performanceCashFlow("2024-01-01", "-1000"),
		performanceCashFlow("2025-01-01", "1100"),
	})
	if err != nil {
		t.Fatalf("CalculateXIRR returned error: %v", err)
	}

	assertDecimalApprox(t, result.Rate, decimal.RequireFromString("10"), decimal.RequireFromString("0.05"))
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculateXIRRHandlesIrregularCashFlows(t *testing.T) {
	result, err := CalculateXIRR([]CashFlow{
		performanceCashFlow("2024-01-01", "-1000"),
		performanceCashFlow("2024-07-01", "-500"),
		performanceCashFlow("2025-01-01", "1700"),
	})
	if err != nil {
		t.Fatalf("CalculateXIRR returned error: %v", err)
	}

	assertDecimalApprox(t, result.Rate, decimal.RequireFromString("16.07"), decimal.RequireFromString("0.05"))
}

func TestCalculateXIRRSortsCashFlowsByDate(t *testing.T) {
	sorted, err := CalculateXIRR([]CashFlow{
		performanceCashFlow("2024-01-01", "-1000"),
		performanceCashFlow("2025-01-01", "1100"),
	})
	if err != nil {
		t.Fatalf("CalculateXIRR sorted returned error: %v", err)
	}

	unsorted, err := CalculateXIRR([]CashFlow{
		performanceCashFlow("2025-01-01", "1100"),
		performanceCashFlow("2024-01-01", "-1000"),
	})
	if err != nil {
		t.Fatalf("CalculateXIRR unsorted returned error: %v", err)
	}

	if !sorted.Rate.Equal(unsorted.Rate) {
		t.Fatalf("sorted rate = %s, unsorted rate = %s", sorted.Rate, unsorted.Rate)
	}
}

func TestCalculateXIRRRejectsInvalidCashFlows(t *testing.T) {
	tests := []struct {
		name        string
		cashFlows   []CashFlow
		wantMessage string
	}{
		{
			name:        "too few flows",
			cashFlows:   []CashFlow{performanceCashFlow("2024-01-01", "-1000")},
			wantMessage: "at least two",
		},
		{
			name: "missing date",
			cashFlows: []CashFlow{
				{Amount: decimal.RequireFromString("-1000")},
				performanceCashFlow("2025-01-01", "1100"),
			},
			wantMessage: "requires a date",
		},
		{
			name: "zero amount",
			cashFlows: []CashFlow{
				performanceCashFlow("2024-01-01", "-1000"),
				performanceCashFlow("2025-01-01", "0"),
			},
			wantMessage: "non-zero",
		},
		{
			name: "no positive flow",
			cashFlows: []CashFlow{
				performanceCashFlow("2024-01-01", "-1000"),
				performanceCashFlow("2025-01-01", "-100"),
			},
			wantMessage: "one positive and one negative",
		},
		{
			name: "same date flows",
			cashFlows: []CashFlow{
				performanceCashFlow("2024-01-01", "-1000"),
				performanceCashFlow("2024-01-01", "1100"),
			},
			wantMessage: "distinct dates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateXIRR(tt.cashFlows)
			if err == nil {
				t.Fatal("CalculateXIRR returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func performanceCashFlow(date string, amount string) CashFlow {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return CashFlow{
		Date:   parsed,
		Amount: decimal.RequireFromString(amount),
	}
}

func mustPerformanceDate(date string) time.Time {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return parsed
}

func assertDecimalApprox(t *testing.T, got decimal.Decimal, want decimal.Decimal, tolerance decimal.Decimal) {
	t.Helper()

	diff := got.Sub(want).Abs()
	if diff.GreaterThan(tolerance) {
		t.Fatalf("value = %s, want %s +/- %s", got, want, tolerance)
	}
}
