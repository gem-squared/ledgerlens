package brightdata_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gem-squared/ledgerlens/internal/brightdata"
)

// These tests are LIVE — they hit Bright Data's network. They are gated on
// the required env vars being set; otherwise they t.Skip. Run with:
//
//	set -a && source .env && set +a && go test -v -count=1 ./internal/brightdata/...
//
// Receipts land in artifacts/fetch_receipts/.

const receiptRoot = "../../artifacts/fetch_receipts"

func requireEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if os.Getenv(k) == "" {
			t.Skipf("env %s not set — skipping live test", k)
		}
	}
}

func TestSERP_Live(t *testing.T) {
	requireEnv(t, "BRIGHTDATA_API_TOKEN", "BRIGHTDATA_SERP_ZONE")

	store, err := brightdata.NewReceiptStore(receiptRoot)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	c := brightdata.NewSERPClient(
		os.Getenv("BRIGHTDATA_API_TOKEN"),
		os.Getenv("BRIGHTDATA_SERP_ZONE"),
		store,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	r, err := c.Search(ctx, "stripe pricing api")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if r.ReceiptID == "" || r.RawRef == "" || r.ContentHash == "" {
		t.Fatalf("receipt incomplete: %+v", r)
	}
	if r.BrightDataProduct != "SERP" {
		t.Fatalf("product=%q want SERP", r.BrightDataProduct)
	}
	t.Logf("SERP receipt: id=%s url=%s hash=%s file=%s",
		r.ReceiptID, r.URL, r.ContentHash, r.RawRef)
}

func TestUnlocker_Live(t *testing.T) {
	requireEnv(t, "BRIGHTDATA_API_TOKEN", "BRIGHTDATA_UNLOCKER_ZONE")

	store, err := brightdata.NewReceiptStore(receiptRoot)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	c := brightdata.NewUnlockerClient(
		os.Getenv("BRIGHTDATA_API_TOKEN"),
		os.Getenv("BRIGHTDATA_UNLOCKER_ZONE"),
		store,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Bright Data's own test target — small, stable, public.
	target := "https://geo.brdtest.com/welcome.txt?product=unlocker&method=api"
	r, err := c.Fetch(ctx, target)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if r.ReceiptID == "" || r.RawRef == "" || r.ContentHash == "" {
		t.Fatalf("receipt incomplete: %+v", r)
	}
	if r.BrightDataProduct != "UNLOCKER" {
		t.Fatalf("product=%q want UNLOCKER", r.BrightDataProduct)
	}
	t.Logf("UNLOCKER receipt: id=%s url=%s hash=%s file=%s",
		r.ReceiptID, r.URL, r.ContentHash, r.RawRef)
}

func TestBrowser_Live(t *testing.T) {
	requireEnv(t, "BRIGHTDATA_BROWSER_HTTPS_URL")

	store, err := brightdata.NewReceiptStore(receiptRoot)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	c, err := brightdata.NewBrowserClient(os.Getenv("BRIGHTDATA_BROWSER_HTTPS_URL"), store)
	if err != nil {
		t.Fatalf("browser client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// example.com — small, stable, public, JS-light → fast smoke test
	r, err := c.FetchPage(ctx, "https://example.com/")
	if err != nil {
		t.Fatalf("fetch page: %v", err)
	}
	if r.ReceiptID == "" || r.RawRef == "" || r.ContentHash == "" {
		t.Fatalf("receipt incomplete: %+v", r)
	}
	if r.BrightDataProduct != "BROWSER" {
		t.Fatalf("product=%q want BROWSER", r.BrightDataProduct)
	}
	// Sanity: example.com page should contain its title
	body, err := os.ReadFile(r.RawRef)
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	if !strings.Contains(strings.ToLower(string(body)), "example domain") {
		t.Fatalf("rendered page does not contain 'Example Domain' — got %d bytes", len(body))
	}
	t.Logf("BROWSER receipt: id=%s url=%s hash=%s file=%s bytes=%d",
		r.ReceiptID, r.URL, r.ContentHash, r.RawRef, len(body))
}

func TestMCP_Smoke(t *testing.T) {
	requireEnv(t, "BRIGHTDATA_API_TOKEN")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	res, err := brightdata.SmokeListTools(
		ctx,
		os.Getenv("BRIGHTDATA_API_TOKEN"),
		os.Getenv("BRIGHTDATA_MCP_COMMAND"),
		nil, // default args
	)
	if err != nil {
		t.Fatalf("mcp smoke: %v", err)
	}
	if !res.Connected {
		t.Fatalf("not connected: %+v", res)
	}
	if len(res.ToolNames) < 1 {
		t.Fatalf("expected ≥1 tool, got %d: %+v", len(res.ToolNames), res)
	}
	t.Logf("MCP smoke: server=%q tools=%d duration=%dms first=%v",
		res.ServerName, len(res.ToolNames), res.DurationMs, res.ToolNames[:min(3, len(res.ToolNames))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestAUP_Blocks(t *testing.T) {
	cases := []struct {
		url   string
		allow bool
	}{
		{"https://example.com/", true},
		{"https://example.com/login", false},
		{"https://example.com/admin/users", false},
		{"https://example.com/account/settings", false},
		{"http://example.com/", true},
		{"ftp://example.com/", false},
		{"https://user:pass@example.com/", false},
		{"not a url", false},
	}
	for _, tc := range cases {
		ok, why := brightdata.IsPublicAllowed(tc.url)
		if ok != tc.allow {
			t.Errorf("%q: allow=%v want=%v (why=%s)", tc.url, ok, tc.allow, why)
		}
	}
}
