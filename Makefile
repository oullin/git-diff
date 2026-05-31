SHELL := /bin/bash

ROOT_PATH := $(shell pwd)
UI_DIR := packages/ui
PORTLESS_APP_NAME := git-diff-ui
PORTLESS_DEFAULT_DEV_SERVER_URL := https://$(PORTLESS_APP_NAME).localhost
FMT_IMAGE := ghcr.io/oullin/go-fmt:v0.2.7
FMT_RUN := docker run --rm \
	-v $(ROOT_PATH):/work \
	-v go-fmt-cache:/cache \
	-w /work \
	-e HOST_PROJECT_PATH=$(ROOT_PATH) \
	-e GOCACHE=/cache/go-build \
	-e GOPATH=/cache/gopath \
	-e GOMODCACHE=/cache/gopath/pkg/mod \
	$(FMT_IMAGE)
TS_GLOBS := '*.ts' '*.tsx' '*.vue' '*.mts' '*.cts'
APP_DATA_DIR := $$HOME/Library/Application Support/git-diff
TURBO_CACHE_DIR := storage/.cache/turbo

API_DIR := packages/api
API_COVERAGE_OUT := $(API_DIR)/coverage.out
API_COVERAGE_FLOOR := 60.0

.DEFAULT_GOAL := help
.PHONY: help dev format format-all format-login fresh test-api test-api-cover

define run_ts_fmt
@files=$$(git ls-files $(1) --exclude-standard -- $(TS_GLOBS) | while IFS= read -r f; do [ -f "$$f" ] && echo "$$f"; done); \
	if [ -z "$$files" ]; then echo "No TS/Vue files to format."; else echo "$$files" | xargs $(FMT_RUN) ts; fi
endef

help: ## Show this help (list of make targets)
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run the UI dev server (https://git-diff-ui.localhost)
	@echo "Dev server will be available at: $(PORTLESS_DEFAULT_DEV_SERVER_URL)"
	@pnpm --dir $(UI_DIR) run dev

format: ## Format changed Go + TS/Vue sources via go-fmt
	@echo "go-fmt: Go formatting (workspace)"
	@$(FMT_RUN) go format
	@echo "go-fmt: TS/Vue formatting (changed files)"
	$(call run_ts_fmt,--others --modified)

format-all: ## Format the entire repo (Go + TS/Vue) via go-fmt
	@echo "go-fmt: formatting whole repository"
	@$(FMT_RUN) format

format-login:
	@gh auth token | docker login ghcr.io -u $$(gh api user -q .login) --password-stdin

test-api: ## Run the packages/api Go test suite with -race
	@cd $(API_DIR) && \
		GOCACHE=$(ROOT_PATH)/storage/.cache/go-build \
		GOPATH=$(ROOT_PATH)/storage/.cache/gopath \
		go test -race ./...

test-api-cover: ## Run packages/api tests with coverage; fails below $(API_COVERAGE_FLOOR)%
	@cd $(API_DIR) && \
		GOCACHE=$(ROOT_PATH)/storage/.cache/go-build \
		GOPATH=$(ROOT_PATH)/storage/.cache/gopath \
		go test -race -covermode=atomic -coverprofile=coverage.out ./...
	@cd $(API_DIR) && \
		go tool cover -func=coverage.out | tail -1
	@cd $(API_DIR) && \
		go tool cover -func=coverage.out | \
		awk -v floor=$(API_COVERAGE_FLOOR) '/^total:/ { v=$$3; sub(/%/,"",v); if (v+0 < floor+0) { printf "coverage %.1f%% below floor %.1f%%\n", v+0, floor+0; exit 1 } }'

fresh: ## Wipe app data + build cache and re-apply DB migrations
	@rm -rf "$(APP_DATA_DIR)" "$(TURBO_CACHE_DIR)"
	@echo "Removed $(APP_DATA_DIR)"
	@echo "Removed $(TURBO_CACHE_DIR)"
	@echo "Applying DB migrations..."
	@cd packages/api && \
		GOCACHE=$(ROOT_PATH)/storage/.cache/go-build \
		GOPATH=$(ROOT_PATH)/storage/.cache/gopath \
		go run ./cmd migrate up
	@echo "Done. Run 'make dev' to start fresh."
