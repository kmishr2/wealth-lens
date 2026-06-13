package snapshots

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

type fakeSnapshotRepo struct {
	getPortfolioID uuid.UUID
	getDate        time.Time
	getPeriod      string
	getCalls       int
	existing       *PortfolioSnapshot
	getErr         error

	created     *PortfolioSnapshot
	createCalls int
	createErr   error

	listPortfolioID uuid.UUID
	listPagination  common.Pagination
	listSnapshots   []PortfolioSnapshot
	listErr         error
}

func (f *fakeSnapshotRepo) Create(snapshot *PortfolioSnapshot) error {
	f.createCalls++
	f.created = snapshot
	return f.createErr
}

func (f *fakeSnapshotRepo) GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*PortfolioSnapshot, error) {
	f.getCalls++
	f.getPortfolioID = portfolioID
	f.getDate = snapshotDate
	f.getPeriod = snapshotPeriod
	if f.existing != nil {
		return f.existing, nil
	}
	if f.getErr != nil {
		return nil, f.getErr
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeSnapshotRepo) ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshot, error) {
	f.listPortfolioID = portfolioID
	f.listPagination = pagination
	return f.listSnapshots, f.listErr
}

type fakeLedgerAsOfReader struct {
	portfolioID uuid.UUID
	asOf        time.Time
	calls       int
	records     []holdings.LedgerEntryRecord
	err         error
}

func (f *fakeLedgerAsOfReader) ListLedgerEntriesAsOf(portfolioID uuid.UUID, asOf time.Time) ([]holdings.LedgerEntryRecord, error) {
	f.calls++
	f.portfolioID = portfolioID
	f.asOf = asOf
	return f.records, f.err
}

type fakeLatestPriceAsOfReader struct {
	assetIDs []uuid.UUID
	asOf     time.Time
	calls    int
	prices   []prices.AssetPrice
	err      error
}

