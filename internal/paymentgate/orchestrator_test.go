package paymentgate_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
	"github.com/gem-squared/ledgerlens/internal/trustgate/memory"
	"github.com/gem-squared/ledgerlens/internal/trustgate/release"
	_ "modernc.org/sqlite"
)

// These are LIVE end-to-end tests of the full LedgerLens pipeline (Evidence
// → L1 → F → L2 → L3 → Settle → AuditBundle). They consume 2 audit-gate
// calls per case (~$0.04 per case). Skipped if env not set.

func requireEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if os.Getenv(k) == "" {
			t.Skipf("env %s not set — skipping live e2e test", k)
		}
	}
}

func buildOrchestrator(t *testing.T) (*paymentgate.Orchestrator, string) {
	t.Helper()
	root := t.TempDir()

	bundles, err := paymentgate.NewBundleStore(root)
	if err != nil {
		t.Fatalf("bundles: %v", err)
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	mem, err := memory.NewStore(db)
	if err != nil {
		t.Fatalf("memory: %v", err)
	}
	return &paymentgate.Orchestrator{
		Audit:      auditgate.NewClient(os.Getenv("GEM2_TPMN_CHECKER_BASE_URL"), os.Getenv("GEM2_API_KEY")),
		Settler:    paymentgate.NewSimulatedSettler(),
		Mem:        mem,
		Bundles:    bundles,
		LLMKeys:    auditgate.LLMKeySet{Anthropic: os.Getenv("ANTHROPIC_API_KEY")},
		Thresholds: release.DefaultThresholds(),
	}, root
}

// strongEvidence is a positive-control evidence corpus — the gate should
// find the seller's claim well-grounded.
var strongEvidence = []schemas.EvidenceReceipt{
	{
		ReceiptID: "ev_strong_1", URL: "https://example-vendor.com/status",
		BrightDataProduct: "BROWSER",
		FetchedAt:         time.Now().UTC().Format(time.RFC3339),
		ContentHash:       "sha256:strong1",
		RawRef:            "", // body inlined below via separate file fixture creation
	},
}

// Run a fresh test fixture per case so each test owns its body files.
func writeEvidenceWithBody(t *testing.T, dir, receiptID, ext, body string) schemas.EvidenceReceipt {
	t.Helper()
	path := dir + "/" + receiptID + "." + ext
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write evidence body: %v", err)
	}
	return schemas.EvidenceReceipt{
		ReceiptID:         receiptID,
		URL:               "https://example-vendor.com/" + receiptID,
		BrightDataProduct: "BROWSER",
		FetchedAt:         time.Now().UTC().Format(time.RFC3339),
		ContentHash:       "sha256:" + receiptID,
		RawRef:            path,
	}
}

// ─── Case A: BLOCKED — stale / contradicting evidence ─────────────────────

