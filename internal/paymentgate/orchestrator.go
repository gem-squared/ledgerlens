package paymentgate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
	"github.com/gem-squared/ledgerlens/internal/trustgate/evidence"
	"github.com/gem-squared/ledgerlens/internal/trustgate/memory"
	"github.com/gem-squared/ledgerlens/internal/trustgate/release"
)

// Orchestrator composes the full LedgerLens pipeline for a single
// (BuyerRequest, SellerOffer, []EvidenceReceipt) tuple:
//
//	① WrapReceipts → evidence chunks
//	② L1 P-check (gem2-tpmn-checker)
//	③ F: compose draft DecisionPacket
//	④ L2 O-check (gem2-tpmn-checker) — skipped iff L1 already DENIED
//	⑤ L3 Release (composite verdict)
//	⑥ SimulatedSettler.Settle
//	⑦ Audit bundle write + Memory store insert
type Orchestrator struct {
	Audit      *auditgate.Client
	Settler    schemas.Settler
	Mem        *memory.Store     // may be nil for tests that don't care
	Bundles    *BundleStore
	LLMKeys    auditgate.LLMKeySet
	Thresholds release.Thresholds
}

// StepEvent is emitted by RunWithEvents at each major pipeline boundary.
// Used by the SSE streaming endpoint to give judges visible progress
// during the 20-45s LIVE verification wait.
type StepEvent struct {
	Step   string `json:"step"`             // "l1" | "l2" | "l3" | "settle" (+ wrapper-level steps from deals handler)
	Status string `json:"status"`           // "running" | "passed" | "blocked" | "failed" | "skipped"
	Label  string `json:"label"`            // human-readable action verb (e.g. "GEM² L1 P-check auditing seller claim…")
	Detail any    `json:"detail,omitempty"` // step-specific payload (score, result_id, etc.)
	Ts     string `json:"ts"`               // ISO8601 UTC
}

// EventEmitter is the SSE-friendly callback shape. Pass nil to RunWithEvents
// to disable emission (matches the pre-Slice-2 Run signature).
type EventEmitter func(StepEvent)

// Run is the backward-compatible entry point for Case A/B replay and any
// caller that doesn't need step-level visibility. Internally delegates to
// RunWithEvents with a nil emitter.
func (o *Orchestrator) Run(
	ctx context.Context,
	buyer schemas.BuyerRequest,
	offer schemas.SellerOffer,
	receipts []schemas.EvidenceReceipt,
) (schemas.DecisionPacket, schemas.SimulatedSettlement, string, error) {
	return o.RunWithEvents(ctx, buyer, offer, receipts, nil)
}

// emitNow is a small helper that fills timestamp + invokes the emitter
// only if non-nil. Kept on the package level so deals handler can reuse.
func emitNow(emit EventEmitter, step, status, label string, detail any) {
	if emit == nil {
		return
	}
	emit(StepEvent{
		Step: step, Status: status, Label: label, Detail: detail,
		Ts: time.Now().UTC().Format(time.RFC3339),
	})
}

