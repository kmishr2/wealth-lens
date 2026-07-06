package notifications

import "github.com/google/uuid"

type Response struct {
	ID             string     `json:"id"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	Title          string     `json:"title"`
	Explanation    string     `json:"explanation"`
	TriggerRule    string     `json:"trigger_rule"`
	AsOfDate       string     `json:"as_of_date"`
	EventDate      string     `json:"event_date"`
	DaysUntilEvent int        `json:"days_until_event"`
	PortfolioID    uuid.UUID  `json:"portfolio_id"`
	PortfolioName  string     `json:"portfolio_name"`
	AccountID      *uuid.UUID `json:"account_id,omitempty"`
	AccountName    string     `json:"account_name,omitempty"`
	EntityID       uuid.UUID  `json:"entity_id"`
	EntityType     string     `json:"entity_type"`
	DataAsOfDate   *string    `json:"data_as_of_date,omitempty"`
	AgeDays        *int       `json:"age_days,omitempty"`
	ThresholdDays  *int       `json:"threshold_days,omitempty"`
}
