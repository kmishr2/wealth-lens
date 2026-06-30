package marketdata

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/shopspring/decimal"
)

const DefaultAMFIURL = "https://portal.amfiindia.com/spages/NAVAll.txt"

type AMFIProvider struct {
	client *http.Client
	url    string
}

func NewAMFIProvider(client *http.Client, url string) *AMFIProvider {
	if url == "" {
		url = DefaultAMFIURL
	}
	return &AMFIProvider{client: client, url: url}
}

func (*AMFIProvider) Name() string { return ProviderAMFI }

func (p *AMFIProvider) Fetch(ctx context.Context, wanted []assets.IdentifiedAsset, from, to time.Time) (map[string]Quote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wealth-lens/market-price-ingestion")
	body, err := get(p.client, req)
	if err != nil {
		return nil, fmt.Errorf("fetch AMFI NAV feed: %w", err)
	}
	return parseAMFI(body, wanted, from, to)
}

func parseAMFI(body []byte, wanted []assets.IdentifiedAsset, from, to time.Time) (map[string]Quote, error) {
	ids := make(map[string]struct{}, len(wanted))
	for _, asset := range wanted {
		ids[strings.TrimSpace(asset.ProviderIdentifier)] = struct{}{}
	}
	quotes := make(map[string]Quote)
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), 2<<20)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ";")
		if len(fields) != 6 {
			continue
		}
		id := strings.TrimSpace(fields[0])
		if _, ok := ids[id]; !ok {
			continue
		}
		value, err := decimal.NewFromString(strings.TrimSpace(fields[4]))
		if err != nil || !value.GreaterThan(decimal.Zero) {
			continue
		}
		date, err := time.Parse("02-Jan-2006", strings.TrimSpace(fields[5]))
		if err != nil || date.Before(day(from)) || date.After(day(to)) {
			continue
		}
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		quotes[id] = Quote{Identifier: id, Price: value, Currency: "INR", MarketDate: date, PricedAt: endOfIndiaDay(date)}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("parse AMFI NAV feed: %w", err)
	}
	return quotes, nil
}

func day(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func endOfIndiaDay(date time.Time) time.Time {
	location, _ := time.LoadLocation("Asia/Kolkata")
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, location).UTC()
}
