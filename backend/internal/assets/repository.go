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

func (r *Repository) ListActiveByProvider(provider string) ([]IdentifiedAsset, error) {
	var result []IdentifiedAsset
	err := r.db.Table("assets").
		Select("assets.*, asset_identifiers.identifier AS provider_identifier").
		Joins("JOIN asset_identifiers ON asset_identifiers.asset_id = assets.id").
		Where("assets.is_active = ? AND asset_identifiers.provider = ?", true, provider).
		Order("assets.id").
		Scan(&result).Error
	return result, err
}
