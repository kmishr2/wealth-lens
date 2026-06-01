package finance

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCalculatePortfolioValuationUsesExplicitPricesAndCashByCurrency(t *testing.T) {
	pricedAt := time.Date(2026, 5, 20, 16, 0, 0, 0, time.UTC)
	holdings := HoldingsResult{
		AssetHoldings: []AssetHolding{
			valuationHolding("asset-b", "VTI", "Vanguard Total Stock Market ETF", "equity", "10", "USD"),
			valuationHolding("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "5", "USD"),
		},
		CashBalances: []CashBalance{
			valuationCash("USD", "100"),
			valuationCash("INR", "500"),
		},
	}
	prices := []ValuationPrice{
		valuationPrice("asset-b", "12", "usd", pricedAt),
		valuationPrice("asset-a", "4", "USD", pricedAt),
	}

	result, err := CalculatePortfolioValuation(holdings, prices)
	if err != nil {
		t.Fatalf("CalculatePortfolioValuation returned error: %v", err)
	}

	if !result.IsFullyValued {
		t.Fatalf("IsFullyValued = false, want true")
	}
	if len(result.AssetValues) != 2 {
		t.Fatalf("asset values length = %d, want 2", len(result.AssetValues))
	}
	assertAssetValue(t, result.AssetValues[0], "asset-a", "BND", "20")
	assertAssetValue(t, result.AssetValues[1], "asset-b", "VTI", "120")
	if len(result.TotalValues) != 2 {
		t.Fatalf("total values length = %d, want 2", len(result.TotalValues))
	}
	assertCurrencyValue(t, result.TotalValues[0], "INR", "500")
	assertCurrencyValue(t, result.TotalValues[1], "USD", "240")
	if result.Definition.Name == "" || result.Definition.Formula == "" || result.ValuationScope == "" {
		t.Fatalf("valuation metadata is incomplete: %+v", result)
	}
}

func TestCalculatePortfolioValuationReportsMissingPrices(t *testing.T) {
	holdings := HoldingsResult{
		AssetHoldings: []AssetHolding{
			valuationHolding("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "5", "USD"),
		},
		CashBalances: []CashBalance{
			valuationCash("USD", "100"),
		},
	}

	result, err := CalculatePortfolioValuation(holdings, nil)
	if err != nil {
		t.Fatalf("CalculatePortfolioValuation returned error: %v", err)
	}

	if result.IsFullyValued {
		t.Fatalf("IsFullyValued = true, want false")
	}
	if len(result.AssetValues) != 0 {
		t.Fatalf("asset values length = %d, want 0", len(result.AssetValues))
	}
	if len(result.MissingPrices) != 1 {
		t.Fatalf("missing prices length = %d, want 1", len(result.MissingPrices))
	}
	if result.MissingPrices[0].AssetID != "asset-a" {
		t.Fatalf("missing asset id = %q, want asset-a", result.MissingPrices[0].AssetID)
	}
	assertCurrencyValue(t, result.TotalValues[0], "USD", "100")
}

func TestCalculatePortfolioValuationRejectsInvalidPrices(t *testing.T) {
	holdings := HoldingsResult{
		AssetHoldings: []AssetHolding{
			valuationHolding("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "5", "USD"),
		},
	}
	validTime := time.Now().UTC()

	tests := []struct {
		name        string
		prices      []ValuationPrice
		wantMessage string
	}{
		{
			name: "missing asset id",
			prices: []ValuationPrice{
				valuationPrice("", "1", "USD", validTime),
			},
			wantMessage: "requires asset_id",
		},
		{
			name: "zero price",
			prices: []ValuationPrice{
				valuationPrice("asset-a", "0", "USD", validTime),
			},
			wantMessage: "greater than zero",
		},
		{
			name: "invalid currency",
			prices: []ValuationPrice{
				valuationPrice("asset-a", "1", "US1", validTime),
			},
			wantMessage: "three-letter uppercase currency code",
		},
		{
			name: "currency mismatch",
			prices: []ValuationPrice{
				valuationPrice("asset-a", "1", "INR", validTime),
			},
			wantMessage: "does not match holding currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculatePortfolioValuation(holdings, tt.prices)
			if err == nil {
				t.Fatal("CalculatePortfolioValuation returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func valuationHolding(assetID string, symbol string, name string, assetClass string, quantity string, currency string) AssetHolding {
	return AssetHolding{
		AssetID:     assetID,
		AssetSymbol: symbol,
		AssetName:   name,
		AssetClass:  assetClass,
		Quantity:    decimal.RequireFromString(quantity),
		Currency:    currency,
	}
}

func valuationCash(currency string, amount string) CashBalance {
	return CashBalance{
		Currency: currency,
		Amount:   decimal.RequireFromString(amount),
	}
}

func valuationPrice(assetID string, price string, currency string, pricedAt time.Time) ValuationPrice {
	return ValuationPrice{
		AssetID:  assetID,
		Price:    decimal.RequireFromString(price),
		Currency: currency,
		PricedAt: pricedAt,
	}
}

func assertAssetValue(t *testing.T, value AssetValuation, assetID string, symbol string, marketValue string) {
	t.Helper()

	if value.AssetID != assetID {
		t.Fatalf("asset id = %q, want %q", value.AssetID, assetID)
	}
	if value.AssetSymbol != symbol {
		t.Fatalf("symbol = %q, want %q", value.AssetSymbol, symbol)
	}
	if !value.MarketValue.Equal(decimal.RequireFromString(marketValue)) {
		t.Fatalf("market value = %s, want %s", value.MarketValue, marketValue)
	}
}

func assertCurrencyValue(t *testing.T, value CurrencyValue, currency string, amount string) {
	t.Helper()

	if value.Currency != currency {
		t.Fatalf("currency = %q, want %q", value.Currency, currency)
	}
	if !value.Amount.Equal(decimal.RequireFromString(amount)) {
		t.Fatalf("amount = %s, want %s", value.Amount, amount)
	}
}
