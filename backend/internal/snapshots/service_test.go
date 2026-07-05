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
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
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

	weeklyExisting        *WeeklyPerformanceSnapshot
	weeklyCreated         *WeeklyPerformanceSnapshot
	weeklyCreateCalls     int
	weeklyCreateErr       error
	weeklyListed          []WeeklyPerformanceSnapshot
	weeklyListPortfolioID uuid.UUID
	weeklyListPagination  common.Pagination

	listPortfolioID uuid.UUID
	listPagination  common.Pagination
	listSnapshots   []PortfolioSnapshot
	listErr         error
}

func (f *fakeSnapshotRepo) ListWeeklyPerformanceByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]WeeklyPerformanceSnapshot, error) {
	f.weeklyListPortfolioID = portfolioID
	f.weeklyListPagination = pagination
	return f.weeklyListed, nil
}

func (f *fakeSnapshotRepo) Create(snapshot *PortfolioSnapshot) error {
	f.createCalls++
	f.created = snapshot
	return f.createErr
}

func (f *fakeSnapshotRepo) CreateWeeklyPerformance(snapshot *WeeklyPerformanceSnapshot) error {
	f.weeklyCreateCalls++
	f.weeklyCreated = snapshot
	return f.weeklyCreateErr
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

func (f *fakeSnapshotRepo) GetWeeklyPerformanceByPortfolioWeekEnd(uuid.UUID, time.Time) (*WeeklyPerformanceSnapshot, error) {
	if f.weeklyExisting != nil {
		return f.weeklyExisting, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeSnapshotRepo) ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshot, error) {
	f.listPortfolioID = portfolioID
	f.listPagination = pagination
	return f.listSnapshots, f.listErr
}

type fakeSnapshotSequenceRepo struct {
	fakeSnapshotRepo
	snapshots map[string]PortfolioSnapshot
}

func (f *fakeSnapshotSequenceRepo) GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*PortfolioSnapshot, error) {
	f.getCalls++
	f.getPortfolioID = portfolioID
	f.getDate = snapshotDate
	f.getPeriod = snapshotPeriod
	snapshot, ok := f.snapshots[snapshotDate.UTC().Format(snapshotDateLayout)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &snapshot, nil
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

type fakeSnapshotCashFlowReader struct {
	portfolioID uuid.UUID
	startAfter  time.Time
	endAt       time.Time
	records     []transactions.ExternalCashFlowRecord
	err         error
}

func (f *fakeSnapshotCashFlowReader) ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error) {
	f.portfolioID = portfolioID
	f.startAfter = startAfter
	f.endAt = endAt
	return f.records, f.err
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

func TestCreateWeeklyPerformanceBuildsSnapshotFromDailySnapshotsAndCashFlows(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	weekStart := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	weekEnd := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	snapshotRepo := &fakeSnapshotSequenceRepo{
		snapshots: map[string]PortfolioSnapshot{
			"2026-01-04": storedSnapshot(t, userID, portfolioID, weekStart, decimal.RequireFromString("1000"), true, nil),
			"2026-01-11": storedSnapshot(t, userID, portfolioID, weekEnd, decimal.RequireFromString("1150"), true, nil),
		},
	}
	cashFlowReader := &fakeSnapshotCashFlowReader{
		records: []transactions.ExternalCashFlowRecord{
			{
				OccurredAt: time.Date(2026, 1, 7, 10, 0, 0, 0, time.UTC),
				Currency:   "USD",
				Amount:     decimal.RequireFromString("50"),
			},
		},
	}
	service := NewServiceWithCashFlows(
		snapshotRepo,
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		cashFlowReader,
	)

	response, err := service.CreateWeeklyPerformance(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-11"})
	if err != nil {
		t.Fatalf("CreateWeeklyPerformance returned error: %v", err)
	}

	if snapshotRepo.weeklyCreateCalls != 1 || snapshotRepo.weeklyCreated == nil {
		t.Fatalf("weekly create calls = %d created = %v", snapshotRepo.weeklyCreateCalls, snapshotRepo.weeklyCreated)
	}
	if response.WeekStartDate != "2026-01-04" || response.WeekEndDate != "2026-01-11" {
		t.Fatalf("week range = %s..%s, want 2026-01-04..2026-01-11", response.WeekStartDate, response.WeekEndDate)
	}
	expectedStartAfter := time.Date(2026, 1, 4, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	expectedEndAt := time.Date(2026, 1, 11, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !cashFlowReader.startAfter.Equal(expectedStartAfter) || !cashFlowReader.endAt.Equal(expectedEndAt) {
		t.Fatalf("cash flow range = %s..%s, want %s..%s", cashFlowReader.startAfter, cashFlowReader.endAt, expectedStartAfter, expectedEndAt)
	}
	if len(response.CurrencyReturns) != 1 {
		t.Fatalf("currency returns = %+v, want one currency", response.CurrencyReturns)
	}
	result := response.CurrencyReturns[0]
	if result.Currency != "USD" {
		t.Fatalf("currency = %s, want USD", result.Currency)
	}
	if !result.NetExternalCashFlow.Equal(decimal.RequireFromString("50")) {
		t.Fatalf("net external cash flow = %s, want 50", result.NetExternalCashFlow)
	}
	if !result.ProfitLoss.Equal(decimal.RequireFromString("100")) {
		t.Fatalf("profit loss = %s, want 100", result.ProfitLoss)
	}
	if result.CashFlowCount != 1 {
		t.Fatalf("cash flow count = %d, want 1", result.CashFlowCount)
	}
	if response.PerformanceScope == "" || response.PnLMetadata.Name == "" || response.CAGRMetadata.Name == "" || response.XIRRMetadata.Name == "" {
		t.Fatalf("metadata is incomplete: %+v", response)
	}
}

func TestCreateWeeklyPerformanceReturnsExistingSnapshotIdempotently(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	existing := storedWeeklyPerformanceSnapshot(t, userID, portfolioID, "2026-01-04", "2026-01-11")
	snapshotRepo := &fakeSnapshotRepo{weeklyExisting: &existing}
	service := NewServiceWithCashFlows(
		snapshotRepo,
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotCashFlowReader{},
	)

	response, err := service.CreateWeeklyPerformance(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-11"})
	if err != nil {
		t.Fatalf("CreateWeeklyPerformance returned error: %v", err)
	}

	if response.ID != existing.ID {
		t.Fatalf("response id = %s, want existing id %s", response.ID, existing.ID)
	}
	if snapshotRepo.weeklyCreateCalls != 0 {
		t.Fatalf("weekly create calls = %d, want 0", snapshotRepo.weeklyCreateCalls)
	}
}

func TestCreateWeeklyPerformanceRejectsNonSundayWeekEnd(t *testing.T) {
	userID := uuid.New()
	portfolioID := uuid.New()
	service := NewServiceWithCashFlows(
		&fakeSnapshotRepo{},
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
		&fakeSnapshotCashFlowReader{},
	)

	_, err := service.CreateWeeklyPerformance(userID, portfolioID, SnapshotCreateRequest{SnapshotDate: "2026-01-10"})
	if err == nil {
		t.Fatal("CreateWeeklyPerformance returned nil error")
	}
	var appErr *common.AppError
	if !errors.As(err, &appErr) || appErr.Message != "Weekly performance snapshot date must be a UTC Sunday" {
		t.Fatalf("error = %v", err)
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

func TestListWeeklyPerformanceReturnsDecodedStoredMetrics(t *testing.T) {
	userID, portfolioID := uuid.New(), uuid.New()
	stored := storedWeeklyPerformanceSnapshot(t, userID, portfolioID, "2026-01-04", "2026-01-11")
	repo := &fakeSnapshotRepo{weeklyListed: []WeeklyPerformanceSnapshot{stored}}
	service := NewService(
		repo,
		&fakeLedgerAsOfReader{},
		&fakeLatestPriceAsOfReader{},
		&fakePortfolioReader{portfolio: &portfolios.Portfolio{ID: portfolioID, UserID: userID}},
	)

	responses, err := service.ListWeeklyPerformance(userID, portfolioID, common.Pagination{Limit: 10, Offset: 20})
	if err != nil {
		t.Fatalf("ListWeeklyPerformance returned error: %v", err)
	}
	if repo.weeklyListPortfolioID != portfolioID || repo.weeklyListPagination.Limit != 10 || repo.weeklyListPagination.Offset != 20 {
		t.Fatalf("weekly list scope=%s pagination=%+v", repo.weeklyListPortfolioID, repo.weeklyListPagination)
	}
	if len(responses) != 1 || responses[0].WeekEndDate != "2026-01-11" || len(responses[0].CurrencyReturns) != 1 {
		t.Fatalf("responses = %+v", responses)
	}
	if responses[0].PnLMetadata.Formula == "" || responses[0].XIRRMetadata.Explanation == "" {
		t.Fatal("stored weekly metric metadata was not decoded")
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

func storedWeeklyPerformanceSnapshot(t *testing.T, userID uuid.UUID, portfolioID uuid.UUID, weekStartDate string, weekEndDate string) WeeklyPerformanceSnapshot {
	t.Helper()

	currencyReturns, err := NewJSONB([]WeeklyCurrencyPerformance{
		{
			Currency:       "USD",
			BeginningValue: decimal.RequireFromString("1000"),
			EndingValue:    decimal.RequireFromString("1100"),
			ProfitLoss:     decimal.RequireFromString("100"),
			CAGR:           decimal.RequireFromString("10"),
			XIRR:           decimal.RequireFromString("10"),
		},
	})
	if err != nil {
		t.Fatalf("currency returns JSON: %v", err)
	}
	pnlMetadata, err := NewJSONB(finance.PeriodPnLDefinition())
	if err != nil {
		t.Fatalf("pnl metadata JSON: %v", err)
	}
	cagrMetadata, err := NewJSONB(finance.CAGRDefinition())
	if err != nil {
		t.Fatalf("cagr metadata JSON: %v", err)
	}
	xirrMetadata, err := NewJSONB(finance.XIRRDefinition())
	if err != nil {
		t.Fatalf("xirr metadata JSON: %v", err)
	}

	return WeeklyPerformanceSnapshot{
		ID:               uuid.New(),
		PortfolioID:      portfolioID,
		WeekStartDate:    mustSnapshotDate(weekStartDate),
		WeekEndDate:      mustSnapshotDate(weekEndDate),
		CurrencyReturns:  currencyReturns,
		PerformanceScope: "test",
		PnLMetadata:      pnlMetadata,
		CAGRMetadata:     cagrMetadata,
		XIRRMetadata:     xirrMetadata,
		CreatedByUserID:  userID,
		CreatedAt:        time.Now().UTC(),
	}
}

func mustSnapshotDate(date string) time.Time {
	parsed, err := time.Parse(snapshotDateLayout, date)
	if err != nil {
		panic(err)
	}
	return parsed
}
