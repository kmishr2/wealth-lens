import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { User } from "@/lib/types";

export async function requireUser(nextPath = "/dashboard"): Promise<User> {
  const accessToken = await getAccessToken();
  if (!accessToken) {
    if (await getRefreshToken()) {
      redirect(`/auth/refresh?next=${encodeURIComponent(nextPath)}`);
    }
    redirect("/login");
  }

  try {
    return await apiRequest<User>("/users/me", { accessToken });
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) {
        redirect(`/auth/refresh?next=${encodeURIComponent(nextPath)}`);
      }
      redirect("/login");
    }
    throw error;
  }
}
