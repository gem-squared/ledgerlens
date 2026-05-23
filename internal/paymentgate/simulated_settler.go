package paymentgate

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// SimulatedSettler is the ONE Settler implementation shipped in the hackathon
// repo. It generates deterministic-shape simulated settlement receipts WITHOUT
// any chain dependency. Every receipt carries:
//
//	mode:             "simulation"
//	network:          "demo-local"
//	asset:            "USDC-demo"
//	real_transaction: false
//	private_keys_used: false
//	real_funds_used:  false
//
// On BLOCKED / ESCALATED decisions, SettlementID is empty (null in JSON),
// AmountUSDC = 0, and Status reflects the decision verdict.
type SimulatedSettler struct {
	// AmountUSDC is the per-decision simulated transfer amount used on
	// APPROVED outcomes. Default 0.001.
	AmountUSDC float64
}

// NewSimulatedSettler returns a settler with the default APPROVED amount.
func NewSimulatedSettler() *SimulatedSettler {
	return &SimulatedSettler{AmountUSDC: 0.001}
}

// Settle implements schemas.Settler. It NEVER fails for the in-memory case;
// errors are reserved for a future real settler implementation.
func (s *SimulatedSettler) Settle(ctx context.Context, decision schemas.DecisionPacket) (schemas.SimulatedSettlement, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	base := schemas.SimulatedSettlement{
		DecisionID:      decision.DecisionID,
		Mode:            schemas.ModeSimulation,
		Network:         "demo-local",
		Asset:           "USDC-demo",
		RealTransaction: false,
		PrivateKeysUsed: false,
		RealFundsUsed:   false,
		Ts:              now,
		Reason:          decision.Reason,
	}

	switch decision.Verdict {
	case schemas.GateApprovedByTrustGate:
		base.SettlementID = "sim_x402_" + randHex(8)
		base.Status = schemas.SimulatedSettled
		base.AmountUSDC = s.amount()
		if decision.Reason == "" {
			base.Reason = "L3 Trust Gate approved grounded claim"
		}
		return base, nil

	case schemas.GateBlockedByTrustGate:
		base.SettlementID = "" // null in JSON via `omitempty`
		base.Status = schemas.BlockedByTrustGate
		base.AmountUSDC = 0
		return base, nil

	case schemas.GateEscalatedToHuman:
		base.SettlementID = ""
		base.Status = schemas.EscalatedToHuman
		base.AmountUSDC = 0
		return base, nil

	default:
		base.SettlementID = ""
		base.Status = schemas.BlockedByTrustGate
		base.AmountUSDC = 0
		base.Reason = "unknown verdict; defaulting to BLOCKED"
		return base, nil
	}
}

func (s *SimulatedSettler) amount() float64 {
	if s.AmountUSDC <= 0 {
		return 0.001
	}
	return s.AmountUSDC
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "deadbeef"
	}
	return hex.EncodeToString(b)
}
