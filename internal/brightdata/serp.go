package brightdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// SERPClient wraps Bright Data's SERP API (POST api.brightdata.com/request).
// Zone is something like "gem2_serp_api1"; APIToken is the customer-level token.
type SERPClient struct {
	APIToken string
	Zone     string
	BaseURL  string // default https://api.brightdata.com
	HTTP     *http.Client
	Store    *ReceiptStore
}

func NewSERPClient(apiToken, zone string, store *ReceiptStore) *SERPClient {
	return &SERPClient{
		APIToken: apiToken,
		Zone:     zone,
		BaseURL:  "https://api.brightdata.com",
		HTTP:     &http.Client{Timeout: 30 * time.Second},
		Store:    store,
	}
}

// Search issues a Google SERP query through Bright Data and writes one
// EvidenceReceipt. Geo is pinned to US/English for deterministic demos.
func (c *SERPClient) Search(ctx context.Context, query string) (schemas.EvidenceReceipt, error) {
	target := buildGoogleURL(query)
	if ok, why := IsPublicAllowed(target); !ok {
		return schemas.EvidenceReceipt{}, fmt.Errorf("serp: aup block: %s", why)
	}

	body := map[string]string{
		"zone":   c.Zone,
		"url":    target,
		"format": "raw",
	}
	raw, err := c.post(ctx, body)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("serp: %w", err)
	}

	return c.Store.Write(schemas.EvidenceReceipt{
		URL:               target,
		BrightDataProduct: "SERP",
	}, raw, "json")
}

func (c *SERPClient) post(ctx context.Context, body map[string]string) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/request", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, snippet(raw))
	}
	return raw, nil
}

// buildGoogleURL pins geo to gl=us&hl=en so demos return deterministic
// English-language results regardless of Bright Data's egress IP location.
func buildGoogleURL(query string) string {
	v := url.Values{}
	v.Set("q", query)
	v.Set("gl", "us")
	v.Set("hl", "en")
	return "https://www.google.com/search?" + v.Encode()
}

func snippet(b []byte) string {
	const n = 200
	if len(b) > n {
		return string(b[:n]) + "..."
	}
	return string(b)
}
