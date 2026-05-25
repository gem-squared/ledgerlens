'use client';

import type { DealRunResult } from '@/lib/types';
import { FinalReportPanel } from './FinalReportPanel';
import { AuditScoreCard } from './AuditScoreCard';
import { SettlementCard } from './SettlementCard';
import { EvidenceList } from './EvidenceList';
import { ClaimAssessmentTable } from './ClaimAssessmentTable';

// One canonical post-run rendering surface. Used by:
//   - JudgeRequestConsole (LIVE deal + REPLAY Case A/B)
//   - BundleViewer (when a judge clicks "View" on a Recent Activity row)
// All entry points get the same Final Report banner + GEM² Audit Score
// rings + X402 Settlement card + Bright Data Evidence list + collapsible
// canonical-EEF claim assessment + footer line.
export function RichRunResult({ result }: { result: DealRunResult }) {
  return (
    <div className="space-y-6">
      <FinalReportPanel report={result.finalReport} durationMs={result.durationMs} />

      <AuditScoreCard l1={result.l1} l2={result.l2} />

      <div className="grid gap-6 lg:grid-cols-2">
        <SettlementCard settlement={result.settlement} />
        <section className="rounded-xl border border-zinc-800 bg-zinc-900/30 p-5">
          <h3 className="mb-3 flex items-baseline justify-between text-xs font-semibold uppercase tracking-wider text-zinc-300">
            <span>Bright Data Evidence</span>
            <span className="font-mono text-zinc-500">
              {result.evidenceReceipts?.length ?? 0} receipts
            </span>
          </h3>
          <EvidenceList receipts={result.evidenceReceipts} />
        </section>
      </div>

      <details className="group rounded-xl border border-zinc-800 bg-zinc-900/30 open:bg-zinc-900/40">
        <summary className="cursor-pointer select-none list-none px-5 py-3 text-sm text-zinc-300 hover:bg-zinc-900/60">
          <span className="inline-block w-4 text-zinc-500 transition-transform group-open:rotate-90">›</span>
          Show claim-by-claim assessment ({result.decision?.claimAssessments?.length ?? 0} claims · canonical EEF tags)
        </summary>
        <div className="border-t border-zinc-800 p-5">
          <ClaimAssessmentTable claims={result.decision?.claimAssessments ?? []} />
        </div>
      </details>

      <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-zinc-500">
        <span>mode: <code>{result.mode}</code></span>
        <span>duration: {result.durationMs} ms</span>
        <span>bundle: <code className="break-all">{result.bundlePath}</code></span>
      </div>
    </div>
  );
}