// RunWithEvents executes the pipeline and emits a StepEvent at each major
// boundary (l1 running/passed/blocked, l2 running/passed/blocked, l3
// running/decision, settle running/result). Total step events on a happy
// path: 8. On L1-deny path (L2 skipped): 6.
func (o *Orchestrator) RunWithEvents(
	ctx context.Context,
	buyer schemas.BuyerRequest,
	offer schemas.SellerOffer,
	receipts []schemas.EvidenceReceipt,
	emit EventEmitter,
) (schemas.DecisionPacket, schemas.SimulatedSettlement, string, error) {

	started := time.Now().UTC().Format(time.RFC3339)

	// ① Wrap evidence
	chunks, _ := evidence.WrapReceipts(receipts)

	// ② L1 P-check
	emitNow(emit, "l1", "running", "GEM² L1 P-check auditing seller claim against evidence…", nil)
	pReq := auditgate.PCheckRequest{
		I:              describeOffer(offer, buyer),
		A:              json.RawMessage(`"SellerOffer{offerId, sellerId, claim, priceUSDC, sourceUrl} for BuyerRequest{query, maxSpendUSDC, policy}"`),
		P:              o.preconditions(buyer, offer),
		T:              o.Thresholds.L1,
		Evidence:       chunks,
		SessionContext: "ledgerlens-decide-" + buyer.RequestID,
		Provider:       "claude",
	}
	pResp, perr := o.Audit.PCheck(ctx, pReq, o.LLMKeys, "")
	if perr != nil {
		emitNow(emit, "l1", "failed", "L1 P-check call failed: "+perr.Error(), nil)
		return schemas.DecisionPacket{}, schemas.SimulatedSettlement{}, "", fmt.Errorf("orchestrator: PCheck: %w", perr)
	}
	{
		l1Status := "passed"
		if pResp.Verdict != "ALLOW" || pResp.Score < o.Thresholds.L1 {
			l1Status = "blocked"
		}
		emitNow(emit, "l1", l1Status,
			fmt.Sprintf("L1 P-check: %s (score %d/100)", pResp.Verdict, pResp.Score),
			map[string]any{"verdict": pResp.Verdict, "score": pResp.Score, "result_id": pResp.Meta.ResultID},
		)
	}

	// ③ F: compose draft decision (claim assessments derived from L1 reasons)
	draft := schemas.DecisionPacket{
		DecisionID:       "dp_" + randHex(8),
		RequestID:        buyer.RequestID,
		OfferID:          offer.OfferID,
		ClaimAssessments: auditgate.ToClaimAssessments(pResp.Reasons, pResp.Score),
		L1ResultID:       pResp.Meta.ResultID,
	}

	var oResp *auditgate.OCheckResponse

	// ④ L2 O-check — only run if L1 ALLOW + above threshold; otherwise skip
	if pResp.Verdict == "ALLOW" && pResp.Score >= o.Thresholds.L1 {
		// We need a draft to validate. Construct a provisionally-APPROVED
		// packet for L2 to evaluate. L3 makes the final call.
		draft.Verdict = schemas.GateApprovedByTrustGate
		draft.Reason = "L1 ALLOW; awaiting L2 + L3"
		draft.PaymentAllowed = true

		emitNow(emit, "l2", "running", "GEM² L2 O-check auditing decision packet postconditions…", nil)
		draftJSON, _ := json.Marshal(draft)
		oReq := auditgate.OCheckRequest{
			O:              string(draftJSON),
			B:              json.RawMessage(`"DecisionPacket{decisionId, requestId, offerId, verdict, reason, claimAssessments, paymentAllowed, auditBundleRef}"`),
			P:              o.postconditions(buyer, offer),
			T:              o.Thresholds.L2,
			Evidence:       chunks,
			SessionContext: "ledgerlens-decide-" + buyer.RequestID,
			Provider:       "claude",
		}
		var oerr error
		oResp, oerr = o.Audit.OCheck(ctx, oReq, o.LLMKeys, "")
		if oerr != nil {
			emitNow(emit, "l2", "failed", "L2 O-check call failed: "+oerr.Error(), nil)
			return schemas.DecisionPacket{}, schemas.SimulatedSettlement{}, "", fmt.Errorf("orchestrator: OCheck: %w", oerr)
		}
		draft.L2ResultID = oResp.Meta.ResultID
		l2Status := "passed"
		if oResp.Verdict != "SUCCESS" || oResp.Score < o.Thresholds.L2 {
			l2Status = "blocked"
		}
		emitNow(emit, "l2", l2Status,
			fmt.Sprintf("L2 O-check: %s (score %d/100)", oResp.Verdict, oResp.Score),
			map[string]any{"verdict": oResp.Verdict, "score": oResp.Score, "result_id": oResp.Meta.ResultID},
		)
	} else {
		emitNow(emit, "l2", "skipped", "L2 O-check skipped — L1 already denied", nil)
	}

	// ⑤ L3 composite
	emitNow(emit, "l3", "running", "L3 Trust Gate composing final verdict from L1 + L2 + policy…", nil)
	l3 := release.Compose(pResp, oResp, buyer.Policy, offer.PriceUSDC, o.Thresholds)
	draft.Verdict = l3.Verdict
	draft.Reason = l3.Reason
	draft.PaymentAllowed = l3.PaymentAllowed
	{
		l3Status := "blocked"
		if l3.Verdict == schemas.GateApprovedByTrustGate {
			l3Status = "passed"
		} else if l3.Verdict == schemas.GateEscalatedToHuman {
			l3Status = "escalated"
		}
		emitNow(emit, "l3", l3Status,
			fmt.Sprintf("L3 Trust Gate: %s — %s", l3.Verdict, l3.Reason),
			map[string]any{"verdict": l3.Verdict, "reason": l3.Reason, "paymentAllowed": l3.PaymentAllowed},
		)
	}

	// ⑥ Settle (simulated)
	emitNow(emit, "settle", "running", "x402-style settlement processing…", nil)
	settlement, serr := o.Settler.Settle(ctx, draft)
	if serr != nil {
		emitNow(emit, "settle", "failed", "Settlement failed: "+serr.Error(), nil)
		return draft, settlement, "", fmt.Errorf("orchestrator: settle: %w", serr)
	}
	{
		settleStatus := "blocked"
		if settlement.Status == schemas.SimulatedSettled {
			settleStatus = "settled"
		}
		emitNow(emit, "settle", settleStatus,
			fmt.Sprintf("x402 settlement: %s · settlement_id=%s · real_transaction=false",
				settlement.Status,
				orDash(settlement.SettlementID)),
			map[string]any{
				"status":           settlement.Status,
				"settlementId":     settlement.SettlementID,
				"amountUSDC":       settlement.AmountUSDC,
				"realTransaction":  settlement.RealTransaction,
			},
		)
	}

	// ⑦ Audit bundle + memory mirror
	bundle := AuditBundle{
		DecisionID:       draft.DecisionID,
		BuyerRequest:     buyer,
		SellerOffer:      offer,
		EvidenceReceipts: receipts,
		L1:               pResp,
		L2:               oResp,
		ClaimAssessments: draft.ClaimAssessments,
		Decision:         draft,
		Settlement:       settlement,
		Timestamps:       BundleTimestamps{StartedAt: started},
	}
	bundlePath, bundle, berr := o.Bundles.Write(bundle, chunks)
	if berr != nil {
		return draft, settlement, "", fmt.Errorf("orchestrator: bundle write: %w", berr)
	}
	draft.AuditBundleRef = bundlePath
	bundle.Decision = draft
	// Re-write so the bundle file also carries the final AuditBundleRef.
	_, _, _ = o.Bundles.Write(bundle, chunks)

	if o.Mem != nil {
		reasonsJSON, _ := json.Marshal(pResp.Reasons)
		_ = o.Mem.Insert(memory.Record{
			ID:          pResp.Meta.ResultID,
			GateType:    "P_GATE",
			Verdict:     pResp.Verdict,
			Score:       pResp.Score,
			Threshold:   o.Thresholds.L1,
			Input:       pReq.I,
			P:           pReq.P,
			Evidence:    chunks,
			Provider:    pResp.Meta.Usage.Provider,
			DurationMs:  pResp.Meta.DurationMs,
			ReasonsJSON: string(reasonsJSON),
		})
		if oResp != nil {
			oReasons, _ := json.Marshal(oResp.Reasons)
			_ = o.Mem.Insert(memory.Record{
				ID:          oResp.Meta.ResultID,
				GateType:    "O_GATE",
				Verdict:     oResp.Verdict,
				Score:       oResp.Score,
				Threshold:   o.Thresholds.L2,
				Input:       draft.DecisionID,
				P:           o.postconditions(buyer, offer),
				Evidence:    chunks,
				Provider:    oResp.Meta.Usage.Provider,
				DurationMs:  oResp.Meta.DurationMs,
				ReasonsJSON: string(oReasons),
			})
		}
	}

	return draft, settlement, bundlePath, nil
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// describeOffer renders the seller offer + buyer request into a single
// natural-language string for the P-check `i` field.
func describeOffer(o schemas.SellerOffer, b schemas.BuyerRequest) string {
	parts := []string{
		fmt.Sprintf("SellerOffer{offerId=%q, sellerId=%q, claim=%q, priceUSDC=%v",
			o.OfferID, o.SellerID, o.Claim, o.PriceUSDC),
	}
	if o.SourceURL != "" {
		parts = append(parts, fmt.Sprintf("sourceUrl=%q", o.SourceURL))
	}
	parts = append(parts, "} for BuyerRequest{query="+fmt.Sprintf("%q", b.Query)+
		fmt.Sprintf(", maxSpendUSDC=%v, policy.spendCap=%v}", b.MaxSpendUSDC, b.Policy.SpendCap))
	return strings.Join(parts, ", ")
}

// preconditions returns the P[] rules for L1 P-check.
func (o *Orchestrator) preconditions(b schemas.BuyerRequest, offer schemas.SellerOffer) []string {
	rules := []string{
		"the offer's claim must be directly supported by at least one evidence chunk",
		fmt.Sprintf("offer.priceUSDC must be ≤ buyer.maxSpendUSDC (cap = %v)", b.MaxSpendUSDC),
		"no SPT violations (S→T, L→G, Δe→∫de) — flag any thin-evidence overclaim",
	}
	if b.Policy.PublicOnly {
		rules = append(rules, "the evidence sources must all be public web URLs (no login-required, no private data)")
	}
	if b.Policy.ClaimGroundedRequired {
		rules = append(rules, "every material claim in the offer must carry [EEF-⊢ grounded] or [EEF-⊨ inferred] — no bare ⊬ extrapolations")
	}
	return rules
}

// postconditions returns the P[] rules for L2 O-check on the draft packet.
func (o *Orchestrator) postconditions(b schemas.BuyerRequest, offer schemas.SellerOffer) []string {
	return []string{
		"decisionPacket.verdict must be one of {APPROVED_BY_TRUST_GATE, BLOCKED_BY_TRUST_GATE, ESCALATED_TO_HUMAN}",
		"decisionPacket.paymentAllowed = true iff decisionPacket.verdict = APPROVED_BY_TRUST_GATE",
		"decisionPacket.claimAssessments must reflect the L1 reasons (one assessment per RULE-N)",
		"if verdict = APPROVED_BY_TRUST_GATE then all claimAssessments must be grounded or inferred — no ⊬-without-basis",
	}
}
