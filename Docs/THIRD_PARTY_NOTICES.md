# Third-Party Notices

LedgerLens application code is licensed under **MIT** (see [`LICENSE`](../LICENSE)).

The following third-party materials may be referenced, quoted, or invoked from this repository under their own licenses. They are listed here for compliance with the Bright Data "Web Data UNLOCKED" Hackathon submission requirement of MIT-compliant deliverables.

## Specifications quoted or referenced (CC-BY-4.0)

| Material | License | Attribution |
|---|---|---|
| **TPMN-PSL** grammar and contract conventions | CC-BY-4.0 | © GEM².AI · [github.com/gem-squared/tpmn-psl](https://github.com/gem-squared/tpmn-psl) |
| **TPMN-SKILL-STANDARD** (skill authoring standard) | CC-BY-4.0 | © GEM².AI |

Any direct quotation of these specs in `Docs/`, `internal/schemas/`, or code comments is used under CC-BY-4.0 with attribution to GEM².AI.

## Go module dependencies

LedgerLens depends on the following Go modules at runtime. Each is used under its own license; consult each module's repo for the canonical license text.

| Module | Purpose | License (typical) |
|---|---|---|
| `github.com/gin-gonic/gin` | HTTP server | MIT |
| `github.com/mark3labs/mcp-go` | MCP client (used in Unit 2) | MIT |
| `modernc.org/sqlite` | Pure-Go SQLite driver | BSD-3-Clause |

`go mod tidy` produces the authoritative list in `go.mod` / `go.sum`.

## Node / npm dependencies (Next.js app)

`apps/web/package.json` declares the canonical list. Major direct deps:

| Package | License (typical) |
|---|---|
| `next` | MIT |
| `react`, `react-dom` | MIT |
| `tailwindcss`, `postcss`, `autoprefixer` | MIT |
| `typescript` | Apache-2.0 |

## External services called at runtime (proprietary)

The following SaaS is **called** by LedgerLens but is **not** part of this MIT-licensed source:

- **GEM² Truth-Filter SaaS** at `https://gem2-tpmn-checker.fly.dev` — proprietary GEM².AI infrastructure. LedgerLens uses the `/api/audit-gate/p-check` and `/api/audit-gate/o-check` endpoints under David's `GEM2_API_KEY`. See the GEM² audit-gate API spec (available from GEM².AI) for the documented surface.
- **Bright Data** APIs (SERP, Web Unlocker, Browser, MCP, Scraper) — proprietary Bright Data infrastructure, used under the participant terms of the Web Data UNLOCKED Hackathon.

## Post-hackathon references (NOT shipped in this repo)

The following are referenced in design docs but are explicitly **excluded** from the hackathon `go.mod`:

- `github.com/coinbase/x402/go` — official Go SDK for x402. Cited as the future `CoinbaseX402Settler` implementation path; not imported in the hackathon build.
- `github.com/mark3labs/mcp-go-x402` — x402 transport for MCP-Go. Cited as a future MCP-tool-pay extension; not imported in the hackathon build.

## Attribution requests

If a third-party material is missing from this notice, please open an issue or email `david@gineers.ai`.
