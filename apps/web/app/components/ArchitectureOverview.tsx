// Static 7-box horizontal flow diagram — the highest-leverage dashboard
// element. Answers "what does this product do?" in under 2 seconds.

const STEPS: Array<{ id: string; title: string; line: string }> = [
  { id: 'judge',     title: 'Judge',          line: 'Asks for trusted web data' },
  { id: 'buyer',     title: 'Buyer Agent',    line: 'Extracts intent and spend policy' },
  { id: 'brightdata',title: 'Bright Data',    line: 'Finds and fetches public evidence' },
  { id: 'seller',    title: 'Seller Offer',   line: 'Forms candidate data deal' },
  { id: 'gem2',      title: 'GEM² Gate',      line: 'Checks if claims are grounded' },
  { id: 'x402',      title: 'x402 Settle',    line: 'Approves or blocks payment' },
  { id: 'bundle',    title: 'Audit Bundle',   line: 'Exports replayable record' },
];

export function ArchitectureOverview() {
  return (
    <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
      <div className="mb-4 flex items-baseline justify-between gap-4">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          Architecture Overview
        </h2>
        <span className="text-[11px] text-zinc-500">
          Verification infrastructure pipeline · one direction
        </span>
      </div>

      {/* Horizontal flow on lg+; vertical stack on small screens */}
      <div className="flex flex-col gap-3 lg:flex-row lg:items-stretch lg:gap-2">
        {STEPS.map((s, i) => (
          <Step key={s.id} step={s} hasArrow={i < STEPS.length - 1} />
        ))}
      </div>

      <p className="mt-4 text-xs italic text-zinc-500">
        Every box has its own contract. Money only moves after the last green check.
      </p>
    </section>
  );
}

function Step({
  step,
  hasArrow,
}: {
  step: { title: string; line: string };
  hasArrow: boolean;
}) {
  return (
    <div className="flex flex-1 items-stretch gap-2 lg:flex-col lg:items-stretch">
      <div className="flex-1 rounded-lg border border-zinc-700 bg-zinc-950 p-3 transition hover:border-zinc-500">
        <div className="text-xs font-semibold uppercase tracking-wider text-zinc-200">
          {step.title}
        </div>
        <p className="mt-1 text-[11px] leading-relaxed text-zinc-400">{step.line}</p>
      </div>
      {hasArrow && (
        <div
          className="flex items-center justify-center text-zinc-600 lg:py-0"
          aria-hidden="true"
        >
          {/* Right arrow on lg, down arrow on small */}
          <span className="hidden text-lg lg:inline">→</span>
          <span className="text-lg lg:hidden">↓</span>
        </div>
      )}
    </div>
  );
}
