.PHONY: run build test seed sqlc ui dev clean lint

# ── Variables ──────────────────────────────────────────────────
BINARY   := api
GO       := go
PORT     ?= 8080
DB_URL   ?= postgres://$(USER)@/gr33n?host=/var/run/postgresql

# ── Development ────────────────────────────────────────────────
run: ## Run the API server
	DATABASE_URL="$(DB_URL)" $(GO) run ./cmd/api/

dev: ## Run API + UI dev server in parallel
	@echo "Starting API on :$(PORT) and UI on :5173"
	@trap 'kill 0' INT; \
		DATABASE_URL="$(DB_URL)" $(GO) run ./cmd/api/ & \
		cd ui && npm run dev & \
		wait

ui: ## Run the Vue dev server
	cd ui && npm run dev

# ── Build ──────────────────────────────────────────────────────
build: ## Build the Go binary
	$(GO) build -o $(BINARY) ./cmd/api/

build-ui: ## Build the Vue frontend for production
	cd ui && npm run build

# ── Test ───────────────────────────────────────────────────────
test: ## Run Go tests
	$(GO) test ./... -v -count=1

lint: ## Run go vet
	$(GO) vet ./...

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
