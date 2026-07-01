package goals

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
)

type fakeActivePortfolioLister struct {
	portfolios []portfolios.Portfolio
	calls      int
	err        error
}

func (f *fakeActivePortfolioLister) ListActiveBatch(afterID *uuid.UUID, limit int) ([]portfolios.Portfolio, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	start := 0
	if afterID != nil {
		for index, portfolio := range f.portfolios {
			if portfolio.ID == *afterID {
				start = index + 1
				break
			}
		}
	}
	end := start + limit
	if end > len(f.portfolios) {
		end = len(f.portfolios)
	}
	return f.portfolios[start:end], nil
}

type fakeMonthlyPortfolioSnapshotCreator struct {
	failPortfolio uuid.UUID
	calls         []uuid.UUID
}

func (f *fakeMonthlyPortfolioSnapshotCreator) CreateMonthlySnapshotsForPortfolio(_ uuid.UUID, portfolioID uuid.UUID, _ string) ([]MonthlyGoalSnapshotResponse, error) {
	f.calls = append(f.calls, portfolioID)
	if portfolioID == f.failPortfolio {
		return nil, errors.New("snapshot failed")
	}
	return []MonthlyGoalSnapshotResponse{}, nil
}

func TestMonthlySnapshotJobBatchesAndContinuesAfterPortfolioFailure(t *testing.T) {
	first := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	second := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	third := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	repo := &fakeActivePortfolioLister{portfolios: []portfolios.Portfolio{first, second, third}}
	creator := &fakeMonthlyPortfolioSnapshotCreator{failPortfolio: second.ID}

	result, err := NewMonthlySnapshotJob(repo, creator, 2).Run("2026-06-30")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Processed != 3 || result.Succeeded != 2 || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Failures) != 1 || result.Failures[0].PortfolioID != second.ID {
		t.Fatalf("failures = %+v", result.Failures)
	}
	if repo.calls != 2 || len(creator.calls) != 3 {
		t.Fatalf("repo calls = %d, creator calls = %d", repo.calls, len(creator.calls))
	}
}

func TestMonthlySnapshotJobRejectsNonMonthEndBeforeReadingPortfolios(t *testing.T) {
	repo := &fakeActivePortfolioLister{}
	creator := &fakeMonthlyPortfolioSnapshotCreator{}

	_, err := NewMonthlySnapshotJob(repo, creator, 0).Run("2026-06-29")
	if err == nil {
		t.Fatal("Run error = nil, want invalid month-end error")
	}
	if repo.calls != 0 {
		t.Fatalf("repo calls = %d, want 0", repo.calls)
	}
}

func TestMonthlySnapshotJobReturnsPortfolioReadError(t *testing.T) {
	repo := &fakeActivePortfolioLister{err: errors.New("database unavailable")}
	creator := &fakeMonthlyPortfolioSnapshotCreator{}

	_, err := NewMonthlySnapshotJob(repo, creator, 10).Run("2026-06-30")
	if err == nil {
		t.Fatal("Run error = nil, want repository error")
	}
}
