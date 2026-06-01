package valuations

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
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
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

func TestGetCurrentCalculatesValuationFromHoldingsAndLatestPrices(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	quantity := decimal.RequireFromString("2.5")
	cash := decimal.RequireFromString("10")
	price := decimal.RequireFromString("4")
	pricedAt := time.Now().UTC().Add(-time.Hour)

	ledgerReader := &fakeLedgerReader{
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
	}
	priceReader := &fakeLatestPriceReader{
		prices: []prices.AssetPrice{
			{
				AssetID:  assetID,
				Price:    price,
				Currency: "USD",
				PricedAt: pricedAt,
			},
		},
	}
	portfolioReader := &fakePortfolioReader{
		portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID},
	}
	service := NewService(ledgerReader, priceReader, portfolioReader)

	response, err := service.GetCurrent(userID, portfolioID)
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if portfolioReader.userID != userID || portfolioReader.portfolioID != portfolioID {
		t.Fatalf("portfolio ownership check used user=%s portfolio=%s", portfolioReader.userID, portfolioReader.portfolioID)
	}
	if ledgerReader.portfolioID != portfolioID {
		t.Fatalf("ledger portfolio = %s, want %s", ledgerReader.portfolioID, portfolioID)
	}
	if len(priceReader.assetIDs) != 1 || priceReader.assetIDs[0] != assetID {
		t.Fatalf("price asset ids = %v, want [%s]", priceReader.assetIDs, assetID)
	}
	if !response.IsFullyValued {
		t.Fatal("IsFullyValued = false, want true")
	}
	if len(response.AssetValues) != 1 || !response.AssetValues[0].MarketValue.Equal(decimal.RequireFromString("10")) {
		t.Fatalf("asset values = %+v, want market value 10", response.AssetValues)
	}
	if len(response.TotalValues) != 1 || !response.TotalValues[0].Amount.Equal(decimal.RequireFromString("20")) {
		t.Fatalf("total values = %+v, want USD 20", response.TotalValues)
	}
	if response.MetricMetadata.Name == "" || response.HoldingsMeta.Name == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestGetCurrentReportsMissingPrices(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	quantity := decimal.RequireFromString("2.5")
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
			},
		},
		&fakeLatestPriceReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	response, err := service.GetCurrent(userID, portfolioID)
	if err != nil {
		t.Fatalf("GetCurrent returned error: %v", err)
	}

	if response.IsFullyValued {
		t.Fatal("IsFullyValued = true, want false")
	}
	if len(response.MissingPrices) != 1 || response.MissingPrices[0].AssetID != assetID.String() {
		t.Fatalf("missing prices = %+v, want asset %s", response.MissingPrices, assetID)
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

func TestAssetIDsForDeduplicatesAndSkipsInvalidIdentifiers(t *testing.T) {
	assetID := uuid.New()
	assetIDs := assetIDsFor([]finance.AssetHolding{
		{AssetID: assetID.String()},
		{AssetID: assetID.String()},
		{AssetID: "not-a-uuid"},
	})

	if len(assetIDs) != 1 || assetIDs[0] != assetID {
		t.Fatalf("assetIDsFor = %v, want [%s]", assetIDs, assetID)
	}
}
