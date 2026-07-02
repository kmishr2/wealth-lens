package contributions

import (
	"errors"
	"fmt"
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

const dateLayout = "2006-01-02"

type portfolioReader interface {
	GetOwned(uuid.UUID, uuid.UUID) (*portfolios.Portfolio, error)
}
type snapshotReader interface {
	GetByPortfolioDatePeriod(uuid.UUID, time.Time, string) (*snapshots.PortfolioSnapshot, error)
}
type cashFlowReader interface {
	ListExternalCashFlows(uuid.UUID, time.Time, time.Time) ([]transactions.ExternalCashFlowRecord, error)
}

type Service struct {
	portfolios portfolioReader
	snapshots  snapshotReader
	cashFlows  cashFlowReader
}

func NewService(portfolios portfolioReader, snapshots snapshotReader, cashFlows cashFlowReader) *Service {
	return &Service{portfolios: portfolios, snapshots: snapshots, cashFlows: cashFlows}
}

func (s *Service) Get(userID, portfolioID uuid.UUID, rawStart, rawEnd, rawCurrency string) (Response, error) {
	if _, err := s.portfolios.GetOwned(userID, portfolioID); errors.Is(err, gorm.ErrRecordNotFound) {
		return Response{}, common.NotFound("Portfolio not found")
	} else if err != nil {
		return Response{}, err
	}
	start, end, err := parseDateRange(rawStart, rawEnd)
	if err != nil {
		return Response{}, err
	}
	currency := common.NormalizeCurrency(rawCurrency)
	if !common.ValidateCurrency(currency) {
		return Response{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	startSnapshot, err := s.getSnapshot(portfolioID, start, "Start daily snapshot not found")
	if err != nil {
		return Response{}, err
	}
	endSnapshot, err := s.getSnapshot(portfolioID, end, "End daily snapshot not found")
	if err != nil {
		return Response{}, err
	}
	startResponse, err := snapshots.ToResponse(*startSnapshot)
	if err != nil {
		return Response{}, err
	}
	endResponse, err := snapshots.ToResponse(*endSnapshot)
	if err != nil {
		return Response{}, err
	}
	if !startResponse.IsFullyValued || !endResponse.IsFullyValued {
		return Response{}, common.BadRequest("Contribution analysis requires fully valued boundary snapshots")
	}
	beginning, ok := totalForCurrency(startResponse.TotalValues, currency)
	if !ok {
		return Response{}, common.BadRequest("Start snapshot does not contain currency " + currency)
	}
	ending, ok := totalForCurrency(endResponse.TotalValues, currency)
	if !ok {
		return Response{}, common.BadRequest("End snapshot does not contain currency " + currency)
	}
	records, err := s.cashFlows.ListExternalCashFlows(portfolioID, endOfDay(start), endOfDay(end))
	if err != nil {
		return Response{}, err
	}
	flows := make([]finance.ExternalContribution, 0, len(records))
	for _, record := range records {
		if record.Currency == currency {
			flows = append(flows, finance.ExternalContribution{Date: record.OccurredAt, Amount: record.Amount})
		}
	}
	result, err := finance.CalculateContributionAnalysis(finance.ContributionAnalysisInput{BeginningValue: beginning, EndingValue: ending,
		StartDate: start, EndDate: end, CashFlows: flows})
	if err != nil {
		return Response{}, common.BadRequest(fmt.Sprintf("Contribution analysis error: %s", err.Error()))
	}
	return Response{PortfolioID: portfolioID, Currency: currency, StartDate: start.Format(dateLayout), EndDate: end.Format(dateLayout),
		ContributionAnalysisResult: result,
		Scope:                      "Analysis uses exact immutable daily snapshot values and ledger-derived deposits and withdrawals for one currency; it does not interpolate values or convert currencies."}, nil
}

func (s *Service) getSnapshot(portfolioID uuid.UUID, date time.Time, message string) (*snapshots.PortfolioSnapshot, error) {
	record, err := s.snapshots.GetByPortfolioDatePeriod(portfolioID, date, snapshots.SnapshotPeriodDaily)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound(message)
	}
	return record, err
}

func parseDateRange(rawStart, rawEnd string) (time.Time, time.Time, error) {
	start, err := time.Parse(dateLayout, rawStart)
	if err != nil {
		return time.Time{}, time.Time{}, common.BadRequest("Start date must use YYYY-MM-DD format")
	}
	end, err := time.Parse(dateLayout, rawEnd)
	if err != nil {
		return time.Time{}, time.Time{}, common.BadRequest("End date must use YYYY-MM-DD format")
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, common.BadRequest("End date must be after start date")
	}
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if end.After(today) {
		return time.Time{}, time.Time{}, common.BadRequest("End date cannot be in the future")
	}
	return start, end, nil
}

func totalForCurrency(values []finance.CurrencyValue, currency string) (decimal.Decimal, bool) {
	for _, value := range values {
		if value.Currency == currency {
			return value.Amount, true
		}
	}
	return decimal.Zero, false
}

func endOfDay(date time.Time) time.Time {
	date = time.Date(date.UTC().Year(), date.UTC().Month(), date.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, 1).Add(-time.Nanosecond)
}
