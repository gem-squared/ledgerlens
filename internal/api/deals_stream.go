package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/agent"
	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"

	"github.com/gin-gonic/gin"
)

// runDealStream is the SSE variant of runDeal. Same body, but the response
// is a text/event-stream of paymentgate.StepEvent frames. The last event
// carries `step: "result"` with the full DealRunResult payload.
//
// Action verb labels per David's directive — judges should see what each
// agent is *doing*, not technical step IDs.
func (s *Server) runDealStream(c *gin.Context) {
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
			"error": "only mode=live is implemented in slice 2; PRE-WARMED is slice 3",
		})
		return
	}
	if s.AnthropicAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ANTHROPIC_API_KEY not configured"})
		return
	}
	if s.SERP == nil || s.Unlocker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bright Data clients not configured"})
		return
	}

	// SSE response headers — Caddy passes these through.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // hint for any upstream proxy
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		// Can't stream — caller will get a partial-but-flat response.
		// Shouldn't happen with gin's default writer.
		fmt.Fprintln(c.Writer, "data: {\"step\":\"error\",\"label\":\"streaming not supported\"}")
		return
	}

	events := make(chan paymentgate.StepEvent, 32)
	resultCh := make(chan DealRunResult, 1)
	errCh := make(chan error, 1)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	// Initial event so the browser knows the connection is live.
	emit := func(e paymentgate.StepEvent) {
		select {
		case events <- e:
		default:
			// drop on full — preserves liveness over completeness
		}
	}

	// Run the pipeline in a goroutine; main loop pumps events to the client.
	go func() {
		defer close(events)
		result, err := s.runDealPipeline(ctx, req, emit)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	writeEvent := func(e paymentgate.StepEvent) {
		data, _ := json.Marshal(e)
		fmt.Fprintf(c.Writer, "event: step\ndata: %s\n\n", data)
		flusher.Flush()
	}
	writeResult := func(r DealRunResult) {
		data, _ := json.Marshal(r)
		fmt.Fprintf(c.Writer, "event: result\ndata: %s\n\n", data)
		flusher.Flush()
	}
	writeError := func(err error) {
		payload := map[string]string{"error": err.Error()}
		data, _ := json.Marshal(payload)
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", data)
		flusher.Flush()
	}

	// Comment line keeps the stream alive through proxies that close idle conns.
	fmt.Fprintln(c.Writer, ": ledgerlens stream open")
	flusher.Flush()

	pendingResult := false
	var finalResult DealRunResult
	for {
		select {
		case <-ctx.Done():
			writeError(ctx.Err())
			return
		case e, more := <-events:
			if !more {
				// goroutine finished — drain result or error
				if pendingResult {
					writeResult(finalResult)
				}
				return
			}
			writeEvent(e)
		case r := <-resultCh:
			finalResult = r
			pendingResult = true
			// don't write yet — let any remaining events flush first
		case err := <-errCh:
			writeError(err)
			return
		}
	}
}

