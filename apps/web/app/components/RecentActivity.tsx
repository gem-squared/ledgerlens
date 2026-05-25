'use client';

import { useEffect, useState } from 'react';
import { listAuditBundles } from '@/lib/api';
import type { BundleSummary } from '@/lib/types';
import { BundleViewer } from './BundleViewer';

interface RecentActivityProps {
  /** Increment to force a refetch (e.g. after a run completes). */
  refreshTrigger?: number;
  /** How many rows to show by default; remainder visible via "Show more". */
  initialLimit?: number;
}

export function RecentActivity({ refreshTrigger, initialLimit = 5 }: RecentActivityProps) {
  const [bundles, setBundles] = useState<BundleSummary[] | null>(null);
  const [expanded, setExpanded] = useState(false);
  const [viewing, setViewing] = useState<string | null>(null);

  async function refetch() {
    try {
      const fresh = await listAuditBundles();
      setBundles(fresh);
    } catch (err) {
      console.error('audit-bundle list failed', err);
    }
  }

  useEffect(() => {
    void refetch();
  }, []);
  useEffect(() => {
    if (refreshTrigger !== undefined && refreshTrigger > 0) void refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshTrigger]);

  if (!bundles) {
    return (
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
        <p className="text-xs text-zinc-500">Loading recent activity…</p>
      </section>
    );
  }

  const visible = expanded ? bundles : bundles.slice(0, initialLimit);

  return (
    <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
      <div className="mb-3 flex items-baseline justify-between gap-3">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          Recent Audit Samples
        </h2>
        <span className="text-[11px] text-zinc-500">
          {bundles.length === 0
            ? 'no runs yet'
            : `last ${visible.length} of ${bundles.length}`}
        </span>
      </div>

      {bundles.length === 0 ? (
        <p className="text-xs text-zinc-500">
          No verification runs recorded yet. Click <strong>▸ Run Autonomous Deal</strong> above
          to create the first audit bundle.
        </p>
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="text-[10px] uppercase tracking-wider text-zinc-500">
                <tr>
                  <th className="px-2 py-1 text-left">Time</th>
                  <th className="px-2 py-1 text-left">Mode · Verdict</th>
                  <th className="px-2 py-1 text-left">Query</th>
                  <th className="px-2 py-1 text-right">Dur</th>
                  <th className="px-2 py-1 text-right">Audit</th>
                  <th className="px-2 py-1 text-right">Bundle</th>
                </tr>
              </thead>
              <tbody>
                {visible.map((b) => (
                  <Row
                    key={b.decisionId}
                    bundle={b}
                    onView={() => setViewing(b.decisionId)}
                  />
                ))}
              </tbody>
            </table>
          </div>
          {bundles.length > initialLimit && (
            <button
              type="button"
              onClick={() => setExpanded((v) => !v)}
              className="mt-3 w-full rounded-md border border-zinc-700 px-2 py-1.5 text-[11px] text-zinc-300 hover:border-zinc-500"
            >
              {expanded ? '▲ collapse' : `▼ show all ${bundles.length}`}
            </button>
          )}
        </>
      )}

      {viewing && (
        <BundleViewer decisionId={viewing} onClose={() => setViewing(null)} />
      )}
    </section>
  );
}

// ── Row ──────────────────────────────────────────────────────────────────

function Row({ bundle, onView }: { bundle: BundleSummary; onView: () => void }) {
  const verdictBadge = badgeFor(bundle.verdict);
  const auditScore =
    bundle.l2Skipped || bundle.l2Score == null
      ? `${bundle.l1Score} / —`
      : `${bundle.l1Score} / ${bundle.l2Score}`;
  return (
    <tr className="border-t border-zinc-800 hover:bg-zinc-900/60">
      <td className="px-2 py-2 align-top text-zinc-400">
        {formatRelative(bundle.createdAt)}
      </td>
      <td className="px-2 py-2 align-top">
        <span className={`mr-1 rounded px-1 py-0.5 font-mono text-[10px] ${modeTone(bundle.mode)}`}>
          {bundle.mode.toUpperCase()}
        </span>
        <span className={`rounded px-1 py-0.5 font-mono text-[10px] ${verdictBadge.cls}`}>
          {verdictBadge.label}
        </span>
      </td>
      <td className="max-w-[28ch] truncate px-2 py-2 align-top text-zinc-300" title={bundle.query}>
        {bundle.query || <span className="italic text-zinc-600">(no query)</span>}
      </td>
      <td className="px-2 py-2 text-right align-top font-mono text-zinc-400">
        {(bundle.durationMs / 1000).toFixed(1)}s
      </td>
      <td className="px-2 py-2 text-right align-top font-mono text-zinc-400">{auditScore}</td>
      <td className="px-2 py-2 text-right align-top">
        <button
          type="button"
          onClick={onView}
          className="rounded border border-zinc-700 px-2 py-0.5 text-[10px] text-zinc-300 hover:border-zinc-500"
        >
          View
        </button>
      </td>
    </tr>
  );
}

function badgeFor(verdict: string): { label: string; cls: string } {
  switch (verdict) {
    case 'APPROVED_BY_TRUST_GATE':
      return { label: 'APPROVED', cls: 'bg-emerald-500/20 text-emerald-300' };
    case 'BLOCKED_BY_TRUST_GATE':
      return { label: 'BLOCKED', cls: 'bg-red-500/20 text-red-300' };
    case 'ESCALATED_TO_HUMAN':
      return { label: 'ESCALATED', cls: 'bg-amber-500/20 text-amber-300' };
    default:
      return { label: verdict, cls: 'bg-zinc-800 text-zinc-300' };
  }
}

function modeTone(mode: string): string {
  switch (mode) {
    case 'live':      return 'bg-indigo-500/20 text-indigo-300';
    case 'prewarmed': return 'bg-sky-500/20 text-sky-300';
    case 'replay':    return 'bg-zinc-700 text-zinc-300';
    default:          return 'bg-zinc-800 text-zinc-400';
  }
}

function formatRelative(iso: string): string {
  if (!iso) return '—';
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
