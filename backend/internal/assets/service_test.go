package assets

import "testing"

func TestNormalizeRiskCategoryUsesUnambiguousDefaults(t *testing.T) {
	tests := []struct {
		assetClass string
		want       string
	}{
		{AssetClassEquity, RiskCategoryEquity},
		{AssetClassBond, RiskCategoryDebt},
		{AssetClassCash, RiskCategoryCashOther},
	}
	for _, tt := range tests {
		result, err := normalizeRiskCategory(nil, tt.assetClass)
		if err != nil || result == nil || *result != tt.want {
			t.Fatalf("normalizeRiskCategory(%s) = %v, %v; want %s", tt.assetClass, result, err, tt.want)
		}
	}
}

func TestNormalizeRiskCategoryLeavesAmbiguousFundUnclassified(t *testing.T) {
	result, err := normalizeRiskCategory(nil, AssetClassFund)
	if err != nil || result != nil {
		t.Fatalf("result = %v, error = %v; want nil", result, err)
	}
}

func TestNormalizeRiskCategoryAcceptsExplicitFundCategory(t *testing.T) {
	value := " Debt "
	result, err := normalizeRiskCategory(&value, AssetClassFund)
	if err != nil || result == nil || *result != RiskCategoryDebt {
		t.Fatalf("result = %v, error = %v", result, err)
	}
}

func TestNormalizeRiskCategoryRejectsUnknownCategory(t *testing.T) {
	value := "balanced"
	if _, err := normalizeRiskCategory(&value, AssetClassFund); err == nil {
		t.Fatal("error = nil, want invalid category error")
	}
}
