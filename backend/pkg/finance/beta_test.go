package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateBetaWithCashFlowAdjustedReturns(t *testing.T) {
	result, err := CalculateBeta([]BetaObservation{
		{Date: betaDate("2026-01-01"), PortfolioValue: decimal.NewFromInt(100), BenchmarkValue: decimal.NewFromInt(100)},
		{Date: betaDate("2026-01-02"), PortfolioValue: decimal.NewFromInt(120), BenchmarkValue: decimal.NewFromInt(110), NetExternalCashFlow: decimal.NewFromInt(10)},
		{Date: betaDate("2026-01-03"), PortfolioValue: decimal.NewFromInt(96), BenchmarkValue: decimal.RequireFromString("104.5")},
	})
	if err != nil {
		t.Fatalf("CalculateBeta returned error: %v", err)
	}
	if !result.Beta.Equal(decimal.NewFromInt(2)) || result.PairedReturnCount != 2 {
		t.Fatalf("result = %+v, want beta 2", result)
	}
	if result.Definition.Name == "" || result.Definition.Formula == "" {
		t.Fatalf("metadata = %+v", result.Definition)
	}
}

func TestCalculateBetaSortsObservations(t *testing.T) {
	result, err := CalculateBeta([]BetaObservation{
		{Date: betaDate("2026-01-03"), PortfolioValue: decimal.NewFromInt(90), BenchmarkValue: decimal.NewFromInt(90)},
		{Date: betaDate("2026-01-01"), PortfolioValue: decimal.NewFromInt(100), BenchmarkValue: decimal.NewFromInt(100)},
		{Date: betaDate("2026-01-02"), PortfolioValue: decimal.NewFromInt(110), BenchmarkValue: decimal.NewFromInt(110)},
	})
	if err != nil || result.PairedReturnCount != 2 {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
}

func TestCalculateBetaRejectsInsufficientDataAndZeroVariance(t *testing.T) {
	_, err := CalculateBeta([]BetaObservation{{}, {}})
	if err == nil || !strings.Contains(err.Error(), "three aligned") {
		t.Fatalf("error = %v", err)
	}
	_, err = CalculateBeta([]BetaObservation{
		{Date: betaDate("2026-01-01"), PortfolioValue: decimal.NewFromInt(100), BenchmarkValue: decimal.NewFromInt(100)},
		{Date: betaDate("2026-01-02"), PortfolioValue: decimal.NewFromInt(101), BenchmarkValue: decimal.NewFromInt(110)},
		{Date: betaDate("2026-01-03"), PortfolioValue: decimal.NewFromInt(102), BenchmarkValue: decimal.NewFromInt(121)},
	})
	if err == nil || !strings.Contains(err.Error(), "zero variance") {
		t.Fatalf("error = %v", err)
	}
}

func betaDate(raw string) time.Time {
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		panic(err)
	}
	return date
}
