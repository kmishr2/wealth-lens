package snapshots

import (
	"errors"
	"fmt"
	"sort"
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

type ledgerEntryReader interface {
	ListLedgerEntriesAsOf(portfolioID uuid.UUID, asOf time.Time) ([]holdings.LedgerEntryRecord, error)
}

type latestPriceReader interface {
	ListLatestByAssetsAsOf(assetIDs []uuid.UUID, asOf time.Time) ([]prices.AssetPrice, error)
}

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type externalCashFlowReader interface {
	ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error)
}

type snapshotReaderWriter interface {
	Create(snapshot *PortfolioSnapshot) error
	CreateWeeklyPerformance(snapshot *WeeklyPerformanceSnapshot) error
	GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*PortfolioSnapshot, error)
	GetWeeklyPerformanceByPortfolioWeekEnd(portfolioID uuid.UUID, weekEndDate time.Time) (*WeeklyPerformanceSnapshot, error)
	ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshot, error)
}

type Service struct {
	snapshotRepo  snapshotReaderWriter
	ledgerRepo    ledgerEntryReader
	priceRepo     latestPriceReader
	portfolioRepo portfolioReader
	cashFlowRepo  externalCashFlowReader
}

func NewService(snapshotRepo snapshotReaderWriter, ledgerRepo ledgerEntryReader, priceRepo latestPriceReader, portfolioRepo portfolioReader) *Service {
	return &Service{
		snapshotRepo:  snapshotRepo,
		ledgerRepo:    ledgerRepo,
		priceRepo:     priceRepo,
		portfolioRepo: portfolioRepo,
	}
}

func NewServiceWithCashFlows(snapshotRepo snapshotReaderWriter, ledgerRepo ledgerEntryReader, priceRepo latestPriceReader, portfolioRepo portfolioReader, cashFlowRepo externalCashFlowReader) *Service {
	service := NewService(snapshotRepo, ledgerRepo, priceRepo, portfolioRepo)
	service.cashFlowRepo = cashFlowRepo
	return service
}

func (s *Service) CreateDaily(userID uuid.UUID, portfolioID uuid.UUID, req SnapshotCreateRequest) (PortfolioSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return PortfolioSnapshotResponse{}, err
	}

	snapshotDate, err := parseSnapshotDate(req.SnapshotDate)
	if err != nil {
		return PortfolioSnapshotResponse{}, err
	}

	existing, err := s.snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, SnapshotPeriodDaily)
	if err == nil {
		return ToResponse(*existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return PortfolioSnapshotResponse{}, err
	}

	snapshot, err := s.buildDailySnapshot(userID, portfolioID, snapshotDate)
	if err != nil {
		return PortfolioSnapshotResponse{}, err
	}

	if err := s.snapshotRepo.Create(snapshot); err != nil {
		if common.IsUniqueViolation(err) {
			existing, getErr := s.snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, SnapshotPeriodDaily)
			if getErr == nil {
				return ToResponse(*existing)
			}
			return PortfolioSnapshotResponse{}, getErr
		}
		return PortfolioSnapshotResponse{}, err
	}

	return ToResponse(*snapshot)
}

func (s *Service) CreateWeeklyPerformance(userID uuid.UUID, portfolioID uuid.UUID, req SnapshotCreateRequest) (WeeklyPerformanceSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}
	if s.cashFlowRepo == nil {
		return WeeklyPerformanceSnapshotResponse{}, common.Internal("Weekly performance snapshots require a cash flow repository")
	}

	weekEndDate, err := parseSnapshotDate(req.SnapshotDate)
	if err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}
	if weekEndDate.Weekday() != time.Sunday {
		return WeeklyPerformanceSnapshotResponse{}, common.BadRequest("Weekly performance snapshot date must be a UTC Sunday")
	}
	weekStartDate := weekEndDate.AddDate(0, 0, -7)

	existing, err := s.snapshotRepo.GetWeeklyPerformanceByPortfolioWeekEnd(portfolioID, weekEndDate)
	if err == nil {
		return ToWeeklyPerformanceResponse(*existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return WeeklyPerformanceSnapshotResponse{}, err
	}

	snapshot, err := s.buildWeeklyPerformanceSnapshot(userID, portfolioID, weekStartDate, weekEndDate)
	if err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}

	if err := s.snapshotRepo.CreateWeeklyPerformance(snapshot); err != nil {
		if common.IsUniqueViolation(err) {
			existing, getErr := s.snapshotRepo.GetWeeklyPerformanceByPortfolioWeekEnd(portfolioID, weekEndDate)
			if getErr == nil {
				return ToWeeklyPerformanceResponse(*existing)
			}
			return WeeklyPerformanceSnapshotResponse{}, getErr
		}
		return WeeklyPerformanceSnapshotResponse{}, err
	}

	return ToWeeklyPerformanceResponse(*snapshot)
}

