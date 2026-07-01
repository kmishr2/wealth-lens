package projections

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type Service struct{ portfolios portfolioReader }

func NewService(portfolios portfolioReader) *Service { return &Service{portfolios: portfolios} }

func (s *Service) CalculateSIP(userID, portfolioID uuid.UUID, req SIPRequest) (SIPResponse, error) {
	if _, err := s.portfolios.GetOwned(userID, portfolioID); errors.Is(err, gorm.ErrRecordNotFound) {
		return SIPResponse{}, common.NotFound("Portfolio not found")
	} else if err != nil {
		return SIPResponse{}, err
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return SIPResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	if req.AnnualReturnPercentage == nil {
		return SIPResponse{}, common.BadRequest("Annual return percentage is required")
	}
	initial := decimal.Zero
	if req.InitialInvestment != nil {
		initial = *req.InitialInvestment
	}
	monthly := decimal.Zero
	if req.MonthlyContribution != nil {
		monthly = *req.MonthlyContribution
	}
	inflation := decimal.Zero
	if req.AnnualInflationPercentage != nil {
		inflation = *req.AnnualInflationPercentage
	}
	result, err := finance.CalculateSIPProjection(finance.SIPProjectionInput{InitialInvestment: initial, MonthlyContribution: monthly,
		AnnualReturnPercentage: *req.AnnualReturnPercentage, AnnualInflationPercentage: inflation, Months: req.Months})
	if err != nil {
		return SIPResponse{}, common.BadRequest(err.Error())
	}
	return SIPResponse{PortfolioID: portfolioID, Currency: currency, SIPProjectionResult: result,
		Scope: "This stateless deterministic scenario uses only the supplied amounts, rates, currency, and horizon. It is not stored as portfolio state or presented as a prediction."}, nil
}
