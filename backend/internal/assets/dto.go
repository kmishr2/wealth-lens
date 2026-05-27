package assets

import "github.com/google/uuid"

type AssetCreateRequest struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	AssetClass string `json:"asset_class"`
	Currency   string `json:"currency"`
	Exchange   string `json:"exchange"`
}

type AssetResponse struct {
	ID         uuid.UUID `json:"id"`
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	AssetClass string    `json:"asset_class"`
	Currency   string    `json:"currency"`
	Exchange   string    `json:"exchange"`
	IsActive   bool      `json:"is_active"`
}

func ToResponse(asset Asset) AssetResponse {
	return AssetResponse{
		ID:         asset.ID,
		Symbol:     asset.Symbol,
		Name:       asset.Name,
		AssetClass: asset.AssetClass,
		Currency:   asset.Currency,
		Exchange:   asset.Exchange,
		IsActive:   asset.IsActive,
	}
}
