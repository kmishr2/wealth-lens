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

export async function updateAccountAction(_state: FormState, formData: FormData): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const accountID = String(formData.get("accountId") ?? "");
  const name = String(formData.get("name") ?? "").trim();
  const institutionName = String(formData.get("institutionName") ?? "").trim();
  if (!name) return { message: "Enter an account name.", fields: { name: "Name is required." } };
  try {
    await authenticatedAccountRequest<Account>(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`, {
      method: "PATCH", body: JSON.stringify({ name, institution_name: institutionName }),
    });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The account could not be updated." };
  }
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}`);
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`);
  return { message: "Account updated.", success: true };
}

export async function deleteAccountAction(_state: FormState, formData: FormData): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "");
  const accountID = String(formData.get("accountId") ?? "");
  const expectedName = String(formData.get("expectedName") ?? "");
  const confirmation = String(formData.get("confirmation") ?? "");
  if (confirmation !== expectedName) return { message: "Type the account name exactly to confirm deletion.", fields: { confirmation: "Name does not match." } };
  try {
    await authenticatedAccountRequest<unknown>(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`, { method: "DELETE" });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The account could not be deleted." };
  }
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}`);
  redirect(`/portfolios/${encodeURIComponent(portfolioID)}`);
}

async function authenticatedAccountRequest<T>(path: string, options: RequestInit): Promise<T> {
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
