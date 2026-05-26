package accounts

import (
	"time"

	"github.com/google/uuid"
)

type AccountCreateRequest struct {
	Name            string `json:"name"`
	AccountType     string `json:"account_type"`
	InstitutionName string `json:"institution_name"`
	Currency        string `json:"currency"`
}

type AccountUpdateRequest struct {
	Name            *string `json:"name"`
	InstitutionName *string `json:"institution_name"`
}

type AccountResponse struct {
	ID              uuid.UUID `json:"id"`
	PortfolioID     uuid.UUID `json:"portfolio_id"`
	Name            string    `json:"name"`
	AccountType     string    `json:"account_type"`
	InstitutionName string    `json:"institution_name"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func ToResponse(account Account) AccountResponse {
	return AccountResponse{
		ID:              account.ID,
		PortfolioID:     account.PortfolioID,
		Name:            account.Name,
		AccountType:     account.AccountType,
		InstitutionName: account.InstitutionName,
		Currency:        account.Currency,
		CreatedAt:       account.CreatedAt,
		UpdatedAt:       account.UpdatedAt,
	}
}
