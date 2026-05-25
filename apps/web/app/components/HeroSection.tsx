export function HeroSection() {
  return (
    <header className="mb-8">
      <div className="flex flex-wrap items-baseline justify-between gap-4">
        <div>
          <span className="inline-block rounded-full bg-simBadge px-3 py-1 text-xs font-semibold uppercase tracking-wider text-white">
            Simulation Mode
          </span>
          <h1 className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-4xl font-bold tracking-tight">
            <span>LedgerLens</span>
            <span aria-hidden className="text-3xl font-light text-zinc-500">×</span>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/brightdata-logo.png"
              alt="Bright Data"
              className="h-7 w-auto sm:h-8"
            />
          </h1>
          <p className="mt-1 text-lg text-zinc-400">No grounded claim, no payment.</p>
        </div>
        <a
          href="https://github.com/gem-squared/ledgerlens"
          className="rounded-md border border-zinc-700 px-3 py-1.5 text-xs text-zinc-400 transition hover:border-zinc-500 hover:text-zinc-200"
        >
          source ↗
        </a>
      </div>

      <p className="mt-4 max-w-3xl text-sm text-zinc-400">
        LedgerLens is a trust-gated commerce layer for autonomous agents. A buyer agent
        interprets the judge&apos;s request, Bright Data collects public evidence, and GEM²
        audits the seller&apos;s claim before any payment is allowed. Settlement is x402-shaped
        simulation: the trust gate is real.
      </p>

      <p className="mt-3 max-w-3xl text-sm italic text-zinc-300">
        Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.
      </p>

      {/* Three product pillars */}
      <div className="mt-5 grid gap-3 sm:grid-cols-3">
        <PillarCard title="Bright Data Evidence">
          SERP + Unlocker + Browser + MCP collect public-web proof.
        </PillarCard>
        <PillarCard title="GEM² Trust Gate">
          L1 P-check + L2 O-check classify claims · ⊢ ⊨ ⊬ ⊥
        </PillarCard>
        <PillarCard title="x402 Settlement">
          L3 releases or blocks simulated payment.
        </PillarCard>
      </div>
    </header>
  );
}

function PillarCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3">
      <div className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
        {title}
      </div>
      <p className="mt-1 text-sm leading-relaxed text-zinc-300">{children}</p>
    </div>
  );
}
