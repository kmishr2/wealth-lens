package marketdata

import (
	"context"
	"time"

	"github.com/kaustubhmishra/wealth-lens/backend/internal/assets"
	"github.com/shopspring/decimal"
)

const (
	ProviderAMFI   = "amfi"
	ProviderUpstox = "upstox"
)

type Quote struct {
	Identifier string
	Price      decimal.Decimal
	Currency   string
	MarketDate time.Time
	PricedAt   time.Time
}

type Provider interface {
	Name() string
	Fetch(context.Context, []assets.IdentifiedAsset, time.Time, time.Time) (map[string]Quote, error)
}
