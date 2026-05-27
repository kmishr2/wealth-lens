package assets

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AssetClassCash        = "cash"
	AssetClassEquity      = "equity"
	AssetClassFund        = "fund"
	AssetClassBond        = "bond"
	AssetClassCrypto      = "crypto"
	AssetClassRealEstate  = "real_estate"
	AssetClassCommodity   = "commodity"
	AssetClassAlternative = "alternative"
	AssetClassOther       = "other"
)

type Asset struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Symbol     string    `gorm:"not null"`
	Name       string    `gorm:"not null"`
	AssetClass string    `gorm:"not null"`
	Currency   string    `gorm:"type:char(3);not null"`
	Exchange   string    `gorm:"not null;default:''"`
	IsActive   bool      `gorm:"not null;default:true"`
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (a *Asset) BeforeCreate(*gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
