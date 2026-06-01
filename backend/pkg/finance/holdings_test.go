package finance

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculateHoldingsSumsAssetQuantitiesAndCashBalances(t *testing.T) {
	entries := []LedgerEntry{
		assetEntry("asset-b", "VTI", "Vanguard Total Stock Market ETF", "equity", "usd", "10"),
		assetEntry("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "USD", "5"),
		assetEntry("asset-b", "VTI", "Vanguard Total Stock Market ETF", "equity", "USD", "-3.25"),
		amountEntry(LedgerEntryKindCash, "USD", "1000"),
		amountEntry(LedgerEntryKindCash, "usd", "-250.25"),
		amountEntry(LedgerEntryKindFee, "USD", "-2.75"),
		amountEntry(LedgerEntryKindTax, "INR", "-100"),
	}

	result, err := CalculateHoldings(entries)
	if err != nil {
		t.Fatalf("CalculateHoldings returned error: %v", err)
	}

	if len(result.AssetHoldings) != 2 {
		t.Fatalf("asset holdings length = %d, want 2", len(result.AssetHoldings))
	}
	assertHolding(t, result.AssetHoldings[0], "asset-a", "BND", "5")
	assertHolding(t, result.AssetHoldings[1], "asset-b", "VTI", "6.75")

	if len(result.CashBalances) != 2 {
		t.Fatalf("cash balances length = %d, want 2", len(result.CashBalances))
	}
	assertCashBalance(t, result.CashBalances[0], "INR", "-100")
	assertCashBalance(t, result.CashBalances[1], "USD", "747")

	if result.Definition.Name == "" || result.Definition.Formula == "" || result.Definition.Explanation == "" {
		t.Fatalf("definition is incomplete: %+v", result.Definition)
	}
}

func TestCalculateHoldingsDropsZeroResultPositions(t *testing.T) {
	entries := []LedgerEntry{
		assetEntry("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "USD", "5"),
		assetEntry("asset-a", "BND", "Vanguard Total Bond Market ETF", "bond", "USD", "-5"),
		amountEntry(LedgerEntryKindCash, "USD", "250"),
		amountEntry(LedgerEntryKindCash, "USD", "-250"),
	}

	result, err := CalculateHoldings(entries)
	if err != nil {
		t.Fatalf("CalculateHoldings returned error: %v", err)
	}

	if len(result.AssetHoldings) != 0 {
		t.Fatalf("asset holdings length = %d, want 0", len(result.AssetHoldings))
	}
	if len(result.CashBalances) != 0 {
		t.Fatalf("cash balances length = %d, want 0", len(result.CashBalances))
	}
}

func TestCalculateHoldingsRejectsInvalidLedgerEntries(t *testing.T) {
	tests := []struct {
		name        string
		entries     []LedgerEntry
		wantMessage string
	}{
		{
			name: "missing asset id",
			entries: []LedgerEntry{
				assetEntry("", "VTI", "Vanguard Total Stock Market ETF", "equity", "USD", "1"),
			},
			wantMessage: "requires asset_id",
		},
		{
			name: "missing asset quantity",
			entries: []LedgerEntry{
				{
					EntryKind: LedgerEntryKindAsset,
					AssetID:   "asset-a",
					Currency:  "USD",
				},
			},
			wantMessage: "requires a non-zero quantity",
		},
		{
			name: "missing amount",
			entries: []LedgerEntry{
				{
					EntryKind: LedgerEntryKindCash,
					Currency:  "USD",
				},
			},
			wantMessage: "requires a non-zero amount",
		},
		{
			name: "invalid currency",
			entries: []LedgerEntry{
				amountEntry(LedgerEntryKindCash, "US1", "1"),
			},
			wantMessage: "three-letter uppercase currency code",
		},
		{
			name: "unsupported entry kind",
			entries: []LedgerEntry{
				{
					EntryKind: "dividend",
				},
			},
			wantMessage: "unsupported entry kind",
		},
		{
			name: "inconsistent asset metadata",
			entries: []LedgerEntry{
				assetEntry("asset-a", "VTI", "Vanguard Total Stock Market ETF", "equity", "USD", "1"),
				assetEntry("asset-a", "VOO", "Vanguard S&P 500 ETF", "equity", "USD", "1"),
			},
			wantMessage: "inconsistent metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateHoldings(tt.entries)
			if err == nil {
				t.Fatal("CalculateHoldings returned nil error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func assetEntry(assetID string, symbol string, name string, assetClass string, currency string, quantity string) LedgerEntry {
	value := decimal.RequireFromString(quantity)
	return LedgerEntry{
		EntryKind:   LedgerEntryKindAsset,
		AssetID:     assetID,
		AssetSymbol: symbol,
		AssetName:   name,
		AssetClass:  assetClass,
		Quantity:    &value,
		Currency:    currency,
	}
}

func amountEntry(entryKind LedgerEntryKind, currency string, amount string) LedgerEntry {
	value := decimal.RequireFromString(amount)
	return LedgerEntry{
		EntryKind: entryKind,
		Amount:    &value,
		Currency:  currency,
	}
}

func assertHolding(t *testing.T, holding AssetHolding, assetID string, symbol string, quantity string) {
	t.Helper()

	if holding.AssetID != assetID {
		t.Fatalf("asset id = %q, want %q", holding.AssetID, assetID)
	}
	if holding.AssetSymbol != symbol {
		t.Fatalf("asset symbol = %q, want %q", holding.AssetSymbol, symbol)
	}
	if !holding.Quantity.Equal(decimal.RequireFromString(quantity)) {
		t.Fatalf("quantity = %s, want %s", holding.Quantity, quantity)
	}
}

func assertCashBalance(t *testing.T, balance CashBalance, currency string, amount string) {
	t.Helper()

	if balance.Currency != currency {
		t.Fatalf("currency = %q, want %q", balance.Currency, currency)
	}
	if !balance.Amount.Equal(decimal.RequireFromString(amount)) {
		t.Fatalf("amount = %s, want %s", balance.Amount, amount)
	}
}
