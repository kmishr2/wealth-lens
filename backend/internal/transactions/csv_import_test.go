package transactions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestParseCSVImportBuildsSignedLedgerRequests(t *testing.T) {
	accountID, assetID := uuid.New(), uuid.New()
	input := "transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key\n" +
		"deposit,2026-01-01T10:00:00Z,Opening balance,,,1000,INR,csv-1\n" +
		"buy,2026-01-02T10:00:00Z,Purchase," + assetID.String() + ",2.5,500,INR,csv-2\n"

	requests, err := parseCSVImport(strings.NewReader(input), accountID)
	if err != nil {
		t.Fatalf("parseCSVImport returned error: %v", err)
	}
	if len(requests) != 2 || requests[0].AccountID != accountID {
		t.Fatalf("requests = %+v", requests)
	}
	if !requests[0].Entries[0].Amount.Equal(decimal.NewFromInt(1000)) {
		t.Fatalf("deposit cash = %v", requests[0].Entries[0].Amount)
	}
	if len(requests[1].Entries) != 2 || !requests[1].Entries[0].Amount.Equal(decimal.NewFromInt(-500)) {
		t.Fatalf("buy entries = %+v", requests[1].Entries)
	}
	if !requests[1].Entries[1].Quantity.Equal(decimal.RequireFromString("2.5")) {
		t.Fatalf("buy quantity = %v", requests[1].Entries[1].Quantity)
	}
}

func TestParseCSVImportRejectsDuplicateIdempotencyKey(t *testing.T) {
	input := "transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key\n" +
		"deposit,2026-01-01T10:00:00Z,One,,,100,INR,repeated\n" +
		"withdrawal,2026-01-02T10:00:00Z,Two,,,50,INR,repeated\n"

	_, err := parseCSVImport(strings.NewReader(input), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "duplicate idempotency key") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseCSVImportRejectsAssetColumnsForCashRow(t *testing.T) {
	input := "transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key\n" +
		"deposit,2026-01-01T10:00:00Z,Invalid," + uuid.NewString() + ",1,100,INR,csv-invalid\n"

	_, err := parseCSVImport(strings.NewReader(input), uuid.New())
	if err == nil || !strings.Contains(err.Error(), "only allowed for buy and sell") {
		t.Fatalf("error = %v", err)
	}
}
