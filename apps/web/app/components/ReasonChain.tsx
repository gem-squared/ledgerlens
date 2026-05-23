import type { GateResponse } from '@/lib/types';

// ReasonChain renders the audit-gate's flat `reasons` array as a vertical
// list with bracket tags surfaced as colored chips. The Go reason parser
// drives ClaimAssessmentTable; this component shows the RAW chain so a
// judge can see EXACTLY what the upstream said.
export function ReasonChain({ title, response }: { title: string; response?: GateResponse }) {
  if (!response) return null;
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400">{title}</h3>
        <div className="flex items-center gap-2 text-xs">
          <span
            className={`rounded px-1.5 py-0.5 font-mono font-bold ${
              response.verdict === 'ALLOW' || response.verdict === 'SUCCESS'
                ? 'bg-emerald-500/20 text-emerald-300'
                : 'bg-red-500/20 text-red-300'
            }`}
          >
            {response.verdict}
          </span>
          <span className="font-mono text-zinc-300">{response.score}/100</span>
        </div>
      </div>
      <ol className="mt-3 space-y-1.5 text-xs">
        {response.reasons.map((line, i) => (
          <li key={i} className="font-mono text-zinc-300">
            <ReasonLine line={line} />
          </li>
        ))}
      </ol>
      <div className="mt-3 flex flex-wrap gap-x-4 text-[10px] text-zinc-500">
        <span>result_id: <code>{response.meta?.result_id}</code></span>
        <span>{response.meta?.duration_ms} ms</span>
        {response.meta?.usage?.estimated_cost_usd && (
          <span>cost: ${response.meta.usage.estimated_cost_usd.toFixed(4)}</span>
        )}
      </div>
    </div>
  );
}

function ReasonLine({ line }: { line: string }) {
  // Highlight the leading bracketed tag.
  const m = /^(\[[^\]]+\])\s+(.*)$/.exec(line);
  if (!m) return <span>{line}</span>;
  const tag = m[1];
  const rest = m[2];
  let cls = 'bg-zinc-800 text-zinc-300';
  if (tag.startsWith('[RULE-'))     cls = 'bg-violet-500/20 text-violet-300';
  if (tag.startsWith('[SPT-'))      cls = 'bg-rose-500/20 text-rose-300';
  if (tag.startsWith('[EEF-'))      cls = 'bg-amber-500/20 text-amber-300';
  if (tag.startsWith('[EVIDENCE-')) cls = 'bg-emerald-500/20 text-emerald-300';
  if (tag.startsWith('[DIM-'))      cls = 'bg-sky-500/20 text-sky-300';
  if (tag === '[TYPE]')             cls = 'bg-zinc-700/40 text-zinc-300';
  return (
    <>
      <span className={`mr-2 rounded px-1.5 py-0.5 text-[10px] font-bold ${cls}`}>{tag}</span>
      <span>{rest}</span>
    </>
  );
}
