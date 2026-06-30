import { redirect } from "next/navigation";
import { getAccessToken, getRefreshToken } from "@/lib/session";

export default async function HomePage() {
  const [accessToken, refreshToken] = await Promise.all([
    getAccessToken(),
    getRefreshToken(),
  ]);

  if (accessToken) {
    redirect("/dashboard");
  }
  if (refreshToken) {
    redirect("/auth/refresh?next=/dashboard");
  }
  redirect("/login");
}
