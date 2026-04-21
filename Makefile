.PHONY: run run-receiver build build-receiver test seed sqlc ui dev dev-auth-test rag-ingest-help clean lint bootstrap-local bootstrap-local-docker install-deps-debian install-pi-edge-deps first-clone first-clone-docker first-clone-install-deps audit-openapi

# ── Variables ──────────────────────────────────────────────────
BINARY   := api
GO       := go
PORT     ?= 8080
DB_URL   ?= postgres://$(USER)@/gr33n?host=/var/run/postgresql

# ── Bootstrap (Phase 15 operator path) ─────────────────────────
bootstrap-local: ## DB schema + migrations, .env from example if missing, npm ci (see docs/local-operator-bootstrap.md)
	@./scripts/bootstrap-local.sh

bootstrap-local-docker: ## docker compose up -d + npm ci (DB/API/UI in containers)
	@./scripts/bootstrap-local.sh --docker

first-clone: ## First git clone: go mod download, .env templates, then bootstrap-local (needs DB — see INSTALL.md)
	@./scripts/setup-first-clone.sh

first-clone-docker: ## Same as first-clone but --docker (no host Postgres required for schema steps)
	@./scripts/setup-first-clone.sh --docker

install-deps-debian: ## Debian/Ubuntu: sudo apt — Postgres 16 (PGDG) + PostGIS + pgvector + TimescaleDB + Node 22 (not Go)
	@./scripts/install-system-deps-debian.sh

install-pi-edge-deps: ## Raspberry Pi OS: sudo apt — Python venv + GPIO helpers + git (+ optional Docker)
	@./scripts/install-pi-edge-deps.sh

install-pi-edge-deps-docker: ## Same + Docker Engine & Compose (full stack on Pi experiments)
	@./scripts/install-pi-edge-deps.sh --with-docker

first-clone-install-deps: ## first-clone + install-deps-debian first (Linux Debian/Ubuntu only)
	@./scripts/setup-first-clone.sh --install-system-deps

# ── Development ────────────────────────────────────────────────
run: ## Run the API server (dev build, auth bypass available)
	AUTH_MODE=dev DATABASE_URL="$(DB_URL)" $(GO) run -tags dev ./cmd/api/

run-auth: ## Run the API server with AUTH_MODE=production (real auth; dev-tagged build)
	AUTH_MODE=production DATABASE_URL="$(DB_URL)" $(GO) run -tags dev ./cmd/api/

run-auth-test: ## Local auth regression: AUTH_MODE=auth_test (requires JWT_SECRET, PI_API_KEY; dev tag only)
	AUTH_MODE=auth_test DATABASE_URL="$(DB_URL)" $(GO) run -tags dev ./cmd/api/

run-receiver: ## Run optional Insert Commons ingest receiver (:8765; set INSERT_COMMONS_SHARED_SECRET or ALLOW_INSECURE)
	DATABASE_URL="$(DB_URL)" $(GO) run ./cmd/insert-commons-receiver/

dev: ## Run API + UI dev server in parallel
	@echo "Starting API on :$(PORT) and UI on :5173"
	@trap 'kill 0' INT; \
		AUTH_MODE=dev DATABASE_URL="$(DB_URL)" $(GO) run -tags dev ./cmd/api/ & \
		cd ui && npm run dev & \
		wait

dev-auth-test: ## API + UI with AUTH_MODE=auth_test — set JWT_SECRET & PI_API_KEY (e.g. in `.env`; see .env.example)
	@echo "Starting API on :$(PORT) with AUTH_MODE=auth_test + UI on :5173"
	@echo "Ensure JWT_SECRET and PI_API_KEY are set (copied from .env.example if needed)."
	@trap 'kill 0' INT; \
		AUTH_MODE=auth_test DATABASE_URL="$(DB_URL)" $(GO) run -tags dev ./cmd/api/ & \
		cd ui && npm run dev & \
		wait

ui: ## Run the Vue dev server
	cd ui && npm run dev

# ── Build ──────────────────────────────────────────────────────
build: ## Build the Go binary
	$(GO) build -o $(BINARY) ./cmd/api/

build-receiver: ## Build Insert Commons receiver binary
	$(GO) build -o insert-commons-receiver ./cmd/insert-commons-receiver/

build-ui: ## Build the Vue frontend for production
	cd ui && npm run build

# ── Test ───────────────────────────────────────────────────────
test: ## Run Go tests (dev build so smoke tests can use auth bypass)
	$(GO) test -tags dev ./... -v -count=1

lint: ## Run go vet
	$(GO) vet -tags dev ./...

audit-openapi: ## Phase 20.95 WS6 — confirm openapi.yaml matches cmd/api/routes.go
	@./scripts/openapi_route_diff.sh

# ── Database ───────────────────────────────────────────────────
sqlc: ## Regenerate sqlc Go code from SQL queries
	sqlc generate

rag-ingest-help: ## Show rag-ingest CLI flags (farm-scoped embedding; see docs/workflow-guide.md §10.6)
	@$(GO) run ./cmd/rag-ingest -help

seed: ## Apply seed data to the database
	psql "$(DB_URL)" -f db/seeds/master_seed.sql

schema: ## Apply the schema to the database
	psql "$(DB_URL)" -f db/schema/gr33n-schema-v2-FINAL.sql

# ── Docker ─────────────────────────────────────────────────────
up: ## Start Docker Compose services
	docker compose up -d

down: ## Stop Docker Compose services
	docker compose down

logs: ## Tail Docker Compose logs
	docker compose logs -f

# ── Cleanup ────────────────────────────────────────────────────
clean: ## Remove build artifacts
	rm -f $(BINARY)
	rm -rf ui/dist

# ── Help ───────────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
