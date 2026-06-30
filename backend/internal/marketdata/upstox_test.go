package marketdata

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
)

func TestUpstoxFetchUsesTokenAndLatestCandle(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if !strings.Contains(r.URL.Path, "/NSE_EQ%7CINE002A01018/day/2026-06-29/2026-06-25") &&
			!strings.Contains(r.URL.Path, "/NSE_EQ|INE002A01018/day/2026-06-29/2026-06-25") {
			t.Fatalf("path = %q", r.URL.EscapedPath())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"status":"success","data":{"candles":[["2026-06-27T00:00:00+05:30",1,2,0.5,1400.25,100,0],["2026-06-29T00:00:00+05:30",1,2,0.5,1450.75,100,0]]}}`)),
		}, nil
	})}

	provider := NewUpstoxProvider(client, "https://example.test/v3/historical-candle", "test-token")
	quotes, err := provider.Fetch(context.Background(), []assets.IdentifiedAsset{{ProviderIdentifier: "NSE_EQ|INE002A01018"}}, mustDate(t, "2026-06-25"), mustDate(t, "2026-06-29"))
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	quote := quotes["NSE_EQ|INE002A01018"]
	if quote.Price.String() != "1450.75" || quote.MarketDate.Format("2006-01-02") != "2026-06-29" {
		t.Fatalf("quote = %+v", quote)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestUpstoxRequiresTokenOnlyWhenAssetsConfigured(t *testing.T) {
	provider := NewUpstoxProvider(http.DefaultClient, "", "")
	_, err := provider.Fetch(context.Background(), []assets.IdentifiedAsset{{ProviderIdentifier: "NSE_EQ|test"}}, mustDate(t, "2026-06-29"), mustDate(t, "2026-06-29"))
	if err == nil || !strings.Contains(err.Error(), "UPSTOX_ACCESS_TOKEN") {
		t.Fatalf("error = %v", err)
	}
}
