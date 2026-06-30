import type { ApiEnvelope } from "@/lib/types";

const API_BASE_URL =
  process.env.BACKEND_API_URL ?? "http://127.0.0.1:8080/api/v1";

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly code: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type ApiRequestOptions = RequestInit & {
  accessToken?: string;
};

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
): Promise<T> {
  const { accessToken, headers, ...requestOptions } = options;
  const requestHeaders = new Headers(headers);
  requestHeaders.set("Accept", "application/json");
  if (requestOptions.body) {
    requestHeaders.set("Content-Type", "application/json");
  }
  if (accessToken) {
    requestHeaders.set("Authorization", `Bearer ${accessToken}`);
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...requestOptions,
      headers: requestHeaders,
      cache: "no-store",
    });
  } catch {
    throw new ApiError(
      "The WealthLens API is unavailable. Check that the backend is running.",
      503,
      "BACKEND_UNAVAILABLE",
    );
  }

  let envelope: ApiEnvelope<T>;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch {
    throw new ApiError(
      "The API returned an unreadable response.",
      response.status,
      "INVALID_API_RESPONSE",
    );
  }

  if (!response.ok || !envelope.success || envelope.data === undefined) {
    throw new ApiError(
      envelope.error?.message ?? "The request could not be completed.",
      response.status,
      envelope.error?.code ?? "API_ERROR",
    );
  }

  return envelope.data;
}
