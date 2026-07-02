"use server";

import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, SIPProjection, WhatIfComparison } from "@/lib/types";

export type SIPState = { message: string; data?: SIPProjection };
export type WhatIfState = { message: string; data?: WhatIfComparison };

export async function calculateSIPAction(_state: SIPState, formData: FormData): Promise<SIPState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const currency = String(formData.get("currency") ?? "").toUpperCase();
  const payload = projectionInput(formData, "");
  const error = validateProjection(payload, currency);
  if (error) return { message: error };
  try {
    const data = await authenticatedRequest<SIPProjection>(`/portfolios/${encodeURIComponent(portfolioID)}/projections/sip`, {
      method: "POST", body: JSON.stringify({ currency, ...payload }),
    });
    return { message: "Projection calculated.", data };
  } catch (requestError) {
    return { message: requestError instanceof Error ? requestError.message : "Projection unavailable." };
  }
}

export async function compareWhatIfAction(_state: WhatIfState, formData: FormData): Promise<WhatIfState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const currency = String(formData.get("currency") ?? "").toUpperCase();
  const baseline = projectionInput(formData, "baseline");
  const alternative = projectionInput(formData, "alternative");
  const error = validateProjection(baseline, currency) || validateProjection(alternative, currency);
  if (error) return { message: error };
  try {
    const data = await authenticatedRequest<WhatIfComparison>(`/portfolios/${encodeURIComponent(portfolioID)}/projections/what-if`, {
      method: "POST",
      body: JSON.stringify({ currency, scenarios: [
        { name: "Baseline", input: baseline },
        { name: "Alternative", input: alternative },
      ] }),
    });
    return { message: "Scenarios compared.", data };
  } catch (requestError) {
    return { message: requestError instanceof Error ? requestError.message : "Comparison unavailable." };
  }
}

function projectionInput(formData: FormData, prefix: string) {
  const key = (name: string) => prefix ? `${prefix}${name[0].toUpperCase()}${name.slice(1)}` : name;
  return {
    initial_investment: String(formData.get(key("initialInvestment")) ?? "0"),
    monthly_contribution: String(formData.get(key("monthlyContribution")) ?? "0"),
    annual_return_percentage: String(formData.get(key("annualReturn")) ?? ""),
    annual_inflation_percentage: String(formData.get(key("inflation")) ?? "0"),
    months: Number(formData.get(key("months")) ?? 0),
  };
}

function validateProjection(input: ReturnType<typeof projectionInput>, currency: string) {
  if (!/^[A-Z]{3}$/.test(currency)) return "Currency must be a three-letter code.";
  if (!Number.isFinite(Number(input.annual_return_percentage)) || input.annual_return_percentage === "") return "Enter an annual return assumption.";
  if (!Number.isInteger(input.months) || input.months < 1 || input.months > 1200) return "Months must be between 1 and 1200.";
  if (![input.initial_investment, input.monthly_contribution, input.annual_inflation_percentage].every((value) => Number.isFinite(Number(value)))) return "Enter valid numeric assumptions.";
  if (Number(input.initial_investment) < 0 || Number(input.monthly_contribution) < 0 || Number(input.annual_inflation_percentage) < 0) return "Amounts and inflation cannot be negative.";
  if (Number(input.initial_investment) === 0 && Number(input.monthly_contribution) === 0) return "Enter an initial investment or monthly contribution.";
  return "";
}

async function authenticatedRequest<T>(path: string, options: RequestInit): Promise<T> {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try {
    return await apiRequest<T>(path, { ...options, accessToken });
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) throw error;
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    const auth = await apiRequest<AuthResponse>("/auth/refresh", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) });
    await setSession(auth);
    return apiRequest<T>(path, { ...options, accessToken: auth.access_token });
  }
}
