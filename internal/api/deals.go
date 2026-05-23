package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/agent"
	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"

	"github.com/gin-gonic/gin"
)

// DealRunRequest is the JSON body for POST /api/v1/deals/run.
type DealRunRequest struct {
	Query           string  `json:"query"           binding:"required"`
	MaxSpendUSDC    float64 `json:"maxSpendUSDC,omitempty"`
	RequireGrounded bool    `json:"requireGrounded,omitempty"`
	Mode            string  `json:"mode,omitempty"` // "live" (default) | "prewarmed" (Slice 3) | "replay" (Case A/B)
}

// DealRunResult is the JSON body returned by POST /api/v1/deals/run.
type DealRunResult struct {
	Mode             string                       `json:"mode"`
	JudgeRequest     string                       `json:"judgeRequest"`
	BuyerIntent      agent.BuyerIntent            `json:"buyerIntent"`
	AgentNarrative   []string                     `json:"agentNarrative"`
	EvidenceReceipts []schemas.EvidenceReceipt    `json:"evidenceReceipts"`
	SellerOffer      schemas.SellerOffer          `json:"sellerOffer"`
	Decision         schemas.DecisionPacket       `json:"decision"`
	Settlement       schemas.SimulatedSettlement  `json:"settlement"`
	L1               *auditgate.PCheckResponse    `json:"l1,omitempty"`
	L2               *auditgate.OCheckResponse    `json:"l2,omitempty"`
	BundlePath       string                       `json:"bundlePath"`
	FinalReport      agent.FinalReport            `json:"finalReport"`
	DurationMs       int64                        `json:"durationMs"`
}

// runDeal handles POST /api/v1/deals/run. LIVE mode only for Slice 1.
// PRE-WARMED + REPLAY land in later slices.
func (s *Server) runDeal(c *gin.Context) {
	var req DealRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	if req.Mode == "" {
		req.Mode = "live"
	}
	if req.Mode != "live" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "mode not implemented in slice 1",
			"mode":  req.Mode,
			"note":  "Slice 2 (sse), Slice 3 (prewarmed) — for now use mode=live",
		})
		return
	}
	if s.AnthropicAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ANTHROPIC_API_KEY not configured on server"})
		return
	}
	if s.SERP == nil || s.Unlocker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bright Data clients not configured (BRIGHTDATA_API_TOKEN + zones)"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	started := time.Now()
	narrative := []string{}
	add := func(line string) { narrative = append(narrative, line) }

	// ─── ① Buyer Agent: intent extraction ────────────────────────────────
	add("Judge request received.")
	intent, err := agent.Extract(ctx, req.Query, s.AnthropicAPIKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "intent extraction: " + err.Error()})
		return
	}
	if intent.OffDomain {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":        "off_domain",
			"politeReject": intent.PoliteReject,
			"judgeRequest": req.Query,
		})
		return
	}
	add(fmt.Sprintf("Buyer Agent interpreted need: %q.", intent.DataNeed))

	// Apply the buyer's max-spend override if provided
	if req.MaxSpendUSDC > 0 {
		intent.MaxSpendUSDC = req.MaxSpendUSDC
	}

	// ─── ② Bright Data: search + fetch ───────────────────────────────────
	add("Bright Data searching public web for candidate sources…")
	receipts, err := s.discoverEvidence(ctx, intent)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "evidence discovery: " + err.Error()})
		return
	}
	if len(receipts) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "no public evidence found for the request"})
		return
	}
	add(fmt.Sprintf("Bright Data collected %d evidence receipt(s).", len(receipts)))

	// ─── ③ Seller Offer Agent: synthesize offer from evidence ───────────
	add("Seller Offer Agent synthesizing candidate offer from evidence…")
	offer, err := agent.Synthesize(ctx, intent, receipts, s.AnthropicAPIKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "offer synthesis: " + err.Error()})
		return
	}
	add(fmt.Sprintf("Candidate offer constructed: %q at %v USDC.", truncate(offer.Claim, 80), offer.PriceUSDC))

	// ─── ④ Orchestrator: L1 + L2 + L3 + Settle ──────────────────────────
	add("GEM² Trust Gate auditing seller claim against evidence…")
	buyer := schemas.BuyerRequest{
		RequestID:    "req_deal_" + agent.RandID8(),
		BuyerID:      "buyer_judge",
		Query:        req.Query,
		MaxSpendUSDC: intent.MaxSpendUSDC,
		Policy: schemas.PaymentPolicy{
			SpendCap:              intent.MaxSpendUSDC,
			PublicOnly:            true,
			ClaimGroundedRequired: req.RequireGrounded || true, // default true for the demo
		},
	}
	decision, settlement, bundlePath, err := s.Orch.Run(ctx, buyer, offer, receipts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "orchestrator: " + err.Error()})
		return
	}

	var bundleL1 *auditgate.PCheckResponse
	var bundleL2 *auditgate.OCheckResponse
	if b, _ := s.BundleStore.Read(decision.DecisionID); b != nil {
		bundleL1 = b.L1
		bundleL2 = b.L2
	}

	// Narrative completion line
	if decision.Verdict == schemas.GateApprovedByTrustGate {
		add("L3 Trust Gate APPROVED. Simulated x402 settlement issued.")
	} else if decision.Verdict == schemas.GateBlockedByTrustGate {
		add("L3 Trust Gate BLOCKED. No settlement issued.")
	} else {
		add(fmt.Sprintf("L3 Trust Gate verdict: %s.", decision.Verdict))
	}

	// ─── ⑤ Final Report (template, no LLM) ──────────────────────────────
	final := agent.Compose(req.Query, intent, offer, receipts, decision, settlement)

	c.JSON(http.StatusOK, DealRunResult{
		Mode:             "live",
		JudgeRequest:     req.Query,
		BuyerIntent:      intent,
		AgentNarrative:   narrative,
		EvidenceReceipts: receipts,
		SellerOffer:      offer,
		Decision:         decision,
		Settlement:       settlement,
		L1:               bundleL1,
		L2:               bundleL2,
		BundlePath:       bundlePath,
		FinalReport:      final,
		DurationMs:       time.Since(started).Milliseconds(),
	})
}

