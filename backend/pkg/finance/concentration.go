package finance

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"
)

const concentrationPrecision int32 = 10

type ConcentrationResult struct {
	Currencies []CurrencyConcentration `json:"currencies"`
	Definition MetricDefinition        `json:"definition"`
	Scope      string                  `json:"scope"`
}

type CurrencyConcentration struct {
	Currency                 string          `json:"currency"`
	AssetCount               int             `json:"asset_count"`
	InvestedValue            decimal.Decimal `json:"invested_value"`
	HerfindahlHirschmanIndex decimal.Decimal `json:"herfindahl_hirschman_index"`
	EffectiveAssetCount      decimal.Decimal `json:"effective_asset_count"`
	LargestAssetID           string          `json:"largest_asset_id"`
	LargestAssetSymbol       string          `json:"largest_asset_symbol"`
	LargestAssetPercentage   decimal.Decimal `json:"largest_asset_percentage"`
}

func ConcentrationDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Asset Concentration",
		Formula: "For each currency, asset weight = asset market value / total invested asset value. HHI = sum((asset weight x 100)^2). Effective asset count = 10000 / HHI.",
		Assumptions: []string{
			"Calculations are performed separately per currency without foreign exchange conversion.",
			"Only valued investment assets are included; cash is excluded from both numerator and denominator.",
			"All asset prices must be available.",
			"No qualitative concentration labels or advisory thresholds are applied.",
		},
		RequiredInputs: []string{"asset identifier", "asset symbol", "asset market value", "currency"},
		Explanation:    "HHI measures observed position concentration from zero toward 10000. The effective asset count is the number of equally weighted assets that would produce the same HHI.",
	}
}

func CalculateConcentration(allocation AllocationResult) (ConcentrationResult, error) {
	if !allocation.IsComplete {
		return ConcentrationResult{}, fmt.Errorf("concentration requires a complete current allocation")
	}
	assetsByCurrency := make(map[string][]AssetAllocation)
	for _, asset := range allocation.AssetAllocations {
		if asset.Currency == "" {
			return ConcentrationResult{}, fmt.Errorf("asset %s requires a currency", asset.AssetID)
		}
		if asset.AssetID == "" {
			return ConcentrationResult{}, fmt.Errorf("asset identifier is required")
		}
		if asset.MarketValue.IsNegative() {
			return ConcentrationResult{}, fmt.Errorf("market value for asset %s must not be negative", asset.AssetID)
		}
		if asset.MarketValue.IsZero() {
			continue
		}
		assetsByCurrency[asset.Currency] = append(assetsByCurrency[asset.Currency], asset)
	}
	if len(assetsByCurrency) == 0 {
		return ConcentrationResult{}, fmt.Errorf("at least one positive asset market value is required")
	}

	currencies := make([]string, 0, len(assetsByCurrency))
	for currency := range assetsByCurrency {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)

	results := make([]CurrencyConcentration, 0, len(currencies))
	hundred := decimal.NewFromInt(100)
	tenThousand := decimal.NewFromInt(10000)
	for _, currency := range currencies {
		assets := assetsByCurrency[currency]
		total := decimal.Zero
		for _, asset := range assets {
			total = total.Add(asset.MarketValue)
		}

		hhi := decimal.Zero
		largest := assets[0]
		largestPercentage := decimal.Zero
		for _, asset := range assets {
			weightPercentage := asset.MarketValue.Div(total).Mul(hundred)
			hhi = hhi.Add(weightPercentage.Mul(weightPercentage))
			if weightPercentage.GreaterThan(largestPercentage) ||
				(weightPercentage.Equal(largestPercentage) && asset.AssetID < largest.AssetID) {
				largest = asset
				largestPercentage = weightPercentage
			}
		}
		hhi = hhi.Round(concentrationPrecision)
		results = append(results, CurrencyConcentration{
			Currency:                 currency,
			AssetCount:               len(assets),
			InvestedValue:            total,
			HerfindahlHirschmanIndex: hhi,
			EffectiveAssetCount:      tenThousand.Div(hhi).Round(concentrationPrecision),
			LargestAssetID:           largest.AssetID,
			LargestAssetSymbol:       largest.AssetSymbol,
			LargestAssetPercentage:   largestPercentage.Round(concentrationPrecision),
		})
	}

	return ConcentrationResult{
		Currencies: results,
		Definition: ConcentrationDefinition(),
		Scope:      "Concentration is calculated from valued investment assets separately per currency; cash and currency conversion are excluded.",
	}, nil
}
