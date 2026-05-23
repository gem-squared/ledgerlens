// Command ledgerlens — LedgerLens API server.
//
// One Go binary that wires:
//   - Bright Data evidence ingestion         (internal/brightdata, Unit 2)
//   - GEM² Trust Gate L0/L1/L2 verification  (internal/trustgate,  Unit 3)
//   - L3 release + x402 simulation           (internal/paymentgate, Unit 4)
//   - HTTP API for the Next.js demo UI       (this file)
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

	r := gin.Default()
	registerRoutes(r, db, cfg)

	addr := cfg.Bind + ":" + cfg.Port
	log.Printf("ledgerlens: listening on %s  (settlement=%s)", addr, cfg.SettlementMode)
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

// ─── Storage ────────────────────────────────────────────────────────────────

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

// ─── Routes ─────────────────────────────────────────────────────────────────

func registerRoutes(r *gin.Engine, _ *sql.DB, cfg Config) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":          "ok",
			"service":         "ledgerlens",
			"settlement_mode": cfg.SettlementMode,
			"real_transaction_capability": false,
		})
	})

	// API v1 — endpoints land in Units 2 / 3 / 4. Stubs respond 501 so the
	// frontend can wire against them.
	api := r.Group("/api/v1")
	api.POST("/requests", notImplemented("Unit 4: POST /requests"))
	api.GET("/decisions/:id", notImplemented("Unit 4: GET /decisions/:id"))
	api.GET("/audit-bundles/:id", notImplemented("Unit 4: GET /audit-bundles/:id"))
	api.GET("/offers", notImplemented("Unit 3: GET /offers"))
	api.POST("/offers", notImplemented("Unit 3: POST /offers"))
}

func notImplemented(label string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "not_implemented",
			"label": label,
			"note":  "scaffolded in Unit 1; implementation belongs to a later unit",
		})
	}
}
