import type { FinalReport } from '@/lib/types';

export function FinalReportPanel({
  report,
  durationMs,
}: {
  report: FinalReport;
  durationMs: number;
}) {
  const isApproved = report.headline.toLowerCase().includes('approve');
  const isBlocked = report.headline.toLowerCase().includes('block');

  const headlineCls = isApproved
    ? 'text-emerald-300'
    : isBlocked
      ? 'text-red-300'
      : 'text-amber-300';

  return (
    <section className="rounded-xl border border-zinc-800 bg-gradient-to-br from-zinc-900/80 to-zinc-950 p-6 shadow-lg">
      <div className="mb-2 flex items-baseline justify-between gap-4">
        <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
          Final Report
        </h2>
        <span className="text-xs text-zinc-500">
          completed in <span className="font-mono text-zinc-300">{(durationMs / 1000).toFixed(1)}s</span>
        </span>
      </div>

      <h1 className={`mb-3 text-3xl font-bold tracking-tight ${headlineCls}`}>
        {report.headline}
      </h1>

      <p className="mb-4 text-lg leading-relaxed text-zinc-200">{report.result}</p>

      <dl className="space-y-3 text-sm">
        <ReportRow label="Request">{report.request}</ReportRow>
        <ReportRow label="Reason">{report.reason}</ReportRow>
        <ReportRow label="Evidence">{report.evidenceSummary}</ReportRow>
        <ReportRow label="Payment">{report.paymentSummary}</ReportRow>
        {report.auditBundleRef && (
          <ReportRow label="Audit bundle">
            <code className="break-all text-xs text-zinc-300">{report.auditBundleRef}</code>
          </ReportRow>
        )}
      </dl>

      <p className="mt-5 border-t border-zinc-800 pt-4 text-sm italic text-zinc-400">
        {report.tagline || 'Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.'}
      </p>
    </section>
  );
}

function ReportRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-[max-content,1fr] gap-x-4">
      <dt className="text-xs font-semibold uppercase tracking-wider text-zinc-500">{label}</dt>
      <dd className="text-zinc-300">{children}</dd>
    </div>
  );
}
