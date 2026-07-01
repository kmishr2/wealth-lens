package goals

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const goalDateLayout = "2006-01-02"

type GoalCreateRequest struct {
	Name         string           `json:"name"`
	TargetAmount *decimal.Decimal `json:"target_amount"`
	Currency     string           `json:"currency"`
	TargetDate   string           `json:"target_date"`
}

type GoalUpdateRequest struct {
	Name         *string          `json:"name"`
	TargetAmount *decimal.Decimal `json:"target_amount"`
	Currency     *string          `json:"currency"`
	TargetDate   *string          `json:"target_date"`
	Status       *string          `json:"status"`
}

type GoalResponse struct {
	ID              uuid.UUID       `json:"id"`
	PortfolioID     uuid.UUID       `json:"portfolio_id"`
	Name            string          `json:"name"`
	TargetAmount    decimal.Decimal `json:"target_amount"`
	Currency        string          `json:"currency"`
	TargetDate      string          `json:"target_date"`
	Status          string          `json:"status"`
	CreatedByUserID uuid.UUID       `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type MonthlyGoalSnapshotResponse struct {
	ID                          uuid.UUID                `json:"id"`
	PortfolioID                 uuid.UUID                `json:"portfolio_id"`
	GoalID                      uuid.UUID                `json:"goal_id"`
	SnapshotMonthEnd            string                   `json:"snapshot_month_end"`
	CurrentValue                decimal.Decimal          `json:"current_value"`
	TargetValue                 decimal.Decimal          `json:"target_value"`
	Currency                    string                   `json:"currency"`
	ProgressPercentage          decimal.Decimal          `json:"progress_percentage"`
	RemainingAmount             decimal.Decimal          `json:"remaining_amount"`
	MonthsRemaining             int                      `json:"months_remaining"`
	RequiredMonthlyContribution decimal.Decimal          `json:"required_monthly_contribution"`
	IsTargetReached             bool                     `json:"is_target_reached"`
	GoalProgressMetadata        finance.MetricDefinition `json:"goal_progress_metadata"`
	CreatedByUserID             uuid.UUID                `json:"created_by_user_id"`
	CreatedAt                   time.Time                `json:"created_at"`
}

func ToGoalResponse(goal Goal) GoalResponse {
	return GoalResponse{
		ID:              goal.ID,
		PortfolioID:     goal.PortfolioID,
		Name:            goal.Name,
		TargetAmount:    goal.TargetAmount,
		Currency:        goal.Currency,
		TargetDate:      goal.TargetDate.UTC().Format(goalDateLayout),
		Status:          goal.Status,
		CreatedByUserID: goal.CreatedByUserID,
		CreatedAt:       goal.CreatedAt,
		UpdatedAt:       goal.UpdatedAt,
	}
}

func ToMonthlyGoalSnapshotResponse(snapshot MonthlyGoalSnapshot) (MonthlyGoalSnapshotResponse, error) {
	var metadata finance.MetricDefinition
	if err := decodeJSONB(snapshot.GoalProgressMetadata, &metadata); err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}

	return MonthlyGoalSnapshotResponse{
		ID:                          snapshot.ID,
		PortfolioID:                 snapshot.PortfolioID,
		GoalID:                      snapshot.GoalID,
		SnapshotMonthEnd:            snapshot.SnapshotMonthEnd.UTC().Format(goalDateLayout),
		CurrentValue:                snapshot.CurrentValue,
		TargetValue:                 snapshot.TargetValue,
		Currency:                    snapshot.Currency,
		ProgressPercentage:          snapshot.ProgressPercentage,
		RemainingAmount:             snapshot.RemainingAmount,
		MonthsRemaining:             snapshot.MonthsRemaining,
		RequiredMonthlyContribution: snapshot.RequiredMonthlyContribution,
		IsTargetReached:             snapshot.IsTargetReached,
		GoalProgressMetadata:        metadata,
		CreatedByUserID:             snapshot.CreatedByUserID,
		CreatedAt:                   snapshot.CreatedAt,
	}, nil
}

func decodeJSONB(raw []byte, target any) error {
	if len(raw) == 0 {
		raw = []byte("null")
	}
	return json.Unmarshal(raw, target)
}
