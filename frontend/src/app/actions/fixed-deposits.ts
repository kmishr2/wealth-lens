"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FixedDeposit, FormState } from "@/lib/types";

export async function createFixedDepositAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const accountID = String(formData.get("accountId") ?? "").trim();
  const name = String(formData.get("name") ?? "").trim();
  const bankReference = String(formData.get("bankReference") ?? "").trim();
  const principal = String(formData.get("principal") ?? "").trim();
  const currency = String(formData.get("currency") ?? "").trim().toUpperCase();
  const annualInterestRate = String(formData.get("annualInterestRate") ?? "").trim();
  const startDate = String(formData.get("startDate") ?? "").trim();
  const maturityDate = String(formData.get("maturityDate") ?? "").trim();
  const currentValue = String(formData.get("currentValue") ?? "").trim();
  const currentValueDate = String(formData.get("currentValueDate") ?? "").trim();
  const fields: Record<string, string> = {};

  if (!name) fields.name = "Enter a fixed deposit name.";
  validatePositive(principal, "principal", "Enter a principal greater than zero.", fields);
  validatePositive(currentValue, "currentValue", "Enter a current value greater than zero.", fields);
  if (!annualInterestRate || !Number.isFinite(Number(annualInterestRate)) || Number(annualInterestRate) <= 0 || Number(annualInterestRate) > 100) {
    fields.annualInterestRate = "Enter an annual rate greater than 0 and no more than 100%.";
  }
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (!validDate(startDate)) fields.startDate = "Choose a valid start date.";
  if (!validDate(maturityDate)) fields.maturityDate = "Choose a valid maturity date.";
  if (!validDate(currentValueDate)) fields.currentValueDate = "Choose a valid value date.";
  if (!fields.startDate && !fields.maturityDate && maturityDate <= startDate) {
    fields.maturityDate = "Maturity date must be after start date.";
  }
  const today = new Date().toISOString().slice(0, 10);
  if (!fields.startDate && startDate > today) fields.startDate = "Start date cannot be in the future.";
  if (!fields.currentValueDate && (currentValueDate < startDate || currentValueDate > today)) {
    fields.currentValueDate = "Value date must be between start date and today.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  const path = `/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}/fixed-deposits`;
  const body = JSON.stringify({
    name,
    bank_reference: bankReference,
    principal,
    currency,
    annual_interest_rate: annualInterestRate,
    start_date: startDate,
    maturity_date: maturityDate,
    current_value: currentValue,
    current_value_date: currentValueDate,
  });
  try {
    await authenticatedRequest<FixedDeposit>(path, { method: "POST", body });
  } catch (error) {
    return {
      message: error instanceof Error ? error.message : "The fixed deposit could not be created.",
    };
  }

  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}`);
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`);
  revalidatePath("/assets");
  return { message: "Fixed deposit added to the ledger.", success: true };
}

export async function recordFixedDepositValueAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const portfolioID = String(formData.get("portfolioId") ?? "").trim();
  const accountID = String(formData.get("accountId") ?? "").trim();
  const fixedDepositID = String(formData.get("fixedDepositId") ?? "").trim();
  const currentValue = String(formData.get("currentValue") ?? "").trim();
  const currentValueDate = String(formData.get("currentValueDate") ?? "").trim();
  const fields: Record<string, string> = {};
  validatePositive(currentValue, "currentValue", "Enter a current value greater than zero.", fields);
  if (!validDate(currentValueDate)) {
    fields.currentValueDate = "Choose a valid value date.";
  } else if (currentValueDate > new Date().toISOString().slice(0, 10)) {
    fields.currentValueDate = "Value date cannot be in the future.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  const path = `/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}/fixed-deposits/${encodeURIComponent(fixedDepositID)}/values`;
  try {
    await authenticatedRequest<FixedDeposit>(path, {
      method: "POST",
      body: JSON.stringify({ current_value: currentValue, current_value_date: currentValueDate }),
    });
  } catch (error) {
    return {
      message: error instanceof Error ? error.message : "The fixed deposit value could not be recorded.",
    };
  }

  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}`);
  revalidatePath(`/portfolios/${encodeURIComponent(portfolioID)}/accounts/${encodeURIComponent(accountID)}`);
  return { message: "Current value recorded.", success: true };
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
    const auth = await apiRequest<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    await setSession(auth);
    return apiRequest<T>(path, { ...options, accessToken: auth.access_token });
  }
}

function validatePositive(value: string, field: string, message: string, fields: Record<string, string>) {
  if (!value || !Number.isFinite(Number(value)) || Number(value) <= 0) fields[field] = message;
}

function validDate(value: string) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const date = new Date(`${value}T00:00:00Z`);
  return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === value;
}
