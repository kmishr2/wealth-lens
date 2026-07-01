package goals

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
)

const DefaultPortfolioBatchSize = 100

type activePortfolioLister interface {
	ListActiveBatch(afterID *uuid.UUID, limit int) ([]portfolios.Portfolio, error)
}

type monthlyPortfolioSnapshotCreator interface {
	CreateMonthlySnapshotsForPortfolio(userID uuid.UUID, portfolioID uuid.UUID, snapshotMonthEnd string) ([]MonthlyGoalSnapshotResponse, error)
}

type MonthlySnapshotFailure struct {
	PortfolioID uuid.UUID `json:"portfolio_id"`
	Error       string    `json:"error"`
}

type MonthlySnapshotJobResult struct {
	SnapshotMonthEnd string                   `json:"snapshot_month_end"`
	Processed        int                      `json:"processed"`
	Succeeded        int                      `json:"succeeded"`
	Failed           int                      `json:"failed"`
	Failures         []MonthlySnapshotFailure `json:"failures"`
}

type MonthlySnapshotJob struct {
	portfolioRepo activePortfolioLister
	goalService   monthlyPortfolioSnapshotCreator
	batchSize     int
}

func NewMonthlySnapshotJob(portfolioRepo activePortfolioLister, goalService monthlyPortfolioSnapshotCreator, batchSize int) *MonthlySnapshotJob {
	if batchSize <= 0 {
		batchSize = DefaultPortfolioBatchSize
	}
	return &MonthlySnapshotJob{portfolioRepo: portfolioRepo, goalService: goalService, batchSize: batchSize}
}

func (j *MonthlySnapshotJob) Run(snapshotMonthEnd string) (MonthlySnapshotJobResult, error) {
	result := MonthlySnapshotJobResult{
		SnapshotMonthEnd: snapshotMonthEnd,
		Failures:         make([]MonthlySnapshotFailure, 0),
	}
	monthEnd, err := parseMonthEnd(snapshotMonthEnd)
	if err != nil {
		return result, err
	}
	result.SnapshotMonthEnd = monthEnd.Format(goalDateLayout)

	var afterID *uuid.UUID
	for {
		batch, err := j.portfolioRepo.ListActiveBatch(afterID, j.batchSize)
		if err != nil {
			return result, fmt.Errorf("list active portfolios: %w", err)
		}
		for _, portfolio := range batch {
			result.Processed++
			_, err := j.goalService.CreateMonthlySnapshotsForPortfolio(portfolio.UserID, portfolio.ID, result.SnapshotMonthEnd)
			if err != nil {
				result.Failed++
				result.Failures = append(result.Failures, MonthlySnapshotFailure{PortfolioID: portfolio.ID, Error: err.Error()})
				continue
			}
			result.Succeeded++
		}
		if len(batch) < j.batchSize {
			break
		}
		lastID := batch[len(batch)-1].ID
		afterID = &lastID
	}
	return result, nil
}
