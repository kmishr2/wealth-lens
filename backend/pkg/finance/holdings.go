package finance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type LedgerEntryKind string

const (
	LedgerEntryKindAsset LedgerEntryKind = "asset"
	LedgerEntryKindCash  LedgerEntryKind = "cash"
	LedgerEntryKindFee   LedgerEntryKind = "fee"
	LedgerEntryKindTax   LedgerEntryKind = "tax"
)

type LedgerEntry struct {
	EntryKind   LedgerEntryKind
	AssetID     string
	AssetSymbol string
	AssetName   string
	AssetClass  string
	Quantity    *decimal.Decimal
	Amount      *decimal.Decimal
	Currency    string
}

type HoldingsResult struct {
	AssetHoldings []AssetHolding   `json:"asset_holdings"`
	CashBalances  []CashBalance    `json:"cash_balances"`
	Definition    MetricDefinition `json:"definition"`
}

type AssetHolding struct {
	AssetID     string          `json:"asset_id"`
	AssetSymbol string          `json:"asset_symbol"`
	AssetName   string          `json:"asset_name"`
	AssetClass  string          `json:"asset_class"`
	Quantity    decimal.Decimal `json:"quantity"`
	Currency    string          `json:"currency"`
}

type CashBalance struct {
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
}

func HoldingsDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Current Holdings",
		Formula: "Asset quantity per asset = sum(signed asset quantities). Cash balance per currency = sum(signed cash, fee, and tax amounts).",
		Assumptions: []string{
			"Ledger entries are immutable and already ordered only for audit review, not for the summation result.",
			"Quantities and amounts are signed: positive values increase a position or balance, negative values reduce it.",
			"Asset entry amounts are not counted as cash; cash movement must be represented by cash, fee, or tax entries.",
		},
		RequiredInputs: []string{
			"entry_kind",
			"asset_id for asset entries",
			"quantity for asset entries",
			"amount for cash, fee, and tax entries",
			"currency",
		},
		Explanation: "Holdings are derived directly from the transaction ledger by summing signed quantities and cash amounts. No portfolio state is stored or inferred outside the ledger inputs.",
	}
}

func CalculateHoldings(entries []LedgerEntry) (HoldingsResult, error) {
	assetTotals := make(map[string]AssetHolding)
	cashTotals := make(map[string]decimal.Decimal)

	for index, entry := range entries {
		switch entry.EntryKind {
		case LedgerEntryKindAsset:
			if err := addAssetEntry(assetTotals, index, entry); err != nil {
				return HoldingsResult{}, err
			}
		case LedgerEntryKindCash, LedgerEntryKindFee, LedgerEntryKindTax:
			if err := addAmountEntry(cashTotals, index, entry); err != nil {
				return HoldingsResult{}, err
			}
		default:
			return HoldingsResult{}, fmt.Errorf("entry %d has unsupported entry kind %q", index, entry.EntryKind)
		}
	}

	assetHoldings := make([]AssetHolding, 0, len(assetTotals))
	for _, holding := range assetTotals {
		if !holding.Quantity.IsZero() {
			assetHoldings = append(assetHoldings, holding)
		}
	}
	sort.Slice(assetHoldings, func(i, j int) bool {
		if assetHoldings[i].AssetSymbol == assetHoldings[j].AssetSymbol {
			return assetHoldings[i].AssetID < assetHoldings[j].AssetID
		}
		return assetHoldings[i].AssetSymbol < assetHoldings[j].AssetSymbol
	})

	cashBalances := make([]CashBalance, 0, len(cashTotals))
	for currency, amount := range cashTotals {
		if !amount.IsZero() {
			cashBalances = append(cashBalances, CashBalance{
				Currency: currency,
				Amount:   amount,
			})
		}
	}
	sort.Slice(cashBalances, func(i, j int) bool {
		return cashBalances[i].Currency < cashBalances[j].Currency
	})

	return HoldingsResult{
		AssetHoldings: assetHoldings,
		CashBalances:  cashBalances,
		Definition:    HoldingsDefinition(),
	}, nil
}

func addAssetEntry(assetTotals map[string]AssetHolding, index int, entry LedgerEntry) error {
	assetID := strings.TrimSpace(entry.AssetID)
	if assetID == "" {
		return fmt.Errorf("asset entry %d requires asset_id", index)
	}
	if entry.Quantity == nil || entry.Quantity.IsZero() {
		return fmt.Errorf("asset entry %d requires a non-zero quantity", index)
	}

	currency, ok := normalizeCurrency(entry.Currency)
	if !ok {
		return fmt.Errorf("asset entry %d requires a three-letter uppercase currency code", index)
	}

	next := AssetHolding{
		AssetID:     assetID,
		AssetSymbol: strings.ToUpper(strings.TrimSpace(entry.AssetSymbol)),
		AssetName:   strings.TrimSpace(entry.AssetName),
		AssetClass:  strings.ToLower(strings.TrimSpace(entry.AssetClass)),
		Quantity:    *entry.Quantity,
		Currency:    currency,
	}

	existing, ok := assetTotals[assetID]
	if !ok {
		assetTotals[assetID] = next
		return nil
	}
	if existing.AssetSymbol != next.AssetSymbol || existing.AssetName != next.AssetName || existing.AssetClass != next.AssetClass || existing.Currency != next.Currency {
		return fmt.Errorf("asset entry %d has inconsistent metadata for asset_id %s", index, assetID)
	}

	existing.Quantity = existing.Quantity.Add(*entry.Quantity)
	assetTotals[assetID] = existing
	return nil
}

func addAmountEntry(cashTotals map[string]decimal.Decimal, index int, entry LedgerEntry) error {
	if entry.Amount == nil || entry.Amount.IsZero() {
		return fmt.Errorf("%s entry %d requires a non-zero amount", entry.EntryKind, index)
	}

	currency, ok := normalizeCurrency(entry.Currency)
	if !ok {
		return fmt.Errorf("%s entry %d requires a three-letter uppercase currency code", entry.EntryKind, index)
	}

	cashTotals[currency] = cashTotals[currency].Add(*entry.Amount)
	return nil
}

func normalizeCurrency(value string) (string, bool) {
	currency := strings.ToUpper(strings.TrimSpace(value))
	if len(currency) != 3 {
		return "", false
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return "", false
		}
	}
	return currency, true
}