// discoverEvidence runs Bright Data SERP for the buyer's search terms and
// fetches the top candidate URL(s) via Web Unlocker. Browser API is
// intentionally NOT used here — it takes ~13s alone and would push live
// latency past 45s. Browser stays in Case A/B replay.
func (s *Server) discoverEvidence(ctx context.Context, intent agent.BuyerIntent) ([]schemas.EvidenceReceipt, error) {
	if len(intent.SearchTerms) == 0 {
		return nil, fmt.Errorf("no search terms in buyer intent")
	}
	query := strings.Join(intent.SearchTerms[:min(3, len(intent.SearchTerms))], " ")

	serpReceipt, err := s.SERP.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SERP search: %w", err)
	}
	receipts := []schemas.EvidenceReceipt{serpReceipt}

	// Pull at most 2 candidate URLs out of the SERP body and Unlocker them.
	urls := extractCandidateURLs(serpReceipt.RawRef, 2)
	for _, u := range urls {
		ev, err := s.Unlocker.Fetch(ctx, u)
		if err != nil {
			// Soft failure — keep the SERP receipt and any prior unlocker
			// fetches; the gate can decide on partial evidence.
			continue
		}
		receipts = append(receipts, ev)
	}
	return receipts, nil
}

// extractCandidateURLs pulls top organic links from a SERP JSON file.
// Defensive: SERP JSON shape can vary; uses string-search to be resilient.
func extractCandidateURLs(serpJSONPath string, maxN int) []string {
	body, err := readFile(serpJSONPath)
	if err != nil {
		return nil
	}
	// Find all "link":"..." substrings — Bright Data's SERP JSON uses this key
	// inside organic[]. Cheap and resilient to schema drift.
	urls := []string{}
	s := string(body)
	for {
		i := strings.Index(s, `"link":"`)
		if i < 0 {
			break
		}
		s = s[i+len(`"link":"`):]
		j := strings.Index(s, `"`)
		if j < 0 {
			break
		}
		candidate := s[:j]
		s = s[j+1:]
		// Filter out Google-internal redirects + noisy domains
		if strings.HasPrefix(candidate, "http") &&
			!strings.Contains(candidate, "google.com/url") &&
			!strings.Contains(candidate, "youtube.com/") {
			urls = append(urls, candidate)
			if len(urls) >= maxN {
				break
			}
		}
	}
	return urls
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func readFile(path string) ([]byte, error) {
	return readFileImpl(path)
}
