import type { CaseListItem, RunResult } from './types';

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
