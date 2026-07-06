import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Notification } from "@/lib/types";

export const metadata: Metadata = { title: "Notices" };

async function listNotifications() {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try {
    return await apiRequest<Notification[]>("/notifications", { accessToken });
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect("/auth/refresh?next=/notifications");
      redirect("/login");
    }
    throw error;
  }
}

export default async function NotificationsPage() {
  const notifications = await listNotifications();
  return (
    <main className="mx-auto max-w-5xl px-6 py-10 lg:px-10 lg:py-14">
      <div className="border-b border-[var(--line)] pb-9">
        <p className="eyebrow">Deterministic reminders</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">Notices</h1>
        <p className="mt-4 max-w-3xl leading-7 text-[var(--muted)]">Current facts derived from your contracts. Notices do not predict markets or recommend financial actions.</p>
      </div>

      {notifications.length === 0 ? (
        <div className="mt-8 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No active notices.</div>
      ) : (
        <div className="mt-8 space-y-4">
          {notifications.map((notification) => (
            <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6" key={notification.id}>
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <p className="eyebrow">{notification.portfolio_name} · {notification.account_name}</p>
                  <h2 className="mt-2 text-xl font-semibold">{notification.title}</h2>
                  <p className="mt-3 leading-7 text-[var(--muted)]">{notification.explanation}</p>
                </div>
                <span className={`rounded-full border px-3 py-1 text-xs font-bold uppercase tracking-[0.08em] ${statusClass(notification.status)}`}>{notification.status}</span>
              </div>
              <div className="mt-5 flex flex-wrap items-center justify-between gap-4 border-t border-[var(--line)] pt-4">
                <p className="text-xs leading-5 text-[var(--muted)]">Rule: {notification.trigger_rule}</p>
                <Link className="focus-ring rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline" href={`/portfolios/${notification.portfolio_id}/accounts/${notification.account_id}`}>Open account →</Link>
              </div>
            </article>
          ))}
        </div>
      )}
    </main>
  );
}

function statusClass(status: Notification["status"]) {
  if (status === "overdue" || status === "due") return "border-[#e3aaa1] bg-[#fff4f2] text-[var(--danger)]";
  if (status === "urgent") return "border-[#dec993] bg-[#fff9e9] text-[#775a12]";
  return "border-[var(--line)] bg-[var(--brand-soft)] text-[var(--brand)]";
}
