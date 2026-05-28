package portfolios

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID 		 uuid.UUID      `gorm:"uniqueIndex:idx_user_portfolio_name"`
    Name   		 string         `gorm:"uniqueIndex:idx_user_portfolio_name"`
	Description  string         `gorm:"not null;default:''"`
	BaseCurrency string         `gorm:"type:char(3);not null"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (p *Portfolio) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
