package benchmarks

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Benchmark struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code            string     `gorm:"not null"`
	Name            string     `gorm:"not null"`
	Currency        string     `gorm:"type:char(3);not null"`
	Source          string     `gorm:"not null"`
	Description     string     `gorm:"not null;default:''"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt       time.Time  `gorm:"not null"`
}

func (b *Benchmark) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type BenchmarkObservation struct {
	ID              uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BenchmarkID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	ObservationDate time.Time       `gorm:"type:date;not null;index"`
	Value           decimal.Decimal `gorm:"type:numeric(28,10);not null"`
	Source          string          `gorm:"not null"`
	Note            string          `gorm:"not null;default:''"`
	CreatedByUserID *uuid.UUID      `gorm:"type:uuid;index"`
	CreatedAt       time.Time       `gorm:"not null"`
}

func (o *BenchmarkObservation) BeforeCreate(*gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
