.PHONY: run run-receiver build build-receiver test seed sqlc migrate merge-legacy-plants ui dev dev-auth-test e2e-browser ollama-smoke ollama-smoke-cpu ollama-smoke-help guardian-eval guardian-qa-smoke guardian-qa-smoke-ec-ph guardian-qa-smoke-unread-alerts guardian-qa-phase127 guardian-qa-regression guardian-qa-manual guardian-qa-smoke-strict guardian-qa-change-requests guardian-laptop-tune rag-ingest-help rag-ingest-demo rag-ingest-platform-docs compose-db-up compose-db-status compose-logging-up compose-logging-down setup-compose-dev dev-stack dev-stack-fresh dev-stack-fresh-rag local-up restart-local restart-local-serve db-sanity-report check-stack check-crop-library check-crop-catalog check-crop-catalog-parity check-catalog-seed-drift add-crop-check check-catalog-release check-ui-domain-parity clean lint bootstrap-local bootstrap-local-docker install-deps-debian install-pi-edge-deps first-clone first-clone-docker first-clone-install-deps audit-openapi audit-env edge-smoke-help edge-actuator-smoke-help recipe-pack-import-help agronomy-seed-pack-help guardian-bootstrap-farm import-agronomy-seed-pack apply-agronomy-overrides rag-ingest-farm-operational

# dash (common default /bin/sh) can report "wait: No child processes" for dev / dev-auth-test;
# bash handles background jobs + wait reliably.
SHELL := /bin/bash

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

test-unit: ## Run unit-testable packages only (excludes cmd/api DB smokes; needs no live Postgres)
	$(GO) test -tags dev $$(go list ./... | grep -v 'gr33n-api/cmd/api$$') -count=1

backup: ## Phase 155 — pg_dump + local file storage backup (see docs/backup-restore-runbook.md)
	@chmod +x ./scripts/backup-gr33n.sh ./scripts/verify-backup-gr33n.sh
	@./scripts/backup-gr33n.sh

verify-backup: ## Phase 155 — restore dump to scratch DB and spot-check (BACKUP=path optional)
	@chmod +x ./scripts/backup-gr33n.sh ./scripts/verify-backup-gr33n.sh
	@./scripts/verify-backup-gr33n.sh $(BACKUP)

vuln-check: ## Phase 156 — govulncheck + npm audit (high+)
	@chmod +x ./scripts/vuln-check.sh
	@./scripts/vuln-check.sh

docs-current-state-hint: ## Phase 157 — counts for regenerating docs/current-state.md
	@chmod +x ./scripts/docs-current-state-hint.sh
	@./scripts/docs-current-state-hint.sh

e2e-browser: ## Playwright browser E2E (requires dev-auth-test stack; see e2e/README.md)
	cd e2e && npm ci && npx playwright install chromium && npm test

ollama-smoke-help: ## Phase 112/118 — print Ollama Guardian E2E smoke commands
	@echo "Ollama Guardian smokes (Phase 112 + 118)"
	@echo "Full notes: INSTALL.md § Ollama E2E smokes"
	@echo ""
	@echo "These are Go tests — they do NOT run inside make dev-auth-test."
	@echo "They start their own test API against DATABASE_URL (like make test)."
	@echo ""
	@echo "One-time DB setup (if smokes fail with auth.users does not exist):"
	@echo "  ./scripts/bootstrap-local.sh --seed"
	@echo "  # or Docker: make dev-stack"
	@echo ""
	@echo "Before each run:"
	@echo "  • Stop make dev-auth-test in other terminals (API + Ollama compete for RAM)"
	@echo "  • Ollama running at LLM_BASE_URL (.env)"
	@echo "  • .env with DATABASE_URL, JWT_SECRET, PI_API_KEY"
	@echo "  • AI_ENABLED=true (smokes default LLM_MODEL=tinyllama; override: LLM_MODEL=… make ollama-smoke-cpu)"
	@echo ""
	@echo "  make ollama-smoke          # standard run"
	@echo "  make ollama-smoke-cpu      # CPU-only box: -timeout 40m + token cap"
	@echo ""
	@echo "Optional pulls before first run:"
	@echo "  ollama pull tinyllama && ollama pull phi3:mini"

ollama-smoke: ## Run Phase 112+118 Ollama smokes (bootstrap-local --seed + .env required)
	@OLLAMA_SMOKE_LLM="$$LLM_MODEL"; \
	if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	if [ -n "$$OLLAMA_SMOKE_LLM" ]; then export LLM_MODEL="$$OLLAMA_SMOKE_LLM"; \
	else export LLM_MODEL=tinyllama; fi; \
	$(GO) test -tags 'dev ollama' ./cmd/api/ -run 'TestPhase112|TestPhase118' -count=1 -v

