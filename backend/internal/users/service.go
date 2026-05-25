package users

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

func (s *Service) GetMe(userID uuid.UUID) (UserResponse, error) {
	user, err := s.repo.GetByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserResponse{}, common.NotFound("User not found")
	}
	if err != nil {
		return UserResponse{}, err
	}
	return ToResponse(*user), nil
}

func (s *Service) UpdateMe(userID uuid.UUID, req UpdateMeRequest) (UserResponse, error) {
	user, err := s.repo.GetByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserResponse{}, common.NotFound("User not found")
	}
	if err != nil {
		return UserResponse{}, err
	}

	if req.DisplayName != nil {
		displayName := strings.TrimSpace(*req.DisplayName)
		if displayName == "" {
			return UserResponse{}, common.BadRequest("Display name is required")
		}
		user.DisplayName = displayName
	}

	if req.BaseCurrency != nil {
		currency := common.NormalizeCurrency(*req.BaseCurrency)
		if !common.ValidateCurrency(currency) {
			return UserResponse{}, common.BadRequest("Base currency must be a three-letter uppercase code")
		}
		user.BaseCurrency = currency
	}

	if req.Timezone != nil {
		timezone := strings.TrimSpace(*req.Timezone)
		if timezone == "" {
			return UserResponse{}, common.BadRequest("Timezone is required")
		}
		user.Timezone = timezone
	}

	if err := s.repo.Update(user); err != nil {
		return UserResponse{}, err
	}

	return ToResponse(*user), nil
}
