package transactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ExternalCashFlowRecord struct {
	OccurredAt time.Time       `gorm:"column:occurred_at"`
	Currency   string          `gorm:"column:currency"`
	Amount     decimal.Decimal `gorm:"column:amount"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTransaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *Repository) CreateWithDB(db *gorm.DB, transaction *Transaction) error {
	return db.Create(transaction).Error
}

func (r *Repository) CreateMany(transactions []*Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, transaction := range transactions {
			if err := tx.Create(transaction).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) GetOwned(userID uuid.UUID, portfolioID uuid.UUID, transactionID uuid.UUID) (*Transaction, error) {
	var transaction Transaction
	err := r.db.
		Preload("Entries").
		Joins("JOIN portfolios ON portfolios.id = transactions.portfolio_id AND portfolios.deleted_at IS NULL").
		Where("transactions.id = ? AND transactions.portfolio_id = ? AND portfolios.user_id = ?", transactionID, portfolioID, userID).
		First(&transaction).Error
	return &transaction, err
}

func (r *Repository) ListOwned(userID uuid.UUID, portfolioID uuid.UUID, pagination common.Pagination) ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.
		Preload("Entries").
		Joins("JOIN portfolios ON portfolios.id = transactions.portfolio_id AND portfolios.deleted_at IS NULL").
		Where("transactions.portfolio_id = ? AND portfolios.user_id = ?", portfolioID, userID).
		Order("transactions.occurred_at desc, transactions.created_at desc, transactions.id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&transactions).Error
	return transactions, err
}

func (r *Repository) GetByIdempotencyKey(portfolioID uuid.UUID, idempotencyKey string) (*Transaction, error) {
	var transaction Transaction
	err := r.db.
		Preload("Entries").
		First(&transaction, "portfolio_id = ? AND idempotency_key = ?", portfolioID, idempotencyKey).Error
	return &transaction, err
}

func (r *Repository) HasReversal(transactionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&Transaction{}).Where("reverses_transaction_id = ?", transactionID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListExternalCashFlows(portfolioID uuid.UUID, startAfter time.Time, endAt time.Time) ([]ExternalCashFlowRecord, error) {
	var cashFlows []ExternalCashFlowRecord
	err := r.db.Raw(`
		SELECT
			t.occurred_at,
			te.currency,
			te.amount
		FROM transaction_entries te
		INNER JOIN transactions t ON t.id = te.transaction_id
		LEFT JOIN transactions reversed ON reversed.id = t.reverses_transaction_id
		WHERE t.portfolio_id = ?
			AND t.occurred_at > ?
			AND t.occurred_at <= ?
			AND te.entry_kind = ?
			AND COALESCE(reversed.transaction_type, t.transaction_type) IN (?, ?)
		ORDER BY t.occurred_at ASC, t.created_at ASC, t.id ASC, te.id ASC
	`, portfolioID, startAfter, endAt, EntryKindCash, TransactionTypeDeposit, TransactionTypeWithdrawal).Scan(&cashFlows).Error
	return cashFlows, err
}
