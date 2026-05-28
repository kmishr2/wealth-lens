package portfolios

import (
	"time"

	"github.com/google/uuid"
)

type PortfolioCreateRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	BaseCurrency string `json:"base_currency"`
}

type PortfolioUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type PortfolioResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	BaseCurrency string    `json:"base_currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToResponse(portfolio Portfolio) PortfolioResponse {
	return PortfolioResponse{
		ID:           portfolio.ID,
		Name:         portfolio.Name,
		Description:  portfolio.Description,
		BaseCurrency: portfolio.BaseCurrency,
		CreatedAt:    portfolio.CreatedAt,
		UpdatedAt:    portfolio.UpdatedAt,
	}
}
