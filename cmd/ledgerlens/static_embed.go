package main

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// staticFS embeds the Next.js static export at build time.
//
// In production: `make web-export` (or `make build-prod` / `make build-linux`)
//   populates cmd/ledgerlens/web_static/ with the contents of apps/web/out/
//   (Next.js `output: 'export'` artifact) BEFORE the Go compile.
// In dev:       cmd/ledgerlens/web_static/ contains only .gitkeep. The
//   embedded FS exists but has no index.html. mountStatic detects this and
//   installs an inline dev-banner handler. For the full interactive UI in
//   dev, run `pnpm dev` in apps/web/.
//
// NOTE: the `all:` prefix is REQUIRED — without it, Go's `//go:embed`
// excludes any directory or filename starting with `_` or `.`. Next.js
// ships all its static assets under `_next/`, so without `all:` the CSS
// and JS chunks 404 at runtime and the page renders unstyled.
//
//go:embed all:web_static
var staticFS embed.FS

// mountStatic registers a NoRoute handler that serves the embedded UI.
// It is wired AFTER all /api/* and /health routes so it does not shadow them.
//
// The handler reads bytes directly from the embed FS rather than delegating
// to http.FileServer. FileServer issues canonical-URL redirects on paths
// ending in "/index.html" (redirect target: "./") which produced an
// infinite-redirect loop on any nested directory like /pitch/ — the SPA
// fallback path set req.URL.Path="/index.html", FileServer redirected to
// "./", browser resolved "./" against the ORIGINAL "/pitch/" and looped.
// Direct read + write sidesteps the redirect dance entirely.
func mountStatic(r *gin.Engine) {
	sub, err := fs.Sub(staticFS, "web_static")
	if err != nil {
		r.NoRoute(devBannerHandler)
		return
	}
	// Detect: empty embed (no index.html). Show a dev banner instead.
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		r.NoRoute(devBannerHandler)
		return
	}

	r.NoRoute(func(c *gin.Context) {
		req := c.Request
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		urlPath := strings.TrimPrefix(req.URL.Path, "/")
		if urlPath == "" {
			urlPath = "index.html"
		}

		// 1) Directory request (path ends with /): try <dir>/index.html
		if strings.HasSuffix(urlPath, "/") {
			candidate := urlPath + "index.html"
			if _, err := fs.Stat(sub, candidate); err == nil {
				serveStatic(c, sub, candidate)
				return
			}
		}

		// 2) Exact file
		if _, err := fs.Stat(sub, urlPath); err == nil {
			// If urlPath is a directory (no trailing slash), serve <dir>/index.html
			info, _ := fs.Stat(sub, urlPath)
			if info != nil && info.IsDir() {
				if _, err := fs.Stat(sub, urlPath+"/index.html"); err == nil {
					serveStatic(c, sub, urlPath+"/index.html")
					return
				}
			} else {
				serveStatic(c, sub, urlPath)
				return
			}
		}

		// 3) Try .html suffix (Next.js export: /case/a → /case/a.html)
		if _, err := fs.Stat(sub, urlPath+".html"); err == nil {
			serveStatic(c, sub, urlPath+".html")
			return
		}

		// 4) SPA fallback — serve root index.html so client-side routes resolve
		serveStatic(c, sub, "index.html")
	})
}

// serveStatic reads name from the embed FS and writes it to the gin response.
// No FileServer, no canonical-URL redirects. Sets cache headers + content-type.
func serveStatic(c *gin.Context, sub fs.FS, name string) {
	data, err := fs.ReadFile(sub, name)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	ct := mime.TypeByExtension(filepath.Ext(name))
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	// Force browsers to revalidate on every request — relevant while the
	// hackathon UI is in active flux. Production-ready setting would
	// distinguish content-hashed /_next/static/* (immutable, long max-age)
	// from index.html (no-cache); for now blanket no-store keeps demos
	// honest after every redeploy.
	c.Writer.Header().Set("Cache-Control", "no-store, max-age=0, must-revalidate")
	c.Writer.Header().Set("Pragma", "no-cache")
	c.Writer.Header().Set("Expires", "0")
	c.Writer.Header().Set("Content-Type", ct)
	if c.Request.Method == http.MethodHead {
		c.Writer.WriteHeader(http.StatusOK)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(data)
}

const devBannerHTML = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8" />
<title>LedgerLens — dev build (UI not embedded)</title>
<style>
body{font-family:ui-sans-serif,system-ui,sans-serif;max-width:720px;margin:4rem auto;padding:0 1.5rem;color:#d4d4d8;background:#09090b}
code{font-family:ui-monospace,monospace;background:#18181b;padding:.2rem .4rem;border-radius:4px;color:#fafafa}
pre{font-family:ui-monospace,monospace;background:#18181b;padding:.75rem 1rem;border-radius:6px;overflow-x:auto}
h1{font-size:2rem;margin-bottom:.25rem}
.pill{display:inline-block;background:#6366f1;color:#fff;font-size:.7rem;padding:.2rem .6rem;border-radius:999px;font-weight:700;letter-spacing:.05em;text-transform:uppercase}
a{color:#818cf8}
</style></head>
<body>
<span class="pill">Simulation Mode · Dev build</span>
<h1>LedgerLens</h1>
<p><em>No grounded claim, no payment.</em></p>
<p>This is a dev build &mdash; the embedded UI is empty (<code>cmd/ledgerlens/web_static/</code> contains only <code>.gitkeep</code>).</p>
<p>For the full interactive UI in dev, run:</p>
<pre>cd apps/web &amp;&amp; pnpm dev</pre>
<p>Then open <a href="http://localhost:3001/">http://localhost:3001/</a>.</p>
<p>For a single-binary build with the embedded UI:</p>
<pre>make build-prod      # darwin
make build-linux     # linux/amd64 for VPS deploy</pre>
<p>The HTTP API is live on this port:</p>
<ul>
<li><a href="/health"><code>/health</code></a></li>
<li><a href="/api/v1/cases"><code>/api/v1/cases</code></a></li>
</ul>
</body></html>`

func devBannerHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(devBannerHTML))
}