ollama-smoke-cpu: ## Ollama smokes for CPU-only hosts (longer timeout, lower max tokens)
	@OLLAMA_SMOKE_LLM="$$LLM_MODEL"; \
	if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	if [ -n "$$OLLAMA_SMOKE_LLM" ]; then export LLM_MODEL="$$OLLAMA_SMOKE_LLM"; \
	else export LLM_MODEL=tinyllama; fi; \
	$(GO) test -tags 'dev ollama' ./cmd/api/ -run 'TestPhase112|TestPhase118' -count=1 -v \
		-timeout 40m LLM_TIMEOUT_SECONDS=150 LLM_MAX_TOKENS=60

guardian-eval: ## Phase 122 — run Guardian model quality eval (API + Ollama must be up; set GUARDIAN_EVAL_TOKEN)
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-all} -farm-id $${FARM_ID:-1} \
		-suite $${SUITE:-regression} \
		-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}

guardian-qa-smoke: ## Phase 131 — 4-prompt smoke suite, sequential, full answers archived
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite smoke \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-smoke-ec-ph: ## Phase 147 — re-run smoke-ec-ph only (post run #4 client timeout)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite smoke -prompt-ids smoke-ec-ph \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-smoke-unread-alerts: ## Phase 149 — re-run smoke-unread-alerts only (post run #7 grounded timeout)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite smoke -prompt-ids smoke-unread-alerts \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-phase127: ## Phase 128 — 4-prompt Phase 127 grounding validation (devices, fert, Pi, triage)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite phase127 \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-regression: ## Phase 131 — full regression fixture set (~24 prompts)
	@if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
	$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-all} -farm-id $${FARM_ID:-1} \
		-suite regression \
		-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}

guardian-qa-manual: ## Phase 131/128 — print UI checklist (SUITE=smoke|phase127|regression)
	@$(GO) run ./cmd/guardian-eval/ -manual -suite $${SUITE:-smoke}

guardian-qa-smoke-strict: ## Smoke suite that exits non-zero on any fixture regression (vs. guardian-qa-smoke's always-0 artifact run)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite $${SUITE:-smoke} -fail-on-regression \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-change-requests: ## Fires write-intent prompts; verifies each proposal in pending queue immediately (before 5m TTL)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite change-requests -check-pending-proposals \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-change-requests-ack: ## Re-run write-ack only (~25 min) — fast change-request smoke
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite change-requests -prompt-ids write-ack -check-pending-proposals \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-change-requests-confirm: ## Phase 162 — propose + per-prompt pending + Confirm→DB for write-intent prompts
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite change-requests -check-pending-proposals -confirm-proposals \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-qa-change-requests-ack-confirm: ## Phase 162 fast path — write-ack propose + Confirm→DB (~25 min)
	@bash -lc 'set -e; cd "$(CURDIR)"; \
		if [ -f .env ]; then set -a && . ./.env && set +a; fi; \
		source scripts/source-local-env.sh --refresh-eval-token; \
		$(GO) run ./cmd/guardian-eval/ -models $${MODEL:-phi3:mini} -farm-id $${FARM_ID:-1} \
			-suite change-requests -prompt-ids write-ack -check-pending-proposals -confirm-proposals \
			-report $${GUARDIAN_EVAL_REPORT:-data/guardian_model_eval.json}'

guardian-laptop-tune: ## Phase 129 — print or apply Guardian laptop .env recommendations (ARGS="--apply")
	@chmod +x ./scripts/tune-guardian-laptop.sh
	@./scripts/tune-guardian-laptop.sh $(ARGS)

lint: ## Run go vet
	$(GO) vet -tags dev ./...

check-crop-library: ## Phase 82 WS4a — validate data/crop_library.yaml (EC mS/cm, growth_stage_enum)
	@./scripts/generate-crop-seed.sql.sh --validate

check-crop-catalog: ## Phase 84 WS-B — validate catalog + field guide seed sources
	@./scripts/generate-crop-catalog-seed.sql.sh --validate

check-crop-catalog-db: ## Phase 84 WS-K — verify Postgres catalog seed (needs migrate + DATABASE_URL)
	@./scripts/check-crop-catalog-db.sh

check-crop-catalog-parity: check-crop-library check-crop-catalog check-crop-catalog-db ## YAML seed + DB parity

check-catalog-seed-drift: ## Phase 95 — regenerated catalog SQL matches db/seed/crop_catalog_from_yaml.sql
	@./scripts/check-catalog-seed-drift.sh

add-crop-check: ## Phase 95 — integrator pre-migrate validation (YAML + seed drift; no DB)
	@./scripts/add-crop-check.sh

