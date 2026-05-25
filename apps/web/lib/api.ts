import type { BundleSummary, CaseListItem, DealRunResult, Stats } from './types';

export async function listCases(): Promise<CaseListItem[]> {
  const r = await fetch('/api/v1/cases', { cache: 'no-store' });
  if (!r.ok) throw new Error(`listCases: HTTP ${r.status}`);
  const json = await r.json();
  return json.cases ?? [];
}

// Case A/B replay shares the LIVE response shape (DealRunResult). The Go
// handler synthesizes finalReport via agent.ComposeFromRequest.
export async function runCase(id: string): Promise<DealRunResult> {
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

// Bundle GET also returns a DealRunResult-shaped envelope (finalReport
// synthesized on read, mode/durationMs derived from bundle metadata).
export async function getAuditBundle(decisionId: string): Promise<DealRunResult> {
  const r = await fetch(`/api/v1/audit-bundles/${decisionId}`, { cache: 'no-store' });
  if (!r.ok) throw new Error(`getAuditBundle(${decisionId}): HTTP ${r.status}`);
  return r.json();
}
