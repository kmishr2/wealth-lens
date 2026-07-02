"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AssetPrice, AuthResponse, FormState } from "@/lib/types";

export async function createPriceAction(_state: FormState, formData: FormData): Promise<FormState> {
  const assetID = String(formData.get("assetId") ?? "");
  const price = String(formData.get("price") ?? "");
  const currency = String(formData.get("currency") ?? "").toUpperCase();
  const pricedAt = String(formData.get("pricedAt") ?? "");
  const source = String(formData.get("source") ?? "manual").trim();
  const note = String(formData.get("note") ?? "").trim();
  const fields: Record<string, string> = {};
  if (!price || !Number.isFinite(Number(price)) || Number(price) <= 0) fields.price = "Enter a price greater than zero.";
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (!pricedAt || Number.isNaN(Date.parse(pricedAt))) fields.pricedAt = "Enter a valid price timestamp.";
  if (pricedAt && Date.parse(pricedAt) > Date.now()) fields.pricedAt = "Price time cannot be in the future.";
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };
  const path = `/assets/${encodeURIComponent(assetID)}/prices`;
  const body = JSON.stringify({ price, currency, priced_at: pricedAt, source, note });
  try {
    await authenticatedRequest<AssetPrice>(path, { method: "POST", body });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The price could not be recorded." };
  }
  revalidatePath(`/assets/${encodeURIComponent(assetID)}`);
  return { message: "Price recorded.", success: true };
}

async function authenticatedRequest<T>(path: string, options: RequestInit): Promise<T> {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try { return await apiRequest<T>(path, { ...options, accessToken }); }
  catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) throw error;
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    const auth = await apiRequest<AuthResponse>("/auth/refresh", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) });
    await setSession(auth);
    return apiRequest<T>(path, { ...options, accessToken: auth.access_token });
  }
}
