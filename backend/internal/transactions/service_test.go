package transactions

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestReverseEntriesNegatesQuantitiesAndAmounts(t *testing.T) {
	assetID := uuid.New()
	quantity := decimal.RequireFromString("12.5000000000")
	amount := decimal.RequireFromString("1024.2500")
	cash := decimal.RequireFromString("-1024.2500")

	entries := []TransactionEntry{
		{
			EntryKind: EntryKindAsset,
			AssetID:   &assetID,
			Quantity:  &quantity,
			Amount:    &amount,
			Currency:  "USD",
		},
		{
			EntryKind: EntryKindCash,
			Amount:    &cash,
			Currency:  "USD",
		},
	}

	reversed := reverseEntries(entries)

	if len(reversed) != len(entries) {
		t.Fatalf("len(reversed) = %d, want %d", len(reversed), len(entries))
	}
	if reversed[0].Quantity == nil || !reversed[0].Quantity.Equal(decimal.RequireFromString("-12.5000000000")) {
		t.Fatalf("asset quantity = %v, want -12.5000000000", reversed[0].Quantity)
	}
	if reversed[0].Amount == nil || !reversed[0].Amount.Equal(decimal.RequireFromString("-1024.2500")) {
		t.Fatalf("asset amount = %v, want -1024.2500", reversed[0].Amount)
	}
	if reversed[1].Amount == nil || !reversed[1].Amount.Equal(decimal.RequireFromString("1024.2500")) {
		t.Fatalf("cash amount = %v, want 1024.2500", reversed[1].Amount)
	}
	if !entries[0].Quantity.Equal(decimal.RequireFromString("12.5000000000")) {
		t.Fatalf("original quantity changed to %v", entries[0].Quantity)
	}
}

func TestReverseEntriesPreservesNilOptionalValues(t *testing.T) {
	entries := []TransactionEntry{
		{
			EntryKind: EntryKindFee,
			Currency:  "USD",
		},
	}

	reversed := reverseEntries(entries)

	if len(reversed) != 1 {
		t.Fatalf("len(reversed) = %d, want 1", len(reversed))
	}
	if reversed[0].Quantity != nil {
		t.Fatalf("quantity = %v, want nil", reversed[0].Quantity)
	}
	if reversed[0].Amount != nil {
		t.Fatalf("amount = %v, want nil", reversed[0].Amount)
	}
}
