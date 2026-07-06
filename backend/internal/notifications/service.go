package notifications

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

const noticeWindowDays = 30

type maturityReader interface {
	ListOpenFixedDepositsMaturingBy(userID uuid.UUID, asOfDate, cutoffDate time.Time) ([]FixedDepositMaturityRecord, error)
	ListActiveGoalsDueBy(userID uuid.UUID, asOfDate, cutoffDate time.Time) ([]GoalTargetRecord, error)
}

type Service struct {
	repo maturityReader
}

func NewService(repo maturityReader) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(userID uuid.UUID, asOfDate time.Time) ([]Response, error) {
	asOfDate = dateOnly(asOfDate)
	cutoff := asOfDate.AddDate(0, 0, noticeWindowDays)
	records, err := s.repo.ListOpenFixedDepositsMaturingBy(userID, asOfDate, cutoff)
	if err != nil {
		return nil, err
	}
	result := make([]Response, 0, len(records))
	for _, record := range records {
		days := int(dateOnly(record.MaturityDate).Sub(asOfDate).Hours() / 24)
		status, explanation := maturityStatus(days, record.DepositName)
		result = append(result, Response{
			ID: "fixed-deposit-maturity:" + record.FixedDepositID.String(), Kind: "fixed_deposit_maturity",
			Status: status, Title: "Fixed deposit maturity", Explanation: explanation,
			TriggerRule: "Open fixed deposit with maturity date no more than 30 calendar days after the as-of date; remains overdue until closure is recorded.",
			AsOfDate:    asOfDate.Format("2006-01-02"), EventDate: record.MaturityDate.UTC().Format("2006-01-02"), DaysUntilEvent: days,
			PortfolioID: record.PortfolioID, PortfolioName: record.PortfolioName,
			AccountID: &record.AccountID, AccountName: record.AccountName,
			EntityID: record.FixedDepositID, EntityType: "fixed_deposit",
		})
	}
	goals, err := s.repo.ListActiveGoalsDueBy(userID, asOfDate, cutoff)
	if err != nil {
		return nil, err
	}
	for _, record := range goals {
		days := int(dateOnly(record.TargetDate).Sub(asOfDate).Hours() / 24)
		status, explanation := goalStatus(days, record.GoalName, record.LatestSnapshotDate)
		var dataAsOfDate *string
		if record.LatestSnapshotDate != nil {
			formatted := record.LatestSnapshotDate.UTC().Format("2006-01-02")
			dataAsOfDate = &formatted
		}
		result = append(result, Response{
			ID: "goal-target-date:" + record.GoalID.String(), Kind: "goal_target_date",
			Status: status, Title: "Financial goal target date", Explanation: explanation,
			TriggerRule: "Active goal not marked reached in its latest monthly snapshot, with target date no more than 30 calendar days after the as-of date; remains overdue until completed, archived, deleted, or reached.",
			AsOfDate:    asOfDate.Format("2006-01-02"), EventDate: record.TargetDate.UTC().Format("2006-01-02"), DaysUntilEvent: days,
			PortfolioID: record.PortfolioID, PortfolioName: record.PortfolioName,
			EntityID: record.GoalID, EntityType: "goal", DataAsOfDate: dataAsOfDate,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].EventDate == result[j].EventDate {
			return result[i].ID < result[j].ID
		}
		return result[i].EventDate < result[j].EventDate
	})
	return result, nil
}

func maturityStatus(days int, name string) (string, string) {
	status := noticeStatus(days)
	switch {
	case days < 0:
		return status, fmt.Sprintf("%s matured %d calendar days ago and has no recorded closure.", name, -days)
	case days == 0:
		return status, fmt.Sprintf("%s matures today and has no recorded closure.", name)
	default:
		return status, fmt.Sprintf("%s matures in %d calendar days.", name, days)
	}
}

func goalStatus(days int, name string, latestSnapshotDate *time.Time) (string, string) {
	status := noticeStatus(days)
	var timing string
	switch {
	case days < 0:
		timing = fmt.Sprintf("%s target date passed %d calendar days ago.", name, -days)
	case days == 0:
		timing = fmt.Sprintf("%s target date is today.", name)
	default:
		timing = fmt.Sprintf("%s target date is in %d calendar days.", name, days)
	}
	if latestSnapshotDate == nil {
		return status, timing + " No monthly goal snapshot is available to determine recorded progress."
	}
	return status, timing + " The latest monthly snapshot on or before the as-of date has not marked the target reached."
}

func noticeStatus(days int) string {
	switch {
	case days < 0:
		return "overdue"
	case days == 0:
		return "due"
	case days <= 7:
		return "urgent"
	default:
		return "upcoming"
	}
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
