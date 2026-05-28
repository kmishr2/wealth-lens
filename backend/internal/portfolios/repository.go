package portfolios

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

func (r *Repository) Create(portfolio *Portfolio) error {
	return r.db.Create(portfolio).Error
}

func (r *Repository) ListByUser(userID uuid.UUID, pagination common.Pagination) ([]Portfolio, error) {
	var portfolios []Portfolio
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&portfolios).Error
	return portfolios, err
}

func (r *Repository) GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*Portfolio, error) {
	var portfolio Portfolio
	err := r.db.First(&portfolio, "id = ? AND user_id = ?", portfolioID, userID).Error
	return &portfolio, err
}

func (r *Repository) Update(portfolio *Portfolio) error {
	return r.db.Save(portfolio).Error
}

func (r *Repository) Delete(portfolio *Portfolio) error {
	return r.db.Delete(portfolio).Error
}
