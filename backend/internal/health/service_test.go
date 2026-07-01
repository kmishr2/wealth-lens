package health

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/allocations"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/risk"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

type fakeAllocationReader struct {
	response allocations.PortfolioAllocationResponse
	err      error
}

func (f *fakeAllocationReader) GetCurrent(uuid.UUID, uuid.UUID) (allocations.PortfolioAllocationResponse, error) {
	return f.response, f.err
}

type fakeRiskReader struct {
	response risk.PortfolioRiskResponse
	err      error
	start    string
	end      string
	periods  string
}

func (f *fakeRiskReader) Get(_ uuid.UUID, _ uuid.UUID, start, end, periods string) (risk.PortfolioRiskResponse, error) {
	f.start, f.end, f.periods = start, end, periods
	return f.response, f.err
}

func TestGetCalculatesPerCurrencyHealthWithModerateDefaults(t *testing.T) {
	riskReader := &fakeRiskReader{response: risk.PortfolioRiskResponse{CurrencyRisk: []risk.CurrencyRiskResponse{{
		Currency: "INR", AnnualizedVolatility: decimal.NewFromInt(12), MaximumDrawdown: decimal.NewFromInt(-10),
	}}}}
	service := NewService(&fakeAllocationReader{response: healthyAllocation()}, riskReader)

	response, err := service.Get(uuid.New(), uuid.New(), ScoreRequest{AsOfDate: "2026-06-30"})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if response.RiskProfile != finance.RiskProfileModerate || len(response.Scores) != 1 || response.Scores[0].Score != 100 {
		t.Fatalf("response = %+v", response)
	}
	if riskReader.start != "2025-06-30" || riskReader.end != "2026-06-30" || riskReader.periods != "252" {
		t.Fatalf("risk arguments = %s, %s, %s", riskReader.start, riskReader.end, riskReader.periods)
	}
}

func TestGetUsesCustomCurrencyThresholds(t *testing.T) {
	riskReader := &fakeRiskReader{response: risk.PortfolioRiskResponse{CurrencyRisk: []risk.CurrencyRiskResponse{{
		Currency: "INR", AnnualizedVolatility: decimal.NewFromInt(12), MaximumDrawdown: decimal.NewFromInt(-10),
	}}}}
	service := NewService(&fakeAllocationReader{response: healthyAllocation()}, riskReader)
	volatility := decimal.NewFromInt(8)
	drawdown := decimal.NewFromInt(-10)

	response, err := service.Get(uuid.New(), uuid.New(), ScoreRequest{AsOfDate: "2026-06-30", CurrencyConfigurations: []CurrencyConfiguration{{
		Currency: "inr", Targets: []finance.RiskCategoryTarget{
			{RiskCategory: "equity", Percentage: decimal.NewFromInt(60)},
			{RiskCategory: "debt", Percentage: decimal.NewFromInt(35)},
			{RiskCategory: "cash_other", Percentage: decimal.NewFromInt(5)},
		}, VolatilityThresholdPercentage: &volatility, DrawdownThresholdPercentage: &drawdown,
	}}})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if response.Scores[0].Score != 82 {
		t.Fatalf("score = %d, want 82", response.Scores[0].Score)
	}
}

func TestGetScoresUnclassifiedAssetAsPartialDataQuality(t *testing.T) {
	allocation := healthyAllocation()
	allocation.AssetAllocations[0].RiskCategory = ""
	service := NewService(&fakeAllocationReader{response: allocation}, &fakeRiskReader{response: risk.PortfolioRiskResponse{CurrencyRisk: []risk.CurrencyRiskResponse{{
		Currency: "INR", AnnualizedVolatility: decimal.NewFromInt(12), MaximumDrawdown: decimal.NewFromInt(-10),
	}}}})

	response, err := service.Get(uuid.New(), uuid.New(), ScoreRequest{AsOfDate: "2026-06-30"})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	components := response.Scores[0].Components
	if components[4].Points != 5 {
		t.Fatalf("data quality points = %d, want 5", components[4].Points)
	}
}

func TestGetRejectsMissingHistoricalCurrency(t *testing.T) {
	service := NewService(&fakeAllocationReader{response: healthyAllocation()}, &fakeRiskReader{})
	if _, err := service.Get(uuid.New(), uuid.New(), ScoreRequest{AsOfDate: "2026-06-30"}); err == nil {
		t.Fatal("Get error = nil, want missing historical currency error")
	}
}

func healthyAllocation() allocations.PortfolioAllocationResponse {
	assets := []finance.AssetAllocation{
		{AssetID: "e1", AssetSymbol: "E1", AssetName: "E1", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "e2", AssetSymbol: "E2", AssetName: "E2", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "e3", AssetSymbol: "E3", AssetName: "E3", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "e4", AssetSymbol: "E4", AssetName: "E4", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "e5", AssetSymbol: "E5", AssetName: "E5", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "e6", AssetSymbol: "E6", AssetName: "E6", Currency: "INR", RiskCategory: "equity", MarketValue: decimal.NewFromInt(10), Percentage: decimal.NewFromInt(10)},
		{AssetID: "d1", AssetSymbol: "D1", AssetName: "D1", Currency: "INR", RiskCategory: "debt", MarketValue: decimal.RequireFromString("17.5"), Percentage: decimal.RequireFromString("17.5")},
		{AssetID: "d2", AssetSymbol: "D2", AssetName: "D2", Currency: "INR", RiskCategory: "debt", MarketValue: decimal.RequireFromString("17.5"), Percentage: decimal.RequireFromString("17.5")},
	}
	return allocations.PortfolioAllocationResponse{AssetAllocations: assets,
		CashAllocations: []finance.CashAllocation{{Currency: "INR", Amount: decimal.NewFromInt(5), Percentage: decimal.NewFromInt(5)}},
		CurrencyTotals:  []finance.CurrencyValue{{Currency: "INR", Amount: decimal.NewFromInt(100)}}, IsComplete: true}
}
