package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gem-squared/ledgerlens/internal/agent"
	"github.com/gem-squared/ledgerlens/internal/brightdata"
	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"

	"github.com/gin-gonic/gin"
)

// Server holds the dependencies the demo HTTP API needs.
type Server struct {
	Orch            *paymentgate.Orchestrator
	BundlesDir      string                     // artifacts/audit_bundles
	EvidenceDir     string                     // artifacts/fetch_receipts
	BundleStore     *paymentgate.BundleStore
	SERP            *brightdata.SERPClient     // Slice 1 (Judge Request Mode) — live SERP search
	Unlocker        *brightdata.UnlockerClient // Slice 1 — live Unlocker fetch
	Browser         *brightdata.BrowserClient  // Slice 1 — live Browser API render (optional; nil → step is skipped)
	AnthropicAPIKey string                     // Slice 1 — intent + offer synthesis LLM calls
}

// RegisterRoutes wires the API onto a gin engine.
func (s *Server) RegisterRoutes(g *gin.RouterGroup) {
	g.GET("/cases", s.listCases)
	g.POST("/cases/:id/run", s.runCase)
	g.GET("/audit-bundles/:decisionId", s.getBundle)
	g.GET("/health", s.health)

	// Slice 1 (v2 Judge Request Mode) — JSON blocking endpoint
	g.POST("/deals/run", s.runDeal)
	// Slice 2 — SSE streaming endpoint (same body, text/event-stream response)
	g.POST("/deals/run-stream", s.runDealStream)
	// Slice 3 (Verification Infrastructure Dashboard) — aggregate stats + bundle list
	g.GET("/stats", s.getStats)
	g.GET("/audit-bundles", s.listAuditBundles)
}

// CaseListItem is the JSON shape for GET /cases.
type CaseListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (s *Server) listCases(c *gin.Context) {
	out := make([]CaseListItem, 0, len(AllCases()))
	for _, ca := range AllCases() {
		out = append(out, CaseListItem{ID: ca.ID, Title: ca.Title, Description: ca.Description})
	}
	c.JSON(http.StatusOK, gin.H{"cases": out})
}

// runCase executes a deterministic Case A/B replay and returns a DealRunResult.
// The response shape matches POST /deals/run so the frontend can render any
// post-run UI uniformly via <RichRunResult />.
func (s *Server) runCase(c *gin.Context) {
	id := c.Param("id")
	def := CaseByID(id)
	if def == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown case", "id": id})
		return
	}

	// Materialize evidence raw bodies into the configured fetch_receipts dir
	// (already gitignored). We use a per-case subdir so files don't collide.
	dir := filepath.Join(s.EvidenceDir, "case_"+def.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "evidence dir: " + err.Error()})
		return
	}
	receipts := def.WriteRawFn(dir)

	// Stamp current time onto offer.
	offer := def.Offer
	offer.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	started := time.Now()
	decision, settlement, bundlePath, err := s.Orch.Run(c.Request.Context(), def.Buyer, offer, receipts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "orchestrator: " + err.Error()})
		return
	}
	durationMs := time.Since(started).Milliseconds()

	// Pull L1/L2 from the persisted bundle so the response carries the full
	// audit-gate verdicts (the orchestrator returns Decision but not L1/L2).
	bundle, _ := s.BundleStore.Read(decision.DecisionID)
	stub := &paymentgate.AuditBundle{
		BuyerRequest:     def.Buyer,
		SellerOffer:      offer,
		EvidenceReceipts: receipts,
		Decision:         decision,
		Settlement:       settlement,
	}
	if bundle != nil {
		stub.L1 = bundle.L1
		stub.L2 = bundle.L2
	}

	c.JSON(http.StatusOK, dealRunResultFromBundle(stub, bundlePath, durationMs))
}

// getBundle returns the persisted audit bundle as a DealRunResult so the
// "View" button in Recent Activity can render the same rich UI the LIVE
// run path produces. Synthesizes finalReport on read (legacy bundles never
// went through agent.Compose) and derives mode/durationMs from metadata.
func (s *Server) getBundle(c *gin.Context) {
	id := c.Param("decisionId")
	bundle, err := s.BundleStore.Read(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	bundlePath := filepath.Join(s.BundlesDir, bundle.DecisionID+".json")
	durationMs := durationMsFromTimestamps(bundle.Timestamps)
	c.JSON(http.StatusOK, dealRunResultFromBundle(bundle, bundlePath, durationMs))
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":                      "ok",
		"service":                     "ledgerlens",
		"settlement_mode":             "simulation",
		"real_transaction_capability": false,
		"cases":                       []string{"a", "b"},
	})
}

// dealRunResultFromBundle constructs a DealRunResult from an AuditBundle. Used
// by runCase (after orchestrator persists the bundle) and getBundle (read from
// disk for any legacy decision). FinalReport is synthesized deterministically
// via agent.ComposeFromRequest — no LLM call, no I/O.
func dealRunResultFromBundle(b *paymentgate.AuditBundle, bundlePath string, durationMs int64) DealRunResult {
	return DealRunResult{
		Mode:             modeFromRequestID(b.BuyerRequest.RequestID),
		JudgeRequest:     b.BuyerRequest.Query,
		BuyerIntent:      agent.BuyerIntent{MaxSpendUSDC: b.BuyerRequest.MaxSpendUSDC},
		EvidenceReceipts: b.EvidenceReceipts,
		SellerOffer:      b.SellerOffer,
		Decision:         b.Decision,
		Settlement:       b.Settlement,
		L1:               b.L1,
		L2:               b.L2,
		BundlePath:       bundlePath,
		FinalReport: agent.ComposeFromRequest(
			b.BuyerRequest, b.SellerOffer, b.EvidenceReceipts, b.Decision, b.Settlement,
		),
		DurationMs: durationMs,
	}
}

// durationMsFromTimestamps parses BundleTimestamps and returns the elapsed
// milliseconds. Returns 0 if either timestamp is missing or unparseable —
// safe for the frontend, which displays "0s" rather than blowing up.
func durationMsFromTimestamps(ts paymentgate.BundleTimestamps) int64 {
	if ts.StartedAt == "" || ts.FinishedAt == "" {
		return 0
	}
	start, err1 := time.Parse(time.RFC3339, ts.StartedAt)
	end, err2 := time.Parse(time.RFC3339, ts.FinishedAt)
	if err1 != nil || err2 != nil {
		return 0
	}
	d := end.Sub(start).Milliseconds()
	if d < 0 {
		return 0
	}
	return d
}

// Silence unused-import warning when SchemasFix is the only remaining
// consumer in this file (defensive: schemas may be referenced only through
// the helper's typed parameters which the compiler tracks separately).
var _ = schemas.GateApprovedByTrustGate
