package portfolios

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"gorm.io/gorm"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(userID uuid.UUID, req PortfolioCreateRequest) (PortfolioResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return PortfolioResponse{}, common.BadRequest("Portfolio name is required")
	}

	baseCurrency := common.NormalizeCurrency(req.BaseCurrency)
	if !common.ValidateCurrency(baseCurrency) {
		return PortfolioResponse{}, common.BadRequest("Base currency must be a three-letter uppercase code")
	}

	portfolio := &Portfolio{
		UserID:       userID,
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		BaseCurrency: baseCurrency,
	}

	if err := s.repo.Create(portfolio); err != nil {
		if common.IsUniqueViolation(err) {
			return PortfolioResponse{}, common.Conflict("Portfolio name already exists")
		}
		return PortfolioResponse{}, err
	}

	return ToResponse(*portfolio), nil
}

func (s *Service) List(userID uuid.UUID, pagination common.Pagination) ([]PortfolioResponse, error) {
	portfolios, err := s.repo.ListByUser(userID, pagination)
	if err != nil {
		return nil, err
	}

	responses := make([]PortfolioResponse, 0, len(portfolios))
	for _, portfolio := range portfolios {
		responses = append(responses, ToResponse(portfolio))
	}
	return responses, nil
}

func (s *Service) Get(userID uuid.UUID, portfolioID uuid.UUID) (PortfolioResponse, error) {
	portfolio, err := s.repo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PortfolioResponse{}, common.NotFound("Portfolio not found")
	}
	if err != nil {
		return PortfolioResponse{}, err
	}
	return ToResponse(*portfolio), nil
}

func (s *Service) Update(userID uuid.UUID, portfolioID uuid.UUID, req PortfolioUpdateRequest) (PortfolioResponse, error) {
	portfolio, err := s.repo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return PortfolioResponse{}, common.NotFound("Portfolio not found")
	}
	if err != nil {
		return PortfolioResponse{}, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return PortfolioResponse{}, common.BadRequest("Portfolio name is required")
		}
		portfolio.Name = name
	}
	if req.Description != nil {
		portfolio.Description = strings.TrimSpace(*req.Description)
	}

	if err := s.repo.Update(portfolio); err != nil {
		if common.IsUniqueViolation(err) {
			return PortfolioResponse{}, common.Conflict("Portfolio name already exists")
		}
		return PortfolioResponse{}, err
	}

	return ToResponse(*portfolio), nil
}

func (s *Service) Delete(userID uuid.UUID, portfolioID uuid.UUID) error {
	portfolio, err := s.repo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NotFound("Portfolio not found")
	}
	if err != nil {
		return err
	}
	return s.repo.Delete(portfolio)
}
