package prices

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func (r *Repository) CreateAutomated(price *AssetPrice) (bool, error) {
	result := r.db.Clauses(clause.OnConflict{
		Columns:     []clause.Column{{Name: "asset_id"}, {Name: "market_date"}, {Name: "source"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Not(clause.Eq{Column: "market_date", Value: nil})}},
		DoNothing:   true,
	}).Create(price)
	return result.RowsAffected == 1, result.Error
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(price *AssetPrice) error {
	return r.db.Create(price).Error
}

func (r *Repository) ListByAsset(assetID uuid.UUID, pagination common.Pagination) ([]AssetPrice, error) {
	var prices []AssetPrice
	err := r.db.
		Where("asset_id = ?", assetID).
		Order("priced_at desc, created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&prices).Error
	return prices, err
}

func (r *Repository) GetLatestByAsset(assetID uuid.UUID) (*AssetPrice, error) {
	var price AssetPrice
	err := r.db.
		Where("asset_id = ?", assetID).
		Order("priced_at desc, created_at desc, id desc").
		First(&price).Error
	return &price, err
}

func (r *Repository) ListLatestByAssets(assetIDs []uuid.UUID) ([]AssetPrice, error) {
	if len(assetIDs) == 0 {
		return []AssetPrice{}, nil
	}

	var prices []AssetPrice
	err := r.db.Raw(`
		SELECT DISTINCT ON (asset_id)
			id,
			asset_id,
			price,
			currency,
			priced_at,
			market_date,
			source,
			note,
			created_by_user_id,
			created_at
		FROM asset_prices
		WHERE asset_id IN ?
		ORDER BY asset_id, priced_at DESC, created_at DESC, id DESC
	`, assetIDs).Scan(&prices).Error
	return prices, err
}

func (r *Repository) ListLatestByAssetsAsOf(assetIDs []uuid.UUID, asOf time.Time) ([]AssetPrice, error) {
	if len(assetIDs) == 0 {
		return []AssetPrice{}, nil
	}

	var prices []AssetPrice
	err := r.db.Raw(`
		SELECT DISTINCT ON (asset_id)
			id,
			asset_id,
			price,
			currency,
			priced_at,
			market_date,
			source,
			note,
			created_by_user_id,
			created_at
		FROM asset_prices
		WHERE asset_id IN ?
			AND priced_at <= ?
		ORDER BY asset_id, priced_at DESC, created_at DESC, id DESC
	`, assetIDs, asOf).Scan(&prices).Error
	return prices, err
}
