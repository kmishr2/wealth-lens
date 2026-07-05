"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, FormState, User } from "@/lib/types";

type ProfilePayload = {
  display_name: string;
  base_currency: string;
  timezone: string;
};

export async function updateProfileAction(
  _state: FormState,
  formData: FormData,
): Promise<FormState> {
  const displayName = String(formData.get("displayName") ?? "").trim();
  const baseCurrency = String(formData.get("baseCurrency") ?? "")
    .trim()
    .toUpperCase();
  const timezone = String(formData.get("timezone") ?? "").trim();
  const fields: Record<string, string> = {};

  if (!displayName) fields.displayName = "Enter a display name.";
  if (!/^[A-Z]{3}$/.test(baseCurrency)) {
    fields.baseCurrency = "Use a three-letter currency code.";
  }
  if (!isValidTimezone(timezone)) {
    fields.timezone = "Enter a valid IANA timezone, such as Asia/Kolkata.";
  }
  if (Object.keys(fields).length > 0) {
    return { message: "Check the highlighted fields.", fields };
  }

  const body: ProfilePayload = {
    display_name: displayName,
    base_currency: baseCurrency,
    timezone,
  };
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");

  try {
    await patchProfile(accessToken, body);
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      return {
        message:
          error instanceof Error
            ? error.message
            : "The profile could not be updated.",
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
      await patchProfile(accessToken, body);
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) {
        redirect("/login");
      }
      return {
        message:
          refreshError instanceof Error
            ? refreshError.message
            : "The profile could not be updated.",
      };
    }
  }

  revalidatePath("/", "layout");
  return { message: "Profile updated.", success: true };
}

function patchProfile(accessToken: string, body: ProfilePayload) {
  return apiRequest<User>("/users/me", {
    method: "PATCH",
    accessToken,
    body: JSON.stringify(body),
  });
}

function isValidTimezone(timezone: string) {
  if (!timezone) return false;
  try {
    new Intl.DateTimeFormat("en", { timeZone: timezone }).format();
    return true;
  } catch {
    return false;
  }
}
