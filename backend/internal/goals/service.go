package goals

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/snapshots"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type portfolioReader interface {
	GetOwned(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error)
}

type snapshotReader interface {
	GetByPortfolioDatePeriod(portfolioID uuid.UUID, snapshotDate time.Time, snapshotPeriod string) (*snapshots.PortfolioSnapshot, error)
}

type goalStore interface {
	Create(goal *Goal) error
	ListByPortfolio(portfolioID uuid.UUID, pagination common.Pagination) ([]Goal, error)
	ListActiveByPortfolio(portfolioID uuid.UUID) ([]Goal, error)
	GetInPortfolio(portfolioID uuid.UUID, goalID uuid.UUID) (*Goal, error)
	Update(goal *Goal) error
	Delete(goal *Goal) error
	CreateMonthlySnapshot(snapshot *MonthlyGoalSnapshot) error
	GetMonthlySnapshotByGoalMonth(goalID uuid.UUID, monthEnd time.Time) (*MonthlyGoalSnapshot, error)
	ListMonthlySnapshots(goalID uuid.UUID, pagination common.Pagination) ([]MonthlyGoalSnapshot, error)
}

type Service struct {
	repo          goalStore
	portfolioRepo portfolioReader
	snapshotRepo  snapshotReader
}

func NewService(repo goalStore, portfolioRepo portfolioReader, snapshotRepo snapshotReader) *Service {
	return &Service{repo: repo, portfolioRepo: portfolioRepo, snapshotRepo: snapshotRepo}
}

func (s *Service) Create(userID uuid.UUID, portfolioID uuid.UUID, req GoalCreateRequest) (GoalResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return GoalResponse{}, err
	}
	goal, err := buildGoal(userID, portfolioID, req)
	if err != nil {
		return GoalResponse{}, err
	}
	if err := s.repo.Create(goal); err != nil {
		if common.IsUniqueViolation(err) {
			return GoalResponse{}, common.Conflict("Goal name already exists")
		}
		return GoalResponse{}, err
	}
	return ToGoalResponse(*goal), nil
}

func (s *Service) List(userID uuid.UUID, portfolioID uuid.UUID, pagination common.Pagination) ([]GoalResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}
	goals, err := s.repo.ListByPortfolio(portfolioID, pagination)
	if err != nil {
		return nil, err
	}
	responses := make([]GoalResponse, 0, len(goals))
	for _, goal := range goals {
		responses = append(responses, ToGoalResponse(goal))
	}
	return responses, nil
}

func (s *Service) Update(userID uuid.UUID, portfolioID uuid.UUID, goalID uuid.UUID, req GoalUpdateRequest) (GoalResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return GoalResponse{}, err
	}
	goal, err := s.getGoal(portfolioID, goalID)
	if err != nil {
		return GoalResponse{}, err
	}
	if err := applyGoalUpdate(goal, req); err != nil {
		return GoalResponse{}, err
	}
	if err := s.repo.Update(goal); err != nil {
		if common.IsUniqueViolation(err) {
			return GoalResponse{}, common.Conflict("Goal name already exists")
		}
		return GoalResponse{}, err
	}
	return ToGoalResponse(*goal), nil
}

func (s *Service) Delete(userID uuid.UUID, portfolioID uuid.UUID, goalID uuid.UUID) error {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return err
	}
	goal, err := s.getGoal(portfolioID, goalID)
	if err != nil {
		return err
	}
	return s.repo.Delete(goal)
}

func (s *Service) CreateMonthlySnapshot(userID uuid.UUID, portfolioID uuid.UUID, goalID uuid.UUID, snapshotMonthEnd string) (MonthlyGoalSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	goal, err := s.getGoal(portfolioID, goalID)
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	monthEnd, err := parseMonthEnd(snapshotMonthEnd)
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	return s.createMonthlySnapshotForGoal(userID, *goal, monthEnd)
}

