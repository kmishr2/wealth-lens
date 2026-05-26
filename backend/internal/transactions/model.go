package transactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypeBuy        = "buy"
	TransactionTypeSell       = "sell"
	TransactionTypeFee        = "fee"
	TransactionTypeTax        = "tax"
	TransactionTypeTransfer   = "transfer"
	TransactionTypeReversal   = "reversal"

	EntryKindCash  = "cash"
	EntryKindAsset = "asset"
	EntryKindFee   = "fee"
	EntryKindTax   = "tax"
)

type Transaction struct {
	ID                    uuid.UUID          `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID           uuid.UUID          `gorm:"type:uuid;not null;index"`
	AccountID             uuid.UUID          `gorm:"type:uuid;not null;index"`
	TransactionType       string             `gorm:"not null"`
	OccurredAt            time.Time          `gorm:"not null;index"`
	Description           string             `gorm:"not null;default:''"`
	IdempotencyKey        *string            `gorm:"index"`
	ReversesTransactionID *uuid.UUID         `gorm:"type:uuid;index"`
	CorrectsTransactionID *uuid.UUID         `gorm:"type:uuid;index"`
	CorrectionGroupID     *uuid.UUID         `gorm:"type:uuid;index"`
	CreatedByUserID       uuid.UUID          `gorm:"type:uuid;not null;index"`
	CreatedAt             time.Time          `gorm:"not null"`
	UpdatedAt             time.Time          `gorm:"not null"`
	Entries               []TransactionEntry `gorm:"foreignKey:TransactionID"`
}

func (t *Transaction) BeforeCreate(*gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type TransactionEntry struct {
	ID            uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID uuid.UUID        `gorm:"type:uuid;not null;index"`
	EntryKind     string           `gorm:"not null"`
	AssetID       *uuid.UUID       `gorm:"type:uuid;index"`
	Quantity      *decimal.Decimal `gorm:"type:numeric(28,10)"`
	Amount        *decimal.Decimal `gorm:"type:numeric(28,4)"`
	Currency      string           `gorm:"type:char(3);not null"`
	CreatedAt     time.Time        `gorm:"not null"`
}

func (e *TransactionEntry) BeforeCreate(*gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
