// Package brightdata wraps Bright Data products into Go interfaces that emit
// canonical schemas.EvidenceReceipt records into artifacts/fetch_receipts/.
//
// Four products are wrapped:
//
//	SERP API           gem2_serp_api1        → serp.Search
//	Web Unlocker       gem2_web_unlocker1    → unlocker.Fetch
//	Scraping Browser   gem2_scraping_browser1 → browser.FetchPage
//	MCP Server         npx @brightdata/mcp-server → mcp_client.SmokeListTools
//
// All page fetches pass through aup.IsPublicAllowed first — login-required,
// admin, and other nonpublic paths are blocked at the Go layer regardless of
// what Bright Data would permit. This is the AUP-aware source policy.
package brightdata
