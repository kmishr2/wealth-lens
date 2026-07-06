import { csvCell, safeSpreadsheetText } from "./csv.ts";
import type { HoldingsResponse, Portfolio, PortfolioAllocation, PortfolioValuation } from "./types.ts";

const headers = [
  "record_type", "portfolio_id", "portfolio_name", "asset_id", "symbol", "name",
  "asset_class", "currency", "quantity", "amount", "market_value", "percentage",
  "price_status", "metric_name", "formula", "assumptions", "required_inputs", "explanation",
];

export function buildPortfolioSummaryCSV(
  portfolio: Portfolio,
  holdings: HoldingsResponse,
  valuation: PortfolioValuation,
  allocation: PortfolioAllocation,
) {
  const rows: string[][] = [];
  const assetAllocations = new Map(allocation.asset_allocations.map((item) => [item.asset_id, item]));
  const missingPrices = new Map(valuation.missing_prices.map((item) => [item.asset_id, item]));

  for (const total of valuation.total_values) {
    rows.push(row("portfolio_total", portfolio, { currency: total.currency, amount: total.amount, priceStatus: valuation.is_fully_valued ? "complete" : "partial", explanation: valuation.valuation_scope }));
  }
  for (const holding of holdings.asset_holdings) {
    const allocated = assetAllocations.get(holding.asset_id);
    const missing = missingPrices.get(holding.asset_id);
    rows.push(row("asset_holding", portfolio, {
      assetID: holding.asset_id, symbol: holding.asset_symbol, name: holding.asset_name,
      assetClass: holding.asset_class, currency: holding.currency, quantity: holding.quantity,
      marketValue: allocated?.market_value, percentage: allocated?.percentage,
      priceStatus: missing ? "missing" : "valued", explanation: missing?.reason,
    }));
  }
  for (const cash of holdings.cash_balances) {
    const allocated = allocation.cash_allocations.find((item) => item.currency === cash.currency);
    rows.push(row("cash_balance", portfolio, { name: "Cash", assetClass: "cash", currency: cash.currency, amount: cash.amount, marketValue: cash.amount, percentage: allocated?.percentage, priceStatus: "not_required" }));
  }
  for (const item of allocation.asset_class_allocations) {
    rows.push(row("asset_class_allocation", portfolio, { name: item.asset_class, assetClass: item.asset_class, currency: item.currency, marketValue: item.market_value, percentage: item.percentage, priceStatus: allocation.is_complete ? "complete" : "partial", explanation: allocation.allocation_scope }));
  }
  for (const missing of valuation.missing_prices) {
    rows.push(row("missing_price", portfolio, { assetID: missing.asset_id, symbol: missing.asset_symbol, name: missing.asset_name, currency: missing.currency, priceStatus: "missing", explanation: missing.reason }));
  }
  for (const metric of [holdings.metric_metadata, valuation.metric_metadata, allocation.metric_metadata]) {
    rows.push(row("metric_definition", portfolio, {
      metricName: metric.name, formula: metric.formula, assumptions: metric.assumptions.join(" | "),
      requiredInputs: metric.required_inputs.join(" | "), explanation: metric.explanation,
    }));
  }

  return [headers, ...rows].map((values) => values.map(csvCell).join(",")).join("\r\n");
}

type RowValues = {
  assetID?: string; symbol?: string; name?: string; assetClass?: string; currency?: string;
  quantity?: string; amount?: string; marketValue?: string; percentage?: string; priceStatus?: string;
  metricName?: string; formula?: string; assumptions?: string; requiredInputs?: string; explanation?: string;
};

function row(recordType: string, portfolio: Portfolio, values: RowValues) {
  return [
    recordType, portfolio.id, safeSpreadsheetText(portfolio.name), values.assetID ?? "",
    safeSpreadsheetText(values.symbol ?? ""), safeSpreadsheetText(values.name ?? ""), values.assetClass ?? "",
    values.currency ?? "", values.quantity ?? "", values.amount ?? "", values.marketValue ?? "",
    values.percentage ?? "", values.priceStatus ?? "", safeSpreadsheetText(values.metricName ?? ""),
    safeSpreadsheetText(values.formula ?? ""), safeSpreadsheetText(values.assumptions ?? ""),
    safeSpreadsheetText(values.requiredInputs ?? ""), safeSpreadsheetText(values.explanation ?? ""),
  ];
}
