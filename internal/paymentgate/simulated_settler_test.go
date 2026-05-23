package paymentgate_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"
)

func TestSimulatedSettler_Approved(t *testing.T) {
	s := paymentgate.NewSimulatedSettler()
	r, err := s.Settle(context.Background(), schemas.DecisionPacket{
		DecisionID: "dp_test_a",
		Verdict:    schemas.GateApprovedByTrustGate,
		Reason:     "ok",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if r.RealTransaction || r.PrivateKeysUsed || r.RealFundsUsed {
		t.Errorf("simulation invariants violated: %+v", r)
	}
	if r.Mode != schemas.ModeSimulation || r.Network != "demo-local" || r.Asset != "USDC-demo" {
		t.Errorf("mode/network/asset wrong: %+v", r)
	}
	if r.Status != schemas.SimulatedSettled {
		t.Errorf("status=%q want SIMULATED_SETTLED", r.Status)
	}
	if r.AmountUSDC <= 0 {
		t.Errorf("AmountUSDC=%v want >0 on APPROVED", r.AmountUSDC)
	}
	if !strings.HasPrefix(r.SettlementID, "sim_x402_") {
		t.Errorf("SettlementID=%q want sim_x402_<hex>", r.SettlementID)
	}
}

func TestSimulatedSettler_Blocked(t *testing.T) {
	s := paymentgate.NewSimulatedSettler()
	r, err := s.Settle(context.Background(), schemas.DecisionPacket{
		DecisionID: "dp_test_b",
		Verdict:    schemas.GateBlockedByTrustGate,
		Reason:     "L1 DENY",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if r.SettlementID != "" {
		t.Errorf("SettlementID=%q want empty on BLOCKED", r.SettlementID)
	}
	if r.Status != schemas.BlockedByTrustGate {
		t.Errorf("status=%q want BLOCKED_BY_TRUST_GATE", r.Status)
	}
	if r.AmountUSDC != 0 {
		t.Errorf("AmountUSDC=%v want 0 on BLOCKED", r.AmountUSDC)
	}
	if r.RealTransaction {
		t.Errorf("RealTransaction=true on BLOCKED — must always be false")
	}
}

func TestSimulatedSettler_Escalated(t *testing.T) {
	s := paymentgate.NewSimulatedSettler()
	r, err := s.Settle(context.Background(), schemas.DecisionPacket{
		DecisionID: "dp_test_c",
		Verdict:    schemas.GateEscalatedToHuman,
		Reason:     "intermediate score band",
	})
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if r.Status != schemas.EscalatedToHuman {
		t.Errorf("status=%q want ESCALATED_TO_HUMAN", r.Status)
	}
	if r.SettlementID != "" {
		t.Errorf("SettlementID should be empty on ESCALATE, got %q", r.SettlementID)
	}
}
