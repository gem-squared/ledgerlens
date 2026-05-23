# Bright Data Integration

Per-product role assignment for LedgerLens.

> ⚠ **SKELETON (Unit 1).** Filled in Unit 2 (Bright Data ingestion layer).
> Until Unit 2 lands, this document records the *intent* per the strategy doc §7.

## Product map

| Bright Data product | Role in LedgerLens | Go code path | Demo moment |
|---|---|---|---|
| **SERP API** | Discovery — find candidate evidence pages, status/pricing/docs URLs for a seller's claim | `internal/brightdata/serp.go` | Pre-flight: "where does the seller's claim live on the public web?" |
| **Web Unlocker** | Static retrieval — trust/pricing/docs pages | `internal/brightdata/unlocker.go` | Case A blocked path: fetches a 6-month-old mirror that L1 marks `[SPT-S→T]` |
| **Browser API** | Interactive / JS-heavy pages | `internal/brightdata/browser.go` | Case B approved path: verifies a JS-rendered live dashboard for the NYSE+NASDAQ feed |
| **MCP Server** | Live agent investigation surface | `internal/brightdata/mcp_client.go` (via `github.com/mark3labs/mcp-go`) | Buyer-agent dialogue: agent invokes Bright Data tools mid-conversation to enrich a claim |
| **Web Scraper API** (optional) | Structured extraction from supported sources | `internal/brightdata/scraper.go` | LinkedIn / Crunchbase / GitHub seller-profile sanity check (if time permits) |

## Sample receipt shape (will be emitted in Unit 2)

```json
{
  "receiptId":         "ev_<uuid8>",
  "url":               "https://example.com/status",
  "brightDataProduct": "BROWSER",
  "fetchedAt":         "2026-05-26T14:22:01Z",
  "contentHash":       "sha256:<hex>",
  "rawRef":            "artifacts/fetch_receipts/ev_<uuid8>.html"
}
```

Receipts are append-only JSONL in `artifacts/fetch_receipts/`. The full set for a given decision becomes the `evidence: []string` payload sent to `gem2-tpmn-checker.fly.dev` in Unit 3.

## AUP posture

- ⊢ Public-only sources. No login attempts, no credentialed access, no nonpublic data.
- ⊢ robots-aware fetch policy in `internal/brightdata/aup.go`.
- ⊢ Source-tier metadata captured per-call for the audit bundle.

## References

- [Bright Data MCP repo](https://github.com/brightdata/brightdata-mcp)
- [Bright Data MCP docs](https://docs.brightdata.com/ai/mcp-server/overview)
- `Docs/Bright-Data-winning-strategy.md` §7 (canonical integration plan)
