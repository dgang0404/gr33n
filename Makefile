.PHONY: run run-receiver build build-receiver test seed sqlc ui dev clean lint bootstrap-local bootstrap-local-docker

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

# ── Database ───────────────────────────────────────────────────
sqlc: ## Regenerate sqlc Go code from SQL queries
	sqlc generate

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
