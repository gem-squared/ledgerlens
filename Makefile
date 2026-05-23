.PHONY: help schemas tidy build run web-install web-dev web-build web-export build-prod build-linux clean preflight

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
	@echo "  clean         remove build artifacts (no source changes)"

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
	@# The cp -R overwrites index.html + adds _next/, 404.html, etc. Those build
	@# artifacts are gitignored.
	mkdir -p cmd/ledgerlens/web_static
	cp -R apps/web/out/. cmd/ledgerlens/web_static/

build-prod: web-export
	@echo "==> go build (darwin/host) → bin/ledgerlens"
	mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o bin/ledgerlens ./cmd/ledgerlens

build-linux: web-export
	@echo "==> cross-compile linux/amd64 → bin/ledgerlens-linux"
	mkdir -p bin
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/ledgerlens-linux ./cmd/ledgerlens
	@ls -la bin/ledgerlens-linux

clean:
	rm -rf bin/ dist/ apps/web/.next apps/web/out
	@echo "cleaned build artifacts"
