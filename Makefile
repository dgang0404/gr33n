.PHONY: run run-receiver build build-receiver test seed sqlc ui dev dev-auth-test rag-ingest-help compose-db-up compose-db-status setup-compose-dev dev-stack local-up check-stack clean lint bootstrap-local bootstrap-local-docker install-deps-debian install-pi-edge-deps first-clone first-clone-docker first-clone-install-deps audit-openapi

# ── Variables ──────────────────────────────────────────────────
BINARY   := api
GO       := go
PORT     ?= 8080
# Optional override for sqlc/run targets only: `make seed DB_URL=postgres://…`
LOCAL_PEER_DSN ?= postgres://$(USER)@/gr33n?host=/var/run/postgresql

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
run: ## Run the API server (dev build, auth bypass available); uses DATABASE_URL from .env unless DB_URL is set (`make run DB_URL=…`)
	AUTH_MODE=dev $(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run -tags dev ./cmd/api/

run-auth: ## Run the API server with AUTH_MODE=production (real auth; dev-tagged build)
	AUTH_MODE=production $(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run -tags dev ./cmd/api/

run-auth-test: ## Local auth regression: AUTH_MODE=auth_test (requires JWT_SECRET, PI_API_KEY; dev tag only)
	AUTH_MODE=auth_test $(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run -tags dev ./cmd/api/

run-receiver: ## Run optional Insert Commons ingest receiver (:8765; set INSERT_COMMONS_SHARED_SECRET or ALLOW_INSECURE)
	$(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run ./cmd/insert-commons-receiver/

dev: ## Run API + UI dev server in parallel (DATABASE_URL from .env unless `make dev DB_URL=…`)
	@echo "Starting API on :$(PORT) and UI on :5173"
	@trap 'kill 0' INT; \
		AUTH_MODE=dev $(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run -tags dev ./cmd/api/ & \
		cd ui && npm run dev & \
		wait

dev-auth-test: ## API + UI with AUTH_MODE=auth_test — set JWT_SECRET & PI_API_KEY (e.g. in `.env`; see .env.example)
	@echo "Starting API on :$(PORT) with AUTH_MODE=auth_test + UI on :5173"
	@echo "Ensure JWT_SECRET and PI_API_KEY are set (copied from .env.example if needed)."
	@trap 'kill 0' INT; \
		AUTH_MODE=auth_test $(if $(DB_URL),DATABASE_URL="$(DB_URL)") $(GO) run -tags dev ./cmd/api/ & \
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

compose-db-up: ## Start only the Postgres image from docker-compose.yml (Timescale + pgvector build)
	docker compose up -d db --build

compose-db-status: ## Show docker compose db container status
	docker compose ps db

setup-compose-dev: ## Alias for dev-stack — Docker db + bootstrap --seed + check-stack (see scripts/dev-stack.sh)
	@./scripts/dev-stack.sh

dev-stack: ## Reliable local Compose DB + bootstrap + seed + verify; auto-retries Docker via sg docker when needed
	@./scripts/dev-stack.sh

local-up: ## dev-stack then API + UI (same as ./scripts/dev-stack.sh --serve)
	@./scripts/dev-stack.sh --serve

check-stack: ## Verify .env DATABASE_URL, pgvector, optional API /health (see docs/local-operator-bootstrap.md)
	@./scripts/check-local-stack.sh

seed: ## Apply seed data to the database (`make seed DB_URL=…` or defaults to LOCAL_PEER_DSN)
	psql "$(if $(DB_URL),$(DB_URL),$(LOCAL_PEER_DSN))" -f db/seeds/master_seed.sql

schema: ## Apply the schema to the database
	psql "$(if $(DB_URL),$(DB_URL),$(LOCAL_PEER_DSN))" -f db/schema/gr33n-schema-v2-FINAL.sql

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
