package transactions

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/accounts"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Service struct {
	repo          *Repository
	portfolioRepo *portfolios.Repository
	accountRepo   *accounts.Repository
	assetRepo     *assets.Repository
}

func NewService(repo *Repository, portfolioRepo *portfolios.Repository, accountRepo *accounts.Repository, assetRepo *assets.Repository) *Service {
	return &Service{
		repo:          repo,
		portfolioRepo: portfolioRepo,
		accountRepo:   accountRepo,
		assetRepo:     assetRepo,
	}
}

func (s *Service) Create(userID uuid.UUID, portfolioID uuid.UUID, req TransactionCreateRequest) (TransactionResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return TransactionResponse{}, err
	}

	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(portfolioID, idempotencyKey)
		if err == nil {
			return ToResponse(*existing), nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return TransactionResponse{}, err
		}
	}

	transaction, err := s.buildTransaction(userID, portfolioID, req, nil, nil, nil)
	if err != nil {
		return TransactionResponse{}, err
	}

	if err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		return s.repo.CreateWithDB(tx, transaction)
	}); err != nil {
		if common.IsUniqueViolation(err) {
			return TransactionResponse{}, common.Conflict("Transaction already exists for idempotency key")
		}
		return TransactionResponse{}, err
	}

	return ToResponse(*transaction), nil
}

func (s *Service) List(userID uuid.UUID, portfolioID uuid.UUID, pagination common.Pagination) ([]TransactionResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}

	transactions, err := s.repo.ListOwned(userID, portfolioID, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, ToResponse(transaction))
	}
	return responses, nil
}

func (s *Service) Get(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID) (TransactionResponse, error) {
	transaction, err := s.getOwnedTransaction(userID, portfolioID, transactionID)
	if err != nil {
		return TransactionResponse{}, err
	}
	return ToResponse(*transaction), nil
}

func (s *Service) Reverse(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID, req TransactionReversalRequest) (TransactionResponse, error) {
	target, err := s.getReversibleTransaction(userID, portfolioID, transactionID)
	if err != nil {
		return TransactionResponse{}, err
	}

	occurredAt := req.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	if !common.ValidateNotFuture(occurredAt) {
		return TransactionResponse{}, common.BadRequest("Transaction occurred_at cannot be in the future")
	}

	groupID := uuid.New()
	reversesID := target.ID
	reversal := &Transaction{
		PortfolioID:           target.PortfolioID,
		AccountID:             target.AccountID,
		TransactionType:       TransactionTypeReversal,
		OccurredAt:            occurredAt,
		Description:           strings.TrimSpace(req.Reason),
		ReversesTransactionID: &reversesID,
		CorrectionGroupID:     &groupID,
		CreatedByUserID:       userID,
		Entries:               reverseEntries(target.Entries),
	}

	if err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		return s.repo.CreateWithDB(tx, reversal)
	}); err != nil {
		if common.IsUniqueViolation(err) {
			return TransactionResponse{}, common.Conflict("Transaction has already been reversed")
		}
		return TransactionResponse{}, err
	}

	return ToResponse(*reversal), nil
}

func (s *Service) Correct(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID, req TransactionCorrectionRequest) (TransactionCorrectionResponse, error) {
	target, err := s.getReversibleTransaction(userID, portfolioID, transactionID)
	if err != nil {
		return TransactionCorrectionResponse{}, err
	}

	groupID := uuid.New()
	reversesID := target.ID
	reversal := &Transaction{
		PortfolioID:           target.PortfolioID,
		AccountID:             target.AccountID,
		TransactionType:       TransactionTypeReversal,
		OccurredAt:            time.Now().UTC(),
		Description:           strings.TrimSpace(req.Reason),
		ReversesTransactionID: &reversesID,
		CorrectionGroupID:     &groupID,
		CreatedByUserID:       userID,
		Entries:               reverseEntries(target.Entries),
	}

	correctsID := target.ID
	replacement, err := s.buildTransaction(userID, portfolioID, req.Replacement, nil, &correctsID, &groupID)
	if err != nil {
		return TransactionCorrectionResponse{}, err
	}

	if err := s.repo.WithTransaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateWithDB(tx, reversal); err != nil {
			return err
		}
		return s.repo.CreateWithDB(tx, replacement)
	}); err != nil {
		if common.IsUniqueViolation(err) {
			return TransactionCorrectionResponse{}, common.Conflict("Transaction has already been reversed or corrected")
		}
		return TransactionCorrectionResponse{}, err
	}

	return TransactionCorrectionResponse{
		Reversal:    ToResponse(*reversal),
		Replacement: ToResponse(*replacement),
	}, nil
}

