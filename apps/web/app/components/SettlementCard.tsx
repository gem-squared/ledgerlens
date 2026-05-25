import type { SimulatedSettlement } from '@/lib/types';

export function SettlementCard({ settlement }: { settlement: SimulatedSettlement }) {
  const settled = settlement.status === 'SIMULATED_SETTLED';
  return (
    <div className={`rounded-lg border ${settled ? 'border-emerald-500/30 bg-emerald-500/5' : 'border-zinc-700 bg-zinc-900/40'} p-4`}>
      <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1">
        <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400">x402 Settlement</h3>
        <span className="inline-block whitespace-nowrap rounded-full bg-simBadge px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-white">
          SIMULATION MODE
        </span>
      </div>
      <div className="mt-3 grid grid-cols-2 gap-x-6 gap-y-1.5 text-sm">
        <Cell k="settlementId" v={settlement.settlementId || '—'} mono />
        <Cell k="status"       v={settlement.status} />
        <Cell k="mode"         v={settlement.mode} />
        <Cell k="network"      v={settlement.network} />
        <Cell k="asset"        v={settlement.asset} />
        <Cell k="amountUSDC"   v={String(settlement.amountUSDC ?? 0)} />
        <Cell k="realTransaction"  v={String(settlement.realTransaction ?? false)} />
        <Cell k="privateKeysUsed"  v={String(settlement.privateKeysUsed ?? false)} />
        <Cell k="realFundsUsed"    v={String(settlement.realFundsUsed ?? false)} />
        <Cell k="ts" v={settlement.ts} />
      </div>
      <p className="mt-3 text-xs italic text-zinc-500">
        We simulate settlement, but the trust gate is real.
      </p>
    </div>
  );
}

function Cell({ k, v, mono }: { k: string; v: string; mono?: boolean }) {
  return (
    <div className="flex items-baseline gap-2">
      <span className="w-32 shrink-0 text-xs text-zinc-500">{k}</span>
      <span className={mono ? 'truncate font-mono text-xs' : 'text-zinc-200'}>{v}</span>
    </div>
  );
}
