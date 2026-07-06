package notifications

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeMaturityReader struct {
	records []FixedDepositMaturityRecord
	asOf    time.Time
	cutoff  time.Time
}

func (f *fakeMaturityReader) ListOpenFixedDepositsMaturingBy(_ uuid.UUID, asOfDate, cutoffDate time.Time) ([]FixedDepositMaturityRecord, error) {
	f.asOf, f.cutoff = asOfDate, cutoffDate
	return f.records, nil
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
