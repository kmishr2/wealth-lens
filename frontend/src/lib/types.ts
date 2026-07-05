export type ApiErrorPayload = {
  code: string;
  message: string;
};

export type ApiEnvelope<T> = {
  success: boolean;
  data?: T;
  error?: ApiErrorPayload;
};

export type User = {
  id: string;
  email: string;
  display_name: string;
  base_currency: string;
  timezone: string;
  created_at: string;
};

export type AuthResponse = {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
  user: User;
};

export type Portfolio = {
  id: string;
  name: string;
  description: string;
  base_currency: string;
  created_at: string;
  updated_at: string;
};

export type AccountType =
  | "brokerage"
  | "retirement"
  | "bank"
  | "wallet"
  | "other";

export type Account = {
  id: string;
  portfolio_id: string;
  name: string;
  account_type: AccountType;
  institution_name: string;
  currency: string;
  created_at: string;
  updated_at: string;
};

export type FixedDeposit = {
  id: string;
  portfolio_id: string;
  account_id: string;
  asset_id: string;
  opening_transaction_id: string;
  name: string;
  bank_reference: string;
  principal: string;
  currency: string;
  annual_interest_rate: string;
  start_date: string;
  maturity_date: string;
  current_value: string;
  current_value_at: string;
  valuation_metadata: MetricDefinition;
  created_at: string;
};

export type Asset = {
  id: string;
  symbol: string;
  name: string;
  asset_class: string;
  risk_category: "equity" | "debt" | "cash_other" | null;
  currency: string;
  exchange: string;
  is_active: boolean;
};

export type AssetPrice = {
  id: string;
  asset_id: string;
  price: string;
  currency: string;
  priced_at: string;
  source: string;
  note: string;
  created_at: string;
};

export type TransactionType =
  | "deposit"
  | "withdrawal"
  | "buy"
  | "sell"
  | "fee"
  | "tax"
  | "transfer"
  | "reversal";

export type TransactionEntry = {
  id: string;
  entry_kind: "cash" | "asset" | "fee" | "tax";
  asset_id?: string;
  quantity?: string;
  amount?: string;
  currency: string;
  created_at: string;
};

export type Transaction = {
  id: string;
  portfolio_id: string;
  account_id: string;
  transaction_type: TransactionType;
  occurred_at: string;
  description: string;
  entries: TransactionEntry[];
  reverses_transaction_id?: string;
  corrects_transaction_id?: string;
  created_at: string;
};

export type MetricDefinition = {
  name: string;
  formula: string;
  assumptions: string[];
  required_inputs: string[];
  explanation: string;
};

export type AssetHolding = {
  asset_id: string;
  asset_symbol: string;
  asset_name: string;
  asset_class: string;
  quantity: string;
  currency: string;
};

export type CashBalance = { currency: string; amount: string };

export type HoldingsResponse = {
  portfolio_id: string;
  asset_holdings: AssetHolding[];
  cash_balances: CashBalance[];
  metric_metadata: MetricDefinition;
};

export type CurrencyValue = { currency: string; amount: string };

export type MissingPrice = {
  asset_id: string;
  asset_symbol: string;
  asset_name: string;
  currency: string;
  reason: string;
};

export type PortfolioValuation = {
  portfolio_id: string;
  total_values: CurrencyValue[];
  missing_prices: MissingPrice[];
  is_fully_valued: boolean;
  valuation_scope: string;
  metric_metadata: MetricDefinition;
};

export type AssetAllocation = {
  asset_id: string;
  asset_symbol: string;
  asset_name: string;
  asset_class: string;
  risk_category: string;
  currency: string;
  market_value: string;
  percentage: string;
};

export type AssetClassAllocation = {
  asset_class: string;
  currency: string;
  market_value: string;
  percentage: string;
};

export type PortfolioAllocation = {
  portfolio_id: string;
  asset_allocations: AssetAllocation[];
  asset_class_allocations: AssetClassAllocation[];
  cash_allocations: Array<{
    currency: string;
    amount: string;
    percentage: string;
  }>;
  currency_totals: CurrencyValue[];
  missing_prices: MissingPrice[];
  is_complete: boolean;
  allocation_scope: string;
  metric_metadata: MetricDefinition;
};

export type PortfolioSnapshot = {
  id: string;
  portfolio_id: string;
  snapshot_date: string;
  snapshot_period: "daily";
  total_values: CurrencyValue[];
  is_fully_valued: boolean;
};

export type WeeklyPerformanceSnapshot = {
  id: string;
  portfolio_id: string;
  week_start_date: string;
  week_end_date: string;
  currency_returns: CurrencyPerformance[];
  performance_scope: string;
  pnl_metadata: MetricDefinition;
  cagr_metadata: MetricDefinition;
  xirr_metadata: MetricDefinition;
  created_at: string;
};

export type CurrencyPerformance = {
  currency: string;
  beginning_value: string;
  ending_value: string;
  net_external_cash_flow: string;
  profit_loss: string;
  cagr: string;
  xirr: string;
  cash_flow_count: number;
};

export type PortfolioPerformance = {
  portfolio_id: string;
  start_date: string;
  end_date: string;
  currency_returns: CurrencyPerformance[];
  performance_scope: string;
  pnl_metadata: MetricDefinition;
  cagr_metadata: MetricDefinition;
  xirr_metadata: MetricDefinition;
};

export type CurrencyRisk = {
  currency: string;
  observation_count: number;
  periodic_return_count: number;
  annualized_volatility: string;
  maximum_drawdown: string;
  peak_date: string;
  trough_date: string;
};

export type PortfolioRisk = {
  portfolio_id: string;
  start_date: string;
  end_date: string;
  periods_per_year: string;
  currency_risk: CurrencyRisk[];
  risk_scope: string;
  volatility_metadata: MetricDefinition;
  drawdown_metadata: MetricDefinition;
};

