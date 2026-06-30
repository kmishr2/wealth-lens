import { cookies } from "next/headers";
import type { AuthResponse } from "@/lib/types";

export const ACCESS_COOKIE = "wealth_lens_access";
export const REFRESH_COOKIE = "wealth_lens_refresh";

const baseCookieOptions = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax" as const,
  path: "/",
};

export async function setSession(auth: AuthResponse) {
  const cookieStore = await cookies();
  cookieStore.set(ACCESS_COOKIE, auth.access_token, {
    ...baseCookieOptions,
    maxAge: auth.expires_in,
  });
  cookieStore.set(REFRESH_COOKIE, auth.refresh_token, baseCookieOptions);
}

export async function clearSession() {
  const cookieStore = await cookies();
  cookieStore.delete(ACCESS_COOKIE);
  cookieStore.delete(REFRESH_COOKIE);
}

export async function getAccessToken() {
  return (await cookies()).get(ACCESS_COOKIE)?.value;
}

export async function getRefreshToken() {
  return (await cookies()).get(REFRESH_COOKIE)?.value;
}
