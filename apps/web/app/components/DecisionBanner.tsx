import type { DecisionPacket } from '@/lib/types';

export function DecisionBanner({ decision }: { decision: DecisionPacket }) {
  const map: Record<string, { bg: string; label: string; emoji: string }> = {
    APPROVED_BY_TRUST_GATE: { bg: 'bg-emerald-600', label: 'PAYMENT APPROVED',     emoji: '✓' },
    BLOCKED_BY_TRUST_GATE:  { bg: 'bg-red-600',     label: 'PAYMENT BLOCKED',      emoji: '✕' },
    ESCALATED_TO_HUMAN:     { bg: 'bg-amber-500',   label: 'ESCALATED TO HUMAN',   emoji: '?' },
  };
  const t = map[decision.verdict] ?? { bg: 'bg-zinc-600', label: decision.verdict, emoji: '·' };
  return (
    <div className={`${t.bg} rounded-lg p-5 text-white shadow-lg`}>
      <div className="flex items-center gap-3">
        <span className="text-3xl font-bold">{t.emoji}</span>
        <div>
          <div className="text-xs uppercase tracking-widest opacity-90">L3 Trust Gate verdict</div>
          <div className="text-2xl font-bold">{t.label}</div>
        </div>
      </div>
      <p className="mt-3 text-sm opacity-95">{decision.reason}</p>
      <div className="mt-3 flex flex-wrap gap-x-6 gap-y-1 text-xs opacity-80">
        <span>decisionId: <code>{decision.decisionId}</code></span>
        <span>L1 result: <code>{decision.l1ResultId || '—'}</code></span>
        <span>L2 result: <code>{decision.l2ResultId || '—'}</code></span>
      </div>
    </div>
  );
}
