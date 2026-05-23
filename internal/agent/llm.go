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

// extractJSON strips common LLM JSON wrapping (markdown fences, leading prose)
// and returns the first balanced JSON object in s. Defensive against models
// that ignore "respond with only JSON" instructions.
func extractJSON(s string) string {
	// Strip markdown fence
	for _, fence := range []string{"```json\n", "```\n", "```"} {
		s = trimAround(s, fence)
	}
	// Find first { and last } — naive but robust enough for one-shot prompts.
	start := -1
	for i, r := range s {
		if r == '{' {
			start = i
			break
		}
	}
	if start < 0 {
		return s
	}
	end := -1
	for i := len(s) - 1; i >= start; i-- {
		if s[i] == '}' {
			end = i
			break
		}
	}
	if end < 0 {
		return s
	}
	return s[start : end+1]
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
