package brightdata

import (
	"net/url"
	"strings"
)

// IsPublicAllowed enforces LedgerLens's public-only source policy at the Go
// layer, BEFORE any Bright Data call. A failing URL never reaches the network.
//
// The rule (per Docs/LEGAL.md): no login-required collection, no nonpublic
// data, no exploit verification, no credential use, no active probing. We
// block by path-segment hint to catch the common cases; the real Bright Data
// AUP is enforced server-side regardless.
func IsPublicAllowed(rawURL string) (bool, string) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false, "url: not a valid absolute URL"
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false, "url: only http/https allowed"
	}
	if u.User != nil {
		return false, "url: must not embed credentials (userinfo present)"
	}

	lowerPath := strings.ToLower(u.Path)
	for _, blocked := range blockedPathPrefixes {
		if strings.HasPrefix(lowerPath, blocked) {
			return false, "path: matches login/admin/private/account pattern: " + blocked
		}
	}
	return true, ""
}

// blockedPathPrefixes are common login/private paths we refuse to fetch.
// Not exhaustive — Bright Data's server-side AUP is the authoritative guard.
var blockedPathPrefixes = []string{
	"/login",
	"/signin",
	"/sign-in",
	"/admin",
	"/account",
	"/private",
	"/auth",
	"/oauth",
	"/sso",
	"/wp-admin",
}
