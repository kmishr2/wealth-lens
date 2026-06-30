package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/shopspring/decimal"
)

const DefaultUpstoxURL = "https://api.upstox.com/v3/historical-candle"

type UpstoxProvider struct {
	client  *http.Client
	baseURL string
	token   string
}

type upstoxResponse struct {
	Status string `json:"status"`
	Data   struct {
		Candles [][]json.RawMessage `json:"candles"`
	} `json:"data"`
}

func NewUpstoxProvider(client *http.Client, baseURL, token string) *UpstoxProvider {
	if baseURL == "" {
		baseURL = DefaultUpstoxURL
	}
	return &UpstoxProvider{client: client, baseURL: baseURL, token: token}
}

func (*UpstoxProvider) Name() string { return ProviderUpstox }

func (p *UpstoxProvider) Fetch(ctx context.Context, wanted []assets.IdentifiedAsset, from, to time.Time) (map[string]Quote, error) {
	if p.token == "" {
		return nil, fmt.Errorf("UPSTOX_ACCESS_TOKEN is required when Upstox assets are configured")
	}
	quotes := make(map[string]Quote, len(wanted))
	for _, asset := range wanted {
		quote, ok, err := p.fetchOne(ctx, asset.ProviderIdentifier, from, to)
		if err != nil {
			return nil, fmt.Errorf("fetch Upstox instrument %q: %w", asset.ProviderIdentifier, err)
		}
		if ok {
			quotes[asset.ProviderIdentifier] = quote
		}
	}
	return quotes, nil
}

func (p *UpstoxProvider) fetchOne(ctx context.Context, identifier string, from, to time.Time) (Quote, bool, error) {
	endpoint := fmt.Sprintf("%s/%s/day/%s/%s", p.baseURL, url.PathEscape(identifier), to.Format("2006-01-02"), from.Format("2006-01-02"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Quote{}, false, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/json")
	body, err := get(p.client, req)
	if err != nil {
		return Quote{}, false, err
	}
	var response upstoxResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Quote{}, false, fmt.Errorf("decode response: %w", err)
	}
	if response.Status != "success" {
		return Quote{}, false, fmt.Errorf("response status is %q", response.Status)
	}
	var latest Quote
	found := false
	for _, candle := range response.Data.Candles {
		if len(candle) < 5 {
			continue
		}
		var timestamp string
		var closeString json.Number
		if json.Unmarshal(candle[0], &timestamp) != nil || json.Unmarshal(candle[4], &closeString) != nil {
			continue
		}
		pricedAt, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			continue
		}
		closePrice, err := decimal.NewFromString(closeString.String())
		if err != nil || !closePrice.GreaterThan(decimal.Zero) {
			continue
		}
		india, _ := time.LoadLocation("Asia/Kolkata")
		local := pricedAt.In(india)
		marketDate := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
		if marketDate.Before(day(from)) || marketDate.After(day(to)) || (found && !marketDate.After(latest.MarketDate)) {
			continue
		}
		latest = Quote{Identifier: identifier, Price: closePrice, Currency: "INR", MarketDate: marketDate, PricedAt: pricedAt.UTC()}
		found = true
	}
	return latest, found, nil
}
