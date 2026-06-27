package holdings

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LedgerEntryRecord struct {
	EntryKind   string           `gorm:"column:entry_kind"`
	AssetID     string           `gorm:"column:asset_id"`
	AssetSymbol string           `gorm:"column:asset_symbol"`
	AssetName   string           `gorm:"column:asset_name"`
	AssetClass  string           `gorm:"column:asset_class"`
	Quantity    *decimal.Decimal `gorm:"column:quantity"`
	Amount      *decimal.Decimal `gorm:"column:amount"`
	Currency    string           `gorm:"column:currency"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListLedgerEntries(portfolioID uuid.UUID) ([]LedgerEntryRecord, error) {
	var entries []LedgerEntryRecord
	err := r.db.Raw(`
		SELECT
			te.entry_kind,
			COALESCE(te.asset_id::text, '') AS asset_id,
			COALESCE(a.symbol, '') AS asset_symbol,
			COALESCE(a.name, '') AS asset_name,
			COALESCE(a.asset_class, '') AS asset_class,
			te.quantity,
			te.amount,
			te.currency
		FROM transaction_entries te
		INNER JOIN transactions t ON t.id = te.transaction_id
		LEFT JOIN assets a ON a.id = te.asset_id
		WHERE t.portfolio_id = ?
		ORDER BY t.occurred_at ASC, t.created_at ASC, t.id ASC, te.id ASC
	`, portfolioID).Scan(&entries).Error
	return entries, err
}

func (r *Repository) ListLedgerEntriesAsOf(portfolioID uuid.UUID, asOf time.Time) ([]LedgerEntryRecord, error) {
	var entries []LedgerEntryRecord
	err := r.db.Raw(`
		SELECT
			te.entry_kind,
			COALESCE(te.asset_id::text, '') AS asset_id,
			COALESCE(a.symbol, '') AS asset_symbol,
			COALESCE(a.name, '') AS asset_name,
			COALESCE(a.asset_class, '') AS asset_class,
			te.quantity,
			te.amount,
			te.currency
		FROM transaction_entries te
		INNER JOIN transactions t ON t.id = te.transaction_id
		LEFT JOIN assets a ON a.id = te.asset_id
		WHERE t.portfolio_id = ?
			AND t.occurred_at <= ?
		ORDER BY t.occurred_at ASC, t.created_at ASC, t.id ASC, te.id ASC
	`, portfolioID, asOf).Scan(&entries).Error
	return entries, err
}
