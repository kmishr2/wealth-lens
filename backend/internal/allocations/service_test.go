package allocations

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakeLedgerReader struct {
	portfolioID uuid.UUID
	records     []holdings.LedgerEntryRecord
	err         error
}

func (f *fakeLedgerReader) ListLedgerEntries(portfolioID uuid.UUID) ([]holdings.LedgerEntryRecord, error) {
	f.portfolioID = portfolioID
	return f.records, f.err
}

type fakeLatestPriceReader struct {
	assetIDs []uuid.UUID
	prices   []prices.AssetPrice
	err      error
}

func (f *fakeLatestPriceReader) ListLatestByAssets(assetIDs []uuid.UUID) ([]prices.AssetPrice, error) {
	f.assetIDs = assetIDs
	return f.prices, f.err
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

func TestGetCurrentCalculatesPortfolioAllocation(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	quantity := decimal.RequireFromString("2")
	cash := decimal.RequireFromString("25")
	price := decimal.RequireFromString("50")
	service := NewService(
		&fakeLedgerReader{
			records: []holdings.LedgerEntryRecord{
				{
					EntryKind:   "asset",
					AssetID:     assetID.String(),
					AssetSymbol: "VTI",
					AssetName:   "Vanguard Total Stock Market ETF",
					AssetClass:  "equity",
					Quantity:    &quantity,
					Currency:    "USD",
				},
				{
					EntryKind: "cash",
					Amount:    &cash,
					Currency:  "USD",
				},
			},
		},
		&fakeLatestPriceReader{
			prices: []prices.AssetPrice{
				{
					AssetID:  assetID,
					Price:    price,
					Currency: "USD",
					PricedAt: time.Now().UTC().Add(-time.Hour),
				},
			},
		},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	response, err := service.GetCurrent(userID, portfolioID)
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if !response.IsComplete {
		t.Fatal("IsComplete = false, want true")
	}
	if len(response.AssetAllocations) != 1 {
		t.Fatalf("asset allocations length = %d, want 1", len(response.AssetAllocations))
	}
	if !response.AssetAllocations[0].Percentage.Equal(decimal.RequireFromString("80")) {
		t.Fatalf("asset allocation percentage = %s, want 80", response.AssetAllocations[0].Percentage)
	}
	if len(response.CashAllocations) != 1 || !response.CashAllocations[0].Percentage.Equal(decimal.RequireFromString("20")) {
		t.Fatalf("cash allocations = %+v, want 20 percent", response.CashAllocations)
	}
	if response.MetricMetadata.Name == "" || response.ValuationMetadata.Name == "" || response.HoldingsMetadata.Name == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestGetCurrentCarriesMissingPricesAsIncompleteAllocation(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	quantity := decimal.RequireFromString("2")
	cash := decimal.RequireFromString("25")
	service := NewService(
		&fakeLedgerReader{
			records: []holdings.LedgerEntryRecord{
				{
					EntryKind:   "asset",
					AssetID:     assetID.String(),
					AssetSymbol: "VTI",
					AssetName:   "Vanguard Total Stock Market ETF",
					AssetClass:  "equity",
					Quantity:    &quantity,
					Currency:    "USD",
				},
				{
					EntryKind: "cash",
					Amount:    &cash,
					Currency:  "USD",
				},
			},
		},
		&fakeLatestPriceReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	response, err := service.GetCurrent(userID, portfolioID)
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if response.IsComplete {
		t.Fatal("IsComplete = true, want false")
	}
	if len(response.MissingPrices) != 1 || response.MissingPrices[0].AssetID != assetID.String() {
		t.Fatalf("missing prices = %+v, want asset %s", response.MissingPrices, assetID)
	}
	if len(response.CashAllocations) != 1 || !response.CashAllocations[0].Percentage.Equal(decimal.RequireFromString("100")) {
		t.Fatalf("cash allocations = %+v, want 100 percent", response.CashAllocations)
	}
}

func TestGetCurrentReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(
		&fakeLedgerReader{},
		&fakeLatestPriceReader{},
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
