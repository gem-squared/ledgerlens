package release_test

import (
	"testing"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
	"github.com/gem-squared/ledgerlens/internal/trustgate/release"
)

func TestCompose_ApprovedHappyPath(t *testing.T) {
	d := release.Compose(
		&auditgate.PCheckResponse{Verdict: "ALLOW", Score: 85},
		&auditgate.OCheckResponse{Verdict: "SUCCESS", Score: 90},
		schemas.PaymentPolicy{SpendCap: 0.01, ClaimGroundedRequired: true},
		0.001,
		release.DefaultThresholds(),
	)
	if d.Verdict != schemas.GateApprovedByTrustGate {
		t.Errorf("verdict=%q want APPROVED_BY_TRUST_GATE (reason=%s)", d.Verdict, d.Reason)
	}
	if !d.PaymentAllowed {
		t.Errorf("PaymentAllowed=false want true")
	}
}

func TestCompose_BlockedByL1Deny(t *testing.T) {
	d := release.Compose(
		&auditgate.PCheckResponse{Verdict: "DENY", Score: 25},
		nil,
		schemas.PaymentPolicy{SpendCap: 0.01},
		0.001,
		release.DefaultThresholds(),
	)
	if d.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("verdict=%q want BLOCKED_BY_TRUST_GATE", d.Verdict)
	}
	if d.PaymentAllowed {
		t.Errorf("PaymentAllowed=true want false on DENY")
	}
}

func TestCompose_BlockedByL1LowScore(t *testing.T) {
	d := release.Compose(
		&auditgate.PCheckResponse{Verdict: "ALLOW", Score: 60}, // below T_L1=70
		&auditgate.OCheckResponse{Verdict: "SUCCESS", Score: 90},
		schemas.PaymentPolicy{SpendCap: 0.01},
		0.001,
		release.DefaultThresholds(),
	)
	if d.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("verdict=%q want BLOCKED (L1 score below threshold)", d.Verdict)
	}
}

func TestCompose_BlockedByL2Failure(t *testing.T) {
	d := release.Compose(
		&auditgate.PCheckResponse{Verdict: "ALLOW", Score: 85},
		&auditgate.OCheckResponse{Verdict: "FAILURE", Score: 40},
		schemas.PaymentPolicy{SpendCap: 0.01},
		0.001,
		release.DefaultThresholds(),
	)
	if d.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("verdict=%q want BLOCKED (L2 FAILURE)", d.Verdict)
	}
}

func TestCompose_BlockedBySpendCap(t *testing.T) {
	d := release.Compose(
		&auditgate.PCheckResponse{Verdict: "ALLOW", Score: 95},
		&auditgate.OCheckResponse{Verdict: "SUCCESS", Score: 95},
		schemas.PaymentPolicy{SpendCap: 0.001},
		0.005, // 5× cap
		release.DefaultThresholds(),
	)
	if d.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("verdict=%q want BLOCKED (spend > cap)", d.Verdict)
	}
}

func TestCompose_NilL1Blocks(t *testing.T) {
	d := release.Compose(nil, nil, schemas.PaymentPolicy{SpendCap: 0.01}, 0.001, release.DefaultThresholds())
	if d.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("verdict=%q want BLOCKED (no L1 response)", d.Verdict)
	}
}
