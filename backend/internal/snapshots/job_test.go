package snapshots

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
)

type fakeActivePortfolioLister struct {
	batches    [][]portfolios.Portfolio
	errAtCall  int
	err        error
	calls      int
	afterIDs   []*uuid.UUID
	batchSizes []int
}

func (f *fakeActivePortfolioLister) ListActiveBatch(afterID *uuid.UUID, limit int) ([]portfolios.Portfolio, error) {
	f.calls++
	f.batchSizes = append(f.batchSizes, limit)
	if afterID == nil {
		f.afterIDs = append(f.afterIDs, nil)
	} else {
		value := *afterID
		f.afterIDs = append(f.afterIDs, &value)
	}
	if f.errAtCall == f.calls {
		return nil, f.err
	}
	index := f.calls - 1
	if index >= len(f.batches) {
		return nil, nil
	}
	return f.batches[index], nil
}

type snapshotCreateCall struct {
	userID      uuid.UUID
	portfolioID uuid.UUID
	date        string
}

type fakeDailySnapshotCreator struct {
	calls  []snapshotCreateCall
	errors map[uuid.UUID]error
}

func (f *fakeDailySnapshotCreator) CreateDaily(userID uuid.UUID, portfolioID uuid.UUID, req SnapshotCreateRequest) (PortfolioSnapshotResponse, error) {
	f.calls = append(f.calls, snapshotCreateCall{userID: userID, portfolioID: portfolioID, date: req.SnapshotDate})
	if err := f.errors[portfolioID]; err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	return PortfolioSnapshotResponse{PortfolioID: portfolioID, SnapshotDate: req.SnapshotDate}, nil
}

func TestDailyJobProcessesActivePortfoliosInBatches(t *testing.T) {
	first := portfolios.Portfolio{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), UserID: uuid.New()}
	second := portfolios.Portfolio{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), UserID: uuid.New()}
	third := portfolios.Portfolio{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), UserID: uuid.New()}
	lister := &fakeActivePortfolioLister{batches: [][]portfolios.Portfolio{{first, second}, {third}}}
	creator := &fakeDailySnapshotCreator{errors: map[uuid.UUID]error{}}

	result, err := NewDailyJob(lister, creator, 2).Run("2026-01-15")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Processed != 3 || result.Succeeded != 3 || result.Failed != 0 {
		t.Fatalf("result = %+v, want processed=3 succeeded=3 failed=0", result)
	}
	if lister.calls != 2 || lister.afterIDs[0] != nil || lister.afterIDs[1] == nil || *lister.afterIDs[1] != second.ID {
		t.Fatalf("batch cursors = %+v, want nil then %s", lister.afterIDs, second.ID)
	}
	if len(creator.calls) != 3 {
		t.Fatalf("create calls = %d, want 3", len(creator.calls))
	}
	for index, call := range creator.calls {
		portfolio := []portfolios.Portfolio{first, second, third}[index]
		if call.userID != portfolio.UserID || call.portfolioID != portfolio.ID || call.date != "2026-01-15" {
			t.Fatalf("call %d = %+v, want owner=%s portfolio=%s date=2026-01-15", index, call, portfolio.UserID, portfolio.ID)
		}
	}
}

func TestDailyJobContinuesAfterPortfolioFailure(t *testing.T) {
	first := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	second := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	creator := &fakeDailySnapshotCreator{errors: map[uuid.UUID]error{first.ID: errors.New("missing prices")}}
	lister := &fakeActivePortfolioLister{batches: [][]portfolios.Portfolio{{first, second}}}

	result, err := NewDailyJob(lister, creator, 10).Run("2026-01-15")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Processed != 2 || result.Succeeded != 1 || result.Failed != 1 || len(result.Failures) != 1 {
		t.Fatalf("result = %+v, want one success and one failure", result)
	}
	if result.Failures[0].PortfolioID != first.ID || result.Failures[0].Error != "missing prices" {
		t.Fatalf("failure = %+v, want portfolio %s missing prices", result.Failures[0], first.ID)
	}
	if len(creator.calls) != 2 {
		t.Fatalf("create calls = %d, want job to continue with second portfolio", len(creator.calls))
	}
}

func TestDailyJobStopsWhenPortfolioEnumerationFails(t *testing.T) {
	first := portfolios.Portfolio{ID: uuid.New(), UserID: uuid.New()}
	lister := &fakeActivePortfolioLister{
		batches:   [][]portfolios.Portfolio{{first}},
		errAtCall: 2,
		err:       errors.New("database unavailable"),
	}
	creator := &fakeDailySnapshotCreator{errors: map[uuid.UUID]error{}}

	result, err := NewDailyJob(lister, creator, 1).Run("2026-01-15")
	if err == nil || err.Error() != "list active portfolios: database unavailable" {
		t.Fatalf("error = %v, want enumeration error", err)
	}
	if result.Processed != 1 || result.Succeeded != 1 {
		t.Fatalf("partial result = %+v, want first batch recorded", result)
	}
}

func TestDailyJobRejectsInvalidDateBeforeListingPortfolios(t *testing.T) {
	lister := &fakeActivePortfolioLister{}
	creator := &fakeDailySnapshotCreator{errors: map[uuid.UUID]error{}}

	_, err := NewDailyJob(lister, creator, 0).Run("not-a-date")
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if lister.calls != 0 || len(creator.calls) != 0 {
		t.Fatalf("lister calls = %d creator calls = %d, want no work", lister.calls, len(creator.calls))
	}
}
