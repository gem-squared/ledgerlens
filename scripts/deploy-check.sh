#!/usr/bin/env bash
# Production smoke checks for LedgerLens. Run after `make deploy`, or
# anytime you want to verify the live deployment looks healthy.
#
# Override SERVICE_URL to point at a different host (e.g. local :8084):
#   SERVICE_URL=http://localhost:8084 bash scripts/deploy-check.sh
#
# Exit code: 0 if all checks pass, 1 if any fails.
set -uo pipefail

SERVICE_URL="${SERVICE_URL:-https://ledgerlens.gemsquared.ai}"
fail=0
pad="                  "

# check_url <label> <path-or-url> <expected-substring-or-empty> [expected-content-type]
check_url() {
  local label="$1" url="$2" expect="${3:-}" expect_ct="${4:-}"
  # If url starts with /, prepend SERVICE_URL.
  if [[ "$url" == /* ]]; then url="$SERVICE_URL$url"; fi

  local body http ct
  body=$(/usr/bin/curl -sS "$url" -m 10 -w "\n__HTTP__%{http_code}\n__CT__%{content_type}\n") || {
    printf "  FAIL  %s%s curl errored\n" "$label" "${pad:${#label}}"
    fail=$((fail+1)); return
  }
  http=$(echo "$body" | grep -oE '__HTTP__[0-9]+' | tail -1 | cut -d_ -f5)
  ct=$(echo "$body" | grep '^__CT__' | tail -1 | sed 's/^__CT__//')
  # Strip the trailing footer lines from the body before grep.
  body=$(echo "$body" | sed '/^__HTTP__/,$d')

  if [ "$http" != "200" ]; then
    printf "  FAIL  %s%s http=%s url=%s\n" "$label" "${pad:${#label}}" "$http" "$url"
    fail=$((fail+1)); return
  fi
  if [ -n "$expect" ] && ! echo "$body" | grep -q "$expect"; then
    printf "  FAIL  %s%s expected '%s' missing  http=%s\n" "$label" "${pad:${#label}}" "$expect" "$http"
    fail=$((fail+1)); return
  fi
  if [ -n "$expect_ct" ] && ! echo "$ct" | grep -qiE "$expect_ct"; then
    printf "  FAIL  %s%s ct='%s' did not match '%s'\n" "$label" "${pad:${#label}}" "$ct" "$expect_ct"
    fail=$((fail+1)); return
  fi
  printf "  OK    %s%s http=%s ct=%s\n" "$label" "${pad:${#label}}" "$http" "$ct"
}

echo "==> deploy-check at $SERVICE_URL"

# Root page
check_url "page /"        "/"                  "LedgerLens"           "text/html"

# JSON API endpoints
check_url "api/health"    "/api/v1/health"     '"status":"ok"'        "application/json"
check_url "api/cases"     "/api/v1/cases"      '"cases":'             "application/json"
check_url "api/stats"     "/api/v1/stats"      '"dealsAudited":'      "application/json"

# Embedded Next.js assets — extract the actual filenames from the HTML,
# then GET them. This catches the //go:embed all:web_static regression
# (where _next/ silently doesn't get embedded if the `all:` prefix is dropped).
HTML=$(/usr/bin/curl -fsS "$SERVICE_URL/" -m 10 2>/dev/null || true)
CSS=$(echo "$HTML" | grep -oE '/_next/static/css/[a-zA-Z0-9._/-]+\.css' | head -1)
CHUNK=$(echo "$HTML" | grep -oE '/_next/static/chunks/[a-zA-Z0-9._/-]+\.js' | head -1)

if [ -z "$CSS" ]; then
  printf "  FAIL  next.css%s no /_next/static/css ref found in HTML\n" "${pad:8}"
  fail=$((fail+1))
else
  check_url "next.css" "$CSS" "" "text/css"
fi

if [ -z "$CHUNK" ]; then
  printf "  FAIL  next.chunk%s no /_next/static/chunks ref found in HTML\n" "${pad:10}"
  fail=$((fail+1))
else
  check_url "next.chunk" "$CHUNK" "" "application/javascript|text/javascript"
fi

echo ""
if [ $fail -gt 0 ]; then
  echo "==> deploy-check FAILED ($fail check(s) failed)"
  exit 1
fi
echo "==> deploy-check OK"
