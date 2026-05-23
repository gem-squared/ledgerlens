// Package brightdata wraps Bright Data products (SERP, Web Unlocker, Browser
// API, MCP Server, optional Web Scraper API) into a single Go interface that
// emits canonical EvidenceReceipt records.
//
// Implementation lands in Unit 2 of WP-ST-1. This file exists in Unit 1 so the
// package compiles and so that github.com/mark3labs/mcp-go is anchored in the
// module's dependency graph for the MCP client work that follows.
package brightdata

import (
	// Anchored for Unit 2 — internal/brightdata/mcp_client.go will use this.
	_ "github.com/mark3labs/mcp-go/mcp"
)