func (s *Service) ListMonthlySnapshots(userID uuid.UUID, portfolioID uuid.UUID, goalID uuid.UUID, pagination common.Pagination) ([]MonthlyGoalSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}
	if _, err := s.getGoal(portfolioID, goalID); err != nil {
		return nil, err
	}
	records, err := s.repo.ListMonthlySnapshots(goalID, pagination)
	if err != nil {
		return nil, err
	}
	responses := make([]MonthlyGoalSnapshotResponse, 0, len(records))
	for _, record := range records {
		response, err := ToMonthlyGoalSnapshotResponse(record)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *Service) CreateMonthlySnapshotsForPortfolio(userID uuid.UUID, portfolioID uuid.UUID, snapshotMonthEnd string) ([]MonthlyGoalSnapshotResponse, error) {
	if _, err := s.getOwnedPortfolio(userID, portfolioID); err != nil {
		return nil, err
	}
	monthEnd, err := parseMonthEnd(snapshotMonthEnd)
	if err != nil {
		return nil, err
	}
	activeGoals, err := s.repo.ListActiveByPortfolio(portfolioID)
	if err != nil {
		return nil, err
	}
	responses := make([]MonthlyGoalSnapshotResponse, 0, len(activeGoals))
	for _, goal := range activeGoals {
		response, err := s.createMonthlySnapshotForGoal(userID, goal, monthEnd)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *Service) createMonthlySnapshotForGoal(userID uuid.UUID, goal Goal, monthEnd time.Time) (MonthlyGoalSnapshotResponse, error) {
	existing, err := s.repo.GetMonthlySnapshotByGoalMonth(goal.ID, monthEnd)
	if err == nil {
		return ToMonthlyGoalSnapshotResponse(*existing)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return MonthlyGoalSnapshotResponse{}, err
	}

	dailySnapshot, err := s.snapshotRepo.GetByPortfolioDatePeriod(goal.PortfolioID, monthEnd, snapshots.SnapshotPeriodDaily)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return MonthlyGoalSnapshotResponse{}, common.NotFound("Month-end daily portfolio snapshot not found")
	}
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	dailyResponse, err := snapshots.ToResponse(*dailySnapshot)
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	currentValue, ok := totalForCurrency(dailyResponse.TotalValues, goal.Currency)
	if !ok {
		return MonthlyGoalSnapshotResponse{}, common.BadRequest("Month-end daily portfolio snapshot does not contain the goal currency")
	}
	result, err := finance.CalculateGoalProgress(finance.GoalProgressInput{
		CurrentValue: currentValue,
		TargetValue:  goal.TargetAmount,
		SnapshotDate: monthEnd,
		TargetDate:   goal.TargetDate,
	})
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, common.BadRequest(fmt.Sprintf("Goal progress error: %s", err.Error()))
	}
	metadata, err := snapshots.NewJSONB(result.Definition)
	if err != nil {
		return MonthlyGoalSnapshotResponse{}, err
	}
	record := &MonthlyGoalSnapshot{
		PortfolioID:                 goal.PortfolioID,
		GoalID:                      goal.ID,
		SnapshotMonthEnd:            monthEnd,
		CurrentValue:                currentValue,
		TargetValue:                 goal.TargetAmount,
		Currency:                    goal.Currency,
		ProgressPercentage:          result.ProgressPercentage,
		RemainingAmount:             result.RemainingAmount,
		MonthsRemaining:             result.MonthsRemaining,
		RequiredMonthlyContribution: result.RequiredMonthlyContribution,
		IsTargetReached:             result.IsTargetReached,
		GoalProgressMetadata:        metadata,
		CreatedByUserID:             userID,
	}
	if err := s.repo.CreateMonthlySnapshot(record); err != nil {
		if common.IsUniqueViolation(err) {
			existing, getErr := s.repo.GetMonthlySnapshotByGoalMonth(goal.ID, monthEnd)
			if getErr == nil {
				return ToMonthlyGoalSnapshotResponse(*existing)
			}
			return MonthlyGoalSnapshotResponse{}, getErr
		}
		return MonthlyGoalSnapshotResponse{}, err
	}
	return ToMonthlyGoalSnapshotResponse(*record)
}

func (s *Service) getOwnedPortfolio(userID uuid.UUID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	portfolio, err := s.portfolioRepo.GetOwned(userID, portfolioID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Portfolio not found")
	}
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (s *Service) getGoal(portfolioID uuid.UUID, goalID uuid.UUID) (*Goal, error) {
	goal, err := s.repo.GetInPortfolio(portfolioID, goalID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NotFound("Goal not found")
	}
	if err != nil {
		return nil, err
	}
	return goal, nil
}

func buildGoal(userID uuid.UUID, portfolioID uuid.UUID, req GoalCreateRequest) (*Goal, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, common.BadRequest("Goal name is required")
	}
	if req.TargetAmount == nil || !req.TargetAmount.GreaterThan(decimal.Zero) {
		return nil, common.BadRequest("Target amount must be greater than zero")
	}
	currency := common.NormalizeCurrency(req.Currency)
	if !common.ValidateCurrency(currency) {
		return nil, common.BadRequest("Currency must be a three-letter uppercase code")
	}
	targetDate, err := parseGoalDate(req.TargetDate, "Target date")
	if err != nil {
		return nil, err
	}

	return &Goal{
		PortfolioID:     portfolioID,
		Name:            name,
		TargetAmount:    req.TargetAmount.Round(10),
		Currency:        currency,
		TargetDate:      targetDate,
		Status:          StatusActive,
		CreatedByUserID: userID,
	}, nil
}

func applyGoalUpdate(goal *Goal, req GoalUpdateRequest) error {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return common.BadRequest("Goal name is required")
		}
		goal.Name = name
	}
	if req.TargetAmount != nil {
		if !req.TargetAmount.GreaterThan(decimal.Zero) {
			return common.BadRequest("Target amount must be greater than zero")
		}
		goal.TargetAmount = req.TargetAmount.Round(10)
	}
	if req.Currency != nil {
		currency := common.NormalizeCurrency(*req.Currency)
		if !common.ValidateCurrency(currency) {
			return common.BadRequest("Currency must be a three-letter uppercase code")
		}
		goal.Currency = currency
	}
	if req.TargetDate != nil {
		targetDate, err := parseGoalDate(*req.TargetDate, "Target date")
		if err != nil {
			return err
		}
		goal.TargetDate = targetDate
	}
	if req.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*req.Status))
		if !common.OneOf(status, StatusActive, StatusCompleted, StatusArchived) {
			return common.BadRequest("Status must be active, completed, or archived")
		}
		goal.Status = status
	}
	return nil
}

func parseGoalDate(raw string, label string) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s is required", label))
	}
	date, err := time.Parse(goalDateLayout, raw)
	if err != nil {
		return time.Time{}, common.BadRequest(fmt.Sprintf("%s must use YYYY-MM-DD format", label))
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC), nil
}

func parseMonthEnd(raw string) (time.Time, error) {
	date, err := parseGoalDate(raw, "Snapshot month end")
	if err != nil {
		return time.Time{}, err
	}
	nextDay := date.AddDate(0, 0, 1)
	if nextDay.Day() != 1 {
		return time.Time{}, common.BadRequest("Snapshot month end must be the last UTC day of a month")
	}
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, common.BadRequest("Snapshot month end cannot be in the future")
	}
	return date, nil
}

func totalForCurrency(values []finance.CurrencyValue, currency string) (decimal.Decimal, bool) {
	for _, value := range values {
		if value.Currency == currency {
			return value.Amount, true
		}
	}
	return decimal.Zero, false
}
