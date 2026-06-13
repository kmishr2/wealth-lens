package finance

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

const (
	annualDays               = 365.25
	performanceResultDecimal = 10
)

type CAGRInput struct {
	BeginningValue decimal.Decimal
	EndingValue    decimal.Decimal
	StartDate      time.Time
	EndDate        time.Time
}

type CAGRResult struct {
	Rate       decimal.Decimal  `json:"rate"`
	Definition MetricDefinition `json:"definition"`
}

type PeriodPnLInput struct {
	BeginningValue      decimal.Decimal
	EndingValue         decimal.Decimal
	NetExternalCashFlow decimal.Decimal
}

type PeriodPnLResult struct {
	Amount     decimal.Decimal  `json:"amount"`
	Definition MetricDefinition `json:"definition"`
}

type CashFlow struct {
	Date   time.Time
	Amount decimal.Decimal
}

type XIRRResult struct {
	Rate       decimal.Decimal  `json:"rate"`
	Definition MetricDefinition `json:"definition"`
}

func CAGRDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Compound Annual Growth Rate",
		Formula: "CAGR = ((ending value / beginning value) ^ (365.25 / elapsed days) - 1) x 100.",
		Assumptions: []string{
			"Beginning and ending values must be greater than zero.",
			"Elapsed time is measured in UTC calendar duration and annualized using 365.25 days.",
			"Cash flows during the period are not included; use XIRR for irregular contributions and withdrawals.",
		},
		RequiredInputs: []string{
			"beginning value",
			"ending value",
			"start date",
			"end date",
		},
		Explanation: "CAGR converts total growth between two positive values into an annualized percentage rate. It is deterministic and does not infer deposits, withdrawals, or benchmark assumptions.",
	}
}

func PeriodPnLDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Period Profit and Loss",
		Formula: "Period PnL = ending portfolio value - beginning portfolio value - net external cash flow.",
		Assumptions: []string{
			"Net external cash flow is positive for contributions and negative for withdrawals.",
			"Beginning and ending values use the same currency.",
			"Internal purchases, sales, fees, and taxes are already reflected in portfolio value and are not treated as external cash flows.",
		},
		RequiredInputs: []string{
			"beginning portfolio value",
			"ending portfolio value",
			"net external cash flow",
		},
		Explanation: "Period PnL isolates the change in portfolio value not explained by external contributions or withdrawals. It is calculated separately per currency without foreign exchange conversion.",
	}
}

func XIRRDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Extended Internal Rate of Return",
		Formula: "Find r where sum(cash flow amount / (1 + r) ^ (days since first flow / 365.25)) = 0; returned rate = r x 100.",
		Assumptions: []string{
			"Cash flow signs are explicit: negative values are invested capital and positive values are returned capital or ending value.",
			"Dates are normalized to UTC calendar days and annualized using 365.25 days.",
			"The calculation uses deterministic bisection over rates greater than -100%.",
		},
		RequiredInputs: []string{
			"cash flow amount",
			"cash flow date",
			"at least one positive cash flow",
			"at least one negative cash flow",
		},
		Explanation: "XIRR calculates the annualized money-weighted return for irregular cash flows by solving the net present value equation. It reports an error when the supplied flows do not produce a solvable rate.",
	}
}

func CalculatePeriodPnL(input PeriodPnLInput) (PeriodPnLResult, error) {
	if input.BeginningValue.IsNegative() {
		return PeriodPnLResult{}, fmt.Errorf("beginning value must not be negative")
	}
	if input.EndingValue.IsNegative() {
		return PeriodPnLResult{}, fmt.Errorf("ending value must not be negative")
	}

	return PeriodPnLResult{
		Amount:     input.EndingValue.Sub(input.BeginningValue).Sub(input.NetExternalCashFlow),
		Definition: PeriodPnLDefinition(),
	}, nil
}

