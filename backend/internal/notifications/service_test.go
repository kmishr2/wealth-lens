package notifications

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeMaturityReader struct {
	records []FixedDepositMaturityRecord
	goals   []GoalTargetRecord
	assets  []HeldAssetPriceRecord
	asOf    time.Time
	cutoff  time.Time
}

func (f *fakeMaturityReader) ListActiveGoalsDueBy(_ uuid.UUID, _, _ time.Time) ([]GoalTargetRecord, error) {
	return f.goals, nil
}

func (f *fakeMaturityReader) ListHeldAssetsWithLatestPrice(_ uuid.UUID, _ time.Time) ([]HeldAssetPriceRecord, error) {
	return f.assets, nil
}

func (f *fakeMaturityReader) ListOpenFixedDepositsMaturingBy(_ uuid.UUID, asOfDate, cutoffDate time.Time) ([]FixedDepositMaturityRecord, error) {
	f.asOf, f.cutoff = asOfDate, cutoffDate
	return f.records, nil
}

func TestListBuildsGoalTargetNoticeFromLatestSnapshot(t *testing.T) {
	goalID := uuid.New()
	snapshotDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	repo := &fakeMaturityReader{goals: []GoalTargetRecord{{
		GoalID: goalID, PortfolioID: uuid.New(), PortfolioName: "Long term", GoalName: "Education",
		TargetDate: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC), LatestSnapshotDate: &snapshotDate,
	}}}
	result, err := NewService(repo).List(uuid.New(), time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Kind != "goal_target_date" || result[0].Status != "upcoming" || result[0].DaysUntilEvent != 14 {
		t.Fatalf("goal notification = %+v", result)
	}
	if result[0].DataAsOfDate == nil || *result[0].DataAsOfDate != "2026-06-30" {
		t.Fatalf("data as-of date = %v", result[0].DataAsOfDate)
	}
}

func TestListBuildsMissingAndStalePriceNotices(t *testing.T) {
	staleDate := time.Date(2026, 6, 29, 0, 0, 0, 0, time.UTC)
	repo := &fakeMaturityReader{assets: []HeldAssetPriceRecord{
		{AssetID: uuid.New(), AssetName: "Missing Asset", AssetSymbol: "MISS", PortfolioID: uuid.New(), PortfolioName: "Portfolio"},
		{AssetID: uuid.New(), AssetName: "Stale Asset", AssetSymbol: "STALE", PortfolioID: uuid.New(), PortfolioName: "Portfolio", LatestPriceAt: &staleDate},
	}}
	result, err := NewService(repo).ListWithStaleAfterDays(uuid.New(), time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("price notices = %+v", result)
	}
	byKind := map[string]Response{}
	for _, notification := range result {
		byKind[notification.Kind] = notification
	}
	if byKind["asset_price_missing"].Status != "missing" {
		t.Fatalf("missing notice = %+v", byKind["asset_price_missing"])
	}
	stale := byKind["asset_price_stale"]
	if stale.Status != "stale" || stale.AgeDays == nil || *stale.AgeDays != 7 || stale.ThresholdDays == nil || *stale.ThresholdDays != 5 {
		t.Fatalf("stale notice = %+v", stale)
	}
}

func TestListBuildsDeterministicMaturityNotifications(t *testing.T) {
	depositID := uuid.New()
	repo := &fakeMaturityReader{records: []FixedDepositMaturityRecord{{
		FixedDepositID: depositID, PortfolioID: uuid.New(), PortfolioName: "Long term",
		AccountID: uuid.New(), AccountName: "Bank", DepositName: "One-year FD",
		MaturityDate: time.Date(2026, 7, 10, 15, 0, 0, 0, time.UTC),
	}}}
	result, err := NewService(repo).List(uuid.New(), time.Date(2026, 7, 6, 22, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("notifications = %d, want 1", len(result))
	}
	notification := result[0]
	if notification.ID != "fixed-deposit-maturity:"+depositID.String() || notification.Status != "urgent" || notification.DaysUntilEvent != 4 {
		t.Fatalf("notification = %+v", notification)
	}
	if repo.asOf.Format("2006-01-02") != "2026-07-06" || repo.cutoff.Format("2006-01-02") != "2026-08-05" {
		t.Fatalf("query window = %s through %s", repo.asOf, repo.cutoff)
	}
	if notification.TriggerRule == "" || notification.Explanation == "" {
		t.Fatal("notification explainability fields are missing")
	}
}

func TestMaturityStatusBands(t *testing.T) {
	tests := []struct {
		days int
		want string
	}{{-1, "overdue"}, {0, "due"}, {1, "urgent"}, {7, "urgent"}, {8, "upcoming"}, {30, "upcoming"}}
	for _, test := range tests {
		got, _ := maturityStatus(test.days, "FD")
		if got != test.want {
			t.Errorf("maturityStatus(%d) = %q, want %q", test.days, got, test.want)
		}
	}
}
