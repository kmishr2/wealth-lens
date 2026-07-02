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

export type FormState = {
  message: string;
  success?: boolean;
  fields?: Record<string, string>;
};
