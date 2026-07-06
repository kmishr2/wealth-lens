package notifications

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/goals"
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

func (r *Repository) ListActiveGoalsDueBy(userID uuid.UUID, asOfDate, cutoffDate time.Time) ([]GoalTargetRecord, error) {
	var records []GoalTargetRecord
	err := r.db.Raw(`
		SELECT g.id AS goal_id, g.portfolio_id, p.name AS portfolio_name,
			g.name AS goal_name, g.target_date, latest.snapshot_month_end AS latest_snapshot_date
		FROM goals g
		INNER JOIN portfolios p ON p.id = g.portfolio_id AND p.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT snapshot_month_end, is_target_reached
			FROM monthly_goal_snapshots
			WHERE goal_id = g.id AND snapshot_month_end <= ?
			ORDER BY snapshot_month_end DESC, id DESC
			LIMIT 1
		) latest ON true
		WHERE p.user_id = ? AND g.deleted_at IS NULL AND g.status = ?
			AND g.target_date <= ? AND COALESCE(latest.is_target_reached, false) = false
		ORDER BY g.target_date ASC, g.id ASC
	`, asOfDate, userID, goals.StatusActive, cutoffDate).Scan(&records).Error
	return records, err
}
