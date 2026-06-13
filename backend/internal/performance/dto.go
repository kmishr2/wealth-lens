package performance

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const performanceDateLayout = "2006-01-02"

type PortfolioPerformanceResponse struct {
	PortfolioID      uuid.UUID                     `json:"portfolio_id"`
	StartDate        string                        `json:"start_date"`
	EndDate          string                        `json:"end_date"`
	CurrencyReturns  []CurrencyPerformanceResponse `json:"currency_returns"`
	PerformanceScope string                        `json:"performance_scope"`
	PnLMetadata      finance.MetricDefinition      `json:"pnl_metadata"`
	CAGRMetadata     finance.MetricDefinition      `json:"cagr_metadata"`
	XIRRMetadata     finance.MetricDefinition      `json:"xirr_metadata"`
}

type CurrencyPerformanceResponse struct {
	Currency            string          `json:"currency"`
	BeginningValue      decimal.Decimal `json:"beginning_value"`
	EndingValue         decimal.Decimal `json:"ending_value"`
	NetExternalCashFlow decimal.Decimal `json:"net_external_cash_flow"`
	ProfitLoss          decimal.Decimal `json:"profit_loss"`
	CAGR                decimal.Decimal `json:"cagr"`
	XIRR                decimal.Decimal `json:"xirr"`
	CashFlowCount       int             `json:"cash_flow_count"`
}

func dateString(date time.Time) string {
	return date.UTC().Format(performanceDateLayout)
}
