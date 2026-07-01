package finance

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type ValuationPrice struct {
	AssetID  string
	Price    decimal.Decimal
	Currency string
	PricedAt time.Time
}

type PortfolioValuationResult struct {
	AssetValues    []AssetValuation `json:"asset_values"`
	CashValues     []CashBalance    `json:"cash_values"`
	TotalValues    []CurrencyValue  `json:"total_values"`
	MissingPrices  []MissingPrice   `json:"missing_prices"`
	Definition     MetricDefinition `json:"definition"`
	IsFullyValued  bool             `json:"is_fully_valued"`
	ValuationScope string           `json:"valuation_scope"`
}

type AssetValuation struct {
	AssetID      string          `json:"asset_id"`
	AssetSymbol  string          `json:"asset_symbol"`
	AssetName    string          `json:"asset_name"`
	AssetClass   string          `json:"asset_class"`
	RiskCategory string          `json:"risk_category"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Currency     string          `json:"currency"`
	MarketValue  decimal.Decimal `json:"market_value"`
	PricedAt     time.Time       `json:"priced_at"`
}

type CurrencyValue struct {
	Currency string          `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
}

type MissingPrice struct {
	AssetID     string `json:"asset_id"`
	AssetSymbol string `json:"asset_symbol"`
	AssetName   string `json:"asset_name"`
	Currency    string `json:"currency"`
	Reason      string `json:"reason"`
}

func PortfolioValuationDefinition() MetricDefinition {
	return MetricDefinition{
		Name:    "Portfolio Valuation",
		Formula: "Asset market value = quantity x explicit asset price. Total value per currency = sum(asset market values in that currency) + cash balance in that currency.",
		Assumptions: []string{
			"Each asset price is an explicit input and must be greater than zero.",
			"Asset prices must use the same currency as the holding being valued.",
			"No foreign exchange conversion is performed; totals are grouped by currency.",
			"Holdings without prices are reported as missing and excluded from total values.",
		},
		RequiredInputs: []string{
			"asset holdings",
			"cash balances",
			"asset_id",
			"quantity",
			"price",
			"price currency",
			"priced_at",
		},
		Explanation: "Portfolio valuation multiplies ledger-derived asset quantities by explicit price inputs, then adds ledger-derived cash balances by currency. Missing prices are surfaced instead of estimated.",
	}
}

func CalculatePortfolioValuation(holdings HoldingsResult, prices []ValuationPrice) (PortfolioValuationResult, error) {
	priceByAssetID := make(map[string]ValuationPrice, len(prices))
	for index, price := range prices {
		assetID := strings.TrimSpace(price.AssetID)
		if assetID == "" {
			return PortfolioValuationResult{}, fmt.Errorf("price %d requires asset_id", index)
		}
		if !price.Price.GreaterThan(decimal.Zero) {
			return PortfolioValuationResult{}, fmt.Errorf("price %d must be greater than zero", index)
		}
		currency, ok := normalizeCurrency(price.Currency)
		if !ok {
			return PortfolioValuationResult{}, fmt.Errorf("price %d requires a three-letter uppercase currency code", index)
		}

		price.AssetID = assetID
		price.Currency = currency
		priceByAssetID[assetID] = price
	}

	totals := make(map[string]decimal.Decimal)
	assetValues := make([]AssetValuation, 0, len(holdings.AssetHoldings))
	missingPrices := make([]MissingPrice, 0)

	for _, holding := range holdings.AssetHoldings {
		price, ok := priceByAssetID[holding.AssetID]
		if !ok {
			missingPrices = append(missingPrices, missingPriceFor(holding, "No explicit price is available for this asset"))
			continue
		}
		if price.Currency != holding.Currency {
			return PortfolioValuationResult{}, fmt.Errorf("price currency %s does not match holding currency %s for asset_id %s", price.Currency, holding.Currency, holding.AssetID)
		}

		marketValue := holding.Quantity.Mul(price.Price)
		assetValues = append(assetValues, AssetValuation{
			AssetID:      holding.AssetID,
			AssetSymbol:  holding.AssetSymbol,
			AssetName:    holding.AssetName,
			AssetClass:   holding.AssetClass,
			RiskCategory: holding.RiskCategory,
			Quantity:     holding.Quantity,
			Price:        price.Price,
			Currency:     price.Currency,
			MarketValue:  marketValue,
			PricedAt:     price.PricedAt,
		})
		totals[price.Currency] = totals[price.Currency].Add(marketValue)
	}

	sort.Slice(assetValues, func(i, j int) bool {
		if assetValues[i].AssetSymbol == assetValues[j].AssetSymbol {
			return assetValues[i].AssetID < assetValues[j].AssetID
		}
		return assetValues[i].AssetSymbol < assetValues[j].AssetSymbol
	})

	cashValues := append([]CashBalance(nil), holdings.CashBalances...)
	sort.Slice(cashValues, func(i, j int) bool {
		return cashValues[i].Currency < cashValues[j].Currency
	})
	for _, balance := range cashValues {
		totals[balance.Currency] = totals[balance.Currency].Add(balance.Amount)
	}

	totalValues := make([]CurrencyValue, 0, len(totals))
	for currency, amount := range totals {
		if !amount.IsZero() {
			totalValues = append(totalValues, CurrencyValue{
				Currency: currency,
				Amount:   amount,
			})
		}
	}
	sort.Slice(totalValues, func(i, j int) bool {
		return totalValues[i].Currency < totalValues[j].Currency
	})

	sort.Slice(missingPrices, func(i, j int) bool {
		if missingPrices[i].AssetSymbol == missingPrices[j].AssetSymbol {
			return missingPrices[i].AssetID < missingPrices[j].AssetID
		}
		return missingPrices[i].AssetSymbol < missingPrices[j].AssetSymbol
	})

	return PortfolioValuationResult{
		AssetValues:    assetValues,
		CashValues:     cashValues,
		TotalValues:    totalValues,
		MissingPrices:  missingPrices,
		Definition:     PortfolioValuationDefinition(),
		IsFullyValued:  len(missingPrices) == 0,
		ValuationScope: "Totals are grouped by currency; no currency conversion is applied.",
	}, nil
}

func missingPriceFor(holding AssetHolding, reason string) MissingPrice {
	return MissingPrice{
		AssetID:     holding.AssetID,
		AssetSymbol: holding.AssetSymbol,
		AssetName:   holding.AssetName,
		Currency:    holding.Currency,
		Reason:      reason,
	}
}
