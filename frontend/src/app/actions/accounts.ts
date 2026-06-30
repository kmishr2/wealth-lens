"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import {
  getAccessToken,
  getRefreshToken,
  setSession,
} from "@/lib/session";
import type {
  Account,
  AccountType,
  AuthResponse,
  FormState,
} from "@/lib/types";

const accountTypes = new Set<AccountType>([
  "brokerage",
  "retirement",
  "bank",
  "wallet",
  "other",
]);

type AccountPayload = {
  name: string;
  account_type: AccountType;
  institution_name: string;
  currency: string;
};

async function submitAccount(
  portfolioID: string,
  accessToken: string,
  body: AccountPayload,
) {
  return apiRequest<Account>(
    `/portfolios/${encodeURIComponent(portfolioID)}/accounts`,
    {
      method: "POST",
      accessToken,
      body: JSON.stringify(body),
    },
  );
}

export async function createAccountAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const rawAccountType = String(formData.get("accountType") ?? "")
    .trim()
    .toLowerCase();
  const institutionName = String(
    formData.get("institutionName") ?? "",
  ).trim();
  const currency = String(formData.get("currency") ?? "")
    .trim()
    .toUpperCase();
  const fields: Record<string, string> = {};

  if (!portfolioID) fields.portfolioId = "Portfolio is required.";
  if (!name) fields.name = "Enter an account name.";
  if (!accountTypes.has(rawAccountType as AccountType)) {
    fields.accountType = "Choose a valid account type.";
  }
  if (!/^[A-Z]{3}$/.test(currency)) {
    fields.currency = "Use a three-letter currency code.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  const body: AccountPayload = {
    name,
    account_type: rawAccountType as AccountType,
    institution_name: institutionName,
    currency,
  };
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");

  try {
    await submitAccount(portfolioID, accessToken, body);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      return {
        message:
          error instanceof Error
            ? error.message
            : "The account could not be created.",
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
      await submitAccount(portfolioID, accessToken, body);
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) {
        redirect("/login");
      }
      return {
        message:
          refreshError instanceof Error
            ? refreshError.message
            : "The account could not be created.",
      };
    }
  }

  const portfolioPath = `/portfolios/${encodeURIComponent(portfolioID)}`;
  revalidatePath(portfolioPath);
  return { message: "Account created.", success: true };
}
