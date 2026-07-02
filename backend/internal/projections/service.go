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
	if err := s.requireOwnedPortfolio(userID, portfolioID); err != nil {
		return SIPResponse{}, err
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return SIPResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	input, err := sipInput(req)
	if err != nil {
		return SIPResponse{}, err
	}
	result, err := finance.CalculateSIPProjection(input)
	if err != nil {
		return SIPResponse{}, common.BadRequest(err.Error())
	}
	return SIPResponse{PortfolioID: portfolioID, Currency: currency, SIPProjectionResult: result,
		Scope: "This stateless deterministic scenario uses only the supplied amounts, rates, currency, and horizon. It is not stored as portfolio state or presented as a prediction."}, nil
}

func (s *Service) CompareWhatIf(userID, portfolioID uuid.UUID, req WhatIfRequest) (WhatIfResponse, error) {
	if err := s.requireOwnedPortfolio(userID, portfolioID); err != nil {
		return WhatIfResponse{}, err
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return WhatIfResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	scenarios := make([]finance.NamedSIPScenario, 0, len(req.Scenarios))
	for index, scenario := range req.Scenarios {
		input, err := sipInput(SIPRequest{InitialInvestment: scenario.Input.InitialInvestment, MonthlyContribution: scenario.Input.MonthlyContribution,
			AnnualReturnPercentage: scenario.Input.AnnualReturnPercentage, AnnualInflationPercentage: scenario.Input.AnnualInflationPercentage, Months: scenario.Input.Months})
		if err != nil {
			return WhatIfResponse{}, common.BadRequest("Scenario input error at index " + decimal.NewFromInt(int64(index)).String() + ": " + err.Error())
		}
		scenarios = append(scenarios, finance.NamedSIPScenario{Name: scenario.Name, Input: input})
	}
	result, err := finance.CompareSIPScenarios(scenarios)
	if err != nil {
		return WhatIfResponse{}, common.BadRequest(err.Error())
	}
	return WhatIfResponse{PortfolioID: portfolioID, Currency: currency, WhatIfComparisonResult: result,
		Scope: "This stateless comparison changes only explicit scenario inputs and reports arithmetic differences from the first scenario; no result is stored or predicted."}, nil
}

func (s *Service) requireOwnedPortfolio(userID, portfolioID uuid.UUID) error {
	if _, err := s.portfolios.GetOwned(userID, portfolioID); errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NotFound("Portfolio not found")
	} else if err != nil {
		return err
	}
	return nil
}

func sipInput(req SIPRequest) (finance.SIPProjectionInput, error) {
	if req.AnnualReturnPercentage == nil {
		return finance.SIPProjectionInput{}, common.BadRequest("Annual return percentage is required")
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
	return finance.SIPProjectionInput{InitialInvestment: initial, MonthlyContribution: monthly,
		AnnualReturnPercentage: *req.AnnualReturnPercentage, AnnualInflationPercentage: inflation, Months: req.Months}, nil
}
