package assets

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

func (r *Repository) Create(asset *Asset) error {
	return r.db.Create(asset).Error
}

func (r *Repository) List(pagination common.Pagination) ([]Asset, error) {
	var assets []Asset
	err := r.db.
		Order("symbol asc, exchange asc, currency asc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&assets).Error
	return assets, err
}

func (r *Repository) GetByID(assetID uuid.UUID) (*Asset, error) {
	var asset Asset
	err := r.db.First(&asset, "id = ?", assetID).Error
	return &asset, err
}
