package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gem-squared/ledgerlens/internal/brightdata"
	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"

	"github.com/gin-gonic/gin"
)

// Server holds the dependencies the demo HTTP API needs.
type Server struct {
	Orch            *paymentgate.Orchestrator
	BundlesDir      string                       // artifacts/audit_bundles
	EvidenceDir     string                       // artifacts/fetch_receipts
	BundleStore     *paymentgate.BundleStore
	SERP            *brightdata.SERPClient       // Slice 1 (Judge Request Mode) — live SERP search
	Unlocker        *brightdata.UnlockerClient   // Slice 1 — live Unlocker fetch
	AnthropicAPIKey string                       // Slice 1 — intent + offer synthesis LLM calls
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

// RunResult is the JSON shape returned by POST /cases/:id/run.
type RunResult struct {
	Case             CaseListItem               `json:"case"`
	BuyerRequest     schemas.BuyerRequest       `json:"buyerRequest"`
	SellerOffer      schemas.SellerOffer        `json:"sellerOffer"`
	EvidenceReceipts []schemas.EvidenceReceipt  `json:"evidenceReceipts"`
	Decision         schemas.DecisionPacket     `json:"decision"`
	Settlement       schemas.SimulatedSettlement `json:"settlement"`
	L1               *auditgate.PCheckResponse  `json:"l1,omitempty"`
	L2               *auditgate.OCheckResponse  `json:"l2,omitempty"`
	BundlePath       string                     `json:"bundlePath"`
	DurationMs       int64                      `json:"durationMs"`
}

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

	// Stamp current time onto offer / buyer.
	offer := def.Offer
	offer.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	started := time.Now()
	decision, settlement, bundlePath, err := s.Orch.Run(c.Request.Context(), def.Buyer, offer, receipts)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "orchestrator: " + err.Error()})
		return
	}
	durationMs := time.Since(started).Milliseconds()

	// Pull L1/L2 from the bundle (orchestrator already wrote them).
	bundle, _ := s.BundleStore.Read(decision.DecisionID)

	result := RunResult{
		Case: CaseListItem{ID: def.ID, Title: def.Title, Description: def.Description},
		BuyerRequest:     def.Buyer,
		SellerOffer:      offer,
		EvidenceReceipts: receipts,
		Decision:         decision,
		Settlement:       settlement,
		BundlePath:       bundlePath,
		DurationMs:       durationMs,
	}
	if bundle != nil {
		result.L1 = bundle.L1
		result.L2 = bundle.L2
	}
	c.JSON(http.StatusOK, result)
}

func (s *Server) getBundle(c *gin.Context) {
	id := c.Param("decisionId")
	bundle, err := s.BundleStore.Read(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bundle)
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":                     "ok",
		"service":                    "ledgerlens",
		"settlement_mode":            "simulation",
		"real_transaction_capability": false,
		"cases":                      []string{"a", "b"},
	})
}

var _ = json.Marshal // keep encoding/json imported for future direct use
