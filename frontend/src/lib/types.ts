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

export type FormState = {
  message: string;
  success?: boolean;
  fields?: Record<string, string>;
};