check-catalog-release: add-crop-check check-crop-catalog-db ## Phase 95 — full catalog release checklist (needs migrate)

check-ui-domain-parity: ## Phase 99 — UI enum lists match backend/OpenAPI (growth stages, lighting presets)
	@./scripts/check-ui-domain-parity.sh

merge-legacy-plants: ## Phase 103 — audit legacy plants (add APPLY=1 to run merge)
	@./scripts/merge-legacy-plants.sh $(if $(APPLY),--apply --audit,)

audit-openapi: ## Phase 20.95 WS6 — confirm openapi.yaml matches cmd/api/routes.go
	@./scripts/openapi_route_diff.sh

audit-env: ## Phase 116 WS1 — confirm env vars are documented
	@./scripts/env_reference_parity.sh

# ── Database ───────────────────────────────────────────────────
sqlc: ## Regenerate sqlc Go code from SQL queries
	sqlc generate

migrate: ## Apply pending db/migrations/*.sql only (skips full schema; uses DATABASE_URL from .env)
	@./scripts/bootstrap-local.sh --skip-schema

rag-ingest-help: ## Show rag-ingest CLI flags (farm-scoped embedding; see docs/workflow-guide.md §10.6)
	@$(GO) run ./cmd/rag-ingest -help

edge-smoke-help: ## Phase 31 WS1 — print laptop stub loop commands (pi_client → dashboard Live Sensors)
	@echo "Edge loop in 5 commands (stub readings → dashboard Live Sensors)"
	@echo "Full notes: docs/local-operator-bootstrap.md § Edge loop in 5 commands"
	@echo ""
	@echo "  1. make dev-stack                 # once: Compose db + schema + master_seed"
	@echo "  2. make dev-auth-test             # terminal 1: API + UI (JWT_SECRET + PI_API_KEY in .env)"
	@echo "  3. ./scripts/print-demo-sensor-ids.sh   # confirm sensor_id ↔ master_seed names"
	@echo "  4. ./scripts/run-edge-stub-client.sh"
	@echo "     # or manual: cd pi_client && cp config.demo-stub.yaml config.yaml"
	@echo "     # set api.api_key = PI_API_KEY from .env, then python3 gr33n_client.py"
	@echo "  5. Dashboard → gr33n Demo Farm → Live Sensors (SSE; values within ~1s)"
	@echo ""
	@echo "Automation stays off the GPIO path by default: AUTOMATION_SIMULATION_MODE=true"
	@echo "  in .env simulates actuator rules in the API; pi_client only posts readings."
	@echo "  Set AUTOMATION_SIMULATION_MODE=false when testing pending_command → GPIO (Phase 31 WS3)."

edge-actuator-smoke-help: ## Phase 31 WS3 — print safe actuator E2E commands (pending_command → pi_client → events)
	@echo "Safe actuator round-trip (Phase 31 WS3)"
	@echo "Full narrative: docs/pi-integration-guide.md §9"
	@echo "Safety: docs/operator-troubleshooting.md §5"
	@echo ""
	@echo "  1. make dev-auth-test                 # API + PI_API_KEY"
	@echo "  2. ./scripts/print-demo-actuator-ids.sh"
	@echo "  3. ./scripts/run-edge-actuator-smoke.sh --direct"
	@echo "     # or manual:"
	@echo "     terminal A: ./scripts/run-edge-actuator-client.sh"
	@echo "     terminal B: ./scripts/enqueue-demo-pending-command.sh on"
	@echo "  4. Guardian path: ./scripts/run-edge-actuator-smoke.sh --guardian"
	@echo ""
	@echo "Ids: demo-veg-relay-01 + Veg Room Grow Light (master_seed)"

recipe-pack-import-help: ## Phase 31 WS5 — print recipe pack promotion commands (commons catalog → farms)
	@echo "Recipe Pack v7 demo promotion (Phase 31 WS5)"
	@echo "Docs: scripts/enterprise/README.md · docs/hypothetical-enterprise-topology.md"
	@echo ""
	@echo "  1. Apply migration: db/migrations/20260527_phase31_commons_recipe_pack_v7.sql"
	@echo "  2. make dev-auth-test"
	@echo "  3. ./scripts/enterprise/import-recipe-pack.sh --dry-run"
	@echo "  4. ./scripts/enterprise/import-recipe-pack.sh --farm-ids 1,2"

