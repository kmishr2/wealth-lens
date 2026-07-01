package finance

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	DataQualityComplete     = "complete"
	DataQualityMinor        = "minor_metadata_missing"
	DataQualityPartial      = "partial_history_or_asset_class"
	DataQualityMajor        = "major_data_missing"
	RiskProfileConservative = "conservative"
	RiskProfileModerate     = "moderate"
	RiskProfileAggressive   = "aggressive"
)

type RiskCategoryTarget struct {
	RiskCategory string          `json:"risk_category"`
	Percentage   decimal.Decimal `json:"percentage"`
}

type HealthProfile struct {
	Name                          string               `json:"name"`
	Targets                       []RiskCategoryTarget `json:"targets"`
	VolatilityThresholdPercentage decimal.Decimal      `json:"volatility_threshold_percentage"`
	DrawdownThresholdPercentage   decimal.Decimal      `json:"drawdown_threshold_percentage"`
}

type HealthScoreInput struct {
	Currency                       string
	LargestAssetPercentage         decimal.Decimal
	HoldingCount                   int
	MaximumAbsoluteDriftPercentage decimal.Decimal
	AnnualizedVolatilityPercentage decimal.Decimal
	VolatilityThresholdPercentage  decimal.Decimal
	MaximumDrawdownPercentage      decimal.Decimal
	DrawdownThresholdPercentage    decimal.Decimal
	DataQuality                    string
}

type HealthScoreComponent struct {
	Category    string          `json:"category"`
	Points      int             `json:"points"`
	Maximum     int             `json:"maximum"`
	Observed    decimal.Decimal `json:"observed"`
	Threshold   decimal.Decimal `json:"threshold"`
	Explanation string          `json:"explanation"`
}

type HealthScoreResult struct {
	Currency   string                 `json:"currency"`
	Score      int                    `json:"score"`
	Maximum    int                    `json:"maximum"`
	Components []HealthScoreComponent `json:"components"`
	Definition MetricDefinition       `json:"definition"`
}

func HealthScoreDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Portfolio Health Score",
		Formula: "Total score = diversification points (max 25) + allocation drift points (max 25) + volatility points (max 20) + drawdown points (max 15) + data quality points (max 15).",
		Assumptions: []string{
			"Scores are calculated separately per currency without currency conversion.",
			"Diversification uses the lower of the largest-asset and holding-count band scores.",
			"Allocation drift uses the maximum absolute asset-class percentage-point drift from explicit targets.",
			"Volatility and drawdown are historical metrics compared with explicit user thresholds or disclosed profile defaults.",
			"The score is a transparent monitoring summary, not advice or a prediction.",
		},
		RequiredInputs: []string{
			"largest asset percentage", "holding count", "maximum absolute allocation drift",
			"annualized volatility and threshold", "maximum drawdown and threshold", "data quality level",
		},
		Explanation: "The health score adds five disclosed rule-based component scores. Every observed input, threshold, and awarded point value is returned.",
	}
}

func DefaultHealthProfile(name string) (HealthProfile, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case RiskProfileConservative:
		return newHealthProfile(RiskProfileConservative, 30, 60, 10, 8, -10), nil
	case "", RiskProfileModerate:
		return newHealthProfile(RiskProfileModerate, 60, 35, 5, 15, -20), nil
	case RiskProfileAggressive:
		return newHealthProfile(RiskProfileAggressive, 85, 10, 5, 25, -35), nil
	default:
		return HealthProfile{}, fmt.Errorf("risk profile must be conservative, moderate, or aggressive")
	}
}

func newHealthProfile(name string, equity, debt, cashOther, volatility, drawdown int64) HealthProfile {
	return HealthProfile{Name: name, Targets: []RiskCategoryTarget{
		{RiskCategory: "equity", Percentage: decimal.NewFromInt(equity)},
		{RiskCategory: "debt", Percentage: decimal.NewFromInt(debt)},
		{RiskCategory: "cash_other", Percentage: decimal.NewFromInt(cashOther)},
	}, VolatilityThresholdPercentage: decimal.NewFromInt(volatility), DrawdownThresholdPercentage: decimal.NewFromInt(drawdown)}
}

