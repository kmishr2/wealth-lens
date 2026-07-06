package fixeddeposits

import (
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const dateLayout = "2006-01-02"

type CreateRequest struct {
	Name               string           `json:"name"`
	BankReference      string           `json:"bank_reference"`
	Principal          *decimal.Decimal `json:"principal"`
	Currency           string           `json:"currency"`
	AnnualInterestRate *decimal.Decimal `json:"annual_interest_rate"`
	StartDate          string           `json:"start_date"`
	MaturityDate       string           `json:"maturity_date"`
	CurrentValue       *decimal.Decimal `json:"current_value"`
	CurrentValueDate   string           `json:"current_value_date"`
}

type ValueCreateRequest struct {
	CurrentValue     *decimal.Decimal `json:"current_value"`
	CurrentValueDate string           `json:"current_value_date"`
}

type CloseRequest struct {
	ClosureType string           `json:"closure_type"`
	ClosedAt    string           `json:"closed_at"`
	Proceeds    *decimal.Decimal `json:"proceeds"`
	Note        string           `json:"note"`
}

type Response struct {
	ID                   uuid.UUID                `json:"id"`
	PortfolioID          uuid.UUID                `json:"portfolio_id"`
	AccountID            uuid.UUID                `json:"account_id"`
	AssetID              uuid.UUID                `json:"asset_id"`
	OpeningTransactionID uuid.UUID                `json:"opening_transaction_id"`
	Name                 string                   `json:"name"`
	BankReference        string                   `json:"bank_reference"`
	Principal            decimal.Decimal          `json:"principal"`
	Currency             string                   `json:"currency"`
	AnnualInterestRate   decimal.Decimal          `json:"annual_interest_rate"`
	StartDate            string                   `json:"start_date"`
	MaturityDate         string                   `json:"maturity_date"`
	CurrentValue         decimal.Decimal          `json:"current_value"`
	CurrentValueAt       time.Time                `json:"current_value_at"`
	ValuationMetadata    finance.MetricDefinition `json:"valuation_metadata"`
	CreatedAt            time.Time                `json:"created_at"`
	Status               string                   `json:"status"`
	DaysToMaturity       int                      `json:"days_to_maturity"`
	ClosureType          *string                  `json:"closure_type,omitempty"`
	ClosedAt             *string                  `json:"closed_at,omitempty"`
	ClosingProceeds      *decimal.Decimal         `json:"closing_proceeds,omitempty"`
	ClosingTransactionID *uuid.UUID               `json:"closing_transaction_id,omitempty"`
}

func applyLifecycle(response *Response, record FixedDeposit, closure *Closure, today time.Time) {
	if closure != nil {
		closedAt := closure.ClosedAt.UTC().Format(dateLayout)
		response.Status = "closed"
		response.ClosureType = &closure.ClosureType
		response.ClosedAt = &closedAt
		response.ClosingProceeds = &closure.Proceeds
		response.ClosingTransactionID = &closure.ClosingTransactionID
		return
	}
	today = utcDate(today)
	if !today.Before(record.MaturityDate) {
		response.Status = "maturity_due"
		return
	}
	response.Status = "active"
	response.DaysToMaturity = int(record.MaturityDate.Sub(today).Hours() / 24)
}

func toResponse(record FixedDeposit, currentValue decimal.Decimal, currentValueAt time.Time) Response {
	return Response{
		ID:                   record.ID,
		PortfolioID:          record.PortfolioID,
		AccountID:            record.AccountID,
		AssetID:              record.AssetID,
		OpeningTransactionID: record.OpeningTransactionID,
		Name:                 record.Name,
		BankReference:        record.BankReference,
		Principal:            record.Principal,
		Currency:             record.Currency,
		AnnualInterestRate:   record.AnnualInterestRate,
		StartDate:            record.StartDate.UTC().Format(dateLayout),
		MaturityDate:         record.MaturityDate.UTC().Format(dateLayout),
		CurrentValue:         currentValue,
		CurrentValueAt:       currentValueAt,
		ValuationMetadata:    finance.FixedDepositValuationDefinition(),
		CreatedAt:            record.CreatedAt,
	}
}
