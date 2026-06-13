package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculateVolatilityAnnualizesSampleStandardDeviation(t *testing.T) {
	result, err := CalculateVolatility(VolatilityInput{
		Values: []PortfolioValuePoint{
			riskValuePoint("2026-01-01", "100"),
			riskValuePoint("2026-01-02", "110"),
			riskValuePoint("2026-01-03", "99"),
		},
		PeriodsPerYear: decimal.RequireFromString("252"),
	})
	if err != nil {
		t.Fatalf("CalculateVolatility returned error: %v", err)
	}

	assertDecimalApprox(t, result.AnnualizedVolatility, decimal.RequireFromString("224.4994"), decimal.RequireFromString("0.001"))
	if result.PeriodicReturnCount != 2 {
		t.Fatalf("periodic return count = %d, want 2", result.PeriodicReturnCount)
	}
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculateVolatilityReturnsZeroForConstantReturns(t *testing.T) {
	result, err := CalculateVolatility(VolatilityInput{
		Values: []PortfolioValuePoint{
			riskValuePoint("2026-01-01", "100"),
			riskValuePoint("2026-01-02", "110"),
			riskValuePoint("2026-01-03", "121"),
		},
		PeriodsPerYear: decimal.RequireFromString("252"),
	})
	if err != nil {
		t.Fatalf("CalculateVolatility returned error: %v", err)
	}

	if !result.AnnualizedVolatility.IsZero() {
		t.Fatalf("annualized volatility = %s, want 0", result.AnnualizedVolatility)
	}
}

func TestRiskCalculationsRemoveExternalContributions(t *testing.T) {
	values := []PortfolioValuePoint{
		riskValuePoint("2026-01-01", "100"),
		{
			Date:                riskDate("2026-01-02"),
			Value:               decimal.RequireFromString("210"),
			NetExternalCashFlow: decimal.RequireFromString("100"),
		},
		{
			Date:  riskDate("2026-01-03"),
			Value: decimal.RequireFromString("231"),
		},
	}

	volatility, err := CalculateVolatility(VolatilityInput{
		Values:         values,
		PeriodsPerYear: decimal.RequireFromString("252"),
	})
	if err != nil {
		t.Fatalf("CalculateVolatility returned error: %v", err)
	}
	if !volatility.AnnualizedVolatility.IsZero() {
		t.Fatalf("annualized volatility = %s, want 0", volatility.AnnualizedVolatility)
	}

	drawdown, err := CalculateMaximumDrawdown(values)
	if err != nil {
		t.Fatalf("CalculateMaximumDrawdown returned error: %v", err)
	}
	if !drawdown.MaximumDrawdown.IsZero() {
		t.Fatalf("maximum drawdown = %s, want 0", drawdown.MaximumDrawdown)
	}
}

func TestCalculateMaximumDrawdownFindsPeakAndTrough(t *testing.T) {
	result, err := CalculateMaximumDrawdown([]PortfolioValuePoint{
		riskValuePoint("2026-01-01", "100"),
		riskValuePoint("2026-01-02", "120"),
		riskValuePoint("2026-01-03", "90"),
		riskValuePoint("2026-01-04", "110"),
		riskValuePoint("2026-01-05", "130"),
	})
	if err != nil {
		t.Fatalf("CalculateMaximumDrawdown returned error: %v", err)
	}

	if !result.MaximumDrawdown.Equal(decimal.RequireFromString("-25")) {
		t.Fatalf("maximum drawdown = %s, want -25", result.MaximumDrawdown)
	}
	if result.PeakDate.Format("2006-01-02") != "2026-01-02" {
		t.Fatalf("peak date = %s, want 2026-01-02", result.PeakDate)
	}
	if result.TroughDate.Format("2006-01-02") != "2026-01-03" {
		t.Fatalf("trough date = %s, want 2026-01-03", result.TroughDate)
	}
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculateMaximumDrawdownReturnsZeroForNewPeaks(t *testing.T) {
	result, err := CalculateMaximumDrawdown([]PortfolioValuePoint{
		riskValuePoint("2026-01-01", "100"),
		riskValuePoint("2026-01-02", "110"),
		riskValuePoint("2026-01-03", "120"),
	})
	if err != nil {
		t.Fatalf("CalculateMaximumDrawdown returned error: %v", err)
	}

	if !result.MaximumDrawdown.IsZero() {
		t.Fatalf("maximum drawdown = %s, want 0", result.MaximumDrawdown)
	}
}

func TestRiskCalculationsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		calculate   func() error
		wantMessage string
	}{
		{
			name: "volatility too few values",
			calculate: func() error {
				_, err := CalculateVolatility(VolatilityInput{
					Values: []PortfolioValuePoint{
						riskValuePoint("2026-01-01", "100"),
						riskValuePoint("2026-01-02", "110"),
					},
					PeriodsPerYear: decimal.RequireFromString("252"),
				})
				return err
			},
			wantMessage: "at least 3",
		},
		{
			name: "invalid annualization",
			calculate: func() error {
				_, err := CalculateVolatility(VolatilityInput{
					Values: []PortfolioValuePoint{
						riskValuePoint("2026-01-01", "100"),
						riskValuePoint("2026-01-02", "110"),
						riskValuePoint("2026-01-03", "120"),
					},
					PeriodsPerYear: decimal.Zero,
				})
				return err
			},
			wantMessage: "periods per year",
		},
		{
			name: "non-positive value",
			calculate: func() error {
				_, err := CalculateMaximumDrawdown([]PortfolioValuePoint{
					riskValuePoint("2026-01-01", "100"),
					riskValuePoint("2026-01-02", "0"),
				})
				return err
			},
			wantMessage: "greater than zero",
		},
		{
			name: "duplicate date",
			calculate: func() error {
				_, err := CalculateMaximumDrawdown([]PortfolioValuePoint{
					riskValuePoint("2026-01-01", "100"),
					riskValuePoint("2026-01-01", "90"),
				})
				return err
			},
			wantMessage: "distinct UTC dates",
		},
		{
			name: "non-positive adjusted value",
			calculate: func() error {
				_, err := CalculateMaximumDrawdown([]PortfolioValuePoint{
					riskValuePoint("2026-01-01", "100"),
					{
						Date:                riskDate("2026-01-02"),
						Value:               decimal.RequireFromString("50"),
						NetExternalCashFlow: decimal.RequireFromString("50"),
					},
				})
				return err
			},
			wantMessage: "cash-flow-adjusted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.calculate()
			if err == nil {
				t.Fatal("calculation returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func riskValuePoint(date string, value string) PortfolioValuePoint {
	return PortfolioValuePoint{
		Date:  riskDate(date),
		Value: decimal.RequireFromString(value),
	}
}

func riskDate(date string) time.Time {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}
	return parsed
}
