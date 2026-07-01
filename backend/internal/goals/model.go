package goals

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusArchived  = "archived"
)

type Goal struct {
	ID              uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	Name            string          `gorm:"not null"`
	TargetAmount    decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	Currency        string          `gorm:"type:char(3);not null"`
	TargetDate      time.Time       `gorm:"type:date;not null"`
	Status          string          `gorm:"not null;default:'active'"`
	CreatedByUserID uuid.UUID       `gorm:"type:uuid;not null;index"`
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"`
}

func (g *Goal) BeforeCreate(*gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

type MonthlyGoalSnapshot struct {
	ID                          uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID                 uuid.UUID       `gorm:"type:uuid;not null;index"`
	GoalID                      uuid.UUID       `gorm:"type:uuid;not null;index"`
	SnapshotMonthEnd            time.Time       `gorm:"type:date;not null;index"`
	CurrentValue                decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	TargetValue                 decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	Currency                    string          `gorm:"type:char(3);not null"`
	ProgressPercentage          decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	RemainingAmount             decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	MonthsRemaining             int             `gorm:"not null"`
	RequiredMonthlyContribution decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	IsTargetReached             bool            `gorm:"not null"`
	GoalProgressMetadata        snapshots.JSONB `gorm:"type:jsonb;not null"`
	CreatedByUserID             uuid.UUID       `gorm:"type:uuid;not null;index"`
	CreatedAt                   time.Time       `gorm:"not null"`
}

func (s *MonthlyGoalSnapshot) BeforeCreate(*gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
