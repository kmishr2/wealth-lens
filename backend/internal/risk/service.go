package risk

import (
	"errors"
	"fmt"
	"sort"
	"strings"
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

type snapshotRangeReader interface {
	ListByPortfolioDateRange(portfolioID uuid.UUID, startDate time.Time, endDate time.Time, snapshotPeriod string) ([]snapshots.PortfolioSnapshot, error)
}

type externalCashFlowReader interface {
	ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]transactions.ExternalCashFlowRecord, error)
}

type Service struct {
	portfolioRepo portfolioReader
	snapshotRepo  snapshotRangeReader
	cashFlowRepo  externalCashFlowReader
}

func NewService(portfolioRepo portfolioReader, snapshotRepo snapshotRangeReader, cashFlowRepo externalCashFlowReader) *Service {
	return &Service{
		portfolioRepo: portfolioRepo,
		snapshotRepo:  snapshotRepo,
		cashFlowRepo:  cashFlowRepo,
	}
}

func (s *Service) Get(userID uuid.UUID, portfolioID uuid.UUID, rawStartDate string, rawEndDate string, rawPeriodsPerYear string) (PortfolioRiskResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return PortfolioRiskResponse{}, err
	}

	startDate, endDate, err := parseDateRange(rawStartDate, rawEndDate)
	if err != nil {
		return PortfolioRiskResponse{}, err
	}
	periodsPerYear, err := parsePeriodsPerYear(rawPeriodsPerYear)
	if err != nil {
		return PortfolioRiskResponse{}, err
	}

	snapshotRecords, err := s.snapshotRepo.ListByPortfolioDateRange(portfolioID, startDate, endDate, snapshots.SnapshotPeriodDaily)
	if err != nil {
		return PortfolioRiskResponse{}, err
	}
	if len(snapshotRecords) < 3 {
		return PortfolioRiskResponse{}, common.BadRequest("Risk calculation requires at least three daily snapshots in the requested range")
	}

	snapshotResponses, err := snapshots.ToResponses(snapshotRecords)
	if err != nil {
		return PortfolioRiskResponse{}, err
	}
	for _, snapshot := range snapshotResponses {
		if !snapshot.IsFullyValued {
			return PortfolioRiskResponse{}, common.BadRequest(fmt.Sprintf("Snapshot %s is not fully valued", snapshot.SnapshotDate))
		}
	}

	cashFlows, err := s.cashFlowRepo.ListExternalCashFlows(portfolioID, endOfUTCDay(startDate), endOfUTCDay(endDate))
	if err != nil {
		return PortfolioRiskResponse{}, err
	}

	currencyRisk, err := calculateCurrencyRisk(snapshotResponses, cashFlows, periodsPerYear)
	if err != nil {
		return PortfolioRiskResponse{}, err
	}

	return PortfolioRiskResponse{
		PortfolioID:        portfolioID,
		StartDate:          formatRiskDate(startDate),
		EndDate:            formatRiskDate(endDate),
		PeriodsPerYear:     periodsPerYear,
		CurrencyRisk:       currencyRisk,
		RiskScope:          "Risk metrics use immutable daily snapshots and cash-flow-adjusted returns, calculated separately per currency without foreign exchange conversion.",
		VolatilityMetadata: finance.VolatilityDefinition(),
		DrawdownMetadata:   finance.DrawdownDefinition(),
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

func calculateCurrencyRisk(snapshotResponses []snapshots.PortfolioSnapshotResponse, cashFlows []transactions.ExternalCashFlowRecord, periodsPerYear decimal.Decimal) ([]CurrencyRiskResponse, error) {
	currencies := currenciesPresentInEverySnapshot(snapshotResponses)
	if len(currencies) == 0 {
		return nil, common.BadRequest("Risk calculation requires at least one currency with positive values in every snapshot")
	}

	sort.SliceStable(cashFlows, func(i, j int) bool {
		return cashFlows[i].OccurredAt.Before(cashFlows[j].OccurredAt)
	})

	responses := make([]CurrencyRiskResponse, 0, len(currencies))
	for _, currency := range currencies {
		points := make([]finance.PortfolioValuePoint, 0, len(snapshotResponses))
		for index, snapshot := range snapshotResponses {
			snapshotDate, err := time.Parse(riskDateLayout, snapshot.SnapshotDate)
			if err != nil {
				return nil, err
			}

			point := finance.PortfolioValuePoint{
				Date:  snapshotDate,
				Value: totalForCurrency(snapshot.TotalValues, currency),
			}
			if index > 0 {
				previousDate, err := time.Parse(riskDateLayout, snapshotResponses[index-1].SnapshotDate)
				if err != nil {
					return nil, err
				}
				point.NetExternalCashFlow = externalCashFlowForInterval(
					cashFlows,
					currency,
					endOfUTCDay(previousDate),
					endOfUTCDay(snapshotDate),
				)
			}
			points = append(points, point)
		}

		volatility, err := finance.CalculateVolatility(finance.VolatilityInput{
			Values:         points,
			PeriodsPerYear: periodsPerYear,
		})
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Volatility error for currency %s: %s", currency, err.Error()))
		}
		drawdown, err := finance.CalculateMaximumDrawdown(points)
		if err != nil {
			return nil, common.BadRequest(fmt.Sprintf("Drawdown error for currency %s: %s", currency, err.Error()))
		}

		responses = append(responses, CurrencyRiskResponse{
			Currency:             currency,
			ObservationCount:     len(points),
			PeriodicReturnCount:  volatility.PeriodicReturnCount,
			AnnualizedVolatility: volatility.AnnualizedVolatility,
			MaximumDrawdown:      drawdown.MaximumDrawdown,
			PeakDate:             formatRiskDate(drawdown.PeakDate),
			TroughDate:           formatRiskDate(drawdown.TroughDate),
		})
	}
	return responses, nil
}

