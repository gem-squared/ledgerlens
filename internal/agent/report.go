package agent

import (
	"fmt"
	"strings"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// FinalReport is the judge-readable summary that appears at the top of the
// Judge Request Mode result. Template-driven (no LLM call) — fast and
// deterministic, no risk of off-domain drift on the closing summary.
type FinalReport struct {
	Headline        string `json:"headline"`        // "Payment Approved" | "Payment Blocked" | "Escalated to Human"
	Request         string `json:"request"`         // the judge's input query
	Result          string `json:"result"`          // one-line result sentence
	Reason          string `json:"reason"`          // why
	EvidenceSummary string `json:"evidenceSummary"` // what Bright Data found
	PaymentSummary  string `json:"paymentSummary"`  // what x402 simulation issued (or didn't)
	AuditBundleRef  string `json:"auditBundleRef"`  // path to artifacts/audit_bundles/<id>.json
	Tagline         string `json:"tagline"`         // canonical: "Fast agents are dangerous..."
}

const taglineCanonical = "Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits."

// Compose builds the FinalReport from the orchestrator's outputs.
func Compose(
	judgeRequest string,
	intent BuyerIntent,
	offer schemas.SellerOffer,
	receipts []schemas.EvidenceReceipt,
	decision schemas.DecisionPacket,
	settlement schemas.SimulatedSettlement,
) FinalReport {
	r := FinalReport{
		Request:        judgeRequest,
		AuditBundleRef: decision.AuditBundleRef,
		Tagline:        taglineCanonical,
	}

	switch decision.Verdict {
	case schemas.GateApprovedByTrustGate:
		r.Headline = "Payment Approved"
		r.Result = fmt.Sprintf("LedgerLens approved a simulated payment of %v USDC-demo to %s.",
			settlement.AmountUSDC, offer.SellerID)
		r.Reason = decision.Reason
		r.PaymentSummary = fmt.Sprintf(
			"Simulated settlement issued · settlementId=%s · mode=%s · network=%s · real_transaction=%v",
			settlement.SettlementID, settlement.Mode, settlement.Network, settlement.RealTransaction,
		)
	case schemas.GateBlockedByTrustGate:
		r.Headline = "Payment Blocked"
		r.Result = fmt.Sprintf("LedgerLens BLOCKED the buyer agent from paying %s.", offer.SellerID)
		// Long-form reason that names the gap concretely (per David's directive).
		r.Reason = fmt.Sprintf(
			"The buyer asked for a provider under $%v/query, but the fetched public evidence did not prove the seller's per-query price. LedgerLens blocked the payment. Trust gate verdict: %s",
			intent.MaxSpendUSDC, decision.Reason,
		)
		r.PaymentSummary = "No settlement issued. The buyer agent's funds are intact. A fast agent would have paid. LedgerLens waited."
	case schemas.GateEscalatedToHuman:
		r.Headline = "Escalated to Human"
		r.Result = "Verdict in intermediate band — operator decision required before settlement."
		r.Reason = decision.Reason
		r.PaymentSummary = "Settlement held pending human approval."
	default:
		r.Headline = "Unknown"
		r.Result = "Unknown verdict; treating as blocked for safety."
		r.Reason = decision.Reason
		r.PaymentSummary = "No settlement issued."
	}

	r.EvidenceSummary = summarizeEvidence(receipts, offer)
	return r
}

func summarizeEvidence(receipts []schemas.EvidenceReceipt, offer schemas.SellerOffer) string {
	if len(receipts) == 0 {
		return "No evidence collected."
	}
	products := map[string]int{}
	for _, r := range receipts {
		products[r.BrightDataProduct]++
	}
	parts := []string{fmt.Sprintf("Bright Data collected %d public-web evidence receipt(s)", len(receipts))}
	if len(products) > 0 {
		products_list := []string{}
		for p, n := range products {
			if n > 1 {
				products_list = append(products_list, fmt.Sprintf("%dx %s", n, p))
			} else {
				products_list = append(products_list, p)
			}
		}
		parts = append(parts, "via "+strings.Join(products_list, " + "))
	}
	if offer.SourceURL != "" {
		parts = append(parts, fmt.Sprintf("primary source: %s", offer.SourceURL))
	}
	return strings.Join(parts, " · ") + "."
}