func TestOrchestrator_CaseA_Blocked(t *testing.T) {
	requireEnv(t, "GEM2_API_KEY", "ANTHROPIC_API_KEY")

	o, _ := buildOrchestrator(t)
	dir := t.TempDir()

	// Seller claims real-time freshness; evidence is a 6-month-old static archive.
	receipts := []schemas.EvidenceReceipt{
		writeEvidenceWithBody(t, dir, "ev_caseA_1", "html",
			`<html><title>Vendor Status — archive from 2025-11-15</title>
<body>
Status page from November 15, 2025 (snapshotted). Last update 6 months ago.
This is a Wayback-style archive page — not a live feed.
Coverage shown: NYSE only. No NASDAQ. No real-time prices visible.
</body></html>`),
	}

	offer := schemas.SellerOffer{
		OfferID:   "offer_caseA",
		SellerID:  "seller_caseA",
		Claim:     "Live real-time NYSE+NASDAQ price feed with 1-second freshness",
		PriceUSDC: 0.001,
		SourceURL: "https://example-vendor.com/status",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	buyer := schemas.BuyerRequest{
		RequestID: "req_caseA", BuyerID: "buyer_caseA",
		Query:        "real-time NYSE NASDAQ price feed",
		MaxSpendUSDC: 0.01,
		Policy: schemas.PaymentPolicy{
			SpendCap: 0.01, PublicOnly: true, ClaimGroundedRequired: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	decision, settlement, bundlePath, err := o.Run(ctx, buyer, offer, receipts)
	if err != nil {
		t.Fatalf("orchestrator.Run: %v", err)
	}
	if decision.Verdict != schemas.GateBlockedByTrustGate {
		t.Errorf("Case A verdict=%q want BLOCKED_BY_TRUST_GATE (reason=%s)", decision.Verdict, decision.Reason)
	}
	if settlement.SettlementID != "" {
		t.Errorf("Case A settlement_id=%q want empty (no settlement on BLOCKED)", settlement.SettlementID)
	}
	if settlement.Status != schemas.BlockedByTrustGate {
		t.Errorf("Case A settlement.Status=%q want BLOCKED_BY_TRUST_GATE", settlement.Status)
	}
	if settlement.RealTransaction {
		t.Errorf("simulation invariant violated: real_transaction=true")
	}
	if bundlePath == "" {
		t.Errorf("audit bundle path empty")
	}
	if _, err := os.Stat(bundlePath); err != nil {
		t.Errorf("audit bundle file missing: %v", err)
	}

	t.Logf("Case A: verdict=%s reason=%q settlement.status=%s bundle=%s claims=%d",
		decision.Verdict, decision.Reason, settlement.Status, bundlePath,
		len(decision.ClaimAssessments))
}

// ─── Case B: APPROVED — clean evidence supporting the claim ──────────────

func TestOrchestrator_CaseB_Approved(t *testing.T) {
	requireEnv(t, "GEM2_API_KEY", "ANTHROPIC_API_KEY")

	o, _ := buildOrchestrator(t)
	dir := t.TempDir()

	// Live status page that supports the claim directly.
	receipts := []schemas.EvidenceReceipt{
		writeEvidenceWithBody(t, dir, "ev_caseB_1", "html",
			`<html><title>Vendor Status — Live</title>
<body>
Status as of `+time.Now().UTC().Format(time.RFC3339)+`:
Live NYSE all symbols feed: GREEN. Latency: 712 ms (last 60 s avg).
Live NASDAQ all symbols feed: GREEN. Latency: 689 ms (last 60 s avg).
Update frequency: every 1 second.
Pricing: $0.001 / query for ≤1M queries / month.
</body></html>`),
		writeEvidenceWithBody(t, dir, "ev_caseB_2", "json",
			`{"organic":[{"link":"https://example-vendor.com/pricing","title":"Example Vendor Pricing","description":"$0.001 per query, NYSE+NASDAQ live feeds, 1-second freshness","global_rank":1}]}`),
	}

	offer := schemas.SellerOffer{
		OfferID:   "offer_caseB",
		SellerID:  "seller_caseB",
		Claim:     "Live NYSE+NASDAQ price feed, 1-second freshness, $0.001/query",
		PriceUSDC: 0.001,
		SourceURL: "https://example-vendor.com/status",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	buyer := schemas.BuyerRequest{
		RequestID: "req_caseB", BuyerID: "buyer_caseB",
		Query:        "real-time NYSE NASDAQ price feed",
		MaxSpendUSDC: 0.01,
		Policy: schemas.PaymentPolicy{
			SpendCap: 0.01, PublicOnly: true, ClaimGroundedRequired: true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	decision, settlement, bundlePath, err := o.Run(ctx, buyer, offer, receipts)
	if err != nil {
		t.Fatalf("orchestrator.Run: %v", err)
	}
	if decision.Verdict != schemas.GateApprovedByTrustGate {
		t.Errorf("Case B verdict=%q want APPROVED_BY_TRUST_GATE (reason=%s)", decision.Verdict, decision.Reason)
	}
	if settlement.SettlementID == "" {
		t.Errorf("Case B settlement_id empty — expected sim_x402_<hex>")
	}
	if !strings.HasPrefix(settlement.SettlementID, "sim_x402_") {
		t.Errorf("Case B settlement_id=%q want sim_x402_<hex> prefix", settlement.SettlementID)
	}
	if settlement.Status != schemas.SimulatedSettled {
		t.Errorf("Case B settlement.Status=%q want SIMULATED_SETTLED", settlement.Status)
	}
	if settlement.AmountUSDC <= 0 {
		t.Errorf("Case B settlement.AmountUSDC=%v want >0", settlement.AmountUSDC)
	}
	if settlement.RealTransaction || settlement.PrivateKeysUsed || settlement.RealFundsUsed {
		t.Errorf("simulation invariants violated: %+v", settlement)
	}
	if decision.L1ResultID == "" || decision.L2ResultID == "" {
		t.Errorf("missing upstream result_ids: L1=%q L2=%q", decision.L1ResultID, decision.L2ResultID)
	}
	if bundlePath == "" {
		t.Errorf("audit bundle path empty")
	}

	t.Logf("Case B: verdict=%s settlement_id=%s status=%s amount=%v bundle=%s claims=%d L1=%s L2=%s",
		decision.Verdict, settlement.SettlementID, settlement.Status, settlement.AmountUSDC,
		bundlePath, len(decision.ClaimAssessments), decision.L1ResultID, decision.L2ResultID)
}
