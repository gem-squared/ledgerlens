# ── root guard ────────────────────────────────────────────────────────────
# Refuse to run if invoked from any directory other than the Makefile's own
# directory. Prevents the `cd apps/web && make build-linux` style mistake
# that caused two 30s outages on 2026-05-25 (Makefile not found → recipe
# silently skipped → subsequent scp/ssh chain destroyed the live binary).
MAKEFILE_DIR := $(abspath $(patsubst %/,%,$(dir $(lastword $(MAKEFILE_LIST)))))
ifneq ($(abspath $(CURDIR)),$(MAKEFILE_DIR))
$(error make must be run from the project root ($(MAKEFILE_DIR)); currently in $(CURDIR))
endif

# ── deploy config (override via env if needed) ────────────────────────────
VPS_HOST    ?= root@173.199.92.236
VPS_DIR     ?= /opt/gem2-ledgerlens
VPS_SERVICE ?= gem2-ledgerlens
SERVICE_URL ?= https://ledgerlens.gemsquared.ai
SSH_KEY     ?= $(HOME)/.ssh/id_ed25519_aio_deploy
LINUX_BIN   := bin/ledgerlens-linux

.PHONY: help preflight schemas tidy build run \
        web-install web-dev web-build web-export \
        build-prod build-linux \
        deploy deploy-check \
        clean

help:
	@echo "LedgerLens — make targets"
	@echo ""
	@echo "  preflight     check Go/Node/pnpm/tygo versions"
	@echo "  schemas       regenerate packages/contracts-ts/types.ts from Go structs via tygo"
	@echo "  tidy          go mod tidy"
	@echo "  build         go build ./..."
	@echo "  run           go run ./cmd/ledgerlens"
	@echo "  web-install   cd apps/web && pnpm install"
	@echo "  web-dev       cd apps/web && pnpm dev"
	@echo "  web-build     cd apps/web && pnpm build"
	@echo "  web-export    Next.js static export → cmd/ledgerlens/web_static/"
	@echo "  build-prod    web-export + native binary  → bin/ledgerlens"
	@echo "  build-linux   web-export + linux/amd64    → bin/ledgerlens-linux"
	@echo "  deploy        build-linux + atomic upload + restart + health check + rollback on fail"
	@echo "  deploy-check  smoke test production: /, /api/v1/{health,cases,stats}, _next/{css,chunks}"
	@echo "  clean         remove build artifacts"
	@echo ""
	@echo "Deploy config (env overrides):"
	@echo "  VPS_HOST=$(VPS_HOST)"
	@echo "  VPS_DIR=$(VPS_DIR)"
	@echo "  VPS_SERVICE=$(VPS_SERVICE)"
	@echo "  SERVICE_URL=$(SERVICE_URL)"

preflight:
	@echo "Go:    $$(go version 2>&1 || echo MISSING)"
	@echo "Node:  $$(node --version 2>&1 || echo MISSING)"
	@echo "pnpm:  $$(pnpm --version 2>&1 || echo MISSING)"
	@echo "git:   $$(git --version 2>&1 || echo MISSING)"
	@echo "tygo:  $$(command -v tygo 2>&1 || echo MISSING - run: go install github.com/gzuidhof/tygo@latest)"

schemas:
	@command -v tygo >/dev/null 2>&1 || { echo "tygo missing — run: go install github.com/gzuidhof/tygo@latest"; exit 1; }
	tygo generate

tidy:
	go mod tidy

build: tidy
	go build ./...

run:
	go run ./cmd/ledgerlens

web-install:
	cd apps/web && pnpm install

web-dev:
	cd apps/web && pnpm dev

web-build:
	cd apps/web && pnpm build

web-export:
	@echo "==> exporting Next.js to apps/web/out/ (static)"
	cd apps/web && NEXT_OUTPUT_MODE=export pnpm build
	@echo "==> syncing apps/web/out/ → cmd/ledgerlens/web_static/"
	@# Overlay copy: do NOT rm -rf cmd/ledgerlens/web_static (it holds the committed
	@# placeholder index.html so `go build` works without first running the export).
	mkdir -p cmd/ledgerlens/web_static
	cp -R apps/web/out/. cmd/ledgerlens/web_static/
	@# Post-export sanity — the //go:embed line needs _next/ to be present.
	@if [ ! -d cmd/ledgerlens/web_static/_next ]; then \
	  echo "FATAL: web-export finished but cmd/ledgerlens/web_static/_next/ missing"; \
	  exit 1; \
	fi
	@echo "==> web-export OK (_next/ present)"

build-prod: web-export
	@echo "==> go build (darwin/host) → bin/ledgerlens"
	mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o bin/ledgerlens ./cmd/ledgerlens

build-linux: web-export
	@echo "==> cross-compile linux/amd64 → $(LINUX_BIN)"
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(LINUX_BIN) ./cmd/ledgerlens
	@ls -la $(LINUX_BIN)
	@# Post-build embed sanity — fail fast if the binary doesn't contain
	@# the Next.js static assets. Catches //go:embed dropping the `all:`
	@# prefix (which silently skips paths starting with _ or .).
	@# Use `grep -aF` directly on the binary, not `strings | grep -q`:
	@# under `set -o pipefail` (used in scripts/deploy.sh), grep -q's
	@# early-exit SIGPIPEs strings, which pipefail then reports as a
	@# pipeline failure even on a successful match.
	@if ! grep -aFq "_next/static/css/" $(LINUX_BIN); then \
	  echo "FATAL: $(LINUX_BIN) missing embedded _next/static/css assets"; \
	  echo "       (//go:embed should be 'all:web_static' — see cmd/ledgerlens/static_embed.go)"; \
	  exit 1; \
	fi
	@if ! grep -aFq "_next/static/chunks/" $(LINUX_BIN); then \
	  echo "FATAL: $(LINUX_BIN) missing embedded _next/static/chunks assets"; \
	  exit 1; \
	fi
	@echo "==> embed check: _next/static/{css,chunks} present in binary"

deploy: build-linux
	@echo ""
	@echo "==> deploy → $(VPS_HOST):$(VPS_DIR) (service: $(VPS_SERVICE))"
	@LINUX_BIN=$(LINUX_BIN) VPS_HOST=$(VPS_HOST) VPS_DIR=$(VPS_DIR) \
	  VPS_SERVICE=$(VPS_SERVICE) SERVICE_URL=$(SERVICE_URL) SSH_KEY=$(SSH_KEY) \
	  bash scripts/deploy.sh
	@echo ""
	@$(MAKE) -s deploy-check

deploy-check:
	@SERVICE_URL=$(SERVICE_URL) bash scripts/deploy-check.sh

clean:
	rm -rf bin/ dist/ apps/web/.next apps/web/out
	@echo "cleaned build artifacts"