func (s *Service) List(userID uuid.UUID, portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}

	snapshots, err := s.snapshotRepo.ListByPortfolio(portfolioID, pagination)
	if err != nil {
		return nil, err
	}
	return ToResponses(snapshots)
}

func (s *Service) buildWeeklyPerformanceSnapshot(userID uuid.UUID, portfolioID uuid.UUID, weekStartDate time.Time, weekEndDate time.Time) (*WeeklyPerformanceSnapshot, error) {
	startSnapshot, err := s.getDailySnapshot(portfolioID, weekStartDate, "Week start daily snapshot not found")
	if err != nil {
		return nil, err
	}
	endSnapshot, err := s.getDailySnapshot(portfolioID, weekEndDate, "Week end daily snapshot not found")
	if err != nil {
		return nil, err
	}

	startResponse, err := ToResponse(*startSnapshot)
	if err != nil {
		return nil, err
	}
	endResponse, err := ToResponse(*endSnapshot)
	if err != nil {
		return nil, err
	}

	cashFlows, err := s.cashFlowRepo.ListExternalCashFlows(portfolioID, endOfUTCDay(weekStartDate), endOfUTCDay(weekEndDate))
	if err != nil {
		return nil, err
	}
	currencyReturns, err := calculateWeeklyCurrencyPerformance(
		weekStartDate,
		weekEndDate,
		totalsByCurrency(startResponse.TotalValues),
		totalsByCurrency(endResponse.TotalValues),
		externalCashFlowsByCurrency(cashFlows),
		netExternalCashFlowsByCurrency(cashFlows),
	)
	if err != nil {
		return nil, err
	}

	return buildWeeklyPerformanceSnapshotModel(userID, portfolioID, weekStartDate, weekEndDate, currencyReturns)
}

func (s *Service) buildDailySnapshot(userID uuid.UUID, portfolioID uuid.UUID, snapshotDate time.Time) (*PortfolioSnapshot, error) {
	asOf := endOfUTCDay(snapshotDate)
	records, err := s.ledgerRepo.ListLedgerEntriesAsOf(portfolioID, asOf)
	if err != nil {
		return nil, err
	}

	holdingsResult, err := finance.CalculateHoldings(holdings.ToFinanceEntries(records))
	if err != nil {
		return nil, common.BadRequest(err.Error())
	}

	priceRecords, err := s.priceRepo.ListLatestByAssetsAsOf(assetIDsFor(holdingsResult.AssetHoldings), asOf)
	if err != nil {
		return nil, err
	}

	valuationResult, err := finance.CalculatePortfolioValuation(holdingsResult, toFinancePrices(priceRecords))
	if err != nil {
		return nil, common.BadRequest(err.Error())
	}
	if len(valuationResult.TotalValues) == 0 {
		return nil, common.BadRequest("Snapshot requires at least one positive total value")
	}

	allocationResult, err := finance.CalculateAllocation(valuationResult)
	if err != nil {
		return nil, common.BadRequest(err.Error())
	}

	return buildSnapshotModel(userID, portfolioID, snapshotDate, holdingsResult, valuationResult, allocationResult)
}

