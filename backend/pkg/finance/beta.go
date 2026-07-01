package finance

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

type BetaObservation struct {
	Date                time.Time
	PortfolioValue      decimal.Decimal
	BenchmarkValue      decimal.Decimal
	NetExternalCashFlow decimal.Decimal
}

type BetaResult struct {
	Beta              decimal.Decimal  `json:"beta"`
	PairedReturnCount int              `json:"paired_return_count"`
	Definition        MetricDefinition `json:"definition"`
}

func BetaDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Portfolio Beta",
		Formula: "Portfolio return = (ending portfolio value - net external cash flow) / beginning portfolio value - 1. Benchmark return = ending benchmark value / beginning benchmark value - 1. Beta = covariance(portfolio returns, benchmark returns) / variance(benchmark returns).",
		Assumptions: []string{
			"Portfolio and benchmark values are aligned on exact observation dates.",
			"Portfolio returns remove external contributions and withdrawals for each interval.",
			"Values use the same currency and no foreign exchange conversion or interpolation is performed.",
			"Beta describes historical co-movement and is not predictive.",
		},
		RequiredInputs: []string{"dated portfolio values", "dated benchmark values", "net external cash flows by interval"},
		Explanation:    "Beta measures historical portfolio return sensitivity to observed benchmark returns over aligned intervals.",
	}
}

func CalculateBeta(observations []BetaObservation) (BetaResult, error) {
	if len(observations) < 3 {
		return BetaResult{}, fmt.Errorf("beta requires at least three aligned observations")
	}
	ordered := append([]BetaObservation(nil), observations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Date.Before(ordered[j].Date) })
	portfolioReturns := make([]decimal.Decimal, 0, len(ordered)-1)
	benchmarkReturns := make([]decimal.Decimal, 0, len(ordered)-1)
	for index := 1; index < len(ordered); index++ {
		previous, current := ordered[index-1], ordered[index]
		if previous.Date.IsZero() || current.Date.IsZero() || !current.Date.After(previous.Date) {
			return BetaResult{}, fmt.Errorf("beta observations require distinct increasing dates")
		}
		if !previous.PortfolioValue.GreaterThan(decimal.Zero) || !current.PortfolioValue.GreaterThan(decimal.Zero) {
			return BetaResult{}, fmt.Errorf("portfolio values must be greater than zero")
		}
		if !previous.BenchmarkValue.GreaterThan(decimal.Zero) || !current.BenchmarkValue.GreaterThan(decimal.Zero) {
			return BetaResult{}, fmt.Errorf("benchmark values must be greater than zero")
		}
		portfolioReturns = append(portfolioReturns, current.PortfolioValue.Sub(current.NetExternalCashFlow).Div(previous.PortfolioValue).Sub(decimal.NewFromInt(1)))
		benchmarkReturns = append(benchmarkReturns, current.BenchmarkValue.Div(previous.BenchmarkValue).Sub(decimal.NewFromInt(1)))
	}

	portfolioMean := decimalMean(portfolioReturns)
	benchmarkMean := decimalMean(benchmarkReturns)
	covarianceNumerator := decimal.Zero
	varianceNumerator := decimal.Zero
	for index := range portfolioReturns {
		portfolioDeviation := portfolioReturns[index].Sub(portfolioMean)
		benchmarkDeviation := benchmarkReturns[index].Sub(benchmarkMean)
		covarianceNumerator = covarianceNumerator.Add(portfolioDeviation.Mul(benchmarkDeviation))
		varianceNumerator = varianceNumerator.Add(benchmarkDeviation.Mul(benchmarkDeviation))
	}
	if varianceNumerator.IsZero() {
		return BetaResult{}, fmt.Errorf("benchmark returns have zero variance")
	}
	return BetaResult{Beta: covarianceNumerator.Div(varianceNumerator).Round(performanceResultDecimal), PairedReturnCount: len(portfolioReturns), Definition: BetaDefinition()}, nil
}

func decimalMean(values []decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, value := range values {
		total = total.Add(value)
	}
	return total.Div(decimal.NewFromInt(int64(len(values))))
}
