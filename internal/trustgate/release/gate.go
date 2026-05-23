// Package release implements the L3 composite verdict — the local
// governance gate that combines L1 P-check + L2 O-check + policy into a
// single APPROVED / BLOCKED / ESCALATE decision before settlement.
package release

import (
	"fmt"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
)

// Thresholds mirror the documented defaults from AUDIT_GATE_API examples.
type Thresholds struct {
	L1 int // P-check score floor — typical 70
	L2 int // O-check score floor — typical 75
}

// DefaultThresholds returns the canonical T_L1=70, T_L2=75 used by the
// AUDIT_GATE_API examples and by LedgerLens's strategy doc §5.2.
func DefaultThresholds() Thresholds {
	return Thresholds{L1: 70, L2: 75}
}

// Decision is the L3 outcome: a verdict + a single-line reason chain.
type Decision struct {
	Verdict        schemas.GateVerdict
	Reason         string
	PaymentAllowed bool
}

// Compose evaluates the composite gate. Either gate denying or scoring
// below threshold or a policy breach blocks the payment; otherwise the
// gate approves at a binary level. (ESCALATED_TO_HUMAN intermediate band
// is reserved for Unit 5's UI surface; this MVP collapses to APPROVED/BLOCKED.)
func Compose(
	l1 *auditgate.PCheckResponse,
	l2 *auditgate.OCheckResponse,
	policy schemas.PaymentPolicy,
	proposedSpend float64,
	t Thresholds,
) Decision {
	// L1 P-check must say ALLOW within threshold.
	if l1 == nil {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         "L1 P-check missing (gate unavailable and no replay)",
			PaymentAllowed: false,
		}
	}
	if l1.Verdict != "ALLOW" {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         fmt.Sprintf("L1 P-check verdict=%s score=%d — claim not grounded", l1.Verdict, l1.Score),
			PaymentAllowed: false,
		}
	}
	if l1.Score < t.L1 {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         fmt.Sprintf("L1 P-check score %d below threshold %d", l1.Score, t.L1),
			PaymentAllowed: false,
		}
	}

	// L2 O-check must say SUCCESS within threshold.
	if l2 == nil {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         "L2 O-check missing — cannot verify proposed decision postconditions",
			PaymentAllowed: false,
		}
	}
	if l2.Verdict != "SUCCESS" {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         fmt.Sprintf("L2 O-check verdict=%s score=%d — postconditions failed", l2.Verdict, l2.Score),
			PaymentAllowed: false,
		}
	}
	if l2.Score < t.L2 {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         fmt.Sprintf("L2 O-check score %d below threshold %d", l2.Score, t.L2),
			PaymentAllowed: false,
		}
	}

	// Policy: spend cap.
	if policy.SpendCap > 0 && proposedSpend > policy.SpendCap {
		return Decision{
			Verdict:        schemas.GateBlockedByTrustGate,
			Reason:         fmt.Sprintf("Proposed spend %.6f USDC exceeds policy cap %.6f", proposedSpend, policy.SpendCap),
			PaymentAllowed: false,
		}
	}

	return Decision{
		Verdict:        schemas.GateApprovedByTrustGate,
		Reason:         fmt.Sprintf("L1 ALLOW (%d) ∧ L2 SUCCESS (%d) ∧ spend %.6f within cap %.6f", l1.Score, l2.Score, proposedSpend, policy.SpendCap),
		PaymentAllowed: true,
	}
}
