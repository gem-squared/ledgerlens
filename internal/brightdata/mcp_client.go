package brightdata

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCPSmokeResult is the outcome of the Unit 2 MCP smoke test:
// connect to the Bright Data MCP server over stdio, initialize the protocol,
// list tools, and report which tools are advertised. We do not invoke any
// tool here (that comes later in Unit 5 for the live buyer-agent path).
type MCPSmokeResult struct {
	Connected   bool     `json:"connected"`
	ToolNames   []string `json:"toolNames"`
	DurationMs  int64    `json:"durationMs"`
	ServerName  string   `json:"serverName,omitempty"`
}

// SmokeListTools spawns `npx -y @brightdata/mcp-server` as a child process,
// connects over stdio via mark3labs/mcp-go, runs Initialize + ListTools, and
// returns the advertised tool names. The Bright Data MCP server requires
// API_TOKEN in its environment.
func SmokeListTools(ctx context.Context, apiToken string, npmCommand string, npmArgs []string) (MCPSmokeResult, error) {
	if apiToken == "" {
		return MCPSmokeResult{}, fmt.Errorf("mcp: empty Bright Data API token")
	}
	if npmCommand == "" {
		npmCommand = "npx"
	}
	if len(npmArgs) == 0 {
		// Canonical Bright Data MCP npm package: @brightdata/mcp (v2.9.5+).
		// (Earlier docs referenced @brightdata/mcp-server, which never existed
		// in the npm registry; LedgerLens uses the correct canonical name.)
		npmArgs = []string{"-y", "@brightdata/mcp"}
	}

	start := time.Now()

	stdioTransport := transport.NewStdio(
		npmCommand,
		[]string{"API_TOKEN=" + apiToken},
		npmArgs...,
	)
	c := client.NewClient(stdioTransport)

	if err := c.Start(ctx); err != nil {
		return MCPSmokeResult{}, fmt.Errorf("mcp: start stdio transport: %w", err)
	}
	defer c.Close()

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "ledgerlens",
		Version: "0.1.0",
	}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	initResp, err := c.Initialize(ctx, initReq)
	if err != nil {
		return MCPSmokeResult{}, fmt.Errorf("mcp: initialize: %w", err)
	}

	listResp, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return MCPSmokeResult{}, fmt.Errorf("mcp: list tools: %w", err)
	}

	names := make([]string, 0, len(listResp.Tools))
	for _, t := range listResp.Tools {
		names = append(names, t.Name)
	}

	return MCPSmokeResult{
		Connected:  true,
		ToolNames:  names,
		DurationMs: time.Since(start).Milliseconds(),
		ServerName: initResp.ServerInfo.Name,
	}, nil
}
