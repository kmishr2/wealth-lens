package allocations

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"gorm.io/gorm"
)

type ledgerEntryReader interface {
	ListLedgerEntries(portfolioID uuid.UUID) ([]holdings.LedgerEntryRecord, error)
}

type latestPriceReader interface {
	ListLatestByAssets(assetIDs []uuid.UUID) ([]prices.AssetPrice, error)
}

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type Service struct {
	ledgerRepo    ledgerEntryReader
	priceRepo     latestPriceReader
	portfolioRepo portfolioReader
}

func NewService(ledgerRepo ledgerEntryReader, priceRepo latestPriceReader, portfolioRepo portfolioReader) *Service {
	return &Service{
		ledgerRepo:    ledgerRepo,
		priceRepo:     priceRepo,
		portfolioRepo: portfolioRepo,
	}
}

func (s *Service) GetCurrent(userID uuid.UUID, portfolioID uuid.UUID) (PortfolioAllocationResponse, error) {
	holdingsResult, valuationResult, allocationResult, err := s.calculateCurrent(userID, portfolioID)
	if err != nil {
		return PortfolioAllocationResponse{}, err
	}
	return ToResponse(portfolioID, holdingsResult, valuationResult, allocationResult), nil
}

func (s *Service) GetConcentration(userID uuid.UUID, portfolioID uuid.UUID) (PortfolioConcentrationResponse, error) {
	holdingsResult, valuationResult, allocationResult, err := s.calculateCurrent(userID, portfolioID)
	if err != nil {
		return PortfolioConcentrationResponse{}, err
	}
	concentration, err := finance.CalculateConcentration(allocationResult)
	if err != nil {
		return PortfolioConcentrationResponse{}, common.BadRequest(err.Error())
	}
	return ToConcentrationResponse(portfolioID, holdingsResult, valuationResult, allocationResult, concentration), nil
}

func (s *Service) CalculateRebalancing(userID uuid.UUID, portfolioID uuid.UUID, req RebalancingRequest) (PortfolioRebalancingResponse, error) {
	holdingsResult, valuationResult, allocationResult, err := s.calculateCurrent(userID, portfolioID)
	if err != nil {
		return PortfolioRebalancingResponse{}, err
	}

	rebalancingResult, err := finance.CalculateRebalancing(finance.RebalancingInput{
		CurrentAllocation:        allocationResult,
		Targets:                  req.Targets,
		DriftTolerancePercentage: req.DriftTolerancePercentage,
	})
	if err != nil {
		return PortfolioRebalancingResponse{}, common.BadRequest(err.Error())
	}

	return ToRebalancingResponse(portfolioID, holdingsResult, valuationResult, allocationResult, rebalancingResult), nil
}

func (s *Service) calculateCurrent(userID uuid.UUID, portfolioID uuid.UUID) (finance.HoldingsResult, finance.PortfolioValuationResult, finance.AllocationResult, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, err
	}

	records, err := s.ledgerRepo.ListLedgerEntries(portfolioID)
	if err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, err
	}

	holdingsResult, err := finance.CalculateHoldings(holdings.ToFinanceEntries(records))
	if err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, common.BadRequest(err.Error())
	}

	priceRecords, err := s.priceRepo.ListLatestByAssets(assetIDsFor(holdingsResult.AssetHoldings))
	if err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, err
	}

	valuationResult, err := finance.CalculatePortfolioValuation(holdingsResult, toFinancePrices(priceRecords))
	if err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, common.BadRequest(err.Error())
	}

	allocationResult, err := finance.CalculateAllocation(valuationResult)
	if err != nil {
		return finance.HoldingsResult{}, finance.PortfolioValuationResult{}, finance.AllocationResult{}, common.BadRequest(err.Error())
	}

	return holdingsResult, valuationResult, allocationResult, nil
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
