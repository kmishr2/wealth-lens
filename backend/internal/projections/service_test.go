package projections

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/portfolios"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type fakePortfolioReader struct{ err error }

func (f *fakePortfolioReader) GetOwned(userID, portfolioID uuid.UUID) (*portfolios.Portfolio, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &portfolios.Portfolio{ID: portfolioID, UserID: userID}, nil
}

func TestCalculateSIPUsesExplicitInputs(t *testing.T) {
	annualReturn := decimal.NewFromInt(12)
	initial := decimal.NewFromInt(1000)
	monthly := decimal.NewFromInt(100)
	response, err := NewService(&fakePortfolioReader{}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{
		Currency: "inr", InitialInvestment: &initial, MonthlyContribution: &monthly, AnnualReturnPercentage: &annualReturn, Months: 2,
	})
	if err != nil {
		t.Fatalf("CalculateSIP returned error: %v", err)
	}
	if response.Currency != "INR" || !response.ProjectedNominalValue.Equal(decimal.RequireFromString("1221.1")) || response.Scope == "" {
		t.Fatalf("response = %+v", response)
	}
}

func TestCalculateSIPRequiresReturnAssumption(t *testing.T) {
	if _, err := NewService(&fakePortfolioReader{}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{Currency: "INR", Months: 12}); err == nil {
		t.Fatal("CalculateSIP error = nil, want missing return error")
	}
}

func TestCalculateSIPRejectsUnownedPortfolio(t *testing.T) {
	annualReturn := decimal.NewFromInt(10)
	if _, err := NewService(&fakePortfolioReader{err: gorm.ErrRecordNotFound}).CalculateSIP(uuid.New(), uuid.New(), SIPRequest{Currency: "INR", AnnualReturnPercentage: &annualReturn, Months: 12}); err == nil {
		t.Fatal("CalculateSIP error = nil, want portfolio error")
	}
}
