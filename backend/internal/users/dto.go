package users

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	BaseCurrency string    `json:"base_currency"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateMeRequest struct {
	DisplayName  *string `json:"display_name"`
	BaseCurrency *string `json:"base_currency"`
	Timezone     *string `json:"timezone"`
}

func ToResponse(user User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		BaseCurrency: user.BaseCurrency,
		Timezone:     user.Timezone,
		CreatedAt:    user.CreatedAt,
	}
}
