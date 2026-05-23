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

// BrowserClient drives Bright Data's Scraping Browser through its Selenium
// WebDriver REST endpoint (port 9515). No external Selenium/Playwright
// dependency — pure net/http against the W3C WebDriver protocol.
//
// The HTTPSURL form embeds credentials as userinfo:
//
//	https://brd-customer-<id>-zone-<zone>:<password>@brd.superproxy.io:9515
//
// We extract userinfo into Basic Auth headers per request because http.Client
// does not promote URL.User automatically.
type BrowserClient struct {
	endpoint *url.URL // parsed BRIGHTDATA_BROWSER_HTTPS_URL
	user     string
	pass     string
	HTTP     *http.Client
	Store    *ReceiptStore
}

func NewBrowserClient(httpsURL string, store *ReceiptStore) (*BrowserClient, error) {
	u, err := url.Parse(httpsURL)
	if err != nil {
		return nil, fmt.Errorf("browser: parse endpoint: %w", err)
	}
	if u.User == nil {
		return nil, fmt.Errorf("browser: endpoint must embed credentials as userinfo")
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	u.User = nil // strip credentials from URL; we set Basic Auth per request

	return &BrowserClient{
		endpoint: u,
		user:     user,
		pass:     pass,
		HTTP:     &http.Client{Timeout: 90 * time.Second},
		Store:    store,
	}, nil
}

// FetchPage opens a Selenium session, navigates to target, captures the page
// source, closes the session, and writes one EvidenceReceipt.
func (c *BrowserClient) FetchPage(ctx context.Context, target string) (schemas.EvidenceReceipt, error) {
	if ok, why := IsPublicAllowed(target); !ok {
		return schemas.EvidenceReceipt{}, fmt.Errorf("browser: aup block: %s", why)
	}

	sessionID, err := c.newSession(ctx)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("browser: %w", err)
	}
	// Always close the session to keep per-minute billing tight.
	defer func() { _ = c.closeSession(context.Background(), sessionID) }()

	if err := c.navigate(ctx, sessionID, target); err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("browser: %w", err)
	}
	source, err := c.pageSource(ctx, sessionID)
	if err != nil {
		return schemas.EvidenceReceipt{}, fmt.Errorf("browser: %w", err)
	}

	return c.Store.Write(schemas.EvidenceReceipt{
		URL:               target,
		BrightDataProduct: "BROWSER",
	}, []byte(source), "html")
}

// ─── WebDriver REST helpers ─────────────────────────────────────────────────

func (c *BrowserClient) newSession(ctx context.Context) (string, error) {
	body := map[string]any{
		"capabilities": map[string]any{
			"alwaysMatch": map[string]any{
				"browserName": "chrome",
			},
		},
	}
	var resp struct {
		Value struct {
			SessionID string `json:"sessionId"`
		} `json:"value"`
	}
	if err := c.do(ctx, http.MethodPost, "/session", body, &resp); err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	if resp.Value.SessionID == "" {
		return "", fmt.Errorf("new session: empty sessionId in response")
	}
	return resp.Value.SessionID, nil
}

func (c *BrowserClient) navigate(ctx context.Context, sessionID, target string) error {
	body := map[string]string{"url": target}
	return c.do(ctx, http.MethodPost, "/session/"+sessionID+"/url", body, nil)
}

func (c *BrowserClient) pageSource(ctx context.Context, sessionID string) (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := c.do(ctx, http.MethodGet, "/session/"+sessionID+"/source", nil, &resp); err != nil {
		return "", fmt.Errorf("page source: %w", err)
	}
	return resp.Value, nil
}

func (c *BrowserClient) closeSession(ctx context.Context, sessionID string) error {
	return c.do(ctx, http.MethodDelete, "/session/"+sessionID, nil, nil)
}

func (c *BrowserClient) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint.String()+path, reqBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.SetBasicAuth(c.user, c.pass)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, snippet(raw))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("unmarshal: %w (body=%s)", err, snippet(raw))
		}
	}
	return nil
}
