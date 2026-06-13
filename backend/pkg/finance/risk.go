package finance

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type PortfolioValuePoint struct {
	Date                time.Time
	Value               decimal.Decimal
	NetExternalCashFlow decimal.Decimal
}

type VolatilityInput struct {
	Values         []PortfolioValuePoint
	PeriodsPerYear decimal.Decimal
}

type VolatilityResult struct {
	AnnualizedVolatility decimal.Decimal  `json:"annualized_volatility"`
	PeriodicReturnCount  int              `json:"periodic_return_count"`
	PeriodsPerYear       decimal.Decimal  `json:"periods_per_year"`
	Definition           MetricDefinition `json:"definition"`
}

type DrawdownResult struct {
	MaximumDrawdown decimal.Decimal  `json:"maximum_drawdown"`
	PeakDate        time.Time        `json:"peak_date"`
	TroughDate      time.Time        `json:"trough_date"`
	Definition      MetricDefinition `json:"definition"`
}

func VolatilityDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Annualized Volatility",
		Formula: "Annualized volatility = sample standard deviation of periodic returns x square root(periods per year) x 100.",
		Assumptions: []string{
			"Periodic return = (current portfolio value - net external cash flow during the period) / previous portfolio value - 1.",
			"Net external cash flow is positive for contributions and negative for withdrawals.",
			"External cash flows are treated as occurring at the end of each observed period.",
			"Portfolio values must be positive and ordered by distinct UTC dates.",
			"Periods per year is an explicit input and is not inferred from missing observations.",
		},
		RequiredInputs: []string{
			"at least three dated portfolio values",
			"net external cash flow per observed period",
			"periods per year",
		},
		Explanation: "Volatility measures the dispersion of observed periodic portfolio returns and annualizes it using an explicit period frequency. It does not predict future risk.",
	}
}

func DrawdownDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Maximum Drawdown",
		Formula: "Cash-flow-adjusted periodic return = (current value - net external cash flow) / previous value - 1. Wealth index compounds these returns from 100. Drawdown = wealth index / prior running peak - 1.",
		Assumptions: []string{
			"Net external cash flow is positive for contributions and negative for withdrawals.",
			"External cash flows are treated as occurring at the end of each observed period.",
			"Portfolio values must be positive and ordered by distinct UTC dates.",
			"Drawdown is reported as a non-positive percentage.",
			"Only observed values are used; no missing dates or values are estimated.",
		},
		RequiredInputs: []string{
			"at least two dated portfolio values",
			"net external cash flow per observed period",
		},
		Explanation: "Maximum drawdown reports the largest observed decline in a normalized portfolio wealth index after removing external contributions and withdrawals.",
	}
}

func CalculateVolatility(input VolatilityInput) (VolatilityResult, error) {
	values, err := validatePortfolioValuePoints(input.Values, 3)
	if err != nil {
		return VolatilityResult{}, err
	}
	if !input.PeriodsPerYear.GreaterThan(decimal.Zero) {
		return VolatilityResult{}, fmt.Errorf("periods per year must be greater than zero")
	}

	returns := make([]float64, 0, len(values)-1)
	for index := 1; index < len(values); index++ {
		periodicReturn, err := cashFlowAdjustedReturn(values[index-1], values[index])
		if err != nil {
			return VolatilityResult{}, err
		}
		value, _ := periodicReturn.Float64()
		returns = append(returns, value)
	}

	mean := 0.0
	for _, periodicReturn := range returns {
		mean += periodicReturn
	}
	mean /= float64(len(returns))

	sumSquaredDeviation := 0.0
	for _, periodicReturn := range returns {
		deviation := periodicReturn - mean
		sumSquaredDeviation += deviation * deviation
	}
	sampleVariance := sumSquaredDeviation / float64(len(returns)-1)
	periodsPerYear, _ := input.PeriodsPerYear.Float64()
	annualized := math.Sqrt(sampleVariance) * math.Sqrt(periodsPerYear) * 100
	if math.IsNaN(annualized) || math.IsInf(annualized, 0) {
		return VolatilityResult{}, fmt.Errorf("volatility calculation produced a non-finite result")
	}

	return VolatilityResult{
		AnnualizedVolatility: decimal.NewFromFloat(annualized).Round(performanceResultDecimal),
		PeriodicReturnCount:  len(returns),
		PeriodsPerYear:       input.PeriodsPerYear,
		Definition:           VolatilityDefinition(),
	}, nil
}

func CalculateMaximumDrawdown(points []PortfolioValuePoint) (DrawdownResult, error) {
	values, err := validatePortfolioValuePoints(points, 2)
	if err != nil {
		return DrawdownResult{}, err
	}

	wealthIndex := decimal.NewFromInt(100)
	peakValue := wealthIndex
	peakDate := values[0].Date
	maximumDrawdown := decimal.Zero
	maximumPeakDate := peakDate
	troughDate := peakDate

	for index := 1; index < len(values); index++ {
		periodicReturn, err := cashFlowAdjustedReturn(values[index-1], values[index])
		if err != nil {
			return DrawdownResult{}, err
		}
		wealthIndex = wealthIndex.Mul(decimal.NewFromInt(1).Add(periodicReturn))

		if wealthIndex.GreaterThan(peakValue) {
			peakValue = wealthIndex
			peakDate = values[index].Date
			continue
		}

		drawdown := wealthIndex.Div(peakValue).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
		if drawdown.LessThan(maximumDrawdown) {
			maximumDrawdown = drawdown
			maximumPeakDate = peakDate
			troughDate = values[index].Date
		}
	}

	return DrawdownResult{
		MaximumDrawdown: maximumDrawdown.Round(performanceResultDecimal),
		PeakDate:        maximumPeakDate,
		TroughDate:      troughDate,
		Definition:      DrawdownDefinition(),
	}, nil
}

func validatePortfolioValuePoints(points []PortfolioValuePoint, minimum int) ([]PortfolioValuePoint, error) {
	if len(points) < minimum {
		return nil, fmt.Errorf("calculation requires at least %d portfolio values", minimum)
	}

	values := append([]PortfolioValuePoint(nil), points...)
	sort.SliceStable(values, func(i, j int) bool {
		return values[i].Date.Before(values[j].Date)
	})

	for index, point := range values {
		if point.Date.IsZero() {
			return nil, fmt.Errorf("portfolio value %d requires a date", index)
		}
		if !point.Value.GreaterThan(decimal.Zero) {
			return nil, fmt.Errorf("portfolio value %d must be greater than zero", index)
		}
		if index > 0 && utcDate(point.Date).Equal(utcDate(values[index-1].Date)) {
			return nil, fmt.Errorf("portfolio values require distinct UTC dates")
		}
	}

	return values, nil
}

func cashFlowAdjustedReturn(previous PortfolioValuePoint, current PortfolioValuePoint) (decimal.Decimal, error) {
	adjustedCurrentValue := current.Value.Sub(current.NetExternalCashFlow)
	if !adjustedCurrentValue.GreaterThan(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("cash-flow-adjusted portfolio value on %s must be greater than zero", utcDate(current.Date).Format("2006-01-02"))
	}
	return adjustedCurrentValue.Div(previous.Value).Sub(decimal.NewFromInt(1)), nil
}