func CalculateMaximumRiskCategoryDrift(allocation AllocationResult, currency string, targets []RiskCategoryTarget) (decimal.Decimal, bool, error) {
	if !allocation.IsComplete {
		return decimal.Zero, false, fmt.Errorf("risk-category drift requires a complete allocation")
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return decimal.Zero, false, fmt.Errorf("currency must be a three-letter code")
	}
	targetByCategory := make(map[string]decimal.Decimal, 3)
	targetSum := decimal.Zero
	for _, target := range targets {
		category := strings.ToLower(strings.TrimSpace(target.RiskCategory))
		if category != "equity" && category != "debt" && category != "cash_other" {
			return decimal.Zero, false, fmt.Errorf("unsupported risk category %q", category)
		}
		if _, exists := targetByCategory[category]; exists {
			return decimal.Zero, false, fmt.Errorf("duplicate target for risk category %s", category)
		}
		if target.Percentage.IsNegative() || target.Percentage.GreaterThan(decimal.NewFromInt(100)) {
			return decimal.Zero, false, fmt.Errorf("target percentage must be between zero and 100")
		}
		targetByCategory[category] = target.Percentage
		targetSum = targetSum.Add(target.Percentage)
	}
	if len(targetByCategory) != 3 || !targetSum.Equal(decimal.NewFromInt(100)) {
		return decimal.Zero, false, fmt.Errorf("equity, debt, and cash_other targets must be present and sum to 100")
	}

	current := map[string]decimal.Decimal{"equity": decimal.Zero, "debt": decimal.Zero, "cash_other": decimal.Zero}
	unclassified := false
	for _, asset := range allocation.AssetAllocations {
		if asset.Currency != currency {
			continue
		}
		if _, ok := current[asset.RiskCategory]; !ok {
			unclassified = true
			continue
		}
		current[asset.RiskCategory] = current[asset.RiskCategory].Add(asset.Percentage)
	}
	for _, cash := range allocation.CashAllocations {
		if cash.Currency == currency {
			current["cash_other"] = current["cash_other"].Add(cash.Percentage)
		}
	}
	maximum := decimal.Zero
	for category, target := range targetByCategory {
		drift := current[category].Sub(target).Abs()
		if drift.GreaterThan(maximum) {
			maximum = drift
		}
	}
	return maximum, unclassified, nil
}

