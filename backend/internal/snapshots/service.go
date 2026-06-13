package snapshots

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/holdings"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
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

type snapshotReaderWriter interface {
	Create(snapshot *PortfolioSnapshot) error
	GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*PortfolioSnapshot, error)
	ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshot, error)
}

type Service struct {
	snapshotRepo  snapshotReaderWriter
	ledgerRepo    ledgerEntryReader
	priceRepo     latestPriceReader
	portfolioRepo portfolioReader
}

func NewService(snapshotRepo snapshotReaderWriter, ledgerRepo ledgerEntryReader, priceRepo latestPriceReader, portfolioRepo portfolioReader) *Service {
	return &Service{
		snapshotRepo:  snapshotRepo,
		ledgerRepo:    ledgerRepo,
		priceRepo:     priceRepo,
		portfolioRepo: portfolioRepo,
	}
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