func buildWeeklyPerformanceSnapshotModel(userID uuid.UUID, portfolioID uuid.UUID, weekStartDate time.Time, weekEndDate time.Time, currencyReturns []WeeklyCurrencyPerformance) (*WeeklyPerformanceSnapshot, error) {
	currencyReturnsJSON, err := NewJSONB(currencyReturns)
	if err != nil {
		return nil, err
	}
	pnlMetadata, err := NewJSONB(finance.PeriodPnLDefinition())
	if err != nil {
		return nil, err
	}
	cagrMetadata, err := NewJSONB(finance.CAGRDefinition())
	if err != nil {
		return nil, err
	}
	xirrMetadata, err := NewJSONB(finance.XIRRDefinition())
	if err != nil {
		return nil, err
	}

	return &WeeklyPerformanceSnapshot{
		PortfolioID:      portfolioID,
		WeekStartDate:    weekStartDate,
		WeekEndDate:      weekEndDate,
		CurrencyReturns:  currencyReturnsJSON,
		PerformanceScope: "Weekly performance is calculated from immutable daily snapshots at the UTC week boundary and external deposit/withdrawal cash flows within the week. Each currency is calculated separately without foreign exchange conversion.",
		PnLMetadata:      pnlMetadata,
		CAGRMetadata:     cagrMetadata,
		XIRRMetadata:     xirrMetadata,
		CreatedByUserID:  userID,
	}, nil
}

func buildSnapshotModel(userID uuid.UUID, portfolioID uuid.UUID, snapshotDate time.Time, holdingsResult finance.HoldingsResult, valuationResult finance.PortfolioValuationResult, allocationResult finance.AllocationResult) (*PortfolioSnapshot, error) {
	totalValues, err := NewJSONB(valuationResult.TotalValues)
	if err != nil {
		return nil, err
	}
	assetAllocations, err := NewJSONB(allocationResult.AssetAllocations)
	if err != nil {
		return nil, err
	}
	assetClassAllocations, err := NewJSONB(allocationResult.AssetClassAllocations)
	if err != nil {
		return nil, err
	}
	cashAllocations, err := NewJSONB(allocationResult.CashAllocations)
	if err != nil {
		return nil, err
	}
	missingPrices, err := NewJSONB(valuationResult.MissingPrices)
	if err != nil {
		return nil, err
	}
	valuationMetadata, err := NewJSONB(valuationResult.Definition)
	if err != nil {
		return nil, err
	}
	allocationMetadata, err := NewJSONB(allocationResult.Definition)
	if err != nil {
		return nil, err
	}
	holdingsMetadata, err := NewJSONB(holdingsResult.Definition)
	if err != nil {
		return nil, err
	}

	return &PortfolioSnapshot{
		PortfolioID:           portfolioID,
		SnapshotDate:          snapshotDate,
		SnapshotPeriod:        SnapshotPeriodDaily,
		TotalValues:           totalValues,
		AssetAllocations:      assetAllocations,
		AssetClassAllocations: assetClassAllocations,
		CashAllocations:       cashAllocations,
		MissingPrices:         missingPrices,
		IsFullyValued:         valuationResult.IsFullyValued,
		ValuationScope:        valuationResult.ValuationScope,
		AllocationScope:       allocationResult.AllocationScope,
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       userID,
	}, nil
}

