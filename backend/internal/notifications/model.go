package notifications

import (
	"time"

	"github.com/google/uuid"
)

type FixedDepositMaturityRecord struct {
	FixedDepositID uuid.UUID `gorm:"column:fixed_deposit_id"`
	PortfolioID    uuid.UUID `gorm:"column:portfolio_id"`
	PortfolioName  string    `gorm:"column:portfolio_name"`
	AccountID      uuid.UUID `gorm:"column:account_id"`
	AccountName    string    `gorm:"column:account_name"`
	DepositName    string    `gorm:"column:deposit_name"`
	MaturityDate   time.Time `gorm:"column:maturity_date"`
}
