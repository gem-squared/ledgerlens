'use client';

import { useEffect, useRef, useState } from 'react';
import { getStats } from '@/lib/api';
import type { Stats } from '@/lib/types';

interface StatsDashboardProps {
  /** Incremented by JudgeRequestConsole when a run completes — triggers refetch + pulse. */
  refreshTrigger?: number;
}

// Five LedgerLens-native KPI tiles. Pulses ONLY tiles whose value changed
// since the previous refetch (no empty pulses).
export function StatsDashboard({ refreshTrigger }: StatsDashboardProps) {
  const [stats, setStats] = useState<Stats | null>(null);
  const [pulsing, setPulsing] = useState<Set<string>>(new Set());
  const prevRef = useRef<Stats | null>(null);

  async function refetch() {
    try {
      const fresh = await getStats();
      if (prevRef.current) {
        const changed = new Set<string>();
        if (fresh.dealsAudited !== prevRef.current.dealsAudited) changed.add('dealsAudited');
        if (fresh.approved !== prevRef.current.approved || fresh.blocked !== prevRef.current.blocked)
          changed.add('split');
        if (fresh.avgAuditScore !== prevRef.current.avgAuditScore) changed.add('avgAuditScore');
        if (fresh.simulatedSpendPreventedUSDC !== prevRef.current.simulatedSpendPreventedUSDC)
          changed.add('simulatedSpendPreventedUSDC');
        if (fresh.avgVerificationLatencyMs !== prevRef.current.avgVerificationLatencyMs)
          changed.add('avgVerificationLatencyMs');
        if (changed.size > 0) {
          setPulsing(changed);
          setTimeout(() => setPulsing(new Set()), 1500);
        }
      }
      prevRef.current = fresh;
      setStats(fresh);
    } catch (err) {
      console.error('stats fetch failed', err);
    }
  }

  useEffect(() => {
    void refetch();
  }, []);
  useEffect(() => {
    if (refreshTrigger !== undefined && refreshTrigger > 0) void refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshTrigger]);

  if (!stats) {
    return (
      <section className="glass p-5">
        <p className="text-xs text-zinc-500">Loading verification infrastructure metrics…</p>
      </section>
    );
  }

  return (
    <section className="glass p-5">
      <div className="mb-4 flex flex-wrap items-baseline justify-between gap-3">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          Verification Infrastructure Dashboard
        </h2>
        <span className="text-[11px] text-zinc-500">
          Aggregated from demo and visitor verification runs · updates after every judge or
          visitor run
        </span>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        <Tile
          id="dealsAudited"
          label="Deals Audited"
          value={stats.dealsAudited}
          sub={`${stats.modesBreakdown.live} live · ${stats.modesBreakdown.replay} replay`}
          pulsing={pulsing}
        />
        <Tile
          id="split"
          label="Approved / Blocked"
          value={`${stats.approved} / ${stats.blocked}`}
          sub={
            stats.escalatedToHuman > 0
              ? `${stats.escalatedToHuman} escalated to human`
              : 'no escalations'
          }
          pulsing={pulsing}
        />
        <Tile
          id="avgAuditScore"
          label="Avg Audit Score"
          value={`${stats.avgAuditScore} / 100`}
          sub={`over ${stats.sampleSize} ${stats.sampleSize === 1 ? 'deal' : 'deals'}`}
          pulsing={pulsing}
        />
        <Tile
          id="simulatedSpendPreventedUSDC"
          label="Simulated Spend Prevented"
          value={`$${stats.simulatedSpendPreventedUSDC.toFixed(4)}`}
          sub={`blocked demo offers · USDC-demo`}
          pulsing={pulsing}
        />
        <Tile
          id="avgVerificationLatencyMs"
          label="Avg Verification Latency"
          value={`${(stats.avgVerificationLatencyMs / 1000).toFixed(1)}s`}
          sub="end-to-end per deal"
          pulsing={pulsing}
        />
      </div>

      {/* Footer status line — audit-gate health + last-updated */}
      <div className="mt-4 flex flex-wrap items-center justify-between gap-x-6 gap-y-1 border-t border-zinc-800 pt-3 text-[11px] text-zinc-500">
        <span>
          Audit gate{' '}
          <code className="text-zinc-300">{stats.auditGateUrl}</code>
          {' · '}
          <span className="text-emerald-400">●</span> responding
          {stats.auditGateAvgLatencyMs > 0 && (
            <>
              {' · '}
              <span className="font-mono">{(stats.auditGateAvgLatencyMs / 1000).toFixed(1)}s</span>{' '}
              avg L1/L2 latency
            </>
          )}
        </span>
        <span>
          {stats.auditBundlesExported} audit bundle{stats.auditBundlesExported === 1 ? '' : 's'}{' '}
          on disk
          {stats.lastUpdatedAt && (
            <>
              {' · updated '}
              <RelativeTime iso={stats.lastUpdatedAt} />
            </>
          )}
        </span>
      </div>
    </section>
  );
}

// ── Tile ─────────────────────────────────────────────────────────────────

interface TileProps {
  id: string;
  label: string;
  value: string | number;
  sub: string;
  pulsing: Set<string>;
}

function Tile({ id, label, value, sub, pulsing }: TileProps) {
  const isPulsing = pulsing.has(id);
  const glowClass = isPulsing
    ? id === 'split' || id === 'simulatedSpendPreventedUSDC'
      ? 'glow-red'
      : id === 'avgAuditScore'
        ? 'glow-green'
        : 'glow-indigo'
    : '';
  return (
    <div
      className={`glass p-3 transition ${
        isPulsing ? 'll-tile-pulsing' : ''
      } ${glowClass}`}
    >
      <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
        {label}
      </div>
      <div className="mt-1 text-2xl font-bold tracking-tight">
        <span className={isPulsing ? 'll-num-bumping' : ''}>{value}</span>
      </div>
      <div className="mt-0.5 text-[11px] text-zinc-500">{sub}</div>
    </div>
  );
}

// ── Relative time ────────────────────────────────────────────────────────

function RelativeTime({ iso }: { iso: string }) {
  const [_, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick((n) => (n + 1) % 1_000_000), 30_000);
    return () => clearInterval(id);
  }, []);
  return <span>{formatRelative(iso)}</span>;
}

function formatRelative(iso: string): string {
  const ts = Date.parse(iso);
  if (Number.isNaN(ts)) return iso;
  const diffSec = Math.max(0, Math.floor((Date.now() - ts) / 1000));
  if (diffSec < 5) return 'just now';
  if (diffSec < 60) return `${diffSec}s ago`;
  const m = Math.floor(diffSec / 60);
  if (m < 60) return `${m} min ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h} hr ago`;
  const d = Math.floor(h / 24);
  return `${d}d ago`;
}
