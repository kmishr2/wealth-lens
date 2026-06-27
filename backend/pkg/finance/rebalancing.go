package finance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type AllocationTarget struct {
	AssetClass       string          `json:"asset_class"`
	Currency         string          `json:"currency"`
	TargetPercentage decimal.Decimal `json:"target_percentage"`
}

type RebalancingInput struct {
	CurrentAllocation        AllocationResult
	Targets                  []AllocationTarget
	DriftTolerancePercentage decimal.Decimal
}

type RebalancingResult struct {
	Items                    []RebalancingItem `json:"items"`
	DriftTolerancePercentage decimal.Decimal   `json:"drift_tolerance_percentage"`
	Definition               MetricDefinition  `json:"definition"`
	DriftDefinition          MetricDefinition  `json:"drift_definition"`
	RebalancingScope         string            `json:"rebalancing_scope"`
}

type RebalancingItem struct {
	AssetClass          string          `json:"asset_class"`
	Currency            string          `json:"currency"`
	CurrentValue        decimal.Decimal `json:"current_value"`
	CurrentPercentage   decimal.Decimal `json:"current_percentage"`
	TargetValue         decimal.Decimal `json:"target_value"`
	TargetPercentage    decimal.Decimal `json:"target_percentage"`
	DriftPercentage     decimal.Decimal `json:"drift_percentage"`
	AbsoluteDrift       decimal.Decimal `json:"absolute_drift"`
	IsOutsideTolerance  bool            `json:"is_outside_tolerance"`
	SuggestedAdjustment decimal.Decimal `json:"suggested_adjustment"`
	Action              string          `json:"action"`
}

func AllocationDriftDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Allocation Drift",
		Formula: "Allocation drift = current allocation percentage - target allocation percentage.",
		Assumptions: []string{
			"Current and target allocations use the same asset class and currency.",
			"Targets sum to exactly 100 percent within each currency.",
			"Positive drift is overweight and negative drift is underweight.",
		},
		RequiredInputs: []string{
			"current allocation percentage",
			"target allocation percentage",
			"currency",
			"asset class",
		},
		Explanation: "Allocation drift measures the signed percentage-point difference between current and explicitly supplied target allocation.",
	}
}

func RebalancingDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Deterministic Rebalancing",
		Formula: "Target value = total portfolio value in currency x target percentage / 100. Suggested adjustment = target value - current value.",
		Assumptions: []string{
			"Calculations are performed separately per currency without foreign exchange conversion.",
			"Targets are explicit user inputs and sum to 100 percent per currency.",
			"An adjustment is surfaced only when absolute percentage drift exceeds the explicit tolerance.",
			"Positive adjustment means increase exposure; negative adjustment means decrease exposure.",
		},
		RequiredInputs: []string{
			"current asset-class allocation",
			"total portfolio value per currency",
			"target allocation percentage",
			"drift tolerance percentage",
		},
		Explanation: "Rebalancing calculates transparent asset-class value differences from explicit targets. It does not execute trades or account for taxes, fees, liquidity, or suitability.",
	}
}

func CalculateRebalancing(input RebalancingInput) (RebalancingResult, error) {
	if !input.CurrentAllocation.IsComplete {
		return RebalancingResult{}, fmt.Errorf("rebalancing requires a complete current allocation")
	}
	if input.DriftTolerancePercentage.IsNegative() {
		return RebalancingResult{}, fmt.Errorf("drift tolerance percentage must not be negative")
	}

	totals, err := rebalancingTotals(input.CurrentAllocation.CurrencyTotals)
	if err != nil {
		return RebalancingResult{}, err
	}
	targets, err := validateAllocationTargets(input.Targets, totals)
	if err != nil {
		return RebalancingResult{}, err
	}

	current := currentClassAllocations(input.CurrentAllocation.AssetClassAllocations)
	keys := make(map[string]allocationClassKey, len(current)+len(targets))
	for key, value := range current {
		keys[key] = allocationClassKey{AssetClass: value.AssetClass, Currency: value.Currency}
	}
	for key, value := range targets {
		keys[key] = allocationClassKey{AssetClass: value.AssetClass, Currency: value.Currency}
	}

	orderedKeys := make([]allocationClassKey, 0, len(keys))
	for _, key := range keys {
		orderedKeys = append(orderedKeys, key)
	}
	sort.Slice(orderedKeys, func(i, j int) bool {
		if orderedKeys[i].Currency == orderedKeys[j].Currency {
			return orderedKeys[i].AssetClass < orderedKeys[j].AssetClass
		}
		return orderedKeys[i].Currency < orderedKeys[j].Currency
	})

	items := make([]RebalancingItem, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		mapKey := allocationKey(key.AssetClass, key.Currency)
		currentAllocation := current[mapKey]
		target := targets[mapKey]
		total := totals[key.Currency]
		targetValue := total.Mul(target.TargetPercentage).Div(decimal.NewFromInt(100))
		drift := currentAllocation.Percentage.Sub(target.TargetPercentage)
		absoluteDrift := drift.Abs()
		outsideTolerance := absoluteDrift.GreaterThan(input.DriftTolerancePercentage)
		adjustment := targetValue.Sub(currentAllocation.MarketValue)
		action := "none"
		if outsideTolerance && adjustment.GreaterThan(decimal.Zero) {
			action = "increase"
		}
		if outsideTolerance && adjustment.LessThan(decimal.Zero) {
			action = "decrease"
		}
		if !outsideTolerance {
			adjustment = decimal.Zero
		}

		items = append(items, RebalancingItem{
			AssetClass:          key.AssetClass,
			Currency:            key.Currency,
			CurrentValue:        currentAllocation.MarketValue,
			CurrentPercentage:   currentAllocation.Percentage,
			TargetValue:         targetValue,
			TargetPercentage:    target.TargetPercentage,
			DriftPercentage:     drift,
			AbsoluteDrift:       absoluteDrift,
			IsOutsideTolerance:  outsideTolerance,
			SuggestedAdjustment: adjustment,
			Action:              action,
		})
	}

	return RebalancingResult{
		Items:                    items,
		DriftTolerancePercentage: input.DriftTolerancePercentage,
		Definition:               RebalancingDefinition(),
		DriftDefinition:          AllocationDriftDefinition(),
		RebalancingScope:         "Asset-class adjustments are calculated separately per currency from explicit targets; no trades are executed and no currency conversion is applied.",
	}, nil
}

