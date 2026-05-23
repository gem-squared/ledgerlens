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

// Run executes the pipeline. Returns the final DecisionPacket, the
// simulated settlement record, and the path to the persisted audit bundle.
func (o *Orchestrator) Run(
	ctx context.Context,
	buyer schemas.BuyerRequest,
	offer schemas.SellerOffer,
	receipts []schemas.EvidenceReceipt,
) (schemas.DecisionPacket, schemas.SimulatedSettlement, string, error) {

	started := time.Now().UTC().Format(time.RFC3339)

	// ① Wrap evidence
	chunks, _ := evidence.WrapReceipts(receipts)

	// ② L1 P-check
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
		return schemas.DecisionPacket{}, schemas.SimulatedSettlement{}, "", fmt.Errorf("orchestrator: PCheck: %w", perr)
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
			return schemas.DecisionPacket{}, schemas.SimulatedSettlement{}, "", fmt.Errorf("orchestrator: OCheck: %w", oerr)
		}
		draft.L2ResultID = oResp.Meta.ResultID
	}

	// ⑤ L3 composite
	l3 := release.Compose(pResp, oResp, buyer.Policy, offer.PriceUSDC, o.Thresholds)
	draft.Verdict = l3.Verdict
	draft.Reason = l3.Reason
	draft.PaymentAllowed = l3.PaymentAllowed

	// ⑥ Settle (simulated)
	settlement, serr := o.Settler.Settle(ctx, draft)
	if serr != nil {
		return draft, settlement, "", fmt.Errorf("orchestrator: settle: %w", serr)
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
