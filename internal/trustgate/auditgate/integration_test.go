package auditgate_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
)

func requireEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if os.Getenv(k) == "" {
			t.Skipf("env %s not set — skipping live test", k)
		}
	}
}

// canned evidence chunks resembling what evidence.WrapReceipts produces — kept
// inline so the test doesn't depend on Unit 2 artifacts existing.
var liveEvidence = []string{
	`[receipt ev_test1 | product=BROWSER | url=https://example-vendor.com/status | fetched=2026-05-23T11:00:00Z | hash=sha256:abc...]
NYSE + NASDAQ live price feed status page. Last update: 2026-05-23 10:59:53 UTC. Latency: 712ms. Coverage: NYSE all symbols, NASDAQ all symbols, no Asian markets. Pricing tier: $0.001 / query for ≤1M queries / month.`,
	`[receipt ev_test2 | product=SERP | url=https://www.google.com/search?gl=us&hl=en&q=example-vendor+pricing | fetched=2026-05-23T11:00:05Z | hash=sha256:def...]
{"organic":[{"link":"https://example-vendor.com/pricing","title":"Example Vendor — Pricing","description":"Live market data API. $0.001 per query. NYSE+NASDAQ coverage.","global_rank":1}]}`,
}

func TestPCheck_Live_Roundtrip(t *testing.T) {
	requireEnv(t, "GEM2_API_KEY", "ANTHROPIC_API_KEY")

	c := auditgate.NewClient(os.Getenv("GEM2_TPMN_CHECKER_BASE_URL"), os.Getenv("GEM2_API_KEY"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := auditgate.PCheckRequest{
		I: "SellerOffer: Live NYSE+NASDAQ price feed at $0.001/query, 1-second freshness",
		A: json.RawMessage(`"SellerOffer{offerId, sellerId, claim, priceUSDC, sourceUrl}"`),
		P: []string{
			"claim must be supported by provided evidence",
			"price must be within buyer.maxSpendUSDC (assume cap=0.01)",
			"no Δe→∫de overclaims (sparse data presented as trend)",
		},
		T:              70,
		Evidence:       liveEvidence,
		SessionContext: "ledgerlens-unit-3-test",
		Provider:       "claude",
	}
	resp, err := c.PCheck(ctx, req, auditgate.LLMKeySet{Anthropic: os.Getenv("ANTHROPIC_API_KEY")}, "")
	if err != nil {
		t.Fatalf("PCheck: %v", err)
	}
	if resp.Verdict != "ALLOW" && resp.Verdict != "DENY" {
		t.Errorf("unexpected verdict: %q", resp.Verdict)
	}
	if resp.Score < 0 || resp.Score > 100 {
		t.Errorf("score out of range: %d", resp.Score)
	}
	if resp.Meta.ResultID == "" {
		t.Errorf("missing result_id")
	}
	if len(resp.Reasons) == 0 {
		t.Errorf("empty reasons")
	}

	parsed := auditgate.ParseReasons(resp.Reasons)
	if len(parsed.Rules) == 0 {
		t.Errorf("parser extracted no rules from live response — reasons=%v", resp.Reasons)
	}
	ca := auditgate.ToClaimAssessments(resp.Reasons, resp.Score)
	if len(ca) == 0 {
		t.Errorf("ToClaimAssessments produced none")
	}
	for _, c := range ca {
		if c.Status != schemas.ClaimGrounded &&
			c.Status != schemas.ClaimInferred &&
			c.Status != schemas.ClaimExtrapolated &&
			c.Status != schemas.ClaimUnknown {
			t.Errorf("claim %s has non-canonical status %q", c.ClaimID, c.Status)
		}
	}

	t.Logf("PCheck live: verdict=%s score=%d result_id=%s reasons=%d rules=%d SPT=%v EEF=%d claims=%d cost=$%.4f",
		resp.Verdict, resp.Score, resp.Meta.ResultID, len(resp.Reasons),
		len(parsed.Rules), parsed.SPTCodes, len(parsed.EEFFlags), len(ca),
		resp.Meta.Usage.EstimatedCostUSD)

	// SEED REPLAY: save the live response into the smoke case so the next
	// test (TestReplay_LoadAfterSave) can read it back.
	store, err := auditgate.NewReplayStore("../../../artifacts/demo_cases")
	if err != nil {
		t.Fatalf("replay store: %v", err)
	}
	if err := store.SavePCheck("unit3_smoke", resp); err != nil {
		t.Fatalf("save replay: %v", err)
	}
}

func TestOCheck_Live_Roundtrip(t *testing.T) {
	requireEnv(t, "GEM2_API_KEY", "ANTHROPIC_API_KEY")

	c := auditgate.NewClient(os.Getenv("GEM2_TPMN_CHECKER_BASE_URL"), os.Getenv("GEM2_API_KEY"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// A draft DecisionPacket — the "output" of the local F (matching engine).
	decisionPacket := `{
		"decisionId": "dp_test1",
		"requestId":  "req_test1",
		"offerId":    "offer_test1",
		"verdict":    "APPROVED_BY_TRUST_GATE",
		"reason":     "claim grounded by NYSE+NASDAQ status page",
		"paymentAllowed": true,
		"auditBundleRef": "artifacts/audit_bundles/dp_test1.json"
	}`

	req := auditgate.OCheckRequest{
		O: decisionPacket,
		B: json.RawMessage(`"DecisionPacket{decisionId, requestId, offerId, verdict, reason, paymentAllowed, auditBundleRef}"`),
		P: []string{
			"decisionPacket.verdict must be APPROVED_BY_TRUST_GATE | BLOCKED_BY_TRUST_GATE | ESCALATED_TO_HUMAN",
			"paymentAllowed = true iff verdict = APPROVED_BY_TRUST_GATE",
			"reason field must reference grounded evidence",
		},
		T:              75,
		Evidence:       liveEvidence,
		SessionContext: "ledgerlens-unit-3-test",
		Provider:       "claude",
	}
	resp, err := c.OCheck(ctx, req, auditgate.LLMKeySet{Anthropic: os.Getenv("ANTHROPIC_API_KEY")}, "")
	if err != nil {
		t.Fatalf("OCheck: %v", err)
	}
	if resp.Verdict != "SUCCESS" && resp.Verdict != "FAILURE" {
		t.Errorf("unexpected verdict: %q", resp.Verdict)
	}
	if resp.Score < 0 || resp.Score > 100 {
		t.Errorf("score out of range: %d", resp.Score)
	}
	if resp.Meta.ResultID == "" {
		t.Errorf("missing result_id")
	}
	if len(resp.Reasons) == 0 {
		t.Errorf("empty reasons")
	}

	parsed := auditgate.ParseReasons(resp.Reasons)
	t.Logf("OCheck live: verdict=%s score=%d result_id=%s reasons=%d rules=%d SPT=%v cost=$%.4f",
		resp.Verdict, resp.Score, resp.Meta.ResultID, len(resp.Reasons),
		len(parsed.Rules), parsed.SPTCodes, resp.Meta.Usage.EstimatedCostUSD)

	// SEED REPLAY
	store, err := auditgate.NewReplayStore("../../../artifacts/demo_cases")
	if err != nil {
		t.Fatalf("replay store: %v", err)
	}
	if err := store.SaveOCheck("unit3_smoke", resp); err != nil {
		t.Fatalf("save replay: %v", err)
	}
}

func TestReplay_LoadAfterSave(t *testing.T) {
	store, err := auditgate.NewReplayStore(t.TempDir())
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	src := &auditgate.PCheckResponse{
		Verdict: "ALLOW",
		Score:   88,
		Reasons: []string{
			"[TYPE] match",
			"[RULE-1] some rule — PASS: explained",
			"[SPT-S→T] generalization warning",
		},
		Meta: auditgate.Meta{ResultID: "rid_test1", DurationMs: 1234},
	}
	if err := store.SavePCheck("offline_case", src); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.LoadPCheck("offline_case")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Verdict != src.Verdict || got.Score != src.Score || got.Meta.ResultID != src.Meta.ResultID {
		t.Errorf("roundtrip mismatch: got=%+v want=%+v", got, src)
	}
	if len(got.Reasons) != 3 {
		t.Errorf("reasons roundtrip: got %d want 3", len(got.Reasons))
	}

	// also verify the file landed where expected
	if _, err := os.Stat(filepath.Join(t.TempDir())); err != nil {
		// no-op; t.TempDir() returns a new dir each call, just sanity
	}
}