// runDealPipeline is the shared inner pipeline used by BOTH the streaming
// (runDealStream) and the non-streaming (runDeal) handlers. The emit
// callback is called at each major boundary — JSON callers pass nil.
func (s *Server) runDealPipeline(
	ctx context.Context,
	req DealRunRequest,
	emit paymentgate.EventEmitter,
) (DealRunResult, error) {
	started := time.Now()
	narrative := []string{}
	add := func(line string) { narrative = append(narrative, line) }

	emitWrap := func(step, status, label string, detail any) {
		if emit != nil {
			emit(paymentgate.StepEvent{
				Step: step, Status: status, Label: label, Detail: detail,
				Ts: time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	// Step 1 — Judge request received
	emitWrap("judge_request", "passed", "Judge request received.", map[string]any{"query": req.Query})
	add("Judge request received.")

	// Step 2 — Buyer Agent: intent extraction
	emitWrap("buyer_intent", "running", "Buyer Agent extracting intent from the request…", nil)
	intent, err := agent.Extract(ctx, req.Query, s.AnthropicAPIKey)
	if err != nil {
		emitWrap("buyer_intent", "failed", "Intent extraction failed: "+err.Error(), nil)
		return DealRunResult{}, fmt.Errorf("intent extraction: %w", err)
	}
	if intent.OffDomain {
		emitWrap("buyer_intent", "rejected", "Off-domain query rejected by Buyer Agent.", map[string]any{"polite_reject": intent.PoliteReject})
		return DealRunResult{}, fmt.Errorf("off_domain: %s", intent.PoliteReject)
	}
	emitWrap("buyer_intent", "passed",
		fmt.Sprintf("Buyer Agent interpreted need: %q (freshness=%s, max=$%v).",
			intent.DataNeed, intent.Freshness, intent.MaxSpendUSDC),
		map[string]any{"intent": intent},
	)
	add(fmt.Sprintf("Buyer Agent interpreted need: %q.", intent.DataNeed))

	if req.MaxSpendUSDC > 0 {
		intent.MaxSpendUSDC = req.MaxSpendUSDC
	}

	// Step 3 — Bright Data SERP discovery
	emitWrap("brightdata_search", "running", "Bright Data SERP API searching the public web for candidate sources…", nil)
	if len(intent.SearchTerms) == 0 {
		emitWrap("brightdata_search", "failed", "No search terms produced by Buyer Agent.", nil)
		return DealRunResult{}, fmt.Errorf("no search terms in buyer intent")
	}
	queryText := strings.Join(intent.SearchTerms[:min(3, len(intent.SearchTerms))], " ")
	serpReceipt, err := s.SERP.Search(ctx, queryText)
	if err != nil {
		emitWrap("brightdata_search", "failed", "SERP search failed: "+err.Error(), nil)
		return DealRunResult{}, fmt.Errorf("SERP search: %w", err)
	}
	urls := extractCandidateURLs(serpReceipt.RawRef, 2)
	emitWrap("brightdata_search", "passed",
		fmt.Sprintf("Bright Data SERP returned %d candidate URL(s).", len(urls)),
		map[string]any{"candidates": urls},
	)

	// Step 4 — Bright Data Unlocker fetch
	emitWrap("brightdata_fetch", "running", "Bright Data Web Unlocker fetching public evidence…", nil)
	receipts := []schemas.EvidenceReceipt{serpReceipt}
	for _, u := range urls {
		ev, err := s.Unlocker.Fetch(ctx, u)
		if err != nil {
			continue
		}
		receipts = append(receipts, ev)
	}
	if len(receipts) == 1 {
		emitWrap("brightdata_fetch", "failed", "No candidate URLs could be fetched.", nil)
		return DealRunResult{}, fmt.Errorf("no candidate URLs fetched")
	}
	emitWrap("brightdata_fetch", "passed",
		fmt.Sprintf("Bright Data collected %d evidence receipt(s).", len(receipts)),
		map[string]any{"receiptCount": len(receipts)},
	)
	add(fmt.Sprintf("Bright Data collected %d evidence receipt(s).", len(receipts)))

	// NOTE: Browser API step intentionally NOT called in the LIVE pipeline.
	// Reason: brd.superproxy.io WebDriver sessions occasionally exceed the
	// deal's 120s context budget (4 WebDriver calls × up to 90s each), which
	// caused "context deadline exceeded" failures with no recovery. The
	// BrowserClient is still instantiated in main.go and the unit tests in
	// internal/brightdata/integration_test.go exercise it — proof of depth
	// without putting the user-facing demo at risk. Re-enable post-hackathon
	// once a per-step Browser budget + graceful-skip pattern is in place.

	// Step 5 — Seller Offer Agent
	emitWrap("seller_offer", "running", "Seller Offer Agent constructing candidate deal from evidence…", nil)
	offer, err := agent.Synthesize(ctx, intent, receipts, s.AnthropicAPIKey)
	if err != nil {
		emitWrap("seller_offer", "failed", "Offer synthesis failed: "+err.Error(), nil)
		return DealRunResult{}, fmt.Errorf("offer synthesis: %w", err)
	}
	emitWrap("seller_offer", "passed",
		fmt.Sprintf("Seller Offer Agent proposed: %q at %v USDC.", truncate(offer.Claim, 80), offer.PriceUSDC),
		map[string]any{"offer": offer},
	)
	add(fmt.Sprintf("Candidate offer constructed: %q at %v USDC.", truncate(offer.Claim, 80), offer.PriceUSDC))

	// Steps 6/7/8/9 — Orchestrator (L1 → L2 → L3 → Settle) emits its own events
	buyer := schemas.BuyerRequest{
		RequestID:    "req_deal_" + agent.RandID8(),
		BuyerID:      "buyer_judge",
		Query:        req.Query,
		MaxSpendUSDC: intent.MaxSpendUSDC,
		Policy: schemas.PaymentPolicy{
			SpendCap:              intent.MaxSpendUSDC,
			PublicOnly:            true,
			ClaimGroundedRequired: req.RequireGrounded || true,
		},
	}
	decision, settlement, bundlePath, err := s.Orch.RunWithEvents(ctx, buyer, offer, receipts, emit)
	if err != nil {
		return DealRunResult{}, fmt.Errorf("orchestrator: %w", err)
	}

	var bundleL1 *auditgate.PCheckResponse
	var bundleL2 *auditgate.OCheckResponse
	if b, _ := s.BundleStore.Read(decision.DecisionID); b != nil {
		bundleL1 = b.L1
		bundleL2 = b.L2
	}

	// Step 10 — Final report
	emitWrap("final_report", "running", "Composing final report for the judge…", nil)
	finalReport := agent.Compose(req.Query, intent, offer, receipts, decision, settlement)
	emitWrap("final_report", "passed", "Final report generated.", nil)

	if decision.Verdict == schemas.GateApprovedByTrustGate {
		add("L3 Trust Gate APPROVED. Simulated x402 settlement issued.")
	} else if decision.Verdict == schemas.GateBlockedByTrustGate {
		add("L3 Trust Gate BLOCKED. No settlement issued. A fast agent would have paid. LedgerLens waited.")
	} else {
		add(fmt.Sprintf("L3 Trust Gate verdict: %s.", decision.Verdict))
	}

	return DealRunResult{
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
		FinalReport:      finalReport,
		DurationMs:       time.Since(started).Milliseconds(),
	}, nil
}
