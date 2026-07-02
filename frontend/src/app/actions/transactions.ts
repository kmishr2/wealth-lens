"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FormState, Transaction } from "@/lib/types";

type EntryPayload = {
  entry_kind: "cash" | "asset" | "fee" | "tax";
  asset_id?: string;
  quantity?: string;
  amount?: string;
  currency: string;
};

type TransactionPayload = {
  account_id: string;
  transaction_type: string;
  occurred_at: string;
  description: string;
  idempotency_key: string;
  entries: EntryPayload[];
};

async function submitTransaction(
  portfolioID: string,
  accessToken: string,
  body: TransactionPayload,
) {
  return apiRequest<Transaction>(
    `/portfolios/${encodeURIComponent(portfolioID)}/transactions`,
    { method: "POST", accessToken, body: JSON.stringify(body) },
  );
}

export async function createTransactionAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const accountID = String(formData.get("accountId") ?? "").trim();
  const type = String(formData.get("transactionType") ?? "").trim();
  const occurredAt = String(formData.get("occurredAt") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const currency = String(formData.get("currency") ?? "").trim().toUpperCase();
  const assetID = String(formData.get("assetId") ?? "").trim();
  const quantity = String(formData.get("quantity") ?? "").trim();
  const amount = String(formData.get("amount") ?? "").trim();
  const fields: Record<string, string> = {};
  const supported = new Set(["deposit", "withdrawal", "buy", "sell", "fee", "tax"]);

  if (!portfolioID || !accountID) fields.account = "Account is required.";
  if (!supported.has(type)) fields.transactionType = "Choose a valid transaction type.";
  if (!occurredAt || Number.isNaN(Date.parse(occurredAt))) fields.occurredAt = "Enter a valid date and time.";
  if (occurredAt && Date.parse(occurredAt) > Date.now()) fields.occurredAt = "The transaction cannot be in the future.";
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (!amount || !Number.isFinite(Number(amount)) || Number(amount) <= 0) fields.amount = "Enter an amount greater than zero.";
  if ((type === "buy" || type === "sell") && !assetID) fields.assetId = "Choose an asset.";
  if ((type === "buy" || type === "sell") && (!quantity || !Number.isFinite(Number(quantity)) || Number(quantity) <= 0)) {
    fields.quantity = "Enter a quantity greater than zero.";
  }
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };

  const entries: EntryPayload[] = [];
  if (type === "buy" || type === "sell") {
    const sign = type === "buy" ? "" : "-";
    entries.push({ entry_kind: "asset", asset_id: assetID, quantity: `${sign}${quantity}`, currency });
    entries.push({ entry_kind: "cash", amount: `${type === "buy" ? "-" : ""}${amount}`, currency });
  } else {
    const entryKind = type === "fee" || type === "tax" ? type : "cash";
    const positive = type === "deposit";
    entries.push({ entry_kind: entryKind, amount: `${positive ? "" : "-"}${amount}`, currency });
  }

  const body: TransactionPayload = {
    account_id: accountID,
    transaction_type: type,
    occurred_at: new Date(occurredAt).toISOString(),
    description,
    idempotency_key: "",
    entries,
  };
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try {
    await submitTransaction(portfolioID, accessToken, body);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      return { message: error instanceof Error ? error.message : "The transaction could not be created." };
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
      await submitTransaction(portfolioID, accessToken, body);
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) redirect("/login");
      return { message: refreshError instanceof Error ? refreshError.message : "The transaction could not be created." };
    }
  }

  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`);
  return { message: "Transaction recorded.", success: true };
}
