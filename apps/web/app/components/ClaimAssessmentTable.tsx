import type { ClaimAssessment } from '@/lib/types';

const STATUS_COLOR: Record<string, string> = {
  grounded:     'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
  inferred:     'bg-sky-500/20 text-sky-300 border-sky-500/30',
  extrapolated: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
  unknown:      'bg-zinc-500/20 text-zinc-300 border-zinc-500/30',
};
const STATUS_GLYPH: Record<string, string> = {
  grounded:     '⊢',
  inferred:     '⊨',
  extrapolated: '⊬',
  unknown:      '⊥',
};

function uiLabel(c: ClaimAssessment): string {
  // "Speculative" UI surface: ⊬ extrapolated WITHOUT a stated Basis.
  if (c.status === 'extrapolated' && !c.basis) return 'Speculative';
  return c.status.charAt(0).toUpperCase() + c.status.slice(1);
}

export function ClaimAssessmentTable({ claims }: { claims: ClaimAssessment[] }) {
  if (!claims?.length) {
    return <p className="text-sm text-zinc-500">No claim assessments produced.</p>;
  }
  return (
    <div className="overflow-x-auto rounded-lg border border-zinc-800">
      <table className="w-full text-sm">
        <thead className="bg-zinc-900 text-xs uppercase tracking-wider text-zinc-400">
          <tr>
            <th className="px-3 py-2 text-left">Claim ID</th>
            <th className="px-3 py-2 text-left">Claim</th>
            <th className="px-3 py-2 text-left">EEF</th>
            <th className="px-3 py-2 text-left">SPT</th>
            <th className="px-3 py-2 text-left">Basis</th>
            <th className="px-3 py-2 text-right">Conf.</th>
          </tr>
        </thead>
        <tbody>
          {claims.map((c) => (
            <tr key={c.claimId} className="border-t border-zinc-800">
              <td className="px-3 py-2 align-top font-mono text-xs text-zinc-400">{c.claimId}</td>
              <td className="px-3 py-2 align-top">{c.claim}</td>
              <td className="px-3 py-2 align-top">
                <span
                  className={`inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs ${
                    STATUS_COLOR[c.status] ?? 'border-zinc-700 bg-zinc-800 text-zinc-300'
                  }`}
                >
                  <span className="font-bold">{STATUS_GLYPH[c.status] ?? '·'}</span>
                  {uiLabel(c)}
                </span>
              </td>
              <td className="px-3 py-2 align-top">
                {c.sptViolations?.length ? (
                  <div className="flex flex-wrap gap-1">
                    {c.sptViolations.map((v) => (
                      <span key={v} className="rounded bg-rose-500/20 px-1.5 py-0.5 font-mono text-[10px] text-rose-300">
                        {v}
                      </span>
                    ))}
                  </div>
                ) : (
                  <span className="text-zinc-600">—</span>
                )}
              </td>
              <td className="px-3 py-2 align-top text-zinc-400">
                {c.basis ? <span className="text-zinc-300">{c.basis}</span> : '—'}
              </td>
              <td className="px-3 py-2 align-top text-right font-mono text-xs">
                {Math.round((c.confidence ?? 0) * 100)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
