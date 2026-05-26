package accounts

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"gorm.io/gorm"
)

type Service struct {
	repo          *Repository
	portfolioRepo *portfolios.Repository
}

func NewService(repo *Repository, portfolioRepo *portfolios.Repository) *Service {
	return &Service{repo: repo, portfolioRepo: portfolioRepo}
}

func (s *Service) Create(userID uuid.UUID, portfolioID uuid.UUID, req AccountCreateRequest) (AccountResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return AccountResponse{}, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return AccountResponse{}, common.BadRequest("Account name is required")
	}

	accountType := strings.ToLower(strings.TrimSpace(req.AccountType))
	if !common.OneOf(accountType, AccountTypeBrokerage, AccountTypeRetirement, AccountTypeBank, AccountTypeWallet, AccountTypeOther) {
		return AccountResponse{}, common.BadRequest("Invalid account type")
	}

	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return AccountResponse{}, common.BadRequest("Currency must be a three-letter uppercase code")
	}

	account := &Account{
		PortfolioID:     portfolioID,
		Name:            name,
		AccountType:     accountType,
		InstitutionName: strings.TrimSpace(req.InstitutionName),
		Currency:        currency,
	}

	if err := s.repo.Create(account); err != nil {
		if common.IsUniqueViolation(err) {
			return AccountResponse{}, common.Conflict("Account name already exists")
		}
		return AccountResponse{}, err
	}

	return ToResponse(*account), nil
}

func (s *Service) List(userID uuid.UUID, portfolioID uuid.UUID, pagination common.Pagination) ([]AccountResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}

	accounts, err := s.repo.ListByPortfolio(portfolioID, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		responses = append(responses, ToResponse(account))
	}
	return responses, nil
}

func (s *Service) Get(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID) (AccountResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return AccountResponse{}, err
	}

	account, err := s.repo.GetInPortfolio(portfolioID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccountResponse{}, common.NotFound("Account not found")
	}
	if err != nil {
		return AccountResponse{}, err
	}
	return ToResponse(*account), nil
}

func (s *Service) Update(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID, req AccountUpdateRequest) (AccountResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return AccountResponse{}, err
	}

	account, err := s.repo.GetInPortfolio(portfolioID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AccountResponse{}, common.NotFound("Account not found")
	}
	if err != nil {
		return AccountResponse{}, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return AccountResponse{}, common.BadRequest("Account name is required")
		}
		account.Name = name
	}
	if req.InstitutionName != nil {
		account.InstitutionName = strings.TrimSpace(*req.InstitutionName)
	}

	if err := s.repo.Update(account); err != nil {
		if common.IsUniqueViolation(err) {
			return AccountResponse{}, common.Conflict("Account name already exists")
		}
		return AccountResponse{}, err
	}

	return ToResponse(*account), nil
}

func (s *Service) Delete(userID uuid.UUID, portfolioID uuid.UUID, accountID uuid.UUID) error {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return err
	}

	account, err := s.repo.GetInPortfolio(portfolioID, accountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NotFound("Account not found")
	}
	if err != nil {
		return err
	}

	return s.repo.Delete(account)
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