func CalculateCAGR(input CAGRInput) (CAGRResult, error) {
	if !input.BeginningValue.GreaterThan(decimal.Zero) {
		return CAGRResult{}, fmt.Errorf("beginning value must be greater than zero")
	}
	if !input.EndingValue.GreaterThan(decimal.Zero) {
		return CAGRResult{}, fmt.Errorf("ending value must be greater than zero")
	}
	if input.StartDate.IsZero() {
		return CAGRResult{}, fmt.Errorf("start date is required")
	}
	if input.EndDate.IsZero() {
		return CAGRResult{}, fmt.Errorf("end date is required")
	}
	if !input.EndDate.After(input.StartDate) {
		return CAGRResult{}, fmt.Errorf("end date must be after start date")
	}

	elapsedDays := input.EndDate.UTC().Sub(input.StartDate.UTC()).Hours() / 24
	if elapsedDays <= 0 {
		return CAGRResult{}, fmt.Errorf("elapsed days must be greater than zero")
	}

	beginning, _ := input.BeginningValue.Float64()
	ending, _ := input.EndingValue.Float64()
	rate := (math.Pow(ending/beginning, annualDays/elapsedDays) - 1) * 100
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return CAGRResult{}, fmt.Errorf("CAGR calculation produced a non-finite result")
	}

	return CAGRResult{
		Rate:       decimal.NewFromFloat(rate).Round(performanceResultDecimal),
		Definition: CAGRDefinition(),
	}, nil
}

func CalculateXIRR(cashFlows []CashFlow) (XIRRResult, error) {
	flows, err := validateAndSortCashFlows(cashFlows)
	if err != nil {
		return XIRRResult{}, err
	}

	low := -0.9999999999
	high := 1.0
	lowNPV := xirrNPV(flows, low)
	highNPV := xirrNPV(flows, high)

	for expansion := 0; sameSign(lowNPV, highNPV) && expansion < 80; expansion++ {
		high *= 2
		highNPV = xirrNPV(flows, high)
		if math.IsInf(high, 0) || math.IsNaN(highNPV) || math.IsInf(highNPV, 0) {
			break
		}
	}

	if sameSign(lowNPV, highNPV) {
		return XIRRResult{}, fmt.Errorf("cash flows do not produce a solvable XIRR")
	}

	var mid float64
	for iteration := 0; iteration < 200; iteration++ {
		mid = (low + high) / 2
		midNPV := xirrNPV(flows, mid)
		if math.Abs(midNPV) < 0.0000001 || math.Abs(high-low) < 0.0000000001 {
			break
		}
		if sameSign(lowNPV, midNPV) {
			low = mid
			lowNPV = midNPV
			continue
		}
		high = mid
	}

	rate := mid * 100
	if math.IsNaN(rate) || math.IsInf(rate, 0) {
		return XIRRResult{}, fmt.Errorf("XIRR calculation produced a non-finite result")
	}

	return XIRRResult{
		Rate:       decimal.NewFromFloat(rate).Round(performanceResultDecimal),
		Definition: XIRRDefinition(),
	}, nil
}

func validateAndSortCashFlows(cashFlows []CashFlow) ([]CashFlow, error) {
	if len(cashFlows) < 2 {
		return nil, fmt.Errorf("XIRR requires at least two cash flows")
	}

	flows := append([]CashFlow(nil), cashFlows...)
	sort.SliceStable(flows, func(i, j int) bool {
		return flows[i].Date.Before(flows[j].Date)
	})

	hasPositive := false
	hasNegative := false
	for index, flow := range flows {
		if flow.Date.IsZero() {
			return nil, fmt.Errorf("cash flow %d requires a date", index)
		}
		if flow.Amount.IsZero() {
			return nil, fmt.Errorf("cash flow %d requires a non-zero amount", index)
		}
		if flow.Amount.GreaterThan(decimal.Zero) {
			hasPositive = true
		}
		if flow.Amount.LessThan(decimal.Zero) {
			hasNegative = true
		}
	}
	if !hasPositive || !hasNegative {
		return nil, fmt.Errorf("XIRR requires at least one positive and one negative cash flow")
	}

	firstDate := utcDate(flows[0].Date)
	lastDate := utcDate(flows[len(flows)-1].Date)
	if !lastDate.After(firstDate) {
		return nil, fmt.Errorf("XIRR requires cash flows on at least two distinct dates")
	}

	return flows, nil
}

func xirrNPV(cashFlows []CashFlow, rate float64) float64 {
	firstDate := utcDate(cashFlows[0].Date)
	npv := 0.0
	for _, flow := range cashFlows {
		amount, _ := flow.Amount.Float64()
		elapsedDays := utcDate(flow.Date).Sub(firstDate).Hours() / 24
		npv += amount / math.Pow(1+rate, elapsedDays/annualDays)
	}
	return npv
}

func sameSign(left float64, right float64) bool {
	return (left < 0 && right < 0) || (left > 0 && right > 0)
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
