package snapshots

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
)

const DefaultPortfolioBatchSize = 100

type activePortfolioLister interface {
	ListActiveBatch(afterID *uuid.UUID, limit int) ([]portfolios.Portfolio, error)
}

type dailySnapshotCreator interface {
	CreateDaily(userID uuid.UUID, portfolioID uuid.UUID, req SnapshotCreateRequest) (PortfolioSnapshotResponse, error)
}

type DailyJobFailure struct {
	PortfolioID uuid.UUID `json:"portfolio_id"`
	Error       string    `json:"error"`
}

type DailyJobResult struct {
	SnapshotDate string            `json:"snapshot_date"`
	Processed    int               `json:"processed"`
	Succeeded    int               `json:"succeeded"`
	Failed       int               `json:"failed"`
	Failures     []DailyJobFailure `json:"failures"`
}

type DailyJob struct {
	portfolioRepo   activePortfolioLister
	snapshotService dailySnapshotCreator
	batchSize       int
}

func NewDailyJob(portfolioRepo activePortfolioLister, snapshotService dailySnapshotCreator, batchSize int) *DailyJob {
	if batchSize <= 0 {
		batchSize = DefaultPortfolioBatchSize
	}
	return &DailyJob{
		portfolioRepo:   portfolioRepo,
		snapshotService: snapshotService,
		batchSize:       batchSize,
	}
}

// Run creates or retrieves the immutable daily snapshot for every active
// portfolio. A failure for one portfolio is reported without stopping the
// remaining portfolios. Repository failures stop the job because enumeration
// can no longer be proven complete.
func (j *DailyJob) Run(snapshotDate string) (DailyJobResult, error) {
	result := DailyJobResult{
		SnapshotDate: snapshotDate,
		Failures:     make([]DailyJobFailure, 0),
	}

	date, err := parseSnapshotDate(snapshotDate)
	if err != nil {
		return result, err
	}
	result.SnapshotDate = date.Format(snapshotDateLayout)

	var afterID *uuid.UUID
	for {
		batch, err := j.portfolioRepo.ListActiveBatch(afterID, j.batchSize)
		if err != nil {
			return result, fmt.Errorf("list active portfolios: %w", err)
		}

		for _, portfolio := range batch {
			result.Processed++
			_, err := j.snapshotService.CreateDaily(
				portfolio.UserID,
				portfolio.ID,
				SnapshotCreateRequest{SnapshotDate: result.SnapshotDate},
			)
			if err != nil {
				result.Failed++
				result.Failures = append(result.Failures, DailyJobFailure{
					PortfolioID: portfolio.ID,
					Error:       err.Error(),
				})
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
