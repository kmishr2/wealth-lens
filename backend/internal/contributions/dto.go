package contributions

import (
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
)

type Response struct {
	PortfolioID uuid.UUID `json:"portfolio_id"`
	Currency    string    `json:"currency"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	finance.ContributionAnalysisResult
	Scope string `json:"scope"`
}
