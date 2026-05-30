SHELL := /bin/bash

ROOT_PATH := $(shell pwd)
UI_DIR := packages/ui
PORTLESS_APP_NAME := git-diff-ui
PORTLESS_DEFAULT_DEV_SERVER_URL := https://$(PORTLESS_APP_NAME).localhost
GO_FMT_COMPOSE_FILE := go-fmt.compose.yaml
GO_FMT_SERVICE := go-fmt
GO_FMT_COMPOSE := docker compose -f $(GO_FMT_COMPOSE_FILE)
GO_FMT_BIN := /usr/local/bin/go-fmt
GO_FMT_EXEC := $(GO_FMT_COMPOSE) exec -T $(GO_FMT_SERVICE) $(GO_FMT_BIN)
OXFMT := pnpm exec oxfmt
OXLINT := pnpm exec oxlint
TSX := pnpm exec tsx
BLANK_LINES := $(ROOT_PATH)/scripts/blank-lines.ts
APP_DATA_DIR := $$HOME/Library/Application Support/git-diff
TURBO_CACHE_DIR := storage/.cache/turbo

API_DIR := packages/api
API_COVERAGE_OUT := $(API_DIR)/coverage.out
API_COVERAGE_FLOOR := 60.0

.DEFAULT_GOAL := help
.PHONY: help dev format format-all format-start format-stop format-login fresh test-api test-api-cover

help: ## Show this help (list of make targets)
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Run the UI dev server (https://git-diff-ui.localhost)
	@echo "Dev server will be available at: $(PORTLESS_DEFAULT_DEV_SERVER_URL)"
	@pnpm --dir $(UI_DIR) run dev

format: format-start ## Format Go, TS, and Vue sources
	@echo "go-fmt format in $(ROOT_PATH)"; \
	$(GO_FMT_EXEC) format --cwd $(ROOT_PATH) --host-path $(ROOT_PATH)
	@echo "blank-lines fix across packages and root scripts"
	@cd $(ROOT_PATH) && $(TSX) $(BLANK_LINES) \
		packages/ui \
		packages/bridge \
		packages/domain \
		scripts
	@echo "oxfmt format in $(ROOT_PATH)"
	@$(OXFMT) --write packages/ui packages/bridge package.json turbo.json
	@echo "oxlint fix in $(ROOT_PATH)"
	@$(OXLINT) --fix --vue-plugin packages/ui packages/bridge

format-all: format-start ## Format Go + all JS/TS/Vue sources (incl. contracts and scripts)
	@echo "go-fmt format in $(ROOT_PATH)"; \
	$(GO_FMT_EXEC) format --cwd $(ROOT_PATH) --host-path $(ROOT_PATH)
	@echo "blank-lines fix across packages and root scripts"
	@cd $(ROOT_PATH) && $(TSX) $(BLANK_LINES) \
		packages/ui \
		packages/bridge \
		packages/domain/src \
		scripts
	@echo "oxfmt format across all JS/TS sources"
	@$(OXFMT) --write \
		packages/ui \
		packages/bridge \
		packages/domain/src \
		scripts \
		package.json \
		turbo.json
	@echo "oxlint fix across all JS/TS sources"
	@$(OXLINT) --fix --vue-plugin \
		packages/ui \
		packages/bridge \
		packages/domain/src \
		scripts

format-start:
	@$(GO_FMT_COMPOSE) up -d $(GO_FMT_SERVICE)

format-stop:
	@$(GO_FMT_COMPOSE) stop $(GO_FMT_SERVICE)

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