func (s *Service) getOwnedPortfolio(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Portfolio not found")
	}
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (s *Service) getDailySnapshot(portfolioID uuid.UUID, snapshotDate time.Time, notFoundMessage string) (*PortfolioSnapshot, error) {
	snapshot, err := s.snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, SnapshotPeriodDaily)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound(notFoundMessage)
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func calculateWeeklyCurrencyPerformance(startDate time.Time, endDate time.Time, startTotals map[string]decimal.Decimal, endTotals map[string]decimal.Decimal, cashFlowsByCurrency map[string][]finance.CashFlow, netExternalCashFlows map[string]decimal.Decimal) ([]WeeklyCurrencyPerformance, error) {
	currencies := make([]string, 0)
	for currency, beginningValue := range startTotals {
		endingValue, ok := endTotals[currency]
		if !ok {
			continue
		}
		if !beginningValue.GreaterThan(decimal.Zero) || !endingValue.GreaterThan(decimal.Zero) {
			continue
		}
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	if len(currencies) == 0 {
		return nil, common.BadRequest("Weekly performance requires at least one currency with positive totals in both week boundary snapshots")
	}

	responses := make([]WeeklyCurrencyPerformance, 0, len(currencies))
	for _, currency := range currencies {
		beginningValue := startTotals[currency]
		endingValue := endTotals[currency]
		netExternalCashFlow := netExternalCashFlows[currency]

		pnl, err := finance.CalculatePeriodPnL(finance.PeriodPnLInput{
			BeginningValue:      beginningValue,
			EndingValue:         endingValue,
			NetExternalCashFlow: netExternalCashFlow,
		})
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Weekly PnL error for currency %s: %s", currency, err.Error()))
		}

		cagr, err := finance.CalculateCAGR(finance.CAGRInput{
			BeginningValue: beginningValue,
			EndingValue:    endingValue,
			StartDate:      startDate,
			EndDate:        endDate,
		})
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Weekly CAGR error for currency %s: %s", currency, err.Error()))
		}

		xirrFlows := make([]finance.CashFlow, 0, len(cashFlowsByCurrency[currency])+2)
		xirrFlows = append(xirrFlows, finance.CashFlow{Date: startDate, Amount: beginningValue.Neg()})
		xirrFlows = append(xirrFlows, cashFlowsByCurrency[currency]...)
		xirrFlows = append(xirrFlows, finance.CashFlow{Date: endDate, Amount: endingValue})

		xirr, err := finance.CalculateXIRR(xirrFlows)
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Weekly XIRR error for currency %s: %s", currency, err.Error()))
		}

		responses = append(responses, WeeklyCurrencyPerformance{
			Currency:            currency,
			BeginningValue:      beginningValue,
			EndingValue:         endingValue,
			NetExternalCashFlow: netExternalCashFlow,
			ProfitLoss:          pnl.Amount,
			CAGR:                cagr.Rate,
			XIRR:                xirr.Rate,
			CashFlowCount:       len(cashFlowsByCurrency[currency]),
		})
	}

	return responses, nil
}

func totalsByCurrency(values []finance.CurrencyValue) map[string]decimal.Decimal {
	totals := make(map[string]decimal.Decimal, len(values))
	for _, value := range values {
		totals[value.Currency] = value.Amount
	}
	return totals
}

func externalCashFlowsByCurrency(cashFlows []transactions.ExternalCashFlowRecord) map[string][]finance.CashFlow {
	byCurrency := make(map[string][]finance.CashFlow)
	for _, cashFlow := range cashFlows {
		byCurrency[cashFlow.Currency] = append(byCurrency[cashFlow.Currency], finance.CashFlow{
			Date:   cashFlow.OccurredAt,
			Amount: cashFlow.Amount,
		})
	}
	return byCurrency
}

func netExternalCashFlowsByCurrency(cashFlows []transactions.ExternalCashFlowRecord) map[string]decimal.Decimal {
	netByCurrency := make(map[string]decimal.Decimal)
	for _, cashFlow := range cashFlows {
		netByCurrency[cashFlow.Currency] = netByCurrency[cashFlow.Currency].Add(cashFlow.Amount)
	}
	return netByCurrency
}

func parseSnapshotDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, common.BadRequest("Snapshot date is required")
	}

	date, err := time.Parse(snapshotDateLayout, raw)
	if err != nil {
		return time.Time{}, common.BadRequest("Snapshot date must use YYYY-MM-DD format")
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, common.BadRequest("Snapshot date cannot be in the future")
	}
	return date, nil
}

func endOfUTCDay(date time.Time) time.Time {
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func assetIDsFor(assetHoldings []finance.AssetHolding) []uuid.UUID {
	assetIDs := make([]uuid.UUID, 0, len(assetHoldings))
	seen := make(map[uuid.UUID]struct{}, len(assetHoldings))
	for _, holding := range assetHoldings {
		assetID, err := uuid.Parse(holding.AssetID)
		if err != nil {
			continue
		}
		if _, ok := seen[assetID]; ok {
			continue
		}
		seen[assetID] = struct{}{}
		assetIDs = append(assetIDs, assetID)
	}
	return assetIDs
}

func toFinancePrices(priceRecords []prices.AssetPrice) []finance.ValuationPrice {
	valuationPrices := make([]finance.ValuationPrice, 0, len(priceRecords))
	for _, price := range priceRecords {
		valuationPrices = append(valuationPrices, finance.ValuationPrice{
			AssetID:  price.AssetID.String(),
			Price:    price.Price,
			Currency: price.Currency,
			PricedAt: price.PricedAt,
		})
	}
	return valuationPrices
}