export type HealthComponent = {
  category: string;
  points: number;
  maximum: number;
  observed: string;
  threshold: string;
  explanation: string;
};

export type CurrencyHealthScore = {
  currency: string;
  score: number;
  maximum: number;
  components: HealthComponent[];
  definition: MetricDefinition;
};

export type PortfolioHealth = {
  portfolio_id: string;
  start_date: string;
  end_date: string;
  risk_profile: string;
  periods_per_year: string;
  scores: CurrencyHealthScore[];
  scope: string;
};

export type Goal = {
  id: string;
  portfolio_id: string;
  name: string;
  target_amount: string;
  currency: string;
  target_date: string;
  status: "active" | "completed" | "archived";
  created_at: string;
  updated_at: string;
};

export type MonthlyGoalSnapshot = {
  id: string;
  portfolio_id: string;
  goal_id: string;
  snapshot_month_end: string;
  current_value: string;
  target_value: string;
  currency: string;
  progress_percentage: string;
  remaining_amount: string;
  months_remaining: number;
  required_monthly_contribution: string;
  is_target_reached: boolean;
  goal_progress_metadata: MetricDefinition;
};

export type SIPProjectionPoint = {
  month: number;
  total_contributions: string;
  projected_nominal_value: string;
  projected_real_value: string;
};

export type SIPProjection = {
  portfolio_id: string;
  currency: string;
  initial_investment: string;
  monthly_contribution: string;
  annual_return_percentage: string;
  annual_inflation_percentage: string;
  months: number;
  total_contributions: string;
  projected_nominal_value: string;
  projected_real_value: string;
  nominal_investment_growth: string;
  schedule: SIPProjectionPoint[];
  definition: MetricDefinition;
  scope: string;
};

export type WhatIfScenarioResult = {
  name: string;
  projection: Omit<SIPProjection, "portfolio_id" | "currency" | "scope">;
  nominal_difference_from_baseline: string;
  real_difference_from_baseline: string;
  contribution_difference_from_baseline: string;
};

export type WhatIfComparison = {
  portfolio_id: string;
  currency: string;
  baseline_name: string;
  months: number;
  scenarios: WhatIfScenarioResult[];
  definition: MetricDefinition;
  scope: string;
};

export type Benchmark = {
  id: string;
  code: string;
  name: string;
  currency: string;
  source: string;
  description: string;
  created_at: string;
};

export type BenchmarkObservation = {
  id: string;
  benchmark_id: string;
  observation_date: string;
  value: string;
  source: string;
  note: string;
  created_at: string;
};

export type BenchmarkComparison = {
  portfolio_id: string;
  benchmark_id: string;
  benchmark_code: string;
  benchmark_name: string;
  currency: string;
  start_date: string;
  end_date: string;
  portfolio_start_value: string;
  portfolio_end_value: string;
  benchmark_start_value: string;
  benchmark_end_value: string;
  portfolio_total_return: string;
  benchmark_total_return: string;
  portfolio_cagr: string;
  benchmark_cagr: string;
  excess_total_return: string;
  excess_cagr: string;
  comparison_scope: string;
  comparison_metadata: MetricDefinition;
};

export type BenchmarkBeta = {
  portfolio_id: string;
  benchmark_id: string;
  benchmark_code: string;
  currency: string;
  start_date: string;
  end_date: string;
  aligned_observation_count: number;
  paired_return_count: number;
  beta: string;
  scope: string;
  metric_metadata: MetricDefinition;
};

export type CurrencyConcentration = {
  currency: string;
  asset_count: number;
  invested_value: string;
  herfindahl_hirschman_index: string;
  effective_asset_count: string;
  largest_asset_id: string;
  largest_asset_symbol: string;
  largest_asset_percentage: string;
};

export type PortfolioConcentration = {
  portfolio_id: string;
  currencies: CurrencyConcentration[];
  concentration_scope: string;
  metric_metadata: MetricDefinition;
};

export type DiversificationAlert = {
  currency: string;
  severity: "none" | "notice" | "warning" | "critical";
  points: number;
  largest_asset_id: string;
  largest_asset_symbol: string;
  largest_asset_percentage: string;
  holding_count: number;
  conditions: string[];
};

export type PortfolioDiversificationAlerts = {
  portfolio_id: string;
  alerts: DiversificationAlert[];
  alert_scope: string;
  alert_metadata: MetricDefinition;
};

export type MonthlyContributionBucket = {
  month: string;
  contributions: string;
  withdrawals: string;
  net_contributions: string;
  event_count: number;
};

export type ContributionAnalysis = {
  portfolio_id: string;
  currency: string;
  start_date: string;
  end_date: string;
  beginning_value: string;
  ending_value: string;
  contributions: string;
  withdrawals: string;
  net_contributions: string;
  investment_growth: string;
  event_count: number;
  monthly_buckets: MonthlyContributionBucket[];
  definition: MetricDefinition;
  scope: string;
};

export type RebalancingItem = {
  asset_class: string;
  currency: string;
  current_value: string;
  current_percentage: string;
  target_value: string;
  target_percentage: string;
  drift_percentage: string;
  absolute_drift: string;
  is_outside_tolerance: boolean;
  suggested_adjustment: string;
  action: "none" | "increase" | "decrease";
};

export type RebalancingResponse = {
  portfolio_id: string;
  items: RebalancingItem[];
  drift_tolerance_percentage: string;
  rebalancing_scope: string;
  metric_metadata: MetricDefinition;
  drift_metadata: MetricDefinition;
};

export type FormState = {
  message: string;
  success?: boolean;
  fields?: Record<string, string>;
};