func CalculateHealthScore(input HealthScoreInput) (HealthScoreResult, error) {
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if len(currency) != 3 {
		return HealthScoreResult{}, fmt.Errorf("currency must be a three-letter code")
	}
	if input.LargestAssetPercentage.IsNegative() || input.LargestAssetPercentage.GreaterThan(decimal.NewFromInt(100)) {
		return HealthScoreResult{}, fmt.Errorf("largest asset percentage must be between zero and 100")
	}
	if input.HoldingCount < 0 {
		return HealthScoreResult{}, fmt.Errorf("holding count must not be negative")
	}
	if input.MaximumAbsoluteDriftPercentage.IsNegative() {
		return HealthScoreResult{}, fmt.Errorf("maximum absolute drift must not be negative")
	}
	if input.AnnualizedVolatilityPercentage.IsNegative() || !input.VolatilityThresholdPercentage.GreaterThan(decimal.Zero) {
		return HealthScoreResult{}, fmt.Errorf("volatility and its threshold must be non-negative and positive respectively")
	}
	if !input.DrawdownThresholdPercentage.Abs().GreaterThan(decimal.Zero) {
		return HealthScoreResult{}, fmt.Errorf("drawdown threshold must not be zero")
	}

	diversification := minInt(diversificationAssetPoints(input.LargestAssetPercentage), diversificationHoldingPoints(input.HoldingCount))
	drift := driftPoints(input.MaximumAbsoluteDriftPercentage)
	volatilityRatio := input.AnnualizedVolatilityPercentage.Div(input.VolatilityThresholdPercentage).Mul(decimal.NewFromInt(100))
	volatility := volatilityPoints(volatilityRatio)
	drawdownRatio := input.MaximumDrawdownPercentage.Abs().Div(input.DrawdownThresholdPercentage.Abs()).Mul(decimal.NewFromInt(100))
	drawdown := drawdownPoints(drawdownRatio)
	quality, err := dataQualityPoints(input.DataQuality)
	if err != nil {
		return HealthScoreResult{}, err
	}

	components := []HealthScoreComponent{
		{Category: "diversification", Points: diversification, Maximum: 25, Observed: input.LargestAssetPercentage, Threshold: decimal.NewFromInt(int64(input.HoldingCount)), Explanation: "Observed is largest asset percentage; threshold field records holding count. The lower band score is used."},
		{Category: "allocation_drift", Points: drift, Maximum: 25, Observed: input.MaximumAbsoluteDriftPercentage, Explanation: "Observed is maximum absolute asset-class drift in percentage points."},
		{Category: "volatility", Points: volatility, Maximum: 20, Observed: input.AnnualizedVolatilityPercentage, Threshold: input.VolatilityThresholdPercentage, Explanation: "Historical annualized volatility is divided by the configured threshold."},
		{Category: "drawdown", Points: drawdown, Maximum: 15, Observed: input.MaximumDrawdownPercentage, Threshold: input.DrawdownThresholdPercentage, Explanation: "Absolute historical maximum drawdown is divided by the absolute configured threshold."},
		{Category: "data_quality", Points: quality, Maximum: 15, Explanation: "Points follow the explicit data quality level: " + input.DataQuality + "."},
	}
	return HealthScoreResult{Currency: currency, Score: diversification + drift + volatility + drawdown + quality, Maximum: 100, Components: components, Definition: HealthScoreDefinition()}, nil
}

func diversificationAssetPoints(value decimal.Decimal) int {
	if value.LessThanOrEqual(decimal.NewFromInt(20)) {
		return 25
	}
	if value.LessThanOrEqual(decimal.NewFromInt(35)) {
		return 18
	}
	if value.LessThanOrEqual(decimal.NewFromInt(50)) {
		return 10
	}
	return 0
}

func diversificationHoldingPoints(count int) int {
	if count >= 8 {
		return 25
	}
	if count >= 5 {
		return 18
	}
	if count >= 3 {
		return 10
	}
	return 0
}

func driftPoints(value decimal.Decimal) int {
	if value.LessThanOrEqual(decimal.NewFromInt(5)) {
		return 25
	}
	if value.LessThanOrEqual(decimal.NewFromInt(10)) {
		return 20
	}
	if value.LessThanOrEqual(decimal.NewFromInt(20)) {
		return 12
	}
	if value.LessThanOrEqual(decimal.NewFromInt(30)) {
		return 6
	}
	return 0
}

func volatilityPoints(ratio decimal.Decimal) int {
	if ratio.LessThanOrEqual(decimal.NewFromInt(80)) {
		return 20
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(100)) {
		return 16
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(125)) {
		return 10
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(150)) {
		return 5
	}
	return 0
}

func drawdownPoints(ratio decimal.Decimal) int {
	if ratio.LessThanOrEqual(decimal.NewFromInt(50)) {
		return 15
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(100)) {
		return 12
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(125)) {
		return 8
	}
	if ratio.LessThanOrEqual(decimal.NewFromInt(150)) {
		return 4
	}
	return 0
}

func dataQualityPoints(level string) (int, error) {
	switch level {
	case DataQualityComplete:
		return 15, nil
	case DataQualityMinor:
		return 10, nil
	case DataQualityPartial:
		return 5, nil
	case DataQualityMajor:
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported data quality level %q", level)
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
