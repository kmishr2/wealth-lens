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
