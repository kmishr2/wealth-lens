package snapshots

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

func (r *Repository) Create(snapshot *PortfolioSnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *Repository) CreateWeeklyPerformance(snapshot *WeeklyPerformanceSnapshot) error {
	return r.db.Create(snapshot).Error
}

func (r *Repository) GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*PortfolioSnapshot, error) {
	var snapshot PortfolioSnapshot
	err := r.db.
		Where("portfolio_id = ? AND snapshot_date = ? AND snapshot_period = ?", portfolioID, snapshotDate, snapshotPeriod).
		First(&snapshot).Error
	return &snapshot, err
}

func (r *Repository) GetWeeklyPerformanceByPortfolioWeekEnd(portfolioID uuid.UUID, weekEndDate time.Time) (*WeeklyPerformanceSnapshot, error) {
	var snapshot WeeklyPerformanceSnapshot
	err := r.db.
		Where("portfolio_id = ? AND week_end_date = ?", portfolioID, weekEndDate).
		First(&snapshot).Error
	return &snapshot, err
}

func (r *Repository) ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]PortfolioSnapshot, error) {
	var snapshots []PortfolioSnapshot
	err := r.db.
		Where("portfolio_id = ?", portfolioID).
		Order("snapshot_date desc, created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *Repository) ListWeeklyPerformanceByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]WeeklyPerformanceSnapshot, error) {
	var snapshots []WeeklyPerformanceSnapshot
	err := r.db.
		Where("portfolio_id = ?", portfolioID).
		Order("week_end_date desc, created_at desc, id desc").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&snapshots).Error
	return snapshots, err
}

func (r *Repository) ListByPortfolioDateRange(portfolioID uuid.UUID, startDate time.Time, endDate time.Time, snapshotPeriod string) ([]PortfolioSnapshot, error) {
	var snapshots []PortfolioSnapshot
	err := r.db.
		Where("portfolio_id = ? AND snapshot_date >= ? AND snapshot_date <= ? AND snapshot_period = ?", portfolioID, startDate, endDate, snapshotPeriod).
		Order("snapshot_date asc, created_at asc, id asc").
		Find(&snapshots).Error
	return snapshots, err
}
