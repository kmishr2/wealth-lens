package fixeddeposits

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type FixedDeposit struct {
	ID                   uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID          uuid.UUID       `gorm:"type:uuid;not null;index"`
	AccountID            uuid.UUID       `gorm:"type:uuid;not null;index"`
	AssetID              uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex"`
	OpeningTransactionID uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex"`
	Name                 string          `gorm:"not null"`
	BankReference        string          `gorm:"not null;default:''"`
	Principal            decimal.Decimal `gorm:"type:numeric(28,4);not null"`
	Currency             string          `gorm:"type:char(3);not null"`
	AnnualInterestRate   decimal.Decimal `gorm:"type:numeric(9,6);not null"`
	StartDate            time.Time       `gorm:"type:date;not null"`
	MaturityDate         time.Time       `gorm:"type:date;not null"`
	CreatedByUserID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	CreatedAt            time.Time       `gorm:"not null"`
}

func (f *FixedDeposit) BeforeCreate(*gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
