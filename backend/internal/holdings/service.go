package holdings

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"gorm.io/gorm"
)

type ledgerEntryReader interface {
	ListLedgerEntries(portfolioID uuid.UUID) ([]LedgerEntryRecord, error)
}

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type Service struct {
	repo          ledgerEntryReader
	portfolioRepo portfolioReader
}

func NewService(repo ledgerEntryReader, portfolioRepo portfolioReader) *Service {
	return &Service{
		repo:          repo,
		portfolioRepo: portfolioRepo,
	}
}

func (s *Service) GetCurrent(userID uuid.UUID, portfolioID uuid.UUID) (HoldingsResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return HoldingsResponse{}, err
	}

	records, err := s.repo.ListLedgerEntries(portfolioID)
	if err != nil {
		return HoldingsResponse{}, err
	}

	result, err := finance.CalculateHoldings(ToFinanceEntries(records))
	if err != nil {
		return HoldingsResponse{}, common.BadRequest(err.Error())
	}

	return ToResponse(portfolioID, result), nil
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

func ToFinanceEntries(records []LedgerEntryRecord) []finance.LedgerEntry {
	entries := make([]finance.LedgerEntry, 0, len(records))
	for _, record := range records {
		entries = append(entries, finance.LedgerEntry{
			EntryKind:   finance.LedgerEntryKind(record.EntryKind),
			AssetID:     record.AssetID,
			AssetSymbol: record.AssetSymbol,
			AssetName:   record.AssetName,
			AssetClass:  record.AssetClass,
			Quantity:    record.Quantity,
			Amount:      record.Amount,
			Currency:    record.Currency,
		})
	}
	return entries
}
