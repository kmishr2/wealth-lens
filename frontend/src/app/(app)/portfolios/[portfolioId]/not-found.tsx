import Link from "next/link";

export default function PortfolioNotFound() {
  return (
    <main className="mx-auto grid min-h-[70vh] max-w-xl place-items-center px-6 text-center">
      <div>
        <p className="eyebrow">Portfolio not found</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-[-0.04em]">
          This portfolio is unavailable.
        </h1>
        <p className="mt-4 leading-7 text-[var(--muted)]">
          It may have been removed, or it does not belong to your account.
        </p>
        <Link
          className="focus-ring mt-7 inline-flex rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white hover:bg-[var(--brand-strong)]"
          href="/dashboard"
        >
          Return to portfolios
        </Link>
      </div>
    </main>
  );
}
