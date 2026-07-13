# Local operator bootstrap — start here

**Run bootstrap and Make targets from the repository root** (`cd /path/to/gr33n-platform` after `git clone`). Commands like `./scripts/bootstrap-local.sh` and `make dev` apply to **this** repo only — not from your home directory (`~`).

**Quick links:** [First session after clone](first-session-after-clone.md) · [Example `.env` (doc copy)](example-env.md) · [Machine setup checklist](machine-setup-checklist.md) · [Operator tour — dashboard narrative](operator-tour.md) · [Tasks-first guide (morning ops, automation, offline queue)](tasks-first-operator-guide.md) · [Operator troubleshooting (401, logs)](operator-troubleshooting.md) · [Sit-in workstream — operator UX + logging + tasks](workstreams/sit-in-operator-experience.md) · [Offline / intranet deployment (LAN, VLAN, local LLM)](offline-or-intranet-deployment.md)

Single happy path for standing up **Postgres → API → dashboard → optional Insert Commons receiver → optional Pi / MQTT bridge**, with explicit env templates and pointers to federation and audit docs. For farm template behavior (blank vs starter pack), see [`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md).

## Laptop daily cheat sheet (Guardian + Ollama)

**Login-and-go:** after reboot you should not need manual `ollama stop` rituals. Tune once per machine, restart the stack, log in, and open Guardian — awakening preloads the counsel model in the background.

**Where to run commands:** `make …` and `./scripts/…` must be run from the **repository root** (`cd ~/gr33n-platform`). `systemctl start ollama` works from **any directory** — you do not need to `cd` into the repo first.

```bash
cd ~/gr33n-platform

# After reboot — one line on your laptop (DB + Ollama + API + UI)
make laptop-up                    # same as make restart-local-serve

# One-time per laptop (CPU timeouts + retry policy)
make guardian-laptop-tune ARGS="--apply"

# Or stepwise:
make restart-local              # → scripts/restart-local.sh (db + sanity only)
make dev-auth-test              # API :8080 + UI :5173

# Login → open Farm Guardian — badge goes amber → green while awakening runs
# Settings → Farm Guardian readiness shows state, corpus counts, Awaken now / Rest now

# Solar/battery: optional auto-rest after idle (unload model RAM — not the Ollama service)
# GUARDIAN_AUTO_DORMANT_MINUTES=45   # in .env, then restart API

# Verify stack (Postgres, API, Ollama probe)
make check-stack                # → scripts/check-local-stack.sh

# Optional: RAG ingest when Ollama idle (not while chatting)
make rag-ingest-farm-operational FARM_ID=1
make rag-ingest-platform-docs
# Full bootstrap: make guardian-bootstrap-farm FARM_ID=1
```

**If awakening stalls > 30s:** Ollama may still be starting — wait a few seconds and tap **Awaken now** in Settings. If you stopped Ollama for testing, `systemctl start ollama` from any terminal (no `cd` into the repo). Use **Quick chat** while the counsel model warms up.

**Laptop vs enterprise:** `make laptop-up` assumes **everything on one machine** (Postgres in Docker, Ollama on loopback, API + UI on localhost). Safe to re-run: already-up DB, Ollama, and API+UI are detected and left alone; if only one of API/UI is up, it starts the missing piece (no duplicate stack). Production / enterprise sites split roles — Postgres, API, UI, and Ollama may each live on different hosts. There is no single “bring up the farm” command across servers; admins start each tier on its own box (see [offline-or-intranet-deployment.md](offline-or-intranet-deployment.md)).

**Power saving (Phase 163):** three tiers — pick what fits your site:

| Tier | What | How |
|------|------|-----|
| **Rest now** | Unload warm chat model from RAM | Settings → Farm Guardian readiness → **Rest now** |
| **Auto-rest** | Same, after N idle minutes | `GUARDIAN_AUTO_DORMANT_MINUTES=45` in `.env`, restart API |
| **Service stop** | Stop Ollama process entirely (admin) | `./scripts/guardian-power.sh sleep` (sudo) — not in web UI |

After service stop, run `./scripts/guardian-power.sh wake` then **Awaken now**.

**Scheduled deep sleep (cron):** on the host where Ollama runs (may differ from the UI/API box), install overnight stop/start:

```bash
./scripts/guardian-power-schedule.sh print-crontab   # example 22:00 sleep / 06:00 wake
./scripts/guardian-power-schedule.sh print-sudoers   # optional NOPASSWD for cron only
```

Cron calls `cron-sleep` / `cron-wake` (logged wrappers around `guardian-power.sh`). Daytime RAM saving stays in the API via `GUARDIAN_AUTO_DORMANT_MINUTES` — no sudo.

**Manual RAM hygiene (rare):** only if the box is wedged after heavy ingest + chat:

```bash
ollama stop phi3:mini
ollama stop rjmalagon/gte-qwen2-1.5b-instruct-embed-f16   # match EMBEDDING_MODEL in .env
```

**Script map:** [`scripts/restart-local.sh`](../scripts/restart-local.sh) · [`scripts/check-local-stack.sh`](../scripts/check-local-stack.sh) · [`scripts/tune-guardian-laptop.sh`](../scripts/tune-guardian-laptop.sh) · [`scripts/guardian-power.sh`](../scripts/guardian-power.sh) · [`scripts/guardian-power-schedule.sh`](../scripts/guardian-power-schedule.sh) · [`scripts/rag-ingest-farm-operational.sh`](../scripts/rag-ingest-farm-operational.sh) · [`scripts/enterprise/guardian-bootstrap-farm.sh`](../scripts/enterprise/guardian-bootstrap-farm.sh)

**Guardian CPU / timeouts / pull vs dropdown:** [guardian-ollama-laptop-playbook.md](guardian-ollama-laptop-playbook.md)

### Guardian QA (Phase 131)

After Guardian changes, validate with the **smoke suite** (4 prompts, sequential, full answers archived):

```bash
# JWT from dev login — export or put in .env
export GUARDIAN_EVAL_TOKEN="<jwt from browser localStorage gr33n_token>"
export GUARDIAN_EVAL_LOG=/tmp/gr33n-api.log   # optional log correlation

make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1
# Archives: data/guardian_qa_runs/<timestamp>_smoke_phi3-mini.json

# Print the same steps for manual UI validation:
make guardian-qa-manual              # smoke checklist (default)
make guardian-qa-manual SUITE=regression

# Full regression (~24 prompts, slow on CPU):
make guardian-qa-regression MODEL=phi3:mini

# Grep logs for walk_farm evidence after morning-walk step:
./scripts/guardian-qa-scrape-logs.sh --expect walk_farm

# After smoke, review thumbs-down feedback (farm admin):
curl -H "Authorization: Bearer $GUARDIAN_EVAL_TOKEN" \
  'http://127.0.0.1:8080/v1/chat/feedback/export?farm_id=1&since=7d'
```

Or use **Settings → Guardian feedback** (Phase 141) — see [guardian-feedback-review-runbook.md](guardian-feedback-review-runbook.md).

See [phase_128 plan](plans/phase_128_validate_phase127_guardian.plan.md) — Phase 127 grounding validation. For the full **129–138** arc (awakening, runtime, QA, router, honesty, feedback, RAG lifecycle, counsel UX, inference policy), see the [Guardian next-level roadmap](plans/phase_129_139_guardian_next_level_roadmap.plan.md).

### Phase 128 — validate Phase 127 grounding (legacy)

Prefer **`make guardian-qa-smoke`** and **`make guardian-qa-manual SUITE=phase127`** over ad-hoc log greps. **`make guardian-qa-smoke`** refreshes the eval JWT automatically (Phase 146). Before first smoke on a fresh DB, run **`make migrate`** so Phase 138 columns (`guardian_counsel_model`, etc.) exist — missing migration causes background tick noise but not chat failure.

Fast automated checks:

```bash
go test ./cmd/api/... -run 'Phase127|Phase128' -count=1
make guardian-qa-manual SUITE=phase127
# Optional live API (slow on CPU):
export GUARDIAN_EVAL_TOKEN="<jwt>"
make guardian-qa-phase127 MODEL=phi3:mini FARM_ID=1
```

### Chat model on a 16 GB CPU laptop — tinyllama vs phi3:mini

**You were right:** for *local* ungrounded chat (farm context **off**), **`tinyllama` is often the better default** than `phi3:mini`. Logs on this profile showed phi3 taking **~9+ minutes to the first token** (`ttft_ms≈558000`) and still **timing out at 777 s** before the answer finished. Tinyllama is much smaller (~638 MB vs ~2.3 GB) and typically answers in seconds to a few minutes on CPU.

| Question type | Farm context | Model | Why |
|---------------|--------------|-------|-----|
| Home garden, general Q&A, “hi”, cherry/forest-garden prompts | **Off** | **tinyllama** (session dropdown or `.env`) | No RAG/embed; fast CPU replies |
| Demo farm beds, alerts, RAG, morning walkthrough | **On** | **phi3:mini** | Needs ≥8192 ctx gate; tinyllama **rejected** (2048 ctx) — UI blocks Send if you pick tinyllama + farm context |
| Best quality, patient wait | Off | phi3:mini | Slower; raise `LLM_TIMEOUT_SECONDS` (e.g. 900–1200) if you insist on phi3 for long answers |

**`.env` examples (repo root, then restart API):**

```bash
# Fast local default (ungrounded chat)
LLM_MODEL=tinyllama:latest
LLM_TIMEOUT_SECONDS=300
LLM_RETRY_MAX_ATTEMPTS=1

# Grounded demo / quality (farm context on) — keep phi3; expect minutes per turn on CPU
# LLM_MODEL=phi3:mini
# LLM_TIMEOUT_SECONDS=900
# LLM_MAX_TOKENS=512          # optional — shorter answers finish sooner
```

**RAG ingest** (`make rag-ingest-farm-operational`, `guardian-bootstrap-farm`) uses **`EMBEDDING_MODEL` only** — not `LLM_MODEL`. Ingest works the same whether chat is tinyllama or phi3.

**Field guides** (`docs/field-guides/`, `make rag-ingest-field-guides`) are only injected when **farm context is on** — not for quick/off-farm chat. After adding guides: `make migrate` then `make rag-ingest-field-guides`. See [field-guides/README.md](field-guides/README.md) and [phase_127 plan](plans/phase_127_snapshot_devices_fertigation_guides.plan.md).

**Warm-up trick for phi3:** if you stay on phi3, send a one-line “hi” first and wait for completion so the model stays in RAM; then send the real question. Cold phi3 on CPU dominates wait time.

## Server & frontier delta (from laptop)

Same **repository root** commands as the laptop sheet (`make migrate`, `make guardian-bootstrap-farm`, `make check-stack`, RAG scripts). What changes is **hardware**, **`.env`**, and **how you run API/UI in production** — not a second script tree.

| Topic | Laptop (Profile A) | Standard server (Profile C/D) | Frontier / multi-site |
|-------|-------------------|--------------------------------|------------------------|
| **Goal** | Dev, demo, CPU phi3 | On-prem Guardian + GPU | Per-site full stack, offline-capable |
| **API + UI** | `make dev-auth-test` | **Production:** built binary + **systemd**; UI static via **nginx/Caddy** — see [farm-guardian-ollama-setup.md](farm-guardian-ollama-setup.md) | Same per site; Pis edge-only |
| **Postgres** | Docker `db` on `:5433` | Dedicated host, 16–32 GB RAM, NVMe | Local DB per site (Topology B) |
| **Ollama** | Same machine, CPU | Often **split host**: `LLM_BASE_URL=http://ollama.farm.local:11434/v1` | Local Ollama per site; optional multi-model pre-pulled |
| **Default chat model** | **`tinyllama`** for speed, or **`phi3:mini`** for grounded demo (slow on CPU) | `llama3.1:8b` (single-box GPU) or **70B** on 24 GB VRAM box | Farm default in Settings + pre-pull 2+ chat models |
| **Embeddings** | Same Ollama host | Same or cloud `EMBEDDING_*` | Same; cron refresh (below) |
| **Pull large models** | Terminal `ollama pull` (UI 600 s often fails) | `ollama pull` on inference box; set `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS=3600` if using UI pull | Pre-pull during site bring-up |
| **RAM hygiene** | Often need `ollama stop` embed before chat | GPU + 32–64 GB: usually **skip** manual stop; use `OLLAMA_KEEP_ALIVE` to keep chat model warm | Keep chat + embed loaded if RAM allows |
| **Guardian UX** | Minutes per grounded turn on CPU | Selector shows **`loaded on GPU`**; seconds to first token | Fast model switch if weights already in RAM |
| **RAG bootstrap** | `make guardian-bootstrap-farm FARM_ID=1` | Same per farm after migrate | + [`scripts/enterprise/apply-site-manifest.sh`](../scripts/enterprise/apply-site-manifest.sh) (`guardian_seed` in YAML) |
| **RAG refresh** | Manual / when idle | Cron: `make rag-ingest-farm-operational FARM_ID=N` every 6h — [enterprise README](../scripts/enterprise/README.md) | Per-farm cron on each site |

### Server bring-up (delta commands)

```bash
cd ~/gr33n-platform   # or /opt/gr33n-platform on the app host

# Schema + seed (once per environment)
make migrate
./scripts/bootstrap-local.sh --seed    # or your prod migration pipeline

# Inference host (often a separate machine) — pull once over LAN/internet
ollama pull llama3.1:8b                # Profile D single-box
# ollama pull llama3.1:70b-instruct-q4_K_M   # Profile C dedicated GPU box

# App host .env (examples — tune for your LAN)
# LLM_BASE_URL=http://192.168.1.50:11434/v1
# LLM_MODEL=llama3.1:8b
# GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS=3600
# LLM_TIMEOUT_SECONDS=666

# Guardian corpus per farm
make guardian-bootstrap-farm FARM_ID=1 ARGS="--smoke"

# Verify (from app host)
make check-stack
```

**Sizing detail:** [recommended-hardware-and-sizing.md](recommended-hardware-and-sizing.md) (Profiles B–D).

**Multi-site / warehouse:** [hypothetical-enterprise-topology.md](hypothetical-enterprise-topology.md) · [scripts/enterprise/README.md](../scripts/enterprise/README.md) — site manifest, agronomy pack, recipe promotion. No third full cheat sheet; extend the enterprise README when you provision farm #2+.

## Prerequisites

| Need | Native install | Docker only |
|------|----------------|-------------|
| Go | 1.23+ | Optional (API runs in container) |
| Node.js | 22+ recommended (`npm` for UI) | Optional if you only use the UI container |
| PostgreSQL | 14+ with **TimescaleDB** and **PostGIS** (schema runs `CREATE EXTENSION`) | Provided by Compose |
| Docker | — | Docker Engine + Compose v2 |

**Ubuntu 22.04 (Jammy) — Docker from Ubuntu repos:** install **`docker.io`** and **`docker-compose-v2`** (provides `docker compose`). The package **`docker-compose-plugin`** is from Docker Inc.’s apt repository and is **not** in the default Ubuntu archive—if `apt` cannot find it, use **`docker-compose-v2`** instead. Install **`docker.io` first** so the **`docker`** group exists, then **`sudo usermod -aG docker "$USER"`** and log out/in (or `newgrp docker`).

Detailed native Postgres steps (peer auth, roles): [`INSTALL.md`](../INSTALL.md).

### Split hosts (DB vs API vs UI / Pi / VPS)

Same codebase everywhere: point **`DATABASE_URL`** at wherever Postgres runs (**host, port, password**, **`sslmode`** for TLS). The DB must provide **TimescaleDB**, **PostGIS**, and **pgvector** (Compose [`db/Dockerfile`](../db/Dockerfile); bare metal [`scripts/install-system-deps-debian.sh`](../scripts/install-system-deps-debian.sh)). Run **`./scripts/bootstrap-local.sh --seed`** (or migrations only) against that URL once per environment; API and UI read **`DATABASE_URL`** / **`VITE_API_URL`** from `.env` like local dev.

## First clone (recommended for new contributors)

From the repository root after `git clone`, run:

```bash
./scripts/setup-first-clone.sh
```

Or **`make first-clone`**. This runs `go mod download`, copies `.env` / `ui/.env` from examples if missing, then **`scripts/bootstrap-local.sh`**. You still need PostgreSQL created with extensions first for the native path — see [INSTALL.md](../INSTALL.md). **Debian/Ubuntu:** install Postgres stack + Node with **`./scripts/install-system-deps-debian.sh`** (sudo apt), or combine with **`./scripts/setup-first-clone.sh --install-system-deps`** (`make first-clone-install-deps`). For a machine without local Postgres, use **`./scripts/setup-first-clone.sh --docker`** or **`make first-clone-docker`**.

For how the schema is defined (and why ad-hoc ERD screenshots may be outdated), see [database-schema-overview.md](database-schema-overview.md).

## One-command bootstrap

From the repository root:

```bash
./scripts/bootstrap-local.sh
```

Options:

| Flag | Meaning |
|------|---------|
| `--docker` | `docker compose up -d` instead of host `psql` schema steps |
| `--seed` | Load [`db/seeds/master_seed.sql`](../db/seeds/master_seed.sql) (legacy demo **farm_id = 1**). Omit if you rely on dashboard **New farm** + template choice. |
| `--skip-schema` | Skip only [`db/schema/gr33n-schema-v2-FINAL.sql`](../db/schema/gr33n-schema-v2-FINAL.sql) (use when enums/tables already exist); **`db/migrations/*.sql` still runs** |

The script copies [`.env.example`](../.env.example) to `.env` **once** if `.env` is missing, then runs `npm ci --legacy-peer-deps` in `ui/` (Capacitor peer ranges need this until versions are aligned).

**Make equivalent:** `make bootstrap-local` (same as the script without flags). Use `make bootstrap-local-docker` for the Docker path.

## After a reboot (same DB volume — no full reinstall)

Typical delay when running **`make dev-auth-test`** is **Go compiling** the API (`go run` builds before listening); that can take **several minutes** on a cold machine and is **not** an infinite loop. The automation worker may also log many rule evaluations shortly after startup — that is normal.

**Quick path:** from the repo root:

```bash
make restart-local        # docker compose db only + wait + db sanity report
make dev-auth-test        # API + UI (compile happens here unless you use a pre-built binary)
```

**`make dev-auth-test`** uses the same port checks as **`restart-local.sh --serve`** ([`scripts/maybe-serve-api-ui.sh`](../scripts/maybe-serve-api-ui.sh)): if `:8080/health` and `:5173` already respond, it leaves them up; if only one is running, it starts the missing piece; if a port is taken but unhealthy, it errors instead of spawning a second stack on `:5174`.

Or one line including servers: **`make restart-local-serve`** (same as `./scripts/restart-local.sh --serve`).

- **`scripts/restart-local.sh`** does **not** run **`bootstrap-local.sh`** — your existing schema and data stay as-is. Use **`make dev-stack`** / **`./scripts/bootstrap-local.sh`** when migrations or seed need applying.
- **`make db-sanity-report`** (or **`scripts/db-sanity-report.sh`**) prints extensions, farm count, duplicate zone names (these break **`master_seed.sql`**), and RAG chunk count. Non-zero exit means fix duplicates or treat DB as unhealthy for seeding.

## When localhost (DB / API / UI) is not running

**Docker:** from the repo root run `docker compose up -d --build` (or `make bootstrap-local-docker`). The **`db`** service only runs Postgres; load schema + optional seed with **`./scripts/bootstrap-local.sh --seed`** (or **`make dev-stack`**, which does that). Dashboard: **http://localhost:5173** · API: **http://localhost:8080** when using full Compose with **`api`+`ui`**. Postgres from Compose is exposed on **localhost:5433** (maps to 5432 inside the container; avoids colliding with OS Postgres on **5432**) — credentials in [`docker-compose.yml`](../docker-compose.yml).

**Native:** follow [INSTALL.md](../INSTALL.md) for Postgres + extensions, then `./scripts/bootstrap-local.sh`, set **`DATABASE_URL`** in `.env`, then **`make dev`** (API + UI together) or **`make run`** and **`make ui`** in two terminals.

### Unblock “API offline” / failed startup (checklist)

1. **`.env` `DATABASE_URL`** must match the Postgres you actually use. Common mistake: leaving the placeholder `user:password` from [`.env.example`](../.env.example). **Compose DB:** after `make compose-db-up`, use `postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable` (host port **5433**). **Native peer:** see [INSTALL.md §2d](../INSTALL.md).
   - **One-shot after Docker is installed:** **`make dev-stack`** (recommended) — runs [`scripts/dev-stack.sh`](../scripts/dev-stack.sh): retries **`docker compose`** through **`sg docker`** when `/var/run/docker.sock` denies access, builds/starts **`db`**, **`bootstrap --seed`**, **`check-stack`**. Same as **`make setup-compose-dev`** (wrapper). **`make local-up`** runs **`dev-stack`** then **`make dev-auth-test`** (full API + UI). **`./scripts/dev-stack.sh --reset-volumes`** wipes Compose volumes before bring-up (destructive — fresh DB).
   - **Docker `permission denied` on `/var/run/docker.sock`:** after `sudo usermod -aG docker "$USER"`, your *current* terminal may still lack the `docker` group. Run **`newgrp docker`**, or **`sg docker -c 'cd …/gr33n-platform && make setup-compose-dev'`**, or **log out and back in**; confirm with **`groups`** (should list `docker`).
2. **`pgvector`** — the API registers the `vector` type; if the extension is missing, startup fails with `vector type not found`. Install packages (e.g. `./scripts/install-system-deps-debian.sh` for PG16 + extensions) or use the Compose `db` image.
3. **Verify without guessing:** **`make check-stack`** (runs [`scripts/check-local-stack.sh`](../scripts/check-local-stack.sh)) — connects with `DATABASE_URL`, checks `vector`, optionally curls `/health`. After a reboot you can use **`make restart-local`** (starts Compose **`db`** only + waits + **`make db-sanity-report`**) before **`make dev-auth-test`**.
4. **UI → API:** [`ui/.env.example`](../ui/.env.example) → `ui/.env` with `VITE_API_URL=http://localhost:8080` if you changed the API port.
5. **Auth test mode:** `JWT_SECRET` and `PI_API_KEY` must be set in `.env` when using **`make dev-auth-test`** (see `.env.example`).
6. **Operational logs (production / LAN):** Set **`LOG_FORMAT=json`** when piping logs to a stack; **`docker-compose.yml`** rotates **json-file** logs per service; optional **`make compose-logging-up`** merges **`docker-compose.logging.yml`** (Loki + Promtail + Grafana demo stack). Details **[operator-logging-runbook.md](operator-logging-runbook.md)**.

## Order of operations

1. **Database** — Full schema: `db/schema/gr33n-schema-v2-FINAL.sql`. Upgrades on older snapshots: apply `db/migrations/*.sql` in **filename sort order** (the bootstrap script does this after the schema).
2. **Environment** — Root [`.env.example`](../.env.example): `DATABASE_URL`, `AUTH_MODE`, and for real auth `JWT_SECRET` / `PI_API_KEY`. The API loads `.env` then `.env.local` from the repo root. To export secrets into your shell: **`source scripts/source-local-env.sh`** (refresh smoke JWT: **`source scripts/source-local-env.sh --refresh-eval-token`**).
3. **API** — `make run` (dev auth bypass) or `make run-auth` / production-style config; see comments in `.env.example`.
4. **UI** — `make ui` or `make dev` (API + UI). Copy [`ui/.env.example`](../ui/.env.example) to `ui/.env` if you need a non-default API URL (`VITE_API_URL`; otherwise the code defaults to `http://localhost:8080`).

**Auth-regression style (real JWT + farm checks, `auth_test` mode):** set in **`.env`** (or export) at least `AUTH_MODE=auth_test`, `JWT_SECRET` (long random), and `PI_API_KEY` (see [`.env.example`](../.env.example)). From the repo root: **`make dev-auth-test`** — same as `make dev` but the API runs with **`AUTH_MODE=auth_test`** (see [`.env.example`](../.env.example) and the `dev-auth-test` target in the [Makefile](../Makefile)). The API still loads `.env` on startup; you must be in the **project root** when you start it.
5. **Optional: Insert Commons receiver** — `make run-receiver`; env and migrations: [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md).
6. **Optional: Pi client / MQTT** — OS packages: [`scripts/install-pi-edge-deps.sh`](../scripts/install-pi-edge-deps.sh). Then [`pi_client/setup.sh`](../pi_client/setup.sh), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md). Full topologies (edge vs all-on-Pi vs split servers): [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md). Python deps: `pi_client/requirements.txt`.

