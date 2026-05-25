package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Anthropic Messages API client — minimal stdlib HTTP.
// We avoid a third-party SDK to keep the binary lean and the dependency
// surface auditable. The Messages API is stable enough for this use.

const (
	anthropicURL     = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"

	// Haiku 4.5 — cheap (~$0.001/call for our prompt sizes), fast (~2-3s).
	// Used for intent extraction + seller-offer synthesis.
	ModelHaiku = "claude-haiku-4-5-20251001"

	// Sonnet 4.6 — reserved for higher-quality synthesis if Haiku produces
	// unreliable JSON. Not used in Slice 1.
	ModelSonnet = "claude-sonnet-4-6"
)

type llmRequest struct {
	Model     string       `json:"model"`
	MaxTokens int          `json:"max_tokens"`
	System    string       `json:"system,omitempty"`
	Messages  []llmMessage `json:"messages"`
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason,omitempty"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// callAnthropic posts a single-user-message completion. Returns the model's
// text content (joined) and basic usage info. Caller is responsible for any
// JSON parsing of structured outputs.
func callAnthropic(ctx context.Context, apiKey, model, system, userMsg string, maxTokens int) (string, *llmResponse, error) {
	if apiKey == "" {
		return "", nil, fmt.Errorf("agent: ANTHROPIC_API_KEY not set")
	}
	body := llmRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []llmMessage{{Role: "user", Content: userMsg}},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", nil, fmt.Errorf("agent: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicURL, bytes.NewReader(raw))
	if err != nil {
		return "", nil, fmt.Errorf("agent: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("agent: do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("agent: read: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return "", nil, fmt.Errorf("agent: anthropic %d: %s", resp.StatusCode, snippet(respBody))
	}
	var parsed llmResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", nil, fmt.Errorf("agent: parse: %w (body=%s)", err, snippet(respBody))
	}
	text := ""
	for _, c := range parsed.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	return text, &parsed, nil
}

func snippet(b []byte) string {
	const n = 240
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}

// extractJSON returns the FIRST balanced top-level JSON object in s.
// Strips markdown fences, then walks character-by-character tracking
// brace depth (skipping string contents) so multi-object responses
// (e.g., model returns "{...},{...}") yield only the first object.
func extractJSON(s string) string {
	// Strip markdown fence wrappers if present.
	for _, fence := range []string{"```json\n", "```\n", "```"} {
		s = trimAround(s, fence)
	}
	// Find first '{'
	start := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '{' {
			start = i
			break
		}
	}
	if start < 0 {
		return s
	}
	// Walk forward, tracking brace depth and string state, until we close
	// the first object. Skip braces inside JSON strings.
	depth := 0
	inStr := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	// Unclosed — return what we have for a clearer parse error.
	return s[start:]
}

func trimAround(s, sub string) string {
	for {
		i := indexOf(s, sub)
		if i < 0 {
			return s
		}
		s = s[:i] + s[i+len(sub):]
	}
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
