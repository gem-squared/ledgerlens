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

export interface RunResult {
  case: CaseListItem;
  buyerRequest: BuyerRequest;
  sellerOffer: SellerOffer;
  evidenceReceipts: EvidenceReceipt[];
  decision: DecisionPacket;
  settlement: SimulatedSettlement;
  l1?: GateResponse;
  l2?: GateResponse;
  bundlePath: string;
  durationMs: number;
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
