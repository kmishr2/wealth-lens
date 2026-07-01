package goals

import (
	"time"

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

func (r *Repository) Create(goal *Goal) error {
	return r.db.Create(goal).Error
}

func (r *Repository) ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]Goal, error) {
	var goals []Goal
	err := r.db.
		Where("portfolio_id = ?", portfolioID).
		Order("created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&goals).Error
	return goals, err
}

func (r *Repository) ListActiveByPortfolio(portfolioID uuid.UUID) ([]Goal, error) {
	var goals []Goal
	err := r.db.
		Where("portfolio_id = ? AND status = ?", portfolioID, StatusActive).
		Order("target_date asc, id asc").
		Find(&goals).Error
	return goals, err
}

func (r *Repository) GetInPortfolio(portfolioID uuid.UUID, goalID uuid.UUID) (*Goal, error) {
	var goal Goal
	err := r.db.First(&goal, "id = ? AND portfolio_id = ?", goalID, portfolioID).Error
	return &goal, err
}

func (r *Repository) Update(goal *Goal) error {
	return r.db.Save(goal).Error
}

func (r *Repository) Delete(goal *Goal) error {
	return r.db.Delete(goal).Error
}

func (r *Repository) CreateMonthlySnapshot(snapshot *MonthlyGoalSnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *Repository) GetMonthlySnapshotByGoalMonth(goalID uuid.UUID, monthEnd time.Time) (*MonthlyGoalSnapshot, error) {
	var snapshot MonthlyGoalSnapshot
	err := r.db.First(&snapshot, "goal_id = ? AND snapshot_month_end = ?", goalID, monthEnd).Error
	return &snapshot, err
}

func (r *Repository) ListMonthlySnapshots(goalID uuid.UUID, pagination common.Pagination) ([]MonthlyGoalSnapshot, error) {
	var snapshots []MonthlyGoalSnapshot
	err := r.db.
		Where("goal_id = ?", goalID).
		Order("snapshot_month_end desc, created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&snapshots).Error
	return snapshots, err
}