func (f *fakeLatestPriceAsOfReader) ListLatestByAssetsAsOf(assetIDs []uuid.UUID, asOf time.Time) ([]prices.AssetPrice, error) {
	f.calls++
	f.assetIDs = assetIDs
	f.asOf = asOf
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

func TestCreateDailyBuildsSnapshotFromAsOfLedgerAndPrices(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	snapshotDate := "2026-01-15"
	quantity := decimal.RequireFromString("2")
	cash := decimal.RequireFromString("25")
	price := decimal.RequireFromString("50")
	expectedAsOf := time.Date(2026, 1, 15, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

	snapshotRepo := &fakeSnapshotRepo{}
	ledgerReader := &fakeLedgerAsOfReader{
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
	priceReader := &fakeLatestPriceAsOfReader{
		prices: []prices.AssetPrice{
			{
				AssetID:  assetID,
				Price:    price,
				Currency: "USD",
				PricedAt: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	portfolioReader := &fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}}
	service := NewService(snapshotRepo, ledgerReader, priceReader, portfolioReader)

	response, err := service.CreateDaily(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: snapshotDate})
	if err != nil {
		t.Fatalf("CreateDaily returned error: %v", err)
	}

	if portfolioReader.userID != userID || portfolioReader.portfolioID != portfolioID {
		t.Fatalf("portfolio ownership check used user=%s portfolio=%s", portfolioReader.userID, portfolioReader.portfolioID)
	}
	if ledgerReader.portfolioID != portfolioID || !ledgerReader.asOf.Equal(expectedAsOf) {
		t.Fatalf("ledger query portfolio=%s asOf=%s, want portfolio=%s asOf=%s", ledgerReader.portfolioID, ledgerReader.asOf, portfolioID, expectedAsOf)
	}
	if priceReader.calls != 1 || len(priceReader.assetIDs) != 1 || priceReader.assetIDs[0] != assetID || !priceReader.asOf.Equal(expectedAsOf) {
		t.Fatalf("price query assetIDs=%v asOf=%s, want [%s] asOf=%s", priceReader.assetIDs, priceReader.asOf, assetID, expectedAsOf)
	}
	if snapshotRepo.createCalls != 1 || snapshotRepo.created == nil {
		t.Fatalf("create calls = %d, created = %v", snapshotRepo.createCalls, snapshotRepo.created)
	}
	if response.SnapshotDate != snapshotDate || response.SnapshotPeriod != SnapshotPeriodDaily {
		t.Fatalf("snapshot date/period = %s/%s, want %s/%s", response.SnapshotDate, response.SnapshotPeriod, snapshotDate, SnapshotPeriodDaily)
	}
	if !response.IsFullyValued {
		t.Fatal("IsFullyValued = false, want true")
	}
	if len(response.TotalValues) != 1 || !response.TotalValues[0].Amount.Equal(decimal.RequireFromString("125")) {
		t.Fatalf("total values = %+v, want USD 125", response.TotalValues)
	}
	if len(response.AssetAllocations) != 1 || !response.AssetAllocations[0].Percentage.Equal(decimal.RequireFromString("80")) {
		t.Fatalf("asset allocations = %+v, want 80 percent", response.AssetAllocations)
	}
	if len(response.CashAllocations) != 1 || !response.CashAllocations[0].Percentage.Equal(decimal.RequireFromString("20")) {
		t.Fatalf("cash allocations = %+v, want 20 percent", response.CashAllocations)
	}
	if response.ValuationMetadata.Name == "" || response.AllocationMetadata.Name == "" || response.HoldingsMetadata.Name == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestCreateDailyReturnsExistingSnapshotIdempotently(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	snapshotDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	existing := storedSnapshot(t, userID, portfolioID, snapshotDate, decimal.RequireFromString("10"), true, nil)
	snapshotRepo := &fakeSnapshotRepo{existing: &existing}
	ledgerReader := &fakeLedgerAsOfReader{}
	priceReader := &fakeLatestPriceAsOfReader{}
	service := NewService(
		snapshotRepo,
		ledgerReader,
		priceReader,
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	response, err := service.CreateDaily(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-15"})
	if err != nil {
		t.Fatalf("CreateDaily returned error: %v", err)
	}

	if response.ID != existing.ID {
		t.Fatalf("response id = %s, want existing id %s", response.ID, existing.ID)
	}
	if snapshotRepo.createCalls != 0 {
		t.Fatalf("create calls = %d, want 0", snapshotRepo.createCalls)
	}
	if ledgerReader.calls != 0 || priceReader.calls != 0 {
		t.Fatalf("ledger calls = %d price calls = %d, want 0", ledgerReader.calls, priceReader.calls)
	}
}

func TestCreateDailyPersistsMissingPricesAsIncompleteSnapshot(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	assetID := uuid.New()
	quantity := decimal.RequireFromString("2")
	cash := decimal.RequireFromString("25")
	service := NewService(
		&fakeSnapshotRepo{},
		&fakeLedgerAsOfReader{
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
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	response, err := service.CreateDaily(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-15"})
	if err != nil {
		t.Fatalf("CreateDaily returned error: %v", err)
	}

	if response.IsFullyValued {
		t.Fatal("IsFullyValued = true, want false")
	}
	if len(response.MissingPrices) != 1 || response.MissingPrices[0].AssetID != assetID.String() {
		t.Fatalf("missing prices = %+v, want asset %s", response.MissingPrices, assetID)
	}
	if len(response.CashAllocations) != 1 || !response.CashAllocations[0].Percentage.Equal(decimal.RequireFromString("100")) {
		t.Fatalf("cash allocations = %+v, want 100 percent", response.CashAllocations)
	}
}

func TestCreateDailyRejectsSnapshotWithoutTotalValue(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	snapshotRepo := &fakeSnapshotRepo{}
	service := NewService(
		snapshotRepo,
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	_, err := service.CreateDaily(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-15"})
	if err == nil {
		t.Fatal("CreateDaily returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusBadRequest)
	}
	if snapshotRepo.createCalls != 0 {
		t.Fatalf("create calls = %d, want 0", snapshotRepo.createCalls)
	}
}

func TestListReturnsOwnedPortfolioSnapshotsNewestFirstFromRepository(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	first := storedSnapshot(t, userID, portfolioID, time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC), decimal.RequireFromString("20"), true, nil)
	second := storedSnapshot(t, userID, portfolioID, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), decimal.RequireFromString("10"), true, nil)
	snapshotRepo := &fakeSnapshotRepo{listSnapshots: []PortfolioSnapshot{first, second}}
	service := NewService(
		snapshotRepo,
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	responses, err := service.List(userID, portfolioID, common.Pagination{Limit: 25, Offset: 50})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if snapshotRepo.listPortfolioID != portfolioID || snapshotRepo.listPagination.Limit != 25 || snapshotRepo.listPagination.Offset != 50 {
		t.Fatalf("list query portfolio=%s pagination=%+v", snapshotRepo.listPortfolioID, snapshotRepo.listPagination)
	}
	if len(responses) != 2 || responses[0].SnapshotDate != "2026-01-16" || responses[1].SnapshotDate != "2026-01-15" {
		t.Fatalf("responses = %+v, want two decoded snapshots in repository order", responses)
	}
}

func TestCreateDailyReturnsNotFoundForUnownedPortfolio(t *testing.T) {
	service := NewService(
		&fakeSnapshotRepo{},
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{err: gorm.ErrRecordNotFound},
	)

	_, err := service.CreateDaily(uuid.New(), uuid.New(), SnapshotCreateRequest{SnapshotDate: "2026-01-15"})
	if err == nil {
		t.Fatal("CreateDaily returned nil error")
	}

	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *common.AppError", err)
	}
	if appErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", appErr.Status, http.StatusNotFound)
	}
}

func storedSnapshot(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, snapshotDate time.Time, total decimal.Decimal, isFullyValued bool, missing []finance.MissingPrice) PortfolioSnapshot {
	t.Helper()

	totalValues, err := NewJSONB([]finance.CurrencyValue{{Currency: "USD", Amount: total}})
	if err != nil {
		t.Fatalf("total values JSON: %v", err)
	}
	emptyAssetAllocations, err := NewJSONB([]finance.AssetAllocation{})
	if err != nil {
		t.Fatalf("asset allocations JSON: %v", err)
	}
	emptyAssetClassAllocations, err := NewJSONB([]finance.AssetClassAllocation{})
	if err != nil {
		t.Fatalf("asset class allocations JSON: %v", err)
	}
	cashAllocations, err := NewJSONB([]finance.CashAllocation{{Currency: "USD", Amount: total, Percentage: decimal.NewFromInt(100)}})
	if err != nil {
		t.Fatalf("cash allocations JSON: %v", err)
	}
	missingPrices, err := NewJSONB(missing)
	if err != nil {
		t.Fatalf("missing prices JSON: %v", err)
	}
	valuationMetadata, err := NewJSONB(finance.PortfolioValuationDefinition())
	if err != nil {
		t.Fatalf("valuation metadata JSON: %v", err)
	}
	allocationMetadata, err := NewJSONB(finance.AllocationDefinition())
	if err != nil {
		t.Fatalf("allocation metadata JSON: %v", err)
	}
	holdingsMetadata, err := NewJSONB(finance.HoldingsDefinition())
	if err != nil {
		t.Fatalf("holdings metadata JSON: %v", err)
	}

	return PortfolioSnapshot{
		ID:                    uuid.New(),
		PortfolioID:           portfolioID,
		SnapshotDate:          snapshotDate,
		SnapshotPeriod:        SnapshotPeriodDaily,
		TotalValues:           totalValues,
		AssetAllocations:      emptyAssetAllocations,
		AssetClassAllocations: emptyAssetClassAllocations,
		CashAllocations:       cashAllocations,
		MissingPrices:         missingPrices,
		IsFullyValued:         isFullyValued,
		ValuationScope:        "Totals are grouped by currency; no currency conversion is applied.",
		AllocationScope:       "Allocation percentages are calculated separately per currency; no currency conversion is applied.",
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       userID,
		CreatedAt:             time.Now().UTC(),
	}
}
