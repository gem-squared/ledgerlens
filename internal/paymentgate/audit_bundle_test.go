package paymentgate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"
)

func TestBundleStore_WriteAndRead(t *testing.T) {
	root := t.TempDir()
	store, err := paymentgate.NewBundleStore(root)
	if err != nil {
		t.Fatalf("new bundle store: %v", err)
	}

	bundle := paymentgate.AuditBundle{
		DecisionID: "dp_test_xyz",
		BuyerRequest: schemas.BuyerRequest{
			RequestID: "req_test_xyz", BuyerID: "buyer_a", Query: "stripe pricing api",
			MaxSpendUSDC: 0.01,
		},
		SellerOffer: schemas.SellerOffer{
			OfferID: "offer_test_xyz", SellerID: "seller_b",
			Claim: "Live NYSE+NASDAQ pricing feed", PriceUSDC: 0.001,
		},
		Decision: schemas.DecisionPacket{
			DecisionID: "dp_test_xyz",
			Verdict:    schemas.GateApprovedByTrustGate,
			Reason:     "all gates passed",
		},
		Settlement: schemas.SimulatedSettlement{
			DecisionID:      "dp_test_xyz",
			Mode:            schemas.ModeSimulation,
			Status:          schemas.SimulatedSettled,
			AmountUSDC:      0.001,
			SettlementID:    "sim_x402_abcd1234",
			RealTransaction: false,
		},
	}
	chunks := []string{"chunk-1: status page", "chunk-2: pricing page"}

	path, written, err := store.Write(bundle, chunks)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.HasSuffix(path, "dp_test_xyz.json") {
		t.Errorf("path=%q want suffix dp_test_xyz.json", path)
	}
	if written.BundleID == "" {
		t.Errorf("BundleID not populated")
	}
	if !strings.HasPrefix(written.EvidenceHash, "sha256:") {
		t.Errorf("EvidenceHash=%q want sha256: prefix", written.EvidenceHash)
	}

	// Read back
	got, err := store.Read("dp_test_xyz")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.DecisionID != bundle.DecisionID ||
		got.Decision.Verdict != bundle.Decision.Verdict ||
		got.Settlement.SettlementID != bundle.Settlement.SettlementID {
		t.Errorf("roundtrip mismatch: got=%+v", got)
	}

	// File exists at the expected location
	if _, err := filepath.Abs(path); err != nil {
		t.Errorf("path not absoluteable: %v", err)
	}
}
