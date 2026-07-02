"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FormState, Goal } from "@/lib/types";

async function submitGoal(portfolioID: string, accessToken: string, body: object) {
  return apiRequest<Goal>(`/portfolios/${encodeURIComponent(portfolioID)}/goals`, {
    method: "POST",
    accessToken,
    body: JSON.stringify(body),
  });
}

export async function createGoalAction(_state: FormState, formData: FormData): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const targetAmount = String(formData.get("targetAmount") ?? "").trim();
  const currency = String(formData.get("currency") ?? "").trim().toUpperCase();
  const targetDate = String(formData.get("targetDate") ?? "").trim();
  const fields: Record<string, string> = {};
  if (!name) fields.name = "Enter a goal name.";
  if (!targetAmount || !Number.isFinite(Number(targetAmount)) || Number(targetAmount) <= 0) fields.targetAmount = "Enter a target greater than zero.";
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (!/^\d{4}-\d{2}-\d{2}$/.test(targetDate)) fields.targetDate = "Choose a target date.";
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };

  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const body = { name, target_amount: targetAmount, currency, target_date: targetDate };
  try {
    await submitGoal(portfolioID, accessToken, body);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) return { message: error instanceof Error ? error.message : "The goal could not be created." };
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    try {
      const auth = await apiRequest<AuthResponse>("/auth/refresh", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) });
      await setSession(auth);
      accessToken = auth.access_token;
      await submitGoal(portfolioID, accessToken, body);
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) redirect("/login");
      return { message: refreshError instanceof Error ? refreshError.message : "The goal could not be created." };
    }
  }
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/planning`);
  return { message: "Goal created.", success: true };
}
