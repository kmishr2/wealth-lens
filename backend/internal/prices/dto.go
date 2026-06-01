package prices

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AssetPriceCreateRequest struct {
	Price    *decimal.Decimal `json:"price"`
	Currency string           `json:"currency"`
	PricedAt time.Time        `json:"priced_at"`
	Source   string           `json:"source"`
	Note     string           `json:"note"`
}

type AssetPriceResponse struct {
	ID              uuid.UUID       `json:"id"`
	AssetID         uuid.UUID       `json:"asset_id"`
	Price           decimal.Decimal `json:"price"`
	Currency        string          `json:"currency"`
	PricedAt        time.Time       `json:"priced_at"`
	Source          string          `json:"source"`
	Note            string          `json:"note"`
	CreatedByUserID uuid.UUID       `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
}

func ToResponse(price AssetPrice) AssetPriceResponse {
	return AssetPriceResponse{
		ID:              price.ID,
		AssetID:         price.AssetID,
		Price:           price.Price,
		Currency:        price.Currency,
		PricedAt:        price.PricedAt,
		Source:          price.Source,
		Note:            price.Note,
		CreatedByUserID: price.CreatedByUserID,
		CreatedAt:       price.CreatedAt,
	}
}
