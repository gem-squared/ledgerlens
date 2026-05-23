// Package auditgate is the Go client for the deployed gem2-tpmn-checker
// audit-gate API (v1.1, see Gem-squared-AI/gem2-TPMN-checker/AUDIT_GATE_API.md).
//
// It exposes the two production endpoints:
//
//	POST /api/audit-gate/p-check   L1 — Pre-execution gate (ALLOW/DENY)
//	POST /api/audit-gate/o-check   L2 — Post-execution gate (SUCCESS/FAILURE)
//
// Both gates are LLM-agnostic; LedgerLens forwards a per-request Anthropic
// (or Gemini / OpenAI) key. The upstream gate stores hash-only audit rows;
// we mirror the same shape locally in internal/trustgate/memory.
//
// REPLAY mode: if the upstream gate is unreachable mid-demo, the client
// falls back to a cached response in artifacts/demo_cases/{case}/{p,o}_check.json.
// The fallback is opt-in via ReplayStore wiring on the Client.
package auditgate

import "encoding/json"

// LLMKeySet bundles per-request LLM provider keys. Exactly one should be
// set; the gem2-tpmn-checker upstream will use it to drive the judge LLM.
type LLMKeySet struct {
	Anthropic string `json:"anthropic_api_key,omitempty"`
	Gemini    string `json:"gemini_api_key,omitempty"`
	OpenAI    string `json:"openai_api_key,omitempty"`
}

// PCheckRequest mirrors AUDIT_GATE_API.md §"L1 P-check / Request".
// `A` accepts either a string description ("TransferRequest{from, to}") or
// a JSON schema object — we use json.RawMessage so callers pass whichever.
type PCheckRequest struct {
	I              string          `json:"i"`
	A              json.RawMessage `json:"a"`
	P              []string        `json:"p"`
	T              int             `json:"t"`
	Evidence       []string        `json:"evidence,omitempty"`
	SessionContext string          `json:"session_context,omitempty"`
	Provider       string          `json:"provider,omitempty"`
	Grammar        string          `json:"grammar,omitempty"`
}

// OCheckRequest mirrors AUDIT_GATE_API.md §"L2 O-check / Request".
type OCheckRequest struct {
	O              string          `json:"o"`
	B              json.RawMessage `json:"b"`
	P              []string        `json:"p"`
	T              int             `json:"t"`
	Evidence       []string        `json:"evidence,omitempty"`
	SessionContext string          `json:"session_context,omitempty"`
	Provider       string          `json:"provider,omitempty"`
	Grammar        string          `json:"grammar,omitempty"`
}

// PCheckResponse / OCheckResponse share shape but differ in Verdict enum.
type PCheckResponse struct {
	Verdict string   `json:"verdict"` // ALLOW | DENY
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
	Meta    Meta     `json:"meta"`
}

type OCheckResponse struct {
	Verdict string   `json:"verdict"` // SUCCESS | FAILURE
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
	Meta    Meta     `json:"meta"`
}

type Meta struct {
	ResultID   string `json:"result_id"`
	DurationMs int64  `json:"duration_ms"`
	Usage      Usage  `json:"usage,omitempty"`
}

type Usage struct {
	Provider           string  `json:"provider,omitempty"`
	Model              string  `json:"model,omitempty"`
	LLMCalls           int     `json:"llm_calls,omitempty"`
	TotalInputTokens   int     `json:"total_input_tokens,omitempty"`
	TotalOutputTokens  int     `json:"total_output_tokens,omitempty"`
	EstimatedCostUSD   float64 `json:"estimated_cost_usd,omitempty"`
}

// PCheckGO is the caller-side GO/STOP rule from AUDIT_GATE_API.md.
func (r *PCheckResponse) GO(threshold int) bool {
	return r.Verdict == "ALLOW" && r.Score >= threshold
}

// OCheckGO is the caller-side GO/STOP rule.
func (r *OCheckResponse) GO(threshold int) bool {
	return r.Verdict == "SUCCESS" && r.Score >= threshold
}
