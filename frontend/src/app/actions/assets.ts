"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { Asset, AuthResponse, FormState } from "@/lib/types";

const assetClasses = new Set(["cash", "equity", "fund", "bond", "crypto", "real_estate", "commodity", "alternative", "other"]);
const riskCategories = new Set(["", "equity", "debt", "cash_other"]);

export async function createAssetAction(_state: FormState, formData: FormData): Promise<FormState> {
  const symbol = String(formData.get("symbol") ?? "").trim().toUpperCase();
  const name = String(formData.get("name") ?? "").trim();
  const assetClass = String(formData.get("assetClass") ?? "").trim();
  const riskCategory = String(formData.get("riskCategory") ?? "").trim();
  const currency = String(formData.get("currency") ?? "").trim().toUpperCase();
  const exchange = String(formData.get("exchange") ?? "").trim().toUpperCase();
  const fields: Record<string, string> = {};
  if (!symbol) fields.symbol = "Enter a symbol.";
  if (!name) fields.name = "Enter an asset name.";
  if (!assetClasses.has(assetClass)) fields.assetClass = "Choose an asset class.";
  if (!riskCategories.has(riskCategory)) fields.riskCategory = "Choose a valid risk category.";
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };

  const body = { symbol, name, asset_class: assetClass, risk_category: riskCategory || null, currency, exchange };
  try {
    await authenticatedRequest<Asset>("/assets", { method: "POST", body: JSON.stringify(body) });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The asset could not be created." };
  }
  revalidatePath("/assets");
  return { message: "Asset created.", success: true };
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
