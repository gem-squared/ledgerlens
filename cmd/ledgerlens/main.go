// Command ledgerlens — LedgerLens API server.
//
// One Go binary that wires:
//   - Bright Data evidence ingestion         (internal/brightdata, Unit 2)
//   - GEM² Trust Gate L0/L1/L2 verification  (internal/trustgate,  Unit 3)
//   - L3 release + x402 simulation           (internal/paymentgate, Unit 4)
//   - HTTP API for the Next.js demo UI       (internal/api, Unit 5)
//
// Settlement is x402 protocol SIMULATION only (v2.2 lock). No private keys,
// no Coinbase account, no Base Sepolia dependency. The Settler interface in
// internal/schemas is the swap point for a post-hackathon real implementation.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gem-squared/ledgerlens/internal/api"
	"github.com/gem-squared/ledgerlens/internal/brightdata"
	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
	"github.com/gem-squared/ledgerlens/internal/trustgate/memory"
	"github.com/gem-squared/ledgerlens/internal/trustgate/release"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := loadConfig()

	db, err := openDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("ledgerlens: open db %q: %v", cfg.DBPath, err)
	}
	defer db.Close()

	mem, err := memory.NewStore(db)
	if err != nil {
		log.Fatalf("ledgerlens: memory store: %v", err)
	}

	bundles, err := paymentgate.NewBundleStore("artifacts/audit_bundles")
	if err != nil {
		log.Fatalf("ledgerlens: bundle store: %v", err)
	}

	if err := os.MkdirAll("artifacts/fetch_receipts", 0o755); err != nil {
		log.Fatalf("ledgerlens: fetch_receipts dir: %v", err)
	}

	orch := &paymentgate.Orchestrator{
		Audit:   auditgate.NewClient(cfg.GEM2TPMNCheckerBaseURL, cfg.GEM2APIKey),
		Settler: paymentgate.NewSimulatedSettler(),
		Mem:     mem,
		Bundles: bundles,
		LLMKeys: auditgate.LLMKeySet{
			Anthropic: cfg.AnthropicAPIKey,
			Gemini:    cfg.GeminiAPIKey,
			OpenAI:    cfg.OpenAIAPIKey,
		},
		Thresholds: release.DefaultThresholds(),
	}

	// Slice-1 (Judge Request Mode) — Bright Data live wrappers for the
	// agent pipeline. Reuses the fetch_receipts/ store that orchestrator
	// also writes to.
	var serpClient *brightdata.SERPClient
	var unlockerClient *brightdata.UnlockerClient
	var browserClient *brightdata.BrowserClient
	if cfg.BrightDataAPIToken != "" {
		bdStore, err := brightdata.NewReceiptStore("artifacts/fetch_receipts")
		if err != nil {
			log.Fatalf("ledgerlens: brightdata receipt store: %v", err)
		}
		if z := os.Getenv("BRIGHTDATA_SERP_ZONE"); z != "" {
			serpClient = brightdata.NewSERPClient(cfg.BrightDataAPIToken, z, bdStore)
		}
		if z := os.Getenv("BRIGHTDATA_UNLOCKER_ZONE"); z != "" {
			unlockerClient = brightdata.NewUnlockerClient(cfg.BrightDataAPIToken, z, bdStore)
		}
		// Browser API authenticates via its embedded-credentials HTTPS URL,
		// not the API token — but in practice all three are set together.
		if u := os.Getenv("BRIGHTDATA_BROWSER_HTTPS_URL"); u != "" {
			bc, berr := brightdata.NewBrowserClient(u, bdStore)
			if berr != nil {
				log.Printf("ledgerlens: browser client init failed: %v (skipping Browser)", berr)
			} else {
				browserClient = bc
			}
		}
	}

	srv := &api.Server{
		Orch:            orch,
		BundlesDir:      "artifacts/audit_bundles",
		EvidenceDir:     "artifacts/fetch_receipts",
		BundleStore:     bundles,
		SERP:            serpClient,
		Unlocker:        unlockerClient,
		Browser:         browserClient,
		AnthropicAPIKey: cfg.AnthropicAPIKey,
	}

	r := gin.Default()

	// CORS for the Next.js dev server (3001 + the dev-rewrite path).
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001", "http://127.0.0.1:3001", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":                      "ok",
			"service":                     "ledgerlens",
			"settlement_mode":             cfg.SettlementMode,
			"real_transaction_capability": false,
		})
	})

	v1 := r.Group("/api/v1")
	srv.RegisterRoutes(v1)

	// Serve embedded Next.js static export at /. Wired last so /api/* and
	// /health take precedence. See cmd/ledgerlens/static_embed.go.
	mountStatic(r)

	addr := cfg.Bind + ":" + cfg.Port
	log.Printf("ledgerlens: listening on %s  (settlement=%s, gem2_configured=%v)",
		addr, cfg.SettlementMode, cfg.GEM2APIKey != "")
	if err := r.Run(addr); err != nil {
		log.Fatalf("ledgerlens: gin run: %v", err)
	}
}

// ─── Config ─────────────────────────────────────────────────────────────────

type Config struct {
	Bind                   string
	Port                   string
	DBPath                 string
	LogLevel               string
	GEM2APIKey             string
	GEM2TPMNCheckerBaseURL string
	GeminiAPIKey           string
	AnthropicAPIKey        string
	OpenAIAPIKey           string
	BrightDataAPIToken     string
	SettlementMode         string // "simulation" — locked for hackathon
}

func loadConfig() Config {
	return Config{
		Bind:                   envOr("LEDGERLENS_BIND", "127.0.0.1"),
		Port:                   envOr("PORT", "8082"),
		DBPath:                 envOr("LEDGERLENS_DB_PATH", "storage/local.db"),
		LogLevel:               envOr("LEDGERLENS_LOG_LEVEL", "info"),
		GEM2APIKey:             os.Getenv("GEM2_API_KEY"),
		GEM2TPMNCheckerBaseURL: envOr("GEM2_TPMN_CHECKER_BASE_URL", "https://gem2-tpmn-checker.fly.dev"),
		GeminiAPIKey:           os.Getenv("GEMINI_API_KEY"),
		AnthropicAPIKey:        os.Getenv("ANTHROPIC_API_KEY"),
		OpenAIAPIKey:           os.Getenv("OPENAI_API_KEY"),
		BrightDataAPIToken:     os.Getenv("BRIGHTDATA_API_TOKEN"),
		SettlementMode:         envOr("LEDGERLENS_SETTLEMENT_MODE", "simulation"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(dir(path), 0o755); err != nil {
		return nil, err
	}
	return sql.Open("sqlite", path)
}

func dir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}
