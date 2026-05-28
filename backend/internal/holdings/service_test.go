package holdings

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakeLedgerEntryReader struct {
	portfolioID uuid.UUID
	records     []LedgerEntryRecord
	err         error
}

func (f *fakeLedgerEntryReader) ListLedgerEntries(portfolioID uuid.UUID) ([]LedgerEntryRecord, error) {
	f.portfolioID = portfolioID
	return f.records, f.err
}

type fakePortfolioReader struct {
	userID      uuid.UUID
	portfolioID uuid.UUID
	portfolio   *portfolios.Portfolio
	err         error
}

func (f *fakePortfolioReader) GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	f.userID = userID
	f.portfolioID = portfolioID
	return f.portfolio, f.err
}

func TestGetCurrentCalculatesHoldingsForOwnedPortfolio(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	quantity := decimal.RequireFromString("4.5")
	amount := decimal.RequireFromString("-125.25")

	ledgerReader := &fakeLedgerEntryReader{
		records: []LedgerEntryRecord{
			{
				EntryKind:   "asset",
				AssetID:     "asset-a",
				AssetSymbol: "VTI",
				AssetName:   "Vanguard Total Stock Market ETF",
				AssetClass:  "equity",
				Quantity:    &quantity,
				Currency:    "USD",
			},
			{
				EntryKind: "cash",
				Amount:    &amount,
				Currency:  "USD",
			},
		},
	}
	portfolioReader := &fakePortfolioReader{
		portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID},
	}
	service := NewService(ledgerReader, portfolioReader)

	response, err := service.GetCurrent(userID, portfolioID)
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if portfolioReader.userID != userID || portfolioReader.portfolioID != portfolioID {
		t.Fatalf("portfolio ownership check used user=%s portfolio=%s", portfolioReader.userID, portfolioReader.portfolioID)
	}
	if ledgerReader.portfolioID != portfolioID {
		t.Fatalf("ledger reader portfolio = %s, want %s", ledgerReader.portfolioID, portfolioID)
	}
	if response.PortfolioID != portfolioID {
		t.Fatalf("response portfolio = %s, want %s", response.PortfolioID, portfolioID)
	}
	if len(response.AssetHoldings) != 1 || !response.AssetHoldings[0].Quantity.Equal(quantity) {
		t.Fatalf("asset holdings = %+v, want one holding with quantity %s", response.AssetHoldings, quantity)
	}
	if len(response.CashBalances) != 1 || !response.CashBalances[0].Amount.Equal(amount) {
		t.Fatalf("cash balances = %+v, want one balance with amount %s", response.CashBalances, amount)
	}
	if response.MetricMetadata.Name == "" || response.MetricMetadata.Formula == "" {
		t.Fatalf("metric metadata is incomplete: %+v", response.MetricMetadata)
	}
}

func TestGetCurrentReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(
		&fakeLedgerEntryReader{},
		&fakePortfolioReader{err: gorm.ErrRecordNotFound},
	)

	_, err := service.GetCurrent(uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("GetCurrent returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
}

func TestToFinanceEntriesMapsLedgerRecordsWithoutCalculating(t *testing.T) {
	quantity := decimal.RequireFromString("2")
	amount := decimal.RequireFromString("-10")
	records := []LedgerEntryRecord{
		{
			EntryKind:   "asset",
			AssetID:     "asset-a",
			AssetSymbol: "VTI",
			AssetName:   "Vanguard Total Stock Market ETF",
			AssetClass:  "equity",
			Quantity:    &quantity,
			Currency:    "USD",
		},
		{
			EntryKind: "fee",
			Amount:    &amount,
			Currency:  "USD",
		},
	}

	entries := ToFinanceEntries(records)

	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
	if string(entries[0].EntryKind) != records[0].EntryKind {
		t.Fatalf("entry kind = %q, want %q", entries[0].EntryKind, records[0].EntryKind)
	}
	if entries[0].AssetID != records[0].AssetID || entries[0].AssetSymbol != records[0].AssetSymbol {
		t.Fatalf("asset mapping = %+v, want %+v", entries[0], records[0])
	}
	if entries[1].Amount == nil || !entries[1].Amount.Equal(amount) {
		t.Fatalf("amount = %v, want %s", entries[1].Amount, amount)
	}
}