agronomy-seed-pack-help: ## Phase 83 WS1 — print agronomy seed pack import + bootstrap commands
	@echo "Cultivator Agronomy Seed Pack v1 (Phase 83 WS1 + WS3)"
	@echo "Docs: docs/crop-catalog-db-cutover-runbook.md · scripts/enterprise/README.md"
	@echo ""
	@echo "  1. make migrate && make check-crop-catalog-parity"
	@echo "  2. make dev-auth-test"
	@echo "  3. ./scripts/enterprise/import-agronomy-seed-pack.sh --dry-run"
	@echo "  4. ./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1"
	@echo "  5. ./scripts/enterprise/guardian-bootstrap-farm.sh --dry-run --farm-id 1"
	@echo "  6. ./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1 [--smoke]"

import-agronomy-seed-pack: ## Phase 83 WS1 — record agronomy pack import + verify DB (FARM_IDS=1)
	@./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids $(or $(FARM_IDS),1)

guardian-bootstrap-farm: ## Phase 83 WS3 — RAG ingest + readiness report (FARM_ID=1)
	@./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id $(or $(FARM_ID),1) $(ARGS)

apply-agronomy-overrides: ## Phase 83 WS2 — farm crop EC/VPD overrides (FARM_ID=1 FILE=data/agronomy-override-pack.example.yaml)
	@./scripts/enterprise/apply-agronomy-overrides.sh --farm-id $(or $(FARM_ID),1) --file $(or $(FILE),data/agronomy-override-pack.example.yaml)

rag-ingest-farm-operational: ## Phase 83 WS5 — incremental operational RAG for farm (FARM_ID=1)
	@./scripts/rag-ingest-farm-operational.sh --farm-id $(or $(FARM_ID),1)

compose-db-up: ## Start only the Postgres image from docker-compose.yml (Timescale + pgvector build)
	docker compose up -d db --build

compose-db-status: ## Show docker compose db container status
	docker compose ps db

compose-logging-up: ## Optional Loki + Promtail + Grafana (merge docker-compose.logging.yml); Linux Docker Engine recommended
	docker compose -f docker-compose.yml -f docker-compose.logging.yml up -d

compose-logging-down: ## Stop merged stack including logging services (use `docker compose down -v` manually to drop Loki/Grafana volumes)
	docker compose -f docker-compose.yml -f docker-compose.logging.yml down

setup-compose-dev: ## Alias for dev-stack — Docker db + bootstrap --seed + check-stack (see scripts/dev-stack.sh)
	@./scripts/dev-stack.sh

dev-stack: ## Idempotent: Compose db + migrations + seed + verify (existing DB: auto-skips schema). See dev-stack-fresh to wipe.
	@./scripts/dev-stack.sh

dev-stack-fresh: ## Wipe Compose DB volumes + full bootstrap + seed (destructive — clean Guardian demo)
	@./scripts/dev-stack.sh --reset-volumes --quick

dev-stack-fresh-rag: ## dev-stack-fresh + rag-ingest demo farm (skip ingest if EMBEDDING_API_KEY unset)
	@./scripts/dev-stack.sh --reset-volumes --quick --rag-ingest

rag-ingest-demo: ## Index farm_id=1 for Guardian RAG (needs EMBEDDING_API_KEY; no-op with skip message if unset)
	@./scripts/rag-ingest-demo.sh

rag-ingest-platform-docs: ## Index operator platform docs for farm 1 (needs EMBEDDING_API_KEY; dry-run without key via --dry-run)
	@./scripts/rag-ingest-platform-docs.sh

rag-ingest-platform-docs-dry-run: ## List manifest files + chunk estimates (no API key required)
	@./scripts/rag-ingest-platform-docs.sh --dry-run

rag-ingest-field-guides: ## Index field guides from DB for farm 1 (default; needs EMBEDDING_API_KEY)
	@./scripts/rag-ingest-field-guides.sh

rag-ingest-field-guides-dry-run: ## Dry-run DB field guides (no API key; AGRONOMY_FIELD_GUIDES_SOURCE=db default)
	@./scripts/rag-ingest-field-guides.sh --dry-run

local-up: ## dev-stack then API + UI (same as ./scripts/dev-stack.sh --serve)
	@./scripts/dev-stack.sh --serve

restart-local: ## After reboot: Compose db up + wait + db sanity report (no migrations). See scripts/restart-local.sh --serve for API+UI
	@./scripts/restart-local.sh

restart-local-serve: ## Same as restart-local then make dev-auth-test (Go compile may be slow cold)
	@./scripts/restart-local.sh --serve

db-sanity-report: ## Read-only SQL checks (duplicate zones/sensors, profile, extensions)
	@./scripts/db-sanity-report.sh

dev-reset-farm: ## Reset farm 1 demo config without volume wipe (Phase 48; DEV_SEED_PROFILE or --profile)
	@./scripts/dev-reset-farm.sh $(ARGS)

apply-dev-retention: ## Apply Timescale retention when TIMESCALE_RETENTION_DAYS is set (dev/staging)
	@./scripts/apply-dev-retention.sh

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
