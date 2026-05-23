// Package memory provides the local SQLite mirror of gem2-tpmn-checker's
// gate_decisions audit trail. Schema matches AUDIT_GATE_API.md §"Audit Trail"
// hash-only storage — raw inputs are NEVER persisted.
//
// Used by Unit 4 (L3 release packet correlates with prior decisions) and
// Unit 5 (memory display). Unit 3 only ships the schema + writer.
package memory

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Store wraps the gate_decisions table.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) (*Store, error) {
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("memory store: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS gate_decisions (
    id            TEXT PRIMARY KEY,        -- matches upstream meta.result_id
    gate_type     TEXT NOT NULL,           -- 'P_GATE' or 'O_GATE'
    verdict       TEXT NOT NULL,           -- ALLOW|DENY|SUCCESS|FAILURE
    score         INTEGER NOT NULL,        -- 0-100
    threshold     INTEGER NOT NULL,        -- T from request
    input_hash    TEXT NOT NULL,           -- SHA-256 of i (or o)
    p_hash        TEXT NOT NULL,           -- SHA-256 of joined p rules
    evidence_hash TEXT,                    -- SHA-256 of joined evidence chunks; NULL if no evidence
    provider      TEXT NOT NULL,
    duration_ms   INTEGER NOT NULL,
    created_at    TEXT NOT NULL,           -- ISO8601 UTC
    reasons_json  TEXT                     -- LedgerLens local extension: full reasons array
);
CREATE INDEX IF NOT EXISTS idx_gate_decisions_created_at ON gate_decisions(created_at);
CREATE INDEX IF NOT EXISTS idx_gate_decisions_gate_type  ON gate_decisions(gate_type);
`

// Record persists a single gate decision. All raw content is hashed; only
// metadata + reasons are stored.
type Record struct {
	ID           string
	GateType     string // P_GATE | O_GATE
	Verdict      string
	Score        int
	Threshold    int
	Input        string   // hashed before insert
	P            []string // hashed (joined) before insert
	Evidence     []string // hashed (joined) before insert; pass nil for no-evidence rows
	Provider     string
	DurationMs   int64
	ReasonsJSON  string
}

func (s *Store) Insert(rec Record) error {
	inputHash := sha256hex(rec.Input)
	pHash := "none"
	if len(rec.P) > 0 {
		pHash = sha256hex(strings.Join(rec.P, "\n"))
	}
	var evHash sql.NullString
	if len(rec.Evidence) > 0 {
		evHash = sql.NullString{String: sha256hex(strings.Join(rec.Evidence, "\n---\n")), Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO gate_decisions
		  (id, gate_type, verdict, score, threshold, input_hash, p_hash, evidence_hash,
		   provider, duration_ms, created_at, reasons_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		rec.ID, rec.GateType, rec.Verdict, rec.Score, rec.Threshold,
		inputHash, pHash, evHash,
		rec.Provider, rec.DurationMs,
		time.Now().UTC().Format(time.RFC3339), rec.ReasonsJSON,
	)
	if err != nil {
		return fmt.Errorf("memory store: insert: %w", err)
	}
	return nil
}

// Count returns the number of rows for a given gate type — used by tests
// and the future audit-log UI.
func (s *Store) Count(gateType string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM gate_decisions WHERE gate_type = ?`, gateType).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("memory store: count: %w", err)
	}
	return n, nil
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return "sha256:" + hex.EncodeToString(sum[:])
}
