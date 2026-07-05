"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import {
  getAccessToken,
  getRefreshToken,
  setSession,
} from "@/lib/session";
import type { AuthResponse, FormState, Portfolio } from "@/lib/types";

async function createPortfolio(
  accessToken: string,
  body: { name: string; description: string; base_currency: string },
) {
  return apiRequest<Portfolio>("/portfolios", {
    method: "POST",
    accessToken,
    body: JSON.stringify(body),
  });
}

export async function createPortfolioAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const name = String(formData.get("name") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const baseCurrency = String(formData.get("baseCurrency") ?? "")
    .trim()
    .toUpperCase();
  const fields: Record<string, string> = {};
  if (!name) fields.name = "Enter a portfolio name.";
  if (!/^[A-Z]{3}$/.test(baseCurrency)) {
    fields.baseCurrency = "Use a three-letter currency code.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");

  const body = {
    name,
    description,
    base_currency: baseCurrency,
  };

  try {
    await createPortfolio(accessToken, body);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      return {
        message:
          error instanceof Error
            ? error.message
            : "The portfolio could not be created.",
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
      await createPortfolio(accessToken, body);
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) {
        redirect("/login");
      }
      return {
        message:
          refreshError instanceof Error
            ? refreshError.message
            : "The portfolio could not be created.",
      };
    }
  }

  revalidatePath("/dashboard");
  return { message: "Portfolio created.", success: true };
}

export async function updatePortfolioAction(_state: FormState, formData: FormData): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const name = String(formData.get("name") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  if (!name) return { message: "Enter a portfolio name.", fields: { name: "Name is required." } };
  try {
    await authenticatedPortfolioRequest<Portfolio>(`/portfolios/${encodeURIComponent(portfolioID)}`, {
      method: "PATCH", body: JSON.stringify({ name, description }),
    });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The portfolio could not be updated." };
  }
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}`);
  revalidatePath("/dashboard");
  return { message: "Portfolio updated.", success: true };
}

export async function deletePortfolioAction(_state: FormState, formData: FormData): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const expectedName = String(formData.get("expectedName") ?? "");
  const confirmation = String(formData.get("confirmation") ?? "");
  if (confirmation !== expectedName) return { message: "Type the portfolio name exactly to confirm deletion.", fields: { confirmation: "Name does not match." } };
  try {
    await authenticatedPortfolioRequest<unknown>(`/portfolios/${encodeURIComponent(portfolioID)}`, { method: "DELETE" });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The portfolio could not be deleted." };
  }
  revalidatePath("/dashboard");
  redirect("/dashboard");
}

async function authenticatedPortfolioRequest<T>(path: string, options: RequestInit): Promise<T> {
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
