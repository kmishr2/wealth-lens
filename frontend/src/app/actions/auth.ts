"use server";

import { redirect } from "next/navigation";
import { apiRequest } from "@/lib/api";
import { clearSession, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FormState } from "@/lib/types";

const emptyState: FormState = { message: "" };

function text(formData: FormData, key: string) {
  return String(formData.get(key) ?? "").trim();
}

function errorMessage(error: unknown) {
  return error instanceof Error
    ? error.message
    : "The request could not be completed.";
}

export async function loginAction(
  previousState: FormState = emptyState,
  formData: FormData,
): Promise<FormState> {
  void previousState;
  const email = text(formData, "email").toLowerCase();
  const password = String(formData.get("password") ?? "");
  const fields: Record<string, string> = {};
  if (!email || !email.includes("@")) fields.email = "Enter a valid email.";
  if (!password) fields.password = "Enter your password.";
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  let auth: AuthResponse;
  try {
    auth = await apiRequest<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
  } catch (error) {
    return { message: errorMessage(error) };
  }

  await setSession(auth);
  redirect("/dashboard");
}

export async function registerAction(
  previousState: FormState = emptyState,
  formData: FormData,
): Promise<FormState> {
  void previousState;
  const displayName = text(formData, "displayName");
  const email = text(formData, "email").toLowerCase();
  const password = String(formData.get("password") ?? "");
  const baseCurrency = text(formData, "baseCurrency").toUpperCase();
  const timezone = text(formData, "timezone") || "UTC";
  const fields: Record<string, string> = {};

  if (!displayName) fields.displayName = "Enter your name.";
  if (!email || !email.includes("@")) fields.email = "Enter a valid email.";
  if (password.length < 12) {
    fields.password = "Use at least 12 characters.";
  }
  if (!/^[A-Z]{3}$/.test(baseCurrency)) {
    fields.baseCurrency = "Use a three-letter currency code.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  let auth: AuthResponse;
  try {
    auth = await apiRequest<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({
        display_name: displayName,
        email,
        password,
        base_currency: baseCurrency,
        timezone,
      }),
    });
  } catch (error) {
    return { message: errorMessage(error) };
  }

  await setSession(auth);
  redirect("/dashboard");
}

export async function logoutAction() {
  const refreshToken = await getRefreshToken();
  if (refreshToken) {
    try {
      await apiRequest<Record<string, never>>("/auth/logout", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    } catch {
      // Local logout remains available when the backend cannot be reached.
    }
  }
  await clearSession();
  redirect("/login");
}
