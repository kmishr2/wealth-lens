"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, Benchmark, BenchmarkObservation, FormState } from "@/lib/types";

export async function createBenchmarkAction(_state: FormState, formData: FormData): Promise<FormState> {
  const code = String(formData.get("code") ?? "").trim().toUpperCase();
  const name = String(formData.get("name") ?? "").trim();
  const currency = String(formData.get("currency") ?? "").trim().toUpperCase();
  const source = String(formData.get("source") ?? "").trim();
  const description = String(formData.get("description") ?? "").trim();
  const fields: Record<string, string> = {};
  if (!code) fields.code = "Enter a benchmark code.";
  if (!name) fields.name = "Enter a benchmark name.";
  if (!/^[A-Z]{3}$/.test(currency)) fields.currency = "Use a three-letter currency code.";
  if (!source) fields.source = "Enter the data source.";
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };
  try {
    await authenticatedRequest<Benchmark>("/benchmarks", { method: "POST", body: JSON.stringify({ code, name, currency, source, description }) });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The benchmark could not be created." };
  }
  revalidatePath("/benchmarks");
  return { message: "Benchmark created.", success: true };
}

export async function createBenchmarkObservationAction(_state: FormState, formData: FormData): Promise<FormState> {
  const benchmarkID = String(formData.get("benchmarkId") ?? "");
  const date = String(formData.get("observationDate") ?? "");
  const value = String(formData.get("value") ?? "");
  const source = String(formData.get("source") ?? "").trim();
  const note = String(formData.get("note") ?? "").trim();
  const fields: Record<string, string> = {};
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) fields.observationDate = "Choose an observation date.";
  if (!value || !Number.isFinite(Number(value)) || Number(value) <= 0) fields.value = "Enter a value greater than zero.";
  if (Object.keys(fields).length > 0) return { message: "Check the highlighted fields.", fields };
  try {
    await authenticatedRequest<BenchmarkObservation>(`/benchmarks/${encodeURIComponent(benchmarkID)}/observations`, {
      method: "POST", body: JSON.stringify({ observation_date: date, value, source, note }),
    });
  } catch (error) {
    return { message: error instanceof Error ? error.message : "The observation could not be recorded." };
  }
  revalidatePath(`/benchmarks/${encodeURIComponent(benchmarkID)}`);
  return { message: "Observation recorded.", success: true };
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
