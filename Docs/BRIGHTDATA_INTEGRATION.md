# Bright Data Integration

Per-product role assignment for LedgerLens. Filled in Unit 2 (live-first).

**Status (2026-05-23):** all four products LIVE and verified via `go test ./internal/brightdata/...`.

## Product map

| Bright Data product | Zone (this account) | Role | Go code path | Smoke status |
|---|---|---|---|---|
| **SERP API** | `gem2_serp_api1` | Discovery — find candidate evidence pages, status / pricing / docs URLs for a seller's claim | [`internal/brightdata/serp.go`](../internal/brightdata/serp.go) | ⊢ PASS — receipt emitted, 3.6s |
| **Web Unlocker** | `gem2_web_unlocker1` | Static retrieval — trust / pricing / docs pages, anti-bot handled server-side | [`internal/brightdata/unlocker.go`](../internal/brightdata/unlocker.go) | ⊢ PASS — receipt emitted, 1.1s |
| **Scraping Browser** | `gem2_scraping_browser1` | Interactive / JS-heavy pages — live dashboard verification | [`internal/brightdata/browser.go`](../internal/brightdata/browser.go) | ⊢ PASS — receipt emitted, 13.2s (incl. session lifecycle) |
| **MCP Server** | (npm `@brightdata/mcp` v2.9.5+) | Live agent investigation surface | [`internal/brightdata/mcp_client.go`](../internal/brightdata/mcp_client.go) | ⊢ PASS — 5 tools listed (`search_engine`, `scrape_as_markdown`, `search_engine_batch`, …), 2.2s after first install |

## Endpoints used

| Product | Method · URL |
|---|---|
| SERP API | POST `https://api.brightdata.com/request`  body: `{zone, url, format: "raw"}` (Google URL pinned with `gl=us&hl=en`) |
| Web Unlocker | POST `https://api.brightdata.com/request`  body: `{zone, url, format: "raw"}` |
| Scraping Browser | Selenium WebDriver REST at `https://brd.superproxy.io:9515`. Session lifecycle: `POST /session` → `POST /session/{id}/url` → `GET /session/{id}/source` → `DELETE /session/{id}` |
| MCP Server | child process `npx -y @brightdata/mcp` (stdio); env `API_TOKEN=<token>`; protocol: `mark3labs/mcp-go` v0.54.0 |

Auth: API token (Bearer) for SERP + Unlocker; basic-auth (zone user/pass embedded in URL, extracted by `NewBrowserClient` and forwarded via `req.SetBasicAuth`) for Browser; env `API_TOKEN` for MCP.

## Receipt shape (emitted to `artifacts/fetch_receipts/receipts.jsonl`)

```json
{
  "receiptId":         "ev_<random hex>",
  "url":               "https://example.com/",
  "brightDataProduct": "BROWSER",
  "fetchedAt":         "2026-05-23T10:53:43Z",
  "contentHash":       "sha256:ee1b911b993f8ea72d99afa57352871948b6d2f7d7a535615f3c85e5dd235e2b",
  "rawRef":            "artifacts/fetch_receipts/ev_<random hex>.html"
}
```

Each receipt has a sibling raw-body file at `rawRef` (`.html` / `.json` / `.txt` depending on product). The full set for a given decision becomes the `evidence: []string` payload sent to `gem2-tpmn-checker.fly.dev` in Unit 3.

## AUP posture (Go layer)

[`internal/brightdata/aup.go`](../internal/brightdata/aup.go) — `IsPublicAllowed(url)` runs **before any Bright Data call**. Rejected categories:

- URL missing scheme/host or not http/https
- URL embedding userinfo (no credentials in caller URLs; the Browser endpoint userinfo is extracted by `NewBrowserClient` BEFORE the URL reaches the public API surface)
- Path matches a blocked prefix: `/login`, `/signin`, `/sign-in`, `/admin`, `/account`, `/private`, `/auth`, `/oauth`, `/sso`, `/wp-admin`

Bright Data's server-side AUP is the authoritative guard; this Go-layer policy is a fast pre-filter and a documented intent.

## Re-running the smoke

```bash
set -a && source .env && set +a
go test -v -count=1 -timeout 300s ./internal/brightdata/...
```

Required env (in gitignored `.env`):

- `BRIGHTDATA_API_TOKEN` — bearer token for SERP + Unlocker + MCP (`API_TOKEN`)
- `BRIGHTDATA_SERP_ZONE` — e.g. `gem2_serp_api1`
- `BRIGHTDATA_UNLOCKER_ZONE` — e.g. `gem2_web_unlocker1`
- `BRIGHTDATA_BROWSER_HTTPS_URL` — full WebDriver URL with embedded userinfo
- `BRIGHTDATA_MCP_COMMAND` (optional, default `npx`)

## Security notes

- ⊢ No API token or zone password is ever logged or stored in receipts. Receipts contain only public URL + content hash + raw-body file ref.
- ⊢ Receipts (`receipts.jsonl` + raw-body files) are gitignored — `artifacts/fetch_receipts/*` (with `.gitkeep` exempted). Source code is the only thing committed.
- ⊢ The `internal/brightdata/browser.go` client strips userinfo from the endpoint URL on `NewBrowserClient` construction and forwards via Basic-Auth header per request — credentials are not present in the in-flight URL string.

## References

- [Bright Data MCP repo](https://github.com/brightdata-com/brightdata-mcp)
- [Bright Data MCP docs](https://docs.brightdata.com/ai/mcp-server/overview)
- `Docs/Bright-Data-winning-strategy.md` §7 (canonical integration plan)
- `WP-ST-1.md` Unit 2 (live-first amendment)
