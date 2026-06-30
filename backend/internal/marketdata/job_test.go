package marketdata

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/prices"
	"github.com/shopspring/decimal"
)

type fakeAssets struct {
	values map[string][]assets.IdentifiedAsset
}

func (f fakeAssets) ListActiveByProvider(provider string) ([]assets.IdentifiedAsset, error) {
	return f.values[provider], nil
}

type fakeWriter struct{ values []*prices.AssetPrice }

func (f *fakeWriter) CreateAutomated(price *prices.AssetPrice) (bool, error) {
	f.values = append(f.values, price)
	return true, nil
}

type fakeProvider struct{ name string }

func (f fakeProvider) Name() string { return f.name }
func (f fakeProvider) Fetch(_ context.Context, configured []assets.IdentifiedAsset, _, _ time.Time) (map[string]Quote, error) {
	date := mustDateNoTest("2026-06-29")
	return map[string]Quote{configured[0].ProviderIdentifier: {
		Identifier: configured[0].ProviderIdentifier,
		Price:      decimal.RequireFromString("10.5"), Currency: "INR", MarketDate: date, PricedAt: endOfIndiaDay(date),
	}}, nil
}

func TestJobStoresAutomatedPriceWithProvenance(t *testing.T) {
	assetID := uuid.New()
	writer := &fakeWriter{}
	job := NewJob(fakeAssets{values: map[string][]assets.IdentifiedAsset{
		ProviderAMFI: {{Asset: assets.Asset{ID: assetID}, ProviderIdentifier: "100001"}},
	}}, writer, fakeProvider{name: ProviderAMFI})
	result, err := job.Run(context.Background(), mustDate(t, "2026-06-25"), mustDate(t, "2026-06-29"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Failed != 0 || result.Providers[0].Inserted != 1 || len(writer.values) != 1 {
		t.Fatalf("result=%+v writes=%d", result, len(writer.values))
	}
	stored := writer.values[0]
	if stored.AssetID != assetID || stored.Source != ProviderAMFI || stored.CreatedByUserID != nil || stored.MarketDate == nil {
		t.Fatalf("stored = %+v", stored)
	}
}

func mustDateNoTest(raw string) time.Time {
	value, _ := time.Parse("2006-01-02", raw)
	return value
}
