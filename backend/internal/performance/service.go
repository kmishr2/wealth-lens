package performance

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type snapshotReader interface {
	GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*snapshots.PortfolioSnapshot, error)
}

type externalCashFlowReader interface {
	ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error)
}

type Service struct {
	portfolioRepo portfolioReader
	snapshotRepo  snapshotReader
	cashFlowRepo  externalCashFlowReader
}

func NewService(portfolioRepo portfolioReader, snapshotRepo snapshotReader, cashFlowRepo externalCashFlowReader) *Service {
	return &Service{
		portfolioRepo: portfolioRepo,
		snapshotRepo:  snapshotRepo,
		cashFlowRepo:  cashFlowRepo,
	}
}

func (s *Service) Get(userID uuid.UUID, portfolioID uuid.UUID, rawStartDate string, rawEndDate string) (PortfolioPerformanceResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return PortfolioPerformanceResponse{}, err
	}

	startDate, endDate, err := parseDateRange(rawStartDate, rawEndDate)
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}

	startSnapshot, err := s.getSnapshot(portfolioID, startDate, "Start snapshot not found")
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}
	endSnapshot, err := s.getSnapshot(portfolioID, endDate, "End snapshot not found")
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}

	startResponse, err := snapshots.ToResponse(*startSnapshot)
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}
	endResponse, err := snapshots.ToResponse(*endSnapshot)
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}

	startTotals := totalsByCurrency(startResponse.TotalValues)
	endTotals := totalsByCurrency(endResponse.TotalValues)
	cashFlows, err := s.cashFlowRepo.ListExternalCashFlows(portfolioID, endOfUTCDay(startDate), endOfUTCDay(endDate))
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}
	cashFlowsByCurrency := externalCashFlowsByCurrency(cashFlows)
	netExternalCashFlows := netExternalCashFlowsByCurrency(cashFlows)

	currencyReturns, err := calculateCurrencyReturns(startDate, endDate, startTotals, endTotals, cashFlowsByCurrency, netExternalCashFlows)
	if err != nil {
		return PortfolioPerformanceResponse{}, err
	}

	return PortfolioPerformanceResponse{
		PortfolioID:      portfolioID,
		StartDate:        dateString(startDate),
		EndDate:          dateString(endDate),
		CurrencyReturns:  currencyReturns,
		PerformanceScope: "Performance is calculated separately per currency from immutable daily snapshots and external deposit/withdrawal cash flows; no currency conversion is applied.",
		PnLMetadata:      finance.PeriodPnLDefinition(),
		CAGRMetadata:     finance.CAGRDefinition(),
		XIRRMetadata:     finance.XIRRDefinition(),
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

func (s *Service) getSnapshot(portfolioID uuid.UUID, snapshotDate time.Time, notFoundMessage string) (*snapshots.PortfolioSnapshot, error) {
	snapshot, err := s.snapshotRepo.GetByPortfolioDatePeriod(portfolioID, snapshotDate, snapshots.SnapshotPeriodDaily)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound(notFoundMessage)
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func calculateCurrencyReturns(startDate time.Time, endDate time.Time, startTotals map[string]decimal.Decimal, endTotals map[string]decimal.Decimal, cashFlowsByCurrency map[string][]finance.CashFlow, netExternalCashFlows map[string]decimal.Decimal) ([]CurrencyPerformanceResponse, error) {
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
		return nil, common.BadRequest("Performance requires at least one currency with positive totals in both start and end snapshots")
	}

	responses := make([]CurrencyPerformanceResponse, 0, len(currencies))
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
			return nil, common.BadRequest(fmt.Sprintf("PnL error for currency %s: %s", currency, err.Error()))
		}

		cagr, err := finance.CalculateCAGR(finance.CAGRInput{
			BeginningValue: beginningValue,
			EndingValue:    endingValue,
			StartDate:      startDate,
			EndDate:        endDate,
		})
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("CAGR error for currency %s: %s", currency, err.Error()))
		}

		xirrFlows := make([]finance.CashFlow, 0, len(cashFlowsByCurrency[currency])+2)
		xirrFlows = append(xirrFlows, finance.CashFlow{Date: startDate, Amount: beginningValue.Neg()})
		xirrFlows = append(xirrFlows, cashFlowsByCurrency[currency]...)
		xirrFlows = append(xirrFlows, finance.CashFlow{Date: endDate, Amount: endingValue})

		xirr, err := finance.CalculateXIRR(xirrFlows)
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("XIRR error for currency %s: %s", currency, err.Error()))
		}

		responses = append(responses, CurrencyPerformanceResponse{
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

func parseDateRange(rawStartDate string, rawEndDate string) (time.Time, time.Time, error) {
	startDate, err := parsePerformanceDate(rawStartDate, "Start date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := parsePerformanceDate(rawEndDate, "End date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !endDate.After(startDate) {
		return time.Time{}, time.Time{}, common.BadRequest("End date must be after start date")
	}
	return startDate, endDate, nil
}

func parsePerformanceDate(raw string, label string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s is required", label))
	}

	date, err := time.Parse(performanceDateLayout, raw)
	if err != nil {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s must use YYYY-MM-DD format", label))
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s cannot be in the future", label))
	}
	return date, nil
}

func totalsByCurrency(totalValues []finance.CurrencyValue) map[string]decimal.Decimal {
	totals := make(map[string]decimal.Decimal, len(totalValues))
	for _, total := range totalValues {
		totals[total.Currency] = total.Amount
	}
	return totals
}

func externalCashFlowsByCurrency(records []transactions.ExternalCashFlowRecord) map[string][]finance.CashFlow {
	flowsByCurrency := make(map[string][]finance.CashFlow)
	for _, record := range records {
		flowsByCurrency[record.Currency] = append(flowsByCurrency[record.Currency], finance.CashFlow{
			Date:   record.OccurredAt,
			Amount: record.Amount.Neg(),
		})
	}
	return flowsByCurrency
}

func netExternalCashFlowsByCurrency(records []transactions.ExternalCashFlowRecord) map[string]decimal.Decimal {
	totals := make(map[string]decimal.Decimal)
	for _, record := range records {
		totals[record.Currency] = totals[record.Currency].Add(record.Amount)
	}
	return totals
}

func endOfUTCDay(date time.Time) time.Time {
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, 1).Add(-time.Nanosecond)
}
