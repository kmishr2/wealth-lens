package accounts

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(account *Account) error {
	return r.db.Create(account).Error
}

func (r *Repository) ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]Account, error) {
	var accounts []Account
	err := r.db.
		Where("portfolio_id = ?", portfolioID).
		Order("created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&accounts).Error
	return accounts, err
}

func (r *Repository) GetInPortfolio(portfolioID uuid.UUID, accountID uuid.UUID) (*Account, error) {
	var account Account
	err := r.db.First(&account, "id = ? AND portfolio_id = ?", accountID, portfolioID).Error
	return &account, err
}

func (r *Repository) Update(account *Account) error {
	return r.db.Save(account).Error
}

func (r *Repository) Delete(account *Account) error {
	return r.db.Delete(account).Error
}
