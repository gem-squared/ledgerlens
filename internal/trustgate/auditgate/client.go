package auditgate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client posts P-check / O-check requests against gem2-tpmn-checker.
type Client struct {
	BaseURL    string // e.g. https://gem2-tpmn-checker.fly.dev
	GEM2APIKey string
	HTTP       *http.Client

	// Replay enables read-through from cached responses on live failure.
	// Set to nil for live-only behavior.
	Replay *ReplayStore
}

func NewClient(baseURL, gem2APIKey string) *Client {
	if baseURL == "" {
		baseURL = "https://gem2-tpmn-checker.fly.dev"
	}
	return &Client{
		BaseURL:    baseURL,
		GEM2APIKey: gem2APIKey,
		HTTP:       &http.Client{Timeout: 60 * time.Second},
	}
}

// ErrUpstreamUnavailable signals a network / 5xx upstream failure for which
// REPLAY mode is appropriate. Configurable callers may type-assert to decide
// on fallback policy.
var ErrUpstreamUnavailable = errors.New("auditgate: upstream unavailable")

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	if c.GEM2APIKey == "" {
		return errors.New("auditgate: GEM2_API_KEY not set")
	}

	// Inject gem2_api_key into the request body (gate auth pattern, per
	// AUDIT_GATE_API.md — Bearer header is NOT used by this endpoint).
	merged, err := mergeAuthIntoBody(body, c.GEM2APIKey)
	if err != nil {
		return fmt.Errorf("auditgate: merge auth: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(merged))
	if err != nil {
		return fmt.Errorf("auditgate: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("auditgate: read: %w", err)
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("%w: status %d body=%s", ErrUpstreamUnavailable, resp.StatusCode, snippet(raw))
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("auditgate: status %d: %s", resp.StatusCode, snippet(raw))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("auditgate: unmarshal: %w (body=%s)", err, snippet(raw))
		}
	}
	return nil
}

// mergeAuthIntoBody marshals body, inserts gem2_api_key, and returns
// the merged JSON. Caller body must marshal to a JSON object.
func mergeAuthIntoBody(body any, gem2APIKey string) ([]byte, error) {
	first, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(first, &obj); err != nil {
		return nil, fmt.Errorf("body did not marshal to a JSON object: %w", err)
	}
	keyJSON, _ := json.Marshal(gem2APIKey)
	obj["gem2_api_key"] = keyJSON
	return json.Marshal(obj)
}

func snippet(b []byte) string {
	const n = 240
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}
