package brightdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// UnlockerClient wraps Bright Data's Web Unlocker (POST api.brightdata.com/request).
// Same endpoint shape as SERP, different zone, different downstream behavior
// (raw HTML for arbitrary public URLs, anti-bot/CAPTCHA handled server-side).
type UnlockerClient struct {
	APIToken string
	Zone     string
	BaseURL  string
	HTTP     *http.Client
	Store    *ReceiptStore
}

func NewUnlockerClient(apiToken, zone string, store *ReceiptStore) *UnlockerClient {
	return &UnlockerClient{
		APIToken: apiToken,
		Zone:     zone,
		BaseURL:  "https://api.brightdata.com",
		HTTP:     &http.Client{Timeout: 45 * time.Second},
		Store:    store,
	}
}

// Fetch retrieves a public URL through the Web Unlocker and writes one
// EvidenceReceipt with the raw response body.
func (c *UnlockerClient) Fetch(ctx context.Context, target string) (schemas.EvidenceReceipt, error) {
	if ok, why := IsPublicAllowed(target); !ok {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: aup block: %s", why)
	}

	body := map[string]string{
		"zone":   c.Zone,
		"url":    target,
		"format": "raw",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/request", bytes.NewReader(payload))
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: do: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: read: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return schemas.EvidenceReceipt{}, fmt.Errorf("unlocker: status %d: %s", resp.StatusCode, snippet(raw))
	}

	// Heuristic: .json/.txt/.html — default to html for unlocker responses
	ext := "html"
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		switch {
		case bytes.Contains([]byte(ct), []byte("json")):
			ext = "json"
		case bytes.Contains([]byte(ct), []byte("text/plain")):
			ext = "txt"
		}
	}

	return c.Store.Write(schemas.EvidenceReceipt{
		URL:               target,
		BrightDataProduct: "UNLOCKER",
	}, raw, ext)
}
