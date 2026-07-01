package health

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/allocations"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/risk"
	"github.com/kaustubhmishra/wealth-lens/backend/pkg/finance"
	"github.com/shopspring/decimal"
)

const dateLayout = "2006-01-02"

var healthPeriodsPerYear = decimal.NewFromInt(252)

type allocationReader interface {
	GetCurrent(userID uuid.UUID, portfolioID uuid.UUID) (allocations.PortfolioAllocationResponse, error)
}

type riskReader interface {
	Get(userID uuid.UUID, portfolioID uuid.UUID, startDate, endDate, periodsPerYear string) (risk.PortfolioRiskResponse, error)
}

type Service struct {
	allocations allocationReader
	risk        riskReader
}

func NewService(allocations allocationReader, risk riskReader) *Service {
	return &Service{allocations: allocations, risk: risk}
}

func (s *Service) Get(userID, portfolioID uuid.UUID, req ScoreRequest) (ScoreResponse, error) {
	asOf, err := parseAsOfDate(req.AsOfDate)
	if err != nil {
		return ScoreResponse{}, err
	}
	profile, err := finance.DefaultHealthProfile(req.RiskProfile)
	if err != nil {
		return ScoreResponse{}, common.BadRequest(err.Error())
	}
	configurations, err := normalizeConfigurations(req.CurrencyConfigurations)
	if err != nil {
		return ScoreResponse{}, err
	}

	allocationResponse, err := s.allocations.GetCurrent(userID, portfolioID)
	if err != nil {
		return ScoreResponse{}, err
	}
	allocation := finance.AllocationResult{
		AssetAllocations: allocationResponse.AssetAllocations, AssetClassAllocations: allocationResponse.AssetClassAllocations,
		CashAllocations: allocationResponse.CashAllocations, CurrencyTotals: allocationResponse.CurrencyTotals,
		MissingPrices: allocationResponse.MissingPrices, IsComplete: allocationResponse.IsComplete,
	}
	concentration, err := finance.CalculateConcentration(allocation)
	if err != nil {
		return ScoreResponse{}, common.BadRequest(err.Error())
	}

	start := asOf.AddDate(-1, 0, 0)
	riskResponse, err := s.risk.Get(userID, portfolioID, start.Format(dateLayout), asOf.Format(dateLayout), healthPeriodsPerYear.String())
	if err != nil {
		return ScoreResponse{}, err
	}
	riskByCurrency := make(map[string]risk.CurrencyRiskResponse, len(riskResponse.CurrencyRisk))
	for _, value := range riskResponse.CurrencyRisk {
		riskByCurrency[value.Currency] = value
	}

	scores := make([]finance.HealthScoreResult, 0, len(concentration.Currencies))
	for _, metric := range concentration.Currencies {
		riskMetric, ok := riskByCurrency[metric.Currency]
		if !ok {
			return ScoreResponse{}, common.BadRequest("Historical risk metrics are unavailable for currency " + metric.Currency)
		}
		settings := settingsFor(metric.Currency, profile, configurations)
		drift, unclassified, err := finance.CalculateMaximumRiskCategoryDrift(allocation, metric.Currency, settings.Targets)
		if err != nil {
			return ScoreResponse{}, common.BadRequest(err.Error())
		}
		quality := dataQuality(allocation, metric.Currency, unclassified)
		score, err := finance.CalculateHealthScore(finance.HealthScoreInput{
			Currency: metric.Currency, LargestAssetPercentage: metric.LargestAssetPercentage, HoldingCount: metric.AssetCount,
			MaximumAbsoluteDriftPercentage: drift, AnnualizedVolatilityPercentage: riskMetric.AnnualizedVolatility,
			VolatilityThresholdPercentage: settings.VolatilityThresholdPercentage, MaximumDrawdownPercentage: riskMetric.MaximumDrawdown,
			DrawdownThresholdPercentage: settings.DrawdownThresholdPercentage, DataQuality: quality,
		})
		if err != nil {
			return ScoreResponse{}, common.BadRequest(fmt.Sprintf("Health score error for currency %s: %s", metric.Currency, err.Error()))
		}
		scores = append(scores, score)
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].Currency < scores[j].Currency })
	return ScoreResponse{PortfolioID: portfolioID, StartDate: start.Format(dateLayout), EndDate: asOf.Format(dateLayout), RiskProfile: profile.Name,
		PeriodsPerYear: healthPeriodsPerYear, Scores: scores,
		Scope: "Scores use current ledger-derived allocation and trailing-12-month snapshot risk, calculated independently per currency with disclosed targets and thresholds."}, nil
}

func parseAsOfDate(raw string) (time.Time, error) {
	date, err := time.Parse(dateLayout, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, common.BadRequest("As-of date must use YYYY-MM-DD format")
	}
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if date.After(today) {
		return time.Time{}, common.BadRequest("As-of date cannot be in the future")
	}
	return date, nil
}

func normalizeConfigurations(configurations []CurrencyConfiguration) (map[string]CurrencyConfiguration, error) {
	result := make(map[string]CurrencyConfiguration, len(configurations))
	for _, configuration := range configurations {
		currency := strings.ToUpper(strings.TrimSpace(configuration.Currency))
		if len(currency) != 3 {
			return nil, common.BadRequest("Configuration currency must be a three-letter code")
		}
		if _, exists := result[currency]; exists {
			return nil, common.BadRequest("Duplicate health configuration for currency " + currency)
		}
		configuration.Currency = currency
		if len(configuration.Targets) == 0 || configuration.VolatilityThresholdPercentage == nil || configuration.DrawdownThresholdPercentage == nil {
			return nil, common.BadRequest("Custom currency configuration requires targets and both risk thresholds")
		}
		result[currency] = configuration
	}
	return result, nil
}

func settingsFor(currency string, profile finance.HealthProfile, configurations map[string]CurrencyConfiguration) finance.HealthProfile {
	configuration, ok := configurations[currency]
	if !ok {
		return profile
	}
	return finance.HealthProfile{Name: "custom", Targets: configuration.Targets,
		VolatilityThresholdPercentage: *configuration.VolatilityThresholdPercentage, DrawdownThresholdPercentage: *configuration.DrawdownThresholdPercentage}
}

func dataQuality(allocation finance.AllocationResult, currency string, unclassified bool) string {
	if len(allocation.MissingPrices) > 0 {
		return finance.DataQualityMajor
	}
	if unclassified {
		return finance.DataQualityPartial
	}
	for _, asset := range allocation.AssetAllocations {
		if asset.Currency == currency && (asset.AssetSymbol == "" || asset.AssetName == "") {
			return finance.DataQualityMinor
		}
	}
	return finance.DataQualityComplete
}
