package fixeddeposits

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/transactions"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBundle(asset *assets.Asset, transaction *transactions.Transaction, price *prices.AssetPrice, record *FixedDeposit) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(asset).Error; err != nil {
			return err
		}
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
		if err := tx.Create(price).Error; err != nil {
			return err
		}
		return tx.Create(record).Error
	})
}

func (r *Repository) ListByAccount(portfolioID uuid.UUID, accountID uuid.UUID) ([]FixedDeposit, error) {
	var records []FixedDeposit
	err := r.db.
		Where("portfolio_id = ? AND account_id = ?", portfolioID, accountID).
		Order("start_date desc, id desc").
		Find(&records).Error
	return records, err
}

func (r *Repository) GetByIDAccount(portfolioID uuid.UUID, accountID uuid.UUID, fixedDepositID uuid.UUID) (*FixedDeposit, error) {
	var record FixedDeposit
	err := r.db.First(&record, "id = ? AND portfolio_id = ? AND account_id = ?", fixedDepositID, portfolioID, accountID).Error
	return &record, err
}

func (r *Repository) CreateValue(price *prices.AssetPrice) error {
	return r.db.Create(price).Error
}

func (r *Repository) GetClosureByFixedDeposit(fixedDepositID uuid.UUID) (*Closure, error) {
	var closure Closure
	err := r.db.First(&closure, "fixed_deposit_id = ?", fixedDepositID).Error
	return &closure, err
}

func (r *Repository) CloseBundle(transaction *transactions.Transaction, closure *Closure) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}
		return tx.Create(closure).Error
	})
}
