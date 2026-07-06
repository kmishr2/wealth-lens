package notifications

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListOpenFixedDepositsMaturingBy(userID uuid.UUID, asOfDate, cutoffDate time.Time) ([]FixedDepositMaturityRecord, error) {
	var records []FixedDepositMaturityRecord
	err := r.db.Raw(`
		SELECT fd.id AS fixed_deposit_id, fd.portfolio_id, p.name AS portfolio_name,
			fd.account_id, a.name AS account_name, fd.name AS deposit_name, fd.maturity_date
		FROM fixed_deposits fd
		INNER JOIN portfolios p ON p.id = fd.portfolio_id AND p.deleted_at IS NULL
		INNER JOIN accounts a ON a.id = fd.account_id AND a.portfolio_id = fd.portfolio_id
		LEFT JOIN fixed_deposit_closures c ON c.fixed_deposit_id = fd.id
		WHERE p.user_id = ? AND c.id IS NULL
			AND fd.start_date <= ? AND fd.maturity_date <= ?
		ORDER BY fd.maturity_date ASC, fd.id ASC
	`, userID, asOfDate, cutoffDate).Scan(&records).Error
	return records, err
}
