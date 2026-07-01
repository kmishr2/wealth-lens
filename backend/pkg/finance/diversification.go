package finance

import "github.com/shopspring/decimal"

type DiversificationAlertResult struct {
	Alerts     []DiversificationAlert `json:"alerts"`
	Definition MetricDefinition       `json:"definition"`
	Scope      string                 `json:"scope"`
}

type DiversificationAlert struct {
	Currency               string          `json:"currency"`
	Severity               string          `json:"severity"`
	Points                 int             `json:"points"`
	LargestAssetID         string          `json:"largest_asset_id"`
	LargestAssetSymbol     string          `json:"largest_asset_symbol"`
	LargestAssetPercentage decimal.Decimal `json:"largest_asset_percentage"`
	HoldingCount           int             `json:"holding_count"`
	Conditions             []string        `json:"conditions"`
}

func DiversificationAlertDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Diversification Alerts",
		Formula: "Diversification points are the lower of two band scores. Largest asset: <=20%=25, >20-35%=18, >35-50%=10, >50%=0. Holding count: >=8=25, 5-7=18, 3-4=10, <3=0. Severity: 25=none, 18=notice, 10=warning, 0=critical.",
		Assumptions: []string{
			"Calculations are performed separately per currency.",
			"Cash is excluded from asset concentration and holding count.",
			"All current asset prices must be available.",
			"Alerts report deterministic observations and are not investment advice.",
		},
		RequiredInputs: []string{"largest asset percentage", "positive holding count", "currency"},
		Explanation:    "The alert severity exposes which approved diversification band the current valued holdings occupy.",
	}
}

func CalculateDiversificationAlerts(concentration ConcentrationResult) DiversificationAlertResult {
	alerts := make([]DiversificationAlert, 0, len(concentration.Currencies))
	for _, metric := range concentration.Currencies {
		assetPoints := diversificationAssetPoints(metric.LargestAssetPercentage)
		holdingPoints := diversificationHoldingPoints(metric.AssetCount)
		points := minInt(assetPoints, holdingPoints)
		conditions := make([]string, 0, 2)
		if assetPoints < 25 {
			conditions = append(conditions, largestAssetCondition(assetPoints))
		}
		if holdingPoints < 25 {
			conditions = append(conditions, holdingCountCondition(holdingPoints))
		}
		alerts = append(alerts, DiversificationAlert{
			Currency: metric.Currency, Severity: diversificationSeverity(points), Points: points,
			LargestAssetID: metric.LargestAssetID, LargestAssetSymbol: metric.LargestAssetSymbol,
			LargestAssetPercentage: metric.LargestAssetPercentage, HoldingCount: metric.AssetCount, Conditions: conditions,
		})
	}
	return DiversificationAlertResult{Alerts: alerts, Definition: DiversificationAlertDefinition(),
		Scope: "Alerts use current valued investment assets separately per currency; cash and currency conversion are excluded."}
}

func diversificationSeverity(points int) string {
	switch points {
	case 25:
		return "none"
	case 18:
		return "notice"
	case 10:
		return "warning"
	default:
		return "critical"
	}
}

func largestAssetCondition(points int) string {
	switch points {
	case 18:
		return "largest asset is greater than 20% and at most 35%"
	case 10:
		return "largest asset is greater than 35% and at most 50%"
	default:
		return "largest asset is greater than 50%"
	}
}

func holdingCountCondition(points int) string {
	switch points {
	case 18:
		return "holding count is between 5 and 7"
	case 10:
		return "holding count is between 3 and 4"
	default:
		return "holding count is fewer than 3"
	}
}
