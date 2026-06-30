package marketdata

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
)

type identifiedAssetLister interface {
	ListActiveByProvider(provider string) ([]assets.IdentifiedAsset, error)
}

type automatedPriceWriter interface {
	CreateAutomated(price *prices.AssetPrice) (bool, error)
}

type ProviderResult struct {
	Provider string   `json:"provider"`
	Assets   int      `json:"assets"`
	Fetched  int      `json:"fetched"`
	Inserted int      `json:"inserted"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

type JobResult struct {
	From      string           `json:"from"`
	To        string           `json:"to"`
	Providers []ProviderResult `json:"providers"`
	Failed    int              `json:"failed"`
}

type Job struct {
	assetRepo identifiedAssetLister
	priceRepo automatedPriceWriter
	providers []Provider
}

func NewJob(assetRepo identifiedAssetLister, priceRepo automatedPriceWriter, providers ...Provider) *Job {
	return &Job{assetRepo: assetRepo, priceRepo: priceRepo, providers: providers}
}

func (j *Job) Run(ctx context.Context, from, to time.Time) (JobResult, error) {
	from, to = day(from), day(to)
	result := JobResult{From: from.Format("2006-01-02"), To: to.Format("2006-01-02")}
	if from.After(to) {
		return result, fmt.Errorf("from date cannot be after to date")
	}
	for _, provider := range j.providers {
		providerResult := ProviderResult{Provider: provider.Name()}
		configured, err := j.assetRepo.ListActiveByProvider(provider.Name())
		if err != nil {
			providerResult.Errors = append(providerResult.Errors, fmt.Sprintf("list configured assets: %v", err))
			result.Failed++
			result.Providers = append(result.Providers, providerResult)
			continue
		}
		providerResult.Assets = len(configured)
		if len(configured) == 0 {
			result.Providers = append(result.Providers, providerResult)
			continue
		}
		quotes, err := provider.Fetch(ctx, configured, from, to)
		if err != nil {
			providerResult.Errors = append(providerResult.Errors, err.Error())
			result.Failed += len(configured)
			result.Providers = append(result.Providers, providerResult)
			continue
		}
		providerResult.Fetched = len(quotes)
		for _, asset := range configured {
			quote, ok := quotes[asset.ProviderIdentifier]
			if !ok {
				providerResult.Skipped++
				continue
			}
			marketDate := quote.MarketDate
			inserted, err := j.priceRepo.CreateAutomated(&prices.AssetPrice{
				ID:         uuid.New(),
				AssetID:    asset.ID,
				Price:      quote.Price,
				Currency:   quote.Currency,
				PricedAt:   quote.PricedAt,
				MarketDate: &marketDate,
				Source:     provider.Name(),
				Note:       "automated end-of-day market data",
			})
			if err != nil {
				providerResult.Errors = append(providerResult.Errors, fmt.Sprintf("store %s: %v", asset.ProviderIdentifier, err))
				result.Failed++
				continue
			}
			if inserted {
				providerResult.Inserted++
			} else {
				providerResult.Skipped++
			}
		}
		result.Providers = append(result.Providers, providerResult)
	}
	return result, nil
}
