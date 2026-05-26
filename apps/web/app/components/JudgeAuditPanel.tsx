export function JudgeAuditPanel() {
  return (
    <section className="glass-panel">
      <div className="mb-3 flex items-baseline justify-between gap-3">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          Judge Audit (Read-Only)
        </h2>
        <span className="text-[11px] text-zinc-500">independent verification</span>
      </div>

      <p className="text-sm leading-relaxed text-zinc-300">
        Every deal attempt produces a read-only audit bundle. Each bundle contains the buyer
        request, seller offer, Bright Data evidence receipts, GEM² audit results, L3 verdict,
        and simulated settlement record.
      </p>

      <pre className="mt-3 overflow-x-auto rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 font-mono text-[11px] leading-relaxed text-zinc-300">
{`# List recent bundles
curl https://ledgerlens.gemsquared.ai/api/v1/audit-bundles

# Fetch a specific bundle (decisionId from any row above)
curl https://ledgerlens.gemsquared.ai/api/v1/audit-bundles/<decisionId>

# Live aggregate stats (no auth required — read-only)
curl https://ledgerlens.gemsquared.ai/api/v1/stats`}
      </pre>

      <ul className="mt-4 space-y-1 text-xs text-zinc-400">
        <li>✓ No private keys. No Coinbase account. No real funds.</li>
        <li>✓ Settlement is simulation-only — every receipt carries <code className="text-zinc-300">real_transaction: false</code>.</li>
        <li>✓ Audit bundles are hash-chained: <code className="text-zinc-300">evidenceHash</code> ties each decision to its evidence corpus.</li>
        <li>✓ Upstream L1/L2 <code className="text-zinc-300">result_id</code>s are preserved for replay through <code className="text-zinc-300">gem2-tpmn-checker.fly.dev</code>.</li>
      </ul>

      <p className="mt-4 text-xs italic text-zinc-500">
        Aggregated metrics on the dashboard are computed from these local audit bundles. Every
        run extends the public verification record.
      </p>
    </section>
  );
}
