import assert from "node:assert/strict";
import test from "node:test";
import { buildPortfolioSummaryCSV } from "../src/lib/portfolio-report.ts";
import type { HoldingsResponse, MetricDefinition, Portfolio, PortfolioAllocation, PortfolioValuation } from "../src/lib/types.ts";

const metric: MetricDefinition = {
  name: "=Ledger metric",
  formula: "quantity × price",
  assumptions: ["No currency conversion"],
  required_inputs: ["Ledger entries", "Explicit prices"],
  explanation: "Deterministic test metric",
};
const portfolio = { id: "portfolio-1", name: "+Summary", base_currency: "INR" } as Portfolio;
const holdings: HoldingsResponse = {
  portfolio_id: portfolio.id,
  asset_holdings: [{ asset_id: "asset-1", asset_symbol: "ABC", asset_name: "ABC Ltd", asset_class: "equity", quantity: "2", currency: "INR" }],
  cash_balances: [{ currency: "USD", amount: "50" }],
  metric_metadata: metric,
};
const valuation: PortfolioValuation = {
  portfolio_id: portfolio.id,
  total_values: [{ currency: "INR", amount: "200" }, { currency: "USD", amount: "50" }],
  missing_prices: [], is_fully_valued: true, valuation_scope: "Per currency", metric_metadata: metric,
};
const allocation: PortfolioAllocation = {
  portfolio_id: portfolio.id,
  asset_allocations: [{ asset_id: "asset-1", asset_symbol: "ABC", asset_name: "ABC Ltd", asset_class: "equity", risk_category: "equity", currency: "INR", market_value: "200", percentage: "100" }],
  asset_class_allocations: [{ asset_class: "equity", currency: "INR", market_value: "200", percentage: "100" }],
  cash_allocations: [{ currency: "USD", amount: "50", percentage: "100" }],
  currency_totals: [{ currency: "INR", amount: "200" }, { currency: "USD", amount: "50" }],
  missing_prices: [], is_complete: true, allocation_scope: "Per currency", metric_metadata: metric,
};

test("portfolio summary keeps currencies separate and includes explainability rows", () => {
  const csv = buildPortfolioSummaryCSV(portfolio, holdings, valuation, allocation);
  assert.match(csv, /"portfolio_total"[^\r\n]*"INR"[^\r\n]*"200"/);
  assert.match(csv, /"portfolio_total"[^\r\n]*"USD"[^\r\n]*"50"/);
  assert.match(csv, /"asset_holding"[^\r\n]*"ABC"/);
  assert.match(csv, /"metric_definition"/);
  assert.match(csv, /"'=Ledger metric"/);
  assert.match(csv, /"'\+Summary"/);
});
