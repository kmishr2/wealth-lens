import { NextRequest, NextResponse } from "next/server";
import { apiRequest, ApiError } from "@/lib/api";
import { safeNextPath } from "@/lib/navigation";
import {
  ACCESS_COOKIE,
  REFRESH_COOKIE,
  getRefreshToken,
} from "@/lib/session";
import type { AuthResponse } from "@/lib/types";

export async function GET(request: NextRequest) {
  const destination = safeNextPath(request.nextUrl.searchParams.get("next"));
  const refreshToken = await getRefreshToken();
  const loginURL = new URL("/login", request.url);

  if (!refreshToken) {
    return NextResponse.redirect(loginURL);
  }

  try {
    const auth = await apiRequest<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    const response = NextResponse.redirect(new URL(destination, request.url));
    const cookieOptions = {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax" as const,
      path: "/",
    };
    response.cookies.set(ACCESS_COOKIE, auth.access_token, {
      ...cookieOptions,
      maxAge: auth.expires_in,
    });
    response.cookies.set(REFRESH_COOKIE, auth.refresh_token, cookieOptions);
    return response;
  } catch (error) {
    const sessionExpired = error instanceof ApiError && error.status === 401;
    loginURL.searchParams.set(
      "reason",
      sessionExpired ? "session-expired" : "backend-unavailable",
    );
    const response = NextResponse.redirect(loginURL);
    if (sessionExpired) {
      response.cookies.delete(ACCESS_COOKIE);
      response.cookies.delete(REFRESH_COOKIE);
    }
    return response;
  }
}
