package marketdata

import (
	"testing"
	"time"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
)

func TestParseAMFIFiltersConfiguredSchemesAndDates(t *testing.T) {
	body := []byte("Scheme Code;ISIN Div Payout/ ISIN Growth;ISIN Div Reinvestment;Scheme Name;Net Asset Value;Date\n" +
		"100001;;;Wanted Fund;123.4567;29-Jun-2026\n" +
		"100002;;;Other Fund;99.00;29-Jun-2026\n" +
		"100001;;;Bad NAV;not-a-number;30-Jun-2026\n")
	from := mustDate(t, "2026-06-25")
	to := mustDate(t, "2026-06-29")

	quotes, err := parseAMFI(body, []assets.IdentifiedAsset{{ProviderIdentifier: "100001"}}, from, to)
	if err != nil {
		t.Fatalf("parseAMFI: %v", err)
	}
	quote, ok := quotes["100001"]
	if !ok {
		t.Fatal("configured scheme quote missing")
	}
	if quote.Price.String() != "123.4567" || quote.Currency != "INR" || quote.MarketDate.Format("2006-01-02") != "2026-06-29" {
		t.Fatalf("quote = %+v", quote)
	}
	if len(quotes) != 1 {
		t.Fatalf("quotes = %d, want 1", len(quotes))
	}
}

func mustDate(t *testing.T, raw string) time.Time {
	t.Helper()
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
