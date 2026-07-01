package finance

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateDiversificationAlertsUsesLowerBand(t *testing.T) {
	result := CalculateDiversificationAlerts(ConcentrationResult{Currencies: []CurrencyConcentration{
		{Currency: "INR", AssetCount: 3, LargestAssetID: "a", LargestAssetSymbol: "AAA", LargestAssetPercentage: decimal.NewFromInt(30)},
	}})
	alert := result.Alerts[0]
	if alert.Points != 10 || alert.Severity != "warning" || len(alert.Conditions) != 2 {
		t.Fatalf("alert = %+v", alert)
	}
	if result.Definition.Formula == "" || result.Scope == "" {
		t.Fatalf("metadata = %+v", result)
	}
}

func TestCalculateDiversificationAlertsBoundarySeverities(t *testing.T) {
	result := CalculateDiversificationAlerts(ConcentrationResult{Currencies: []CurrencyConcentration{
		{Currency: "A", AssetCount: 8, LargestAssetPercentage: decimal.NewFromInt(20)},
		{Currency: "B", AssetCount: 7, LargestAssetPercentage: decimal.NewFromInt(35)},
		{Currency: "C", AssetCount: 4, LargestAssetPercentage: decimal.NewFromInt(50)},
		{Currency: "D", AssetCount: 2, LargestAssetPercentage: decimal.RequireFromString("50.1")},
	}})
	want := []string{"none", "notice", "warning", "critical"}
	for index, alert := range result.Alerts {
		if alert.Severity != want[index] {
			t.Fatalf("alert %d severity = %s, want %s", index, alert.Severity, want[index])
		}
	}
}
