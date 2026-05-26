package accounts

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AccountTypeBrokerage  = "brokerage"
	AccountTypeRetirement = "retirement"
	AccountTypeBank       = "bank"
	AccountTypeWallet     = "wallet"
	AccountTypeOther      = "other"
)

type Account struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID     uuid.UUID      `gorm:"uniqueIndex:idx_portfolio_account_name"`
	Name            string         `gorm:"not null;uniqueIndex:idx_portfolio_account_name"`
	AccountType     string         `gorm:"not null"`
	InstitutionName string         `gorm:"not null;default:''"`
	Currency        string         `gorm:"type:char(3);not null"`
	CreatedAt       time.Time      `gorm:"not null"`
	UpdatedAt       time.Time      `gorm:"not null"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

func (a *Account) BeforeCreate(*gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