func currenciesPresentInEverySnapshot(snapshotResponses []snapshots.PortfolioSnapshotResponse) []string {
	counts := make(map[string]int)
	for _, snapshot := range snapshotResponses {
		for _, total := range snapshot.TotalValues {
			if total.Amount.GreaterThan(decimal.Zero) {
				counts[total.Currency]++
			}
		}
	}

	currencies := make([]string, 0)
	for currency, count := range counts {
		if count == len(snapshotResponses) {
			currencies = append(currencies, currency)
		}
	}
	sort.Strings(currencies)
	return currencies
}

func totalForCurrency(totals []finance.CurrencyValue, currency string) decimal.Decimal {
	for _, total := range totals {
		if total.Currency == currency {
			return total.Amount
		}
	}
	return decimal.Zero
}

func externalCashFlowForInterval(records []transactions.ExternalCashFlowRecord, currency string, startAfter time.Time, endAt time.Time) decimal.Decimal {
	total := decimal.Zero
	for _, record := range records {
		if record.Currency != currency {
			continue
		}
		if record.OccurredAt.After(startAfter) && !record.OccurredAt.After(endAt) {
			total = total.Add(record.Amount)
		}
	}
	return total
}

func parseDateRange(rawStartDate string, rawEndDate string) (time.Time, time.Time, error) {
	startDate, err := parseRiskDate(rawStartDate, "Start date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endDate, err := parseRiskDate(rawEndDate, "End date")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !endDate.After(startDate) {
		return time.Time{}, time.Time{}, common.BadRequest("End date must be after start date")
	}
	return startDate, endDate, nil
}

func parseRiskDate(raw string, label string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s is required", label))
	}
	date, err := time.Parse(riskDateLayout, raw)
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

func parsePeriodsPerYear(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, common.BadRequest("Periods per year is required")
	}
	value, err := decimal.NewFromString(raw)
	if err != nil || !value.GreaterThan(decimal.Zero) {
		return decimal.Zero, common.BadRequest("Periods per year must be a number greater than zero")
	}
	return value, nil
}

func endOfUTCDay(date time.Time) time.Time {
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, 1).Add(-time.Nanosecond)
}