func (s *Service) buildTransaction(userID uuid.UUID, portfolioID uuid.UUID, req TransactionCreateRequest, reversesID *uuid.UUID, correctsID *uuid.UUID, correctionGroupID *uuid.UUID) (*Transaction, error) {
	account, err := s.accountRepo.GetInPortfolio(portfolioID, req.AccountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Account not found")
	}
	if err != nil {
		return nil, err
	}

	transactionType := strings.ToLower(strings.TrimSpace(req.TransactionType))
	if !common.OneOf(transactionType, TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeBuy, TransactionTypeSell, TransactionTypeFee, TransactionTypeTax, TransactionTypeTransfer) {
		return nil, common.BadRequest("Invalid transaction type")
	}

	if req.OccurredAt.IsZero() {
		return nil, common.BadRequest("Transaction occurred_at is required")
	}
	if !common.ValidateNotFuture(req.OccurredAt) {
		return nil, common.BadRequest("Transaction occurred_at cannot be in the future")
	}

	if len(req.Entries) == 0 {
		return nil, common.BadRequest("Transaction entries are required")
	}

	entries := make([]TransactionEntry, 0, len(req.Entries))
	for idx, entryReq := range req.Entries {
		entry, err := s.buildEntry(idx, entryReq)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	var idempotencyKeyPtr *string
	if idempotencyKey != "" {
		idempotencyKeyPtr = &idempotencyKey
	}

	return &Transaction{
		PortfolioID:           portfolioID,
		AccountID:             account.ID,
		TransactionType:       transactionType,
		OccurredAt:            req.OccurredAt.UTC(),
		Description:           strings.TrimSpace(req.Description),
		IdempotencyKey:        idempotencyKeyPtr,
		ReversesTransactionID: reversesID,
		CorrectsTransactionID: correctsID,
		CorrectionGroupID:     correctionGroupID,
		CreatedByUserID:       userID,
		Entries:               entries,
	}, nil
}

func (s *Service) buildEntry(index int, req TransactionEntryRequest) (TransactionEntry, error) {
	entryKind := strings.ToLower(strings.TrimSpace(req.EntryKind))
	if !common.OneOf(entryKind, EntryKindCash, EntryKindAsset, EntryKindFee, EntryKindTax) {
		return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Invalid entry kind at index %d", index))
	}

	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Entry currency must be a three-letter uppercase code at index %d", index))
	}

	entry := TransactionEntry{
		EntryKind: entryKind,
		Currency:  currency,
	}

	if entryKind == EntryKindAsset {
		if req.AssetID == nil || *req.AssetID == uuid.Nil {
			return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Asset entry requires asset_id at index %d", index))
		}
		if req.Quantity == nil || req.Quantity.IsZero() {
			return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Asset entry requires non-zero quantity at index %d", index))
		}

		asset, err := s.assetRepo.GetByID(*req.AssetID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Asset not found at index %d", index))
		}
		if err != nil {
			return TransactionEntry{}, err
		}
		if !asset.IsActive {
			return TransactionEntry{}, common.BadRequest(fmt.Sprintf("Asset is inactive at index %d", index))
		}

		entry.AssetID = req.AssetID
		entry.Quantity = req.Quantity
		entry.Amount = req.Amount
		return entry, nil
	}

	if req.Amount == nil || req.Amount.IsZero() {
		return TransactionEntry{}, common.BadRequest(fmt.Sprintf("%s entry requires non-zero amount at index %d", entryKind, index))
	}
	entry.Amount = req.Amount
	entry.AssetID = req.AssetID
	entry.Quantity = req.Quantity
	return entry, nil
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

func (s *Service) getOwnedTransaction(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID) (*Transaction, error) {
	transaction, err := s.repo.GetOwned(userID, portfolioID, transactionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Transaction not found")
	}
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) getReversibleTransaction(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID) (*Transaction, error) {
	target, err := s.getOwnedTransaction(userID, portfolioID, transactionID)
	if err != nil {
		return nil, err
	}
	if target.TransactionType == TransactionTypeReversal {
		return nil, common.BadRequest("Reversal transactions cannot be reversed")
	}

	hasReversal, err := s.repo.HasReversal(target.ID)
	if err != nil {
		return nil, err
	}
	if hasReversal {
		return nil, common.Conflict("Transaction has already been reversed")
	}
	return target, nil
}

func reverseEntries(entries []TransactionEntry) []TransactionEntry {
	reversed := make([]TransactionEntry, 0, len(entries))
	for _, entry := range entries {
		var quantity *decimal.Decimal
		if entry.Quantity != nil {
			value := entry.Quantity.Neg()
			quantity = &value
		}

		var amount *decimal.Decimal
		if entry.Amount != nil {
			value := entry.Amount.Neg()
			amount = &value
		}

		reversed = append(reversed, TransactionEntry{
			EntryKind: entry.EntryKind,
			AssetID:   entry.AssetID,
			Quantity:  quantity,
			Amount:    amount,
			Currency:  entry.Currency,
		})
	}
	return reversed
}
