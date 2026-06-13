package snapshots

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const SnapshotPeriodDaily = "daily"

type JSONB []byte

func NewJSONB(value any) (JSONB, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return JSONB(raw), nil
}

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = JSONB("null")
		return nil
	}

	switch typed := value.(type) {
	case []byte:
		*j = append((*j)[0:0], typed...)
	case string:
		*j = append((*j)[0:0], typed...)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
	return nil
}

type PortfolioSnapshot struct {
	ID                    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PortfolioID           uuid.UUID `gorm:"type:uuid;not null;index"`
	SnapshotDate          time.Time `gorm:"type:date;not null;index"`
	SnapshotPeriod        string    `gorm:"not null"`
	TotalValues           JSONB     `gorm:"type:jsonb;not null"`
	AssetAllocations      JSONB     `gorm:"type:jsonb;not null"`
	AssetClassAllocations JSONB     `gorm:"type:jsonb;not null"`
	CashAllocations       JSONB     `gorm:"type:jsonb;not null"`
	MissingPrices         JSONB     `gorm:"type:jsonb;not null"`
	IsFullyValued         bool      `gorm:"not null"`
	ValuationScope        string    `gorm:"not null"`
	AllocationScope       string    `gorm:"not null"`
	ValuationMetadata     JSONB     `gorm:"type:jsonb;not null"`
	AllocationMetadata    JSONB     `gorm:"type:jsonb;not null"`
	HoldingsMetadata      JSONB     `gorm:"type:jsonb;not null"`
	CreatedByUserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt             time.Time `gorm:"not null"`
}

func (s *PortfolioSnapshot) BeforeCreate(*gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
