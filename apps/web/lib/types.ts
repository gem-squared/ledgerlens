// Re-exports of Go-generated TS types so we have a stable @/lib/types entry
// regardless of where `tygo generate` lands the file. Run `make schemas`
// to regenerate ../../packages/contracts-ts/types.ts from Go.
export * from '../../../packages/contracts-ts/types';

// API response shapes — these live ONLY on the Go side (internal/api/handlers.go)
// so we re-declare them here to keep the UI strongly typed.
import type {
  BuyerRequest,
  SellerOffer,
  EvidenceReceipt,
  DecisionPacket,
  SimulatedSettlement,
  ClaimAssessment,
} from '../../../packages/contracts-ts/types';

export interface CaseListItem {
  id: string;
  title: string;
  description: string;
}

export interface GateResponse {
  verdict: string;
  score: number;
  reasons: string[];
  meta: {
    result_id: string;
    duration_ms: number;
    usage?: {
      provider?: string;
      model?: string;
      estimated_cost_usd?: number;
    };
  };
}

// Re-export the named structs for convenience.
export type {
  BuyerRequest,
  SellerOffer,
  EvidenceReceipt,
  DecisionPacket,
  SimulatedSettlement,
  ClaimAssessment,
};

// ─── Slice 1/2: Judge Request Mode types ──────────────────────────────────

export interface BuyerIntent {
  offDomain: boolean;
  politeReject?: string;
  dataNeed?: string;
  freshness?: string;
  maxSpendUSDC?: number;
  policyRequirements?: string[];
  searchTerms?: string[];
}

export interface FinalReport {
  headline: string;
  request: string;
  result: string;
  reason: string;
  evidenceSummary: string;
  paymentSummary: string;
  auditBundleRef: string;
  tagline: string;
}

export interface DealRunResult {
  mode: string;
  judgeRequest: string;
  buyerIntent: BuyerIntent;
  agentNarrative: string[];
  evidenceReceipts: EvidenceReceipt[];
  sellerOffer: SellerOffer;
  decision: DecisionPacket;
  settlement: SimulatedSettlement;
  l1?: GateResponse;
  l2?: GateResponse;
  bundlePath: string;
  finalReport: FinalReport;
  durationMs: number;
}

// ─── Slice 2: SSE step events ──────────────────────────────────────────────

export type StepStatus =
  | 'idle'
  | 'running'
  | 'passed'
  | 'blocked'
  | 'failed'
  | 'skipped'
  | 'settled'
  | 'rejected'
  | 'escalated';

export interface StepEvent {
  step: string;
  status: StepStatus;
  label: string;
  detail?: unknown;
  ts: string;
}

export type RunMode = 'live' | 'prewarmed' | 'replay';

// ─── Slice 3: Verification Infrastructure Dashboard ────────────────────────

export interface Modes {
  live: number;
  preWarmed: number;
  replay: number;
  unknown: number;
}

export interface Stats {
  dealsAudited: number;
  approved: number;
  blocked: number;
  escalatedToHuman: number;
  avgAuditScore: number;
  avgVerificationLatencyMs: number;
  simulatedSpendPreventedUSDC: number;
  brightDataReceipts: number;
  auditBundlesExported: number;
  sampleSize: number;
  lastUpdatedAt?: string;
  auditGateUrl: string;
  auditGateAvgLatencyMs: number;
  modesBreakdown: Modes;
}

export interface BundleSummary {
  decisionId: string;
  bundleId: string;
  verdict: string;
  mode: string;
  query: string;
  durationMs: number;
  l1Score: number;
  l2Score?: number;
  l2Skipped: boolean;
  paymentAllowed: boolean;
  settlementStatus?: string;
  settlementId?: string;
  amountUSDC: number;
  createdAt: string;
  evidenceCount: number;
}
