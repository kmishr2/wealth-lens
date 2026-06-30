export function AuthShell({
  eyebrow,
  title,
  description,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <main className="grid min-h-screen lg:grid-cols-[minmax(320px,0.9fr)_minmax(520px,1.1fr)]">
      <section className="relative hidden overflow-hidden bg-[#173c30] p-12 text-white lg:flex lg:flex-col lg:justify-between xl:p-16">
        <div className="absolute -right-28 -top-28 h-96 w-96 rounded-full border border-white/10" />
        <div className="absolute -bottom-36 -left-20 h-[30rem] w-[30rem] rounded-full bg-[#215943]" />
        <BrandMark inverse />
        <div className="relative max-w-lg">
          <p className="mb-5 text-xs font-bold uppercase tracking-[0.2em] text-[#d9b56c]">
            Clarity over prediction
          </p>
          <blockquote className="text-4xl font-medium leading-[1.15] tracking-[-0.035em] xl:text-5xl">
            Every number should be traceable back to what actually happened.
          </blockquote>
          <p className="mt-8 max-w-md text-base leading-7 text-white/65">
            WealthLens turns an immutable transaction ledger into transparent,
            formula-based portfolio analytics.
          </p>
        </div>
        <p className="relative text-xs tracking-wide text-white/45">
          No predictions. No hidden scoring. No trading incentives.
        </p>
      </section>

      <section className="flex min-h-screen items-center justify-center px-6 py-12 sm:px-10">
        <div className="w-full max-w-md">
          <div className="mb-12 lg:hidden">
            <BrandMark />
          </div>
          <p className="eyebrow">{eyebrow}</p>
          <h1 className="mt-3 text-4xl font-semibold tracking-[-0.04em] sm:text-[2.75rem]">
            {title}
          </h1>
          <p className="mt-4 max-w-sm leading-7 text-[var(--muted)]">
            {description}
          </p>
          {children}
        </div>
      </section>
    </main>
  );
}

export function BrandMark({ inverse = false }: { inverse?: boolean }) {
  return (
    <div className="relative flex items-center gap-3">
      <span
        className={`grid h-10 w-10 place-items-center rounded-xl border text-sm font-black ${
          inverse
            ? "border-white/20 bg-white/10 text-white"
            : "border-[var(--brand)]/15 bg-[var(--brand-soft)] text-[var(--brand)]"
        }`}
      >
        WL
      </span>
      <span className="text-lg font-bold tracking-[-0.025em]">WealthLens</span>
    </div>
  );
}