type allocationClassKey struct {
	AssetClass string
	Currency   string
}

func rebalancingTotals(currencyTotals []CurrencyValue) (map[string]decimal.Decimal, error) {
	totals := make(map[string]decimal.Decimal, len(currencyTotals))
	for _, total := range currencyTotals {
		currency, ok := normalizeCurrency(total.Currency)
		if !ok {
			return nil, fmt.Errorf("currency total requires a three-letter uppercase currency code")
		}
		if !total.Amount.GreaterThan(decimal.Zero) {
			return nil, fmt.Errorf("total value for currency %s must be greater than zero", currency)
		}
		if _, exists := totals[currency]; exists {
			return nil, fmt.Errorf("duplicate total value for currency %s", currency)
		}
		totals[currency] = total.Amount
	}
	if len(totals) == 0 {
		return nil, fmt.Errorf("at least one currency total is required")
	}
	return totals, nil
}

func validateAllocationTargets(targets []AllocationTarget, totals map[string]decimal.Decimal) (map[string]AllocationTarget, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one allocation target is required")
	}

	normalized := make(map[string]AllocationTarget, len(targets))
	sums := make(map[string]decimal.Decimal)
	for index, target := range targets {
		assetClass := strings.ToLower(strings.TrimSpace(target.AssetClass))
		if assetClass == "" {
			return nil, fmt.Errorf("allocation target %d requires asset_class", index)
		}
		currency, ok := normalizeCurrency(target.Currency)
		if !ok {
			return nil, fmt.Errorf("allocation target %d requires a three-letter uppercase currency code", index)
		}
		if _, ok := totals[currency]; !ok {
			return nil, fmt.Errorf("allocation target %d has no portfolio total for currency %s", index, currency)
		}
		if target.TargetPercentage.IsNegative() || target.TargetPercentage.GreaterThan(decimal.NewFromInt(100)) {
			return nil, fmt.Errorf("allocation target %d percentage must be between zero and 100", index)
		}

		key := allocationKey(assetClass, currency)
		if _, exists := normalized[key]; exists {
			return nil, fmt.Errorf("duplicate allocation target for asset class %s and currency %s", assetClass, currency)
		}
		target.AssetClass = assetClass
		target.Currency = currency
		normalized[key] = target
		sums[currency] = sums[currency].Add(target.TargetPercentage)
	}

	for currency := range totals {
		if !sums[currency].Equal(decimal.NewFromInt(100)) {
			return nil, fmt.Errorf("allocation targets for currency %s must sum to 100", currency)
		}
	}
	return normalized, nil
}

func currentClassAllocations(allocations []AssetClassAllocation) map[string]AssetClassAllocation {
	current := make(map[string]AssetClassAllocation, len(allocations))
	for _, allocation := range allocations {
		assetClass := strings.ToLower(strings.TrimSpace(allocation.AssetClass))
		currency, ok := normalizeCurrency(allocation.Currency)
		if !ok || assetClass == "" {
			continue
		}
		allocation.AssetClass = assetClass
		allocation.Currency = currency
		current[allocationKey(assetClass, currency)] = allocation
	}
	return current
}

func allocationKey(assetClass string, currency string) string {
	return currency + ":" + assetClass
}
