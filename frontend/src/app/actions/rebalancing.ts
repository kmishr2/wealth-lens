"use server";

import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, RebalancingResponse } from "@/lib/types";

export type RebalancingState = { message: string; data?: RebalancingResponse };

export async function calculateRebalancingAction(
  _state: RebalancingState,
  formData: FormData,
): Promise<RebalancingState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const tolerance = String(formData.get("tolerance") ?? "");
  const targets: Array<{ asset_class: string; currency: string; target_percentage: string }> = [];
  for (const [key, value] of formData.entries()) {
    if (!key.startsWith("target::")) continue;
    const [, currency, assetClass] = key.split("::");
    targets.push({ asset_class: assetClass, currency, target_percentage: String(value) });
  }
  if (!Number.isFinite(Number(tolerance)) || Number(tolerance) < 0) return { message: "Tolerance must be zero or greater." };
  if (targets.length === 0 || targets.some((target) => !Number.isFinite(Number(target.target_percentage)) || Number(target.target_percentage) < 0 || Number(target.target_percentage) > 100)) {
    return { message: "Every target must be between zero and 100%." };
  }

  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const path = `/portfolios/${encodeURIComponent(portfolioID)}/rebalancing`;
  const body = JSON.stringify({ targets, drift_tolerance_percentage: tolerance });
  try {
    const data = await apiRequest<RebalancingResponse>(path, { method: "POST", accessToken, body });
    return { message: "Rebalancing calculated.", data };
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) return { message: error instanceof Error ? error.message : "Rebalancing unavailable." };
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    try {
      const auth = await apiRequest<AuthResponse>("/auth/refresh", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) });
      await setSession(auth);
      accessToken = auth.access_token;
      const data = await apiRequest<RebalancingResponse>(path, { method: "POST", accessToken, body });
      return { message: "Rebalancing calculated.", data };
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) redirect("/login");
      return { message: refreshError instanceof Error ? refreshError.message : "Rebalancing unavailable." };
    }
  }
}
