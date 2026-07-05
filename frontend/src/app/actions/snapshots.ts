"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FormState, PortfolioSnapshot } from "@/lib/types";

export async function createDailySnapshotAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const snapshotDate = String(formData.get("snapshotDate") ?? "").trim();
  const fields: Record<string, string> = {};
  const parsedDate = new Date(`${snapshotDate}T00:00:00Z`);

  if (
    !/^\d{4}-\d{2}-\d{2}$/.test(snapshotDate) ||
    Number.isNaN(parsedDate.getTime()) ||
    parsedDate.toISOString().slice(0, 10) !== snapshotDate
  ) {
    fields.snapshotDate = "Choose a valid snapshot date.";
  } else if (parsedDate.getTime() > startOfTodayUTC()) {
    fields.snapshotDate = "Snapshot date cannot be in the future.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the snapshot date.", fields };
  }

  const path = `/portfolios/${encodeURIComponent(portfolioID)}/snapshots`;
  const body = JSON.stringify({ snapshot_date: snapshotDate });
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");

  try {
    await apiRequest<PortfolioSnapshot>(path, {
      method: "POST",
      accessToken,
      body,
    });
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      return {
        message:
          error instanceof Error
            ? error.message
            : "The daily snapshot could not be created.",
      };
    }

    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    try {
      const auth = await apiRequest<AuthResponse>("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      await setSession(auth);
      accessToken = auth.access_token;
      await apiRequest<PortfolioSnapshot>(path, {
        method: "POST",
        accessToken,
        body,
      });
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) {
        redirect("/login");
      }
      return {
        message:
          refreshError instanceof Error
            ? refreshError.message
            : "The daily snapshot could not be created.",
      };
    }
  }

  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/analytics`);
  return {
    message: `Daily snapshot available for ${snapshotDate}.`,
    success: true,
  };
}

function startOfTodayUTC() {
  const now = new Date();
  return Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate());
}