### Edge loop in 5 commands (Phase 31 WS1)

Prove the **field path on a laptop** before wiring a Pi: `pi_client` uses **stub drivers** when GPIO libraries are absent, posts readings with **`X-API-Key`**, and the dashboard **Live Sensors** card updates via SSE (`GET /farms/{id}/sensors/stream`).

| Step | Command |
|------|---------|
| 1 | `make dev-stack` — DB, schema, [`master_seed.sql`](../db/seeds/master_seed.sql) (demo **farm_id = 1**) |
| 2 | `make dev-auth-test` — API + UI in one terminal; requires **`JWT_SECRET`** and **`PI_API_KEY`** in [`.env`](../.env.example) |
| 3 | `./scripts/print-demo-sensor-ids.sh` — list numeric **`sensor_id`** values for master_seed sensor names (re-run if you seeded more than once) |
| 4 | `./scripts/run-edge-stub-client.sh` — resolves **`sensor_id`** from DB names, installs Python deps, runs stub **`pi_client`** (or manual: `cp config.demo-stub.yaml` + edit **`api.api_key`**) |
| 5 | Open **http://localhost:5173** → **gr33n Demo Farm** → confirm **Live Sensors** show values (not **NO DATA**) within ~1s |

Shortcut: **`make edge-smoke-help`** prints the same steps.

