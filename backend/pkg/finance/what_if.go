package finance

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type NamedSIPScenario struct {
	Name  string
	Input SIPProjectionInput
}

type WhatIfScenarioResult struct {
	Name                               string              `json:"name"`
	Projection                         SIPProjectionResult `json:"projection"`
	NominalDifferenceFromBaseline      decimal.Decimal     `json:"nominal_difference_from_baseline"`
	RealDifferenceFromBaseline         decimal.Decimal     `json:"real_difference_from_baseline"`
	ContributionDifferenceFromBaseline decimal.Decimal     `json:"contribution_difference_from_baseline"`
}

type WhatIfComparisonResult struct {
	BaselineName string                 `json:"baseline_name"`
	Months       int                    `json:"months"`
	Scenarios    []WhatIfScenarioResult `json:"scenarios"`
	Definition   MetricDefinition       `json:"definition"`
}

func WhatIfComparisonDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "SIP What-If Comparison",
		Formula: "Each scenario uses the SIP projection formula. Difference from baseline = scenario final value or contributions - first scenario final value or contributions.",
		Assumptions: []string{
			"The first explicitly supplied scenario is the baseline.",
			"All scenarios use the same horizon and contribution timing.",
			"Every return, inflation, contribution, and initial investment assumption is supplied explicitly.",
			"Results are deterministic simulations, not predictions or advice.",
		},
		RequiredInputs: []string{"two to ten uniquely named SIP scenarios", "common number of months"},
		Explanation:    "The comparison isolates the arithmetic effect of changing explicit scenario inputs while reporting differences against the first scenario.",
	}
}

func CompareSIPScenarios(scenarios []NamedSIPScenario) (WhatIfComparisonResult, error) {
	if len(scenarios) < 2 || len(scenarios) > 10 {
		return WhatIfComparisonResult{}, fmt.Errorf("what-if comparison requires between two and ten scenarios")
	}
	seen := make(map[string]struct{}, len(scenarios))
	projections := make([]SIPProjectionResult, len(scenarios))
	baselineMonths := scenarios[0].Input.Months
	for index, scenario := range scenarios {
		name := strings.TrimSpace(scenario.Name)
		if name == "" {
			return WhatIfComparisonResult{}, fmt.Errorf("scenario %d requires a name", index)
		}
		key := strings.ToLower(name)
		if _, exists := seen[key]; exists {
			return WhatIfComparisonResult{}, fmt.Errorf("scenario names must be unique")
		}
		seen[key] = struct{}{}
		if scenario.Input.Months != baselineMonths {
			return WhatIfComparisonResult{}, fmt.Errorf("all scenarios must use the same number of months")
		}
		projection, err := CalculateSIPProjection(scenario.Input)
		if err != nil {
			return WhatIfComparisonResult{}, fmt.Errorf("scenario %q: %w", name, err)
		}
		scenarios[index].Name = name
		projections[index] = projection
	}
	baseline := projections[0]
	results := make([]WhatIfScenarioResult, 0, len(scenarios))
	for index, scenario := range scenarios {
		projection := projections[index]
		results = append(results, WhatIfScenarioResult{Name: scenario.Name, Projection: projection,
			NominalDifferenceFromBaseline:      projection.ProjectedNominalValue.Sub(baseline.ProjectedNominalValue).Round(sipPrecision),
			RealDifferenceFromBaseline:         projection.ProjectedRealValue.Sub(baseline.ProjectedRealValue).Round(sipPrecision),
			ContributionDifferenceFromBaseline: projection.TotalContributions.Sub(baseline.TotalContributions).Round(sipPrecision)})
	}
	return WhatIfComparisonResult{BaselineName: scenarios[0].Name, Months: baselineMonths, Scenarios: results, Definition: WhatIfComparisonDefinition()}, nil
}
