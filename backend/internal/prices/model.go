package prices

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AssetPrice struct {
	ID              uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AssetID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	Price           decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	Currency        string          `gorm:"type:char(3);not null"`
	PricedAt        time.Time       `gorm:"not null;index"`
	Source          string          `gorm:"not null;default:'manual'"`
	Note            string          `gorm:"not null;default:''"`
	CreatedByUserID uuid.UUID       `gorm:"type:uuid;not null;index"`
	CreatedAt       time.Time       `gorm:"not null"`
}

func (p *AssetPrice) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
