package snapshots

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const snapshotDateLayout = "2006-01-02"

type SnapshotCreateRequest struct {
	SnapshotDate string `json:"snapshot_date"`
}

type PortfolioSnapshotResponse struct {
	ID                    uuid.UUID                      `json:"id"`
	PortfolioID           uuid.UUID                      `json:"portfolio_id"`
	SnapshotDate          string                         `json:"snapshot_date"`
	SnapshotPeriod        string                         `json:"snapshot_period"`
	TotalValues           []finance.CurrencyValue        `json:"total_values"`
	AssetAllocations      []finance.AssetAllocation      `json:"asset_allocations"`
	AssetClassAllocations []finance.AssetClassAllocation `json:"asset_class_allocations"`
	CashAllocations       []finance.CashAllocation       `json:"cash_allocations"`
	MissingPrices         []finance.MissingPrice         `json:"missing_prices"`
	IsFullyValued         bool                           `json:"is_fully_valued"`
	ValuationScope        string                         `json:"valuation_scope"`
	AllocationScope       string                         `json:"allocation_scope"`
	ValuationMetadata     finance.MetricDefinition       `json:"valuation_metadata"`
	AllocationMetadata    finance.MetricDefinition       `json:"allocation_metadata"`
	HoldingsMetadata      finance.MetricDefinition       `json:"holdings_metadata"`
	CreatedByUserID       uuid.UUID                      `json:"created_by_user_id"`
	CreatedAt             time.Time                      `json:"created_at"`
}

type WeeklyPerformanceSnapshotResponse struct {
	ID               uuid.UUID                   `json:"id"`
	PortfolioID      uuid.UUID                   `json:"portfolio_id"`
	WeekStartDate    string                      `json:"week_start_date"`
	WeekEndDate      string                      `json:"week_end_date"`
	CurrencyReturns  []WeeklyCurrencyPerformance `json:"currency_returns"`
	PerformanceScope string                      `json:"performance_scope"`
	PnLMetadata      finance.MetricDefinition    `json:"pnl_metadata"`
	CAGRMetadata     finance.MetricDefinition    `json:"cagr_metadata"`
	XIRRMetadata     finance.MetricDefinition    `json:"xirr_metadata"`
	CreatedByUserID  uuid.UUID                   `json:"created_by_user_id"`
	CreatedAt        time.Time                   `json:"created_at"`
}

type WeeklyCurrencyPerformance struct {
	Currency            string          `json:"currency"`
	BeginningValue      decimal.Decimal `json:"beginning_value"`
	EndingValue         decimal.Decimal `json:"ending_value"`
	NetExternalCashFlow decimal.Decimal `json:"net_external_cash_flow"`
	ProfitLoss          decimal.Decimal `json:"profit_loss"`
	CAGR                decimal.Decimal `json:"cagr"`
	XIRR                decimal.Decimal `json:"xirr"`
	CashFlowCount       int             `json:"cash_flow_count"`
}

func ToResponse(snapshot PortfolioSnapshot) (PortfolioSnapshotResponse, error) {
	var totalValues []finance.CurrencyValue
	var assetAllocations []finance.AssetAllocation
	var assetClassAllocations []finance.AssetClassAllocation
	var cashAllocations []finance.CashAllocation
	var missingPrices []finance.MissingPrice
	var valuationMetadata finance.MetricDefinition
	var allocationMetadata finance.MetricDefinition
	var holdingsMetadata finance.MetricDefinition

	if err := decodeJSONB(snapshot.TotalValues, &totalValues); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.AssetAllocations, &assetAllocations); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.AssetClassAllocations, &assetClassAllocations); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.CashAllocations, &cashAllocations); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.MissingPrices, &missingPrices); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.ValuationMetadata, &valuationMetadata); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.AllocationMetadata, &allocationMetadata); err != nil {
		return PortfolioSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.HoldingsMetadata, &holdingsMetadata); err != nil {
		return PortfolioSnapshotResponse{}, err
	}

	return PortfolioSnapshotResponse{
		ID:                    snapshot.ID,
		PortfolioID:           snapshot.PortfolioID,
		SnapshotDate:          snapshot.SnapshotDate.UTC().Format(snapshotDateLayout),
		SnapshotPeriod:        snapshot.SnapshotPeriod,
		TotalValues:           totalValues,
		AssetAllocations:      assetAllocations,
		AssetClassAllocations: assetClassAllocations,
		CashAllocations:       cashAllocations,
		MissingPrices:         missingPrices,
		IsFullyValued:         snapshot.IsFullyValued,
		ValuationScope:        snapshot.ValuationScope,
		AllocationScope:       snapshot.AllocationScope,
		ValuationMetadata:     valuationMetadata,
		AllocationMetadata:    allocationMetadata,
		HoldingsMetadata:      holdingsMetadata,
		CreatedByUserID:       snapshot.CreatedByUserID,
		CreatedAt:             snapshot.CreatedAt,
	}, nil
}

func ToResponses(snapshots []PortfolioSnapshot) ([]PortfolioSnapshotResponse, error) {
	responses := make([]PortfolioSnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		response, err := ToResponse(snapshot)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func ToWeeklyPerformanceResponse(snapshot WeeklyPerformanceSnapshot) (WeeklyPerformanceSnapshotResponse, error) {
	var currencyReturns []WeeklyCurrencyPerformance
	var pnlMetadata finance.MetricDefinition
	var cagrMetadata finance.MetricDefinition
	var xirrMetadata finance.MetricDefinition

	if err := decodeJSONB(snapshot.CurrencyReturns, &currencyReturns); err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.PnLMetadata, &pnlMetadata); err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.CAGRMetadata, &cagrMetadata); err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}
	if err := decodeJSONB(snapshot.XIRRMetadata, &xirrMetadata); err != nil {
		return WeeklyPerformanceSnapshotResponse{}, err
	}

	return WeeklyPerformanceSnapshotResponse{
		ID:               snapshot.ID,
		PortfolioID:      snapshot.PortfolioID,
		WeekStartDate:    snapshot.WeekStartDate.UTC().Format(snapshotDateLayout),
		WeekEndDate:      snapshot.WeekEndDate.UTC().Format(snapshotDateLayout),
		CurrencyReturns:  currencyReturns,
		PerformanceScope: snapshot.PerformanceScope,
		PnLMetadata:      pnlMetadata,
		CAGRMetadata:     cagrMetadata,
		XIRRMetadata:     xirrMetadata,
		CreatedByUserID:  snapshot.CreatedByUserID,
		CreatedAt:        snapshot.CreatedAt,
	}, nil
}

func decodeJSONB(raw JSONB, target any) error {
	if len(raw) == 0 {
		raw = JSONB("null")
	}
	return json.Unmarshal(raw, target)
}
