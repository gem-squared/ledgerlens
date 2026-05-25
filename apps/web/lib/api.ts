import type { BundleSummary, CaseListItem, RunResult, Stats } from './types';

export async function listCases(): Promise<CaseListItem[]> {
  const r = await fetch('/api/v1/cases', { cache: 'no-store' });
  if (!r.ok) throw new Error(`listCases: HTTP ${r.status}`);
  const json = await r.json();
  return json.cases ?? [];
}

export async function runCase(id: string): Promise<RunResult> {
  const r = await fetch(`/api/v1/cases/${id}/run`, { method: 'POST', cache: 'no-store' });
  if (!r.ok) {
    const body = await r.text();
    throw new Error(`runCase(${id}): HTTP ${r.status} — ${body}`);
  }
  return r.json();
}

// ─── Slice 3 ───────────────────────────────────────────────────────────────

export async function getStats(): Promise<Stats> {
  const r = await fetch('/api/v1/stats', { cache: 'no-store' });
  if (!r.ok) throw new Error(`getStats: HTTP ${r.status}`);
  return r.json();
}

export async function listAuditBundles(): Promise<BundleSummary[]> {
  const r = await fetch('/api/v1/audit-bundles', { cache: 'no-store' });
  if (!r.ok) throw new Error(`listAuditBundles: HTTP ${r.status}`);
  const json = await r.json();
  return json.bundles ?? [];
}

export async function getAuditBundle(decisionId: string): Promise<unknown> {
  const r = await fetch(`/api/v1/audit-bundles/${decisionId}`, { cache: 'no-store' });
  if (!r.ok) throw new Error(`getAuditBundle(${decisionId}): HTTP ${r.status}`);
  return r.json();
}
