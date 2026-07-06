package notifications

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const maturityWindowDays = 30

type maturityReader interface {
	ListOpenFixedDepositsMaturingBy(userID uuid.UUID, asOfDate, cutoffDate time.Time) ([]FixedDepositMaturityRecord, error)
}

type Service struct {
	repo maturityReader
}

func NewService(repo maturityReader) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(userID uuid.UUID, asOfDate time.Time) ([]Response, error) {
	asOfDate = dateOnly(asOfDate)
	cutoff := asOfDate.AddDate(0, 0, maturityWindowDays)
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
			AccountID: record.AccountID, AccountName: record.AccountName,
			EntityID: record.FixedDepositID, EntityType: "fixed_deposit",
		})
	}
	return result, nil
}

func maturityStatus(days int, name string) (string, string) {
	switch {
	case days < 0:
		return "overdue", fmt.Sprintf("%s matured %d calendar days ago and has no recorded closure.", name, -days)
	case days == 0:
		return "due", fmt.Sprintf("%s matures today and has no recorded closure.", name)
	case days <= 7:
		return "urgent", fmt.Sprintf("%s matures in %d calendar days.", name, days)
	default:
		return "upcoming", fmt.Sprintf("%s matures in %d calendar days.", name, days)
	}
}

func dateOnly(value time.Time) time.Time {
	year, month, day := value.UTC().Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