**Real Pi on a bench:** after the stub loop works, follow **[`pi-integration-guide.md` §8 — Field checklist](pi-integration-guide.md#8-field-checklist--first-pi-on-a-real-bench-phase-31-ws2)** (power, relay safety, `PI_API_KEY`, three-tier zone naming, offline queue drill, `TestPiContract*` links).

**Actuator round-trip (Phase 31 WS3):** [`pi-integration-guide.md` §9](pi-integration-guide.md#9-safe-actuator-e2e--pending_command-round-trip-phase-31-ws3) — `./scripts/run-edge-actuator-smoke.sh --direct` or two-terminal `./scripts/run-edge-actuator-client.sh` + `./scripts/enqueue-demo-pending-command.sh on`. Safety: [operator-troubleshooting.md §5](operator-troubleshooting.md#5-edge-actuator-safety-phase-31-ws3).

**Sensor IDs:** [`pi_client/config.demo-stub.yaml`](../pi_client/config.demo-stub.yaml) maps **`sensor_id`** to master_seed names for a **fresh** `make dev-stack-fresh` (e.g. **3** = Air Temp Indoor, **5** = Air Humidity Indoor). Duplicate seed runs can shift ids — align with step 3 or use a clean volume via **`make dev-stack-fresh`**.

**Automation simulation (off path for WS1):** [`.env.example`](../.env.example) sets **`AUTOMATION_SIMULATION_MODE=true`**. The automation worker then records **simulated** actuator events and does **not** enqueue **`pending_command`** on devices. That is intentional for laptop demos: **`pi_client`** supplies **real ingest** for readings only. To exercise GPIO / **`pending_command`** round-trip (Phase 31 WS3), set **`AUTOMATION_SIMULATION_MODE=false`** and bind actuators to real **`device_id`** rows — see [`pi-integration-guide.md`](pi-integration-guide.md).

**Verify without the UI:** after step 4, log in via the dashboard once, then `GET /sensors/{id}/readings/latest` with a JWT returns the stub value (Pi key alone is ingest-only).

## API integration smoke tests

Run from repo root: `go test -tags dev ./cmd/api/...` (or `make test`, which includes this package). The `cmd/api` tests build an in-memory `httptest` server wired like production, with **`AUTH_MODE=auth_test`** and fixed test-only `JWT_SECRET` / `PI_API_KEY` (not read from your `.env`).

| Requirement | Notes |
|---------------|--------|
| **`DATABASE_URL`** | Must point at Postgres that already has **full schema** (`db/schema/gr33n-schema-v2-FINAL.sql`) and **migrations** applied (same order as bootstrap). Export it in the shell before `go test`, or rely on the Makefile default `DB_URL` when you run `make test`. |
| **`-tags dev`** | Required so `auth_test` mode is allowed (`make test` sets this). |
| **Seed data** | Recommended: [`db/seeds/master_seed.sql`](../db/seeds/master_seed.sql) (demo **farm_id = 1**, sensors, NF inputs, crop cycles, etc.). A few tests **skip** if expected rows are missing (e.g. “no sensors in seed”, “no NF inputs in seed data”). |
| **Smoke pollution** | Repeated `make test` against one DB accumulates junk (extra `bootstrap_farm_*` rows, `ph208_zone_*`, mass automation-rule alerts). For a clean Guardian demo, reset with `make dev-stack` or a fresh Compose volume + bootstrap. |
| **OpenAPI parity** | `make audit-openapi` (shell diff) **and** `go test ./cmd/api/ -run TestOpenAPI_AllRoutesDocumented` (Go guard in [`openapi_parity_test.go`](../cmd/api/openapi_parity_test.go)). Both should pass before push. |
| **No database** | If the pool cannot open or ping, `TestMain` prints a **stderr hint** and exits **0** locally (so `go test ./...` without Postgres does not fail every package). In **CI** (`CI=true` or `GITHUB_ACTIONS`), the same condition exits **1** so a forgotten DB service does not look green. |
| **Unset `DATABASE_URL`** | Tests use a **Linux peer-auth default** (`postgres://davidg@/gr33n?host=/var/run/postgresql`). Override with `DATABASE_URL` if your user or socket path differs. |

Do not use `go test -shuffle=on` on this package as a gate — smoke tests share package-level state (see Phase 20.95 plan notes).

## First user and auth

- **`AUTH_MODE=dev`** (default in `make run` / `make dev`): use the UI **Register** flow or `POST /auth/register` with `email`, `password` (minimum 8 characters), optional `full_name`.
- **Production**: set `AUTH_MODE=production`, `JWT_SECRET`, and `PI_API_KEY`; optional env-admin login via `ADMIN_USERNAME` + `ADMIN_PASSWORD_HASH` in `.env` (see `.env.example`).

## Insert Commons and custom integrators

Farm-side pipeline and **strict ingest JSON** (only six top-level keys; no extra fields): [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) — read **Custom senders** before POSTing from scripts or third-party tools.

## Audit and operator index

- Farm audit API and sensitive actions: [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md).
- Phase 14 playbook index (MQTT, commons catalog, notifications, etc.): [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).

## OpenAPI route audit

From the repo root, `make audit-openapi` runs [`scripts/openapi_route_diff.sh`](../scripts/openapi_route_diff.sh). It diffs **(HTTP method, path)** pairs from [`cmd/api/routes.go`](../cmd/api/routes.go) against [`openapi.yaml`](../openapi.yaml) and exits non-zero on any mismatch — run it after you add or rename HTTP routes.

Additionally, **`go test ./cmd/api/ -run TestOpenAPI_AllRoutesDocumented`** ([`cmd/api/openapi_parity_test.go`](../cmd/api/openapi_parity_test.go)) runs as part of `make test` and enforces the same contract from Go — catches drift even if the shell script is skipped.

## Guardian-ready demo (after seed)

Farm Guardian layers three knowledge sources ([`farm-guardian-architecture.md`](farm-guardian-architecture.md)):

1. **Llama weights** — install Ollama + pull model ([`farm-guardian-ollama-setup.md`](farm-guardian-ollama-setup.md)).
2. **RAG corpus** — seed loads operational rows but **not** embeddings. After `make seed`, run **`make rag-ingest-demo`** (needs `EMBEDDING_API_KEY`; skips with a message if unset). One-shot fresh demo with embeddings: **`make dev-stack-fresh-rag`**.

   **Platform how-to docs** (Confirm workflow, Pi setup, operator tour, **greenhouse climate §5b**, lighting §5) are a separate index: **`make rag-ingest-platform-docs`**. Re-run whenever you change curated markdown in [`docs/rag/platform-doc-manifest.yaml`](rag/platform-doc-manifest.yaml) — ingest is idempotent per file. Dry-run without API key: `./scripts/rag-ingest-platform-docs.sh --dry-run`.

   **Full bootstrap (field guides + platform + operational):** `make guardian-bootstrap-farm FARM_ID=1`. **CPU laptop pitfalls** (slow phi3, embed contention, manual `ollama stop`): [guardian-ollama-laptop-playbook.md](guardian-ollama-laptop-playbook.md).

   **Phase 135 — Settings corpus card:** After first bootstrap, **Settings → Field memories (RAG corpus)** shows chunk counts, last-ingest age, and amber staleness when operational rows are >7d old. Farm admins can **Re-ingest** per scope (`field_guides`, `platform_docs`, `operational`) without terminal scripts — requires LAN `EMBEDDING_BASE_URL` + `EMBEDDING_API_KEY`. `GET /v1/chat/health?farm_id=N` exposes the same `awakening.corpus` block.

3. **Live snapshot** — built automatically on each grounded chat turn (zones, active cycles, unread alerts).

**Phase 29 WS7 — sample unread alerts:** [`db/seeds/master_seed.sql`](../db/seeds/master_seed.sql) inserts three unread `alerts_notifications` rows for demo **farm_id = 1** (OHN inventory low, Flower Room humidity high, 12/12 light transition reminder). Re-run **`make seed`** or **`make dev-stack-fresh`** to apply; subjects are idempotent.

### Guardian agent demo in 3 commands

From the repo root (destructive DB wipe — use only when you want a clean demo farm):

```bash
make dev-stack-fresh-rag    # or: make dev-stack-fresh  (skip RAG if EMBEDDING_API_KEY unset)
make restart-local-serve    # API + UI (or: make dev-auth-test in one terminal)
# Dashboard → select gr33n Demo Farm → toggle Guardian (sidebar, ✨ TopBar, or right-edge tab)
# 1) Ask: "What unread alerts do I have?" or use ✨ Ask Guardian on the humidity alert row
# 2) Then: "acknowledge the humidity alert" → proposal card → Confirm
```

With **AI_ENABLED** and Ollama running, grounded chat includes the three seed alerts in the live snapshot. **Change requests** use proposal cards + **Confirm** (`POST /v1/chat/confirm`); pending items also appear in the drawer **Pending** tab and **`/guardian/requests`**. Audit rows: `guardian_tool_executed`. See [operator tour §6](operator-tour.md#6-farm-guardian-change-requests-with-your-ok) and [farm-guardian-architecture §8](farm-guardian-architecture.md#8-operator-expectations-at-phase-30-ship).

**Hardware expectations:** Guardian chat is GPU/LLM-bound on weak laptops — see [recommended-hardware-and-sizing.md](recommended-hardware-and-sizing.md) (dev vs production profiles, Lite mode without GPU). **CPU-only laptops:** [guardian-ollama-laptop-playbook.md](guardian-ollama-laptop-playbook.md) (RAG bring-up, `ollama stop`, stale `ollama run`, what “CPU” means in the UI).

**Real grow (live plants):** do not skip **[guardian-real-grow-readiness.md](guardian-real-grow-readiness.md)** — ingest checklist, Confirm vs automation, bench actuators first, Phase 82/83 bootstrap when shipped.

If your DB has been used for smoke tests for weeks, you may see hundreds of thousands of stale automation alerts and extra test farms — reset with **`make dev-stack-fresh`** for a clean demo farm.

## Slow UI and dev DB hygiene

Local slowness after many dev sessions is usually **data volume**, not Vue bundle size:

| Symptom | Likely cause | Quick fix |
|---------|--------------|-----------|
| Dashboard takes seconds to load | Duplicate sensors from re-running **`make seed`** / bootstrap on the same Docker volume; each refresh hits many `/readings/latest` rows | **`./scripts/dev-reset-farm.sh --profile small_indoor`** (Phase 48) or **`make dev-stack-fresh`** (volume wipe) |
| Hundreds of `/sensors/{id}/readings/latest` 404s | Stale sensor IDs in UI cache vs DB | **`make dev-reset-farm`** or fresh seed; UI batch-loads via `GET /farms/{id}/sensors/readings/latest` |
| Guardian drawer sluggish | Mass automation-rule alerts from smoke tests | **`make db-sanity-report`** then **`make dev-reset-farm`** |

**Phase 48 profiles:** [`dev-farm-profiles.md`](dev-farm-profiles.md) — `small_indoor` (daily dev / sit-in) vs `demo_showcase` (operator tour). Farm 1 tag: `farms.meta_data.dev_seed_profile`.

```bash
make dev-reset-farm ARGS="--farm-id 1 --profile small_indoor"
make dev-reset-farm ARGS="--farm-id 1 --profile demo_showcase --include-readings"
make db-sanity-report
# Optional hypertable retention (dev only):
TIMESCALE_RETENTION_DAYS=90 make apply-dev-retention
```

**Do not** use **`./scripts/restart-local.sh`** alone when the API is down — it only starts Postgres unless you pass **`--serve`**. From repo root use **`make dev-auth-test`** (API + UI) or **`make local-up`**.

Canonical plan: [`docs/plans/phase_48_dev_seed_and_small_farm_profiles.plan.md`](plans/phase_48_dev_seed_and_small_farm_profiles.plan.md).

**Multi-site / enterprise (hypothetical):** how 500 warehouse-scale sites map onto org/farm/zone + commons recipe packs — no core software changes required: [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md). **Phase 30** — Guardian PR queue (config + Pi via confirm): [`plans/phase_30_guardian_change_requests.plan.md`](plans/phase_30_guardian_change_requests.plan.md). **Phase 31** — Pi/breadboard field validation: [`plans/phase_31_field_validation_and_edge.plan.md`](plans/phase_31_field_validation_and_edge.plan.md).

**Edge vs dashboard auth in the spec:** paths wrapped with `requireAPIKey` in `routes.go` are **Pi / bridge** calls using header **`X-API-Key`** (same secret as `PI_API_KEY` in `.env`). `GET /farms/{id}/devices` uses **`requireJWTOrPiEdge`**: OpenAPI lists **both** `bearerAuth` and `apiKeyAuth` so operators know the Pi may poll device `config` (including `pending_command`) with the API key while the dashboard uses a JWT.

## Security notes

Bootstrap keeps **secrets and TLS** in your hands: the script does not generate passwords or certificates. Use real secrets in production; do not commit `.env`.
