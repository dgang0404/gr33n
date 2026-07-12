# gr33n — current state

> **Generated:** 2026-07-11 · Regenerate after major phase ship · **Canonical history:** [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md) · **Numbers hint:** `make docs-current-state-hint`

---

## What it is

**gr33n** is an AGPL v3, self-hosted farm operating system: PostgreSQL + Go API + Vue 3 SPA + optional Raspberry Pi edge client. Data stays on your LAN; Guardian chat can run fully local via Ollama (`LLM_BASE_URL`).

**New clone path:** [README](../README.md) → this page → [operator tour](operator-tour.md) → [first session after clone](first-session-after-clone.md).

---

## Shipped capabilities (at a glance)

| Area | What works today |
|------|------------------|
| **Sensors & alerts** | Live dashboards, SSE stream, rules, unread alert inbox |
| **Control** | Manual toggles, cron schedules, automation rules, Pi `device_commands` FIFO queue |
| **Zones** | Zone cockpit — Water / Light / Climate tabs, plants, tasks, grow cycles |
| **Guardian** | Farm Counsel (RAG + live data) vs Quick Chat; proposals → Confirm; **full citation deep links** (schedule, alert, docs); accuracy banners **persist on reload** |
| **Crops** | Postgres catalog (~50 crops), `crop_key` on plants, Guardian `lookup_crop_targets` |
| **Edge** | Pi client, MQTT bridge, Virtual Pi wiring, `/pi-setup-wizard` |
| **Ops** | Costs/receipts, tasks, audit events, optional Insert Commons export |
| **Quality** | `make test-unit`, `make backup`, `make vuln-check`, `make guardian-qa-smoke` |

---

## UI workspaces & routes

| Route | Workspace |
|-------|-----------|
| `/` | Today dashboard |
| `/zones`, `/zones/:id` | Zones (inline hub: overview, water, light, climate, plants, tasks, alerts) |
| `/feed-water`, `/money`, `/hardware`, `/comfort-targets` | Legacy workspace entry points (zone-first redirects where applicable) |
| `/chat`, `/guardian/requests` | Farm Guardian + pending change-request tab |
| `/settings` | Farm, Guardian, crops, QA, feedback |
| `/virtual-pi`, `/pi-setup`, `/pi-setup-wizard` | Pi wiring & config |
| `/catalog`, `/farm-knowledge`, `/symptom-guide` | Commons, RAG knowledge, symptoms |
| `/crop-cycles/:id/summary` | Grow run summary (Guardian citation target) |

Source: [`ui/src/router/index.js`](../ui/src/router/index.js).

---

## API surface

**OpenAPI tags:** health, auth, farms, zones, sensors, devices, actuators, automation, lighting, tasks, plants, costs, fertigation, naturalfarming, alerts, profiles, rag, **chat**, capabilities, commons, organizations, units, crop-cycle-analytics.

**Guardian (`/v1/chat`):** grounded chat, proposals queue (`GET /v1/chat/proposals`), feedback export, model list/pull, QA run metadata.

Spec: [`openapi.yaml`](../openapi.yaml) · live Redoc when API is up at `/openapi`.

---

## Postgres schemas

| Schema | Role |
|--------|------|
| `auth` | Users, invites, sessions |
| `gr33ncore` | Farms, zones, sensors, devices, tasks, alerts, RAG, Guardian turns |
| `gr33nfertigation` | Programs, crop cycles, mixing |
| `gr33ncrops` | Plants, crop catalog (DB source of truth) |
| `gr33nnaturalfarming` | JADAM / natural farming batches |
| `gr33nanimals`, `gr33naquaponics` | Opt-in domain stubs (`farm_active_modules`) |

Migrations: `db/migrations/` · overview: [`database-schema-overview.md`](database-schema-overview.md).

---

## Farm Guardian

| Mode | Behavior |
|------|----------|
| **Farm Counsel** | Grounded chat — RAG chunks, live read tools, `[n]` citations, optional proposal cards |
| **Quick Chat** | LLM-only (smaller models allowed; no grounded minimum context) |
| **Change requests** | Propose → operator **Confirm** → audited write; inbox at `/guardian/requests` |

**Smoke & QA**

```bash
make guardian-qa-smoke              # artifact run (always exits 0)
make guardian-qa-smoke-strict       # pass/fail heuristics
make guardian-qa-change-requests    # internal proposal queue persistence
make guardian-qa-change-requests-confirm  # propose → Confirm → DB (Phase 162)
make guardian-eval -manual          # UI checklist
```

Architecture: [`farm-guardian-architecture.md`](farm-guardian-architecture.md) · CI (opt-in): [`ci-guardian-qa.md`](ci-guardian-qa.md).

---

## Edge / Pi

- **Telemetry:** `POST /sensors/readings/batch`, MQTT bridge ([`pi_client/`](../pi_client/))
- **Actuation:** `device_commands` queue (FIFO) + legacy `pending_command` mirror
- **Config:** Virtual Pi export, push-config to device, Pi setup wizard

Playbooks: [`pi-integration-guide.md`](pi-integration-guide.md) · [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md).

---

## Operator env knobs (top)

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | Postgres connection |
| `JWT_SECRET`, `AUTH_MODE` | Auth (`dev` / `auth_test` / `production`) |
| `LLM_BASE_URL`, `LLM_MODEL` | Guardian provider (Ollama default) |
| `CROP_CATALOG_SOURCE` | `db` (default) or `yaml` |
| `FILE_STORAGE_DIR` | Receipt blobs (local) |
| `GUARDIAN_COST_GUARD` | Token budget (`off` in dev) |

Full list: [`environment-variables.md`](environment-variables.md).

---

## Infra & trust (Phases 154–158)

| Command | Purpose |
|---------|---------|
| `make test-unit` | Fast Go tests (no DB smokes) |
| `make backup` / `make verify-backup` | Automated farm backup |
| `make vuln-check` | govulncheck + npm audit |
| `make docs-current-state-hint` | Regenerate OpenAPI/migration counts for this page |

Accessibility: skip link, Guardian drawer focus trap, zone tab semantics — [`a11y-audit-2026-07-11.md`](a11y-audit-2026-07-11.md).

---

## Not shipped / partial

| Item | Notes |
|------|--------|
| **Insert Commons** | Opt-in federation; not required for single-farm LAN |
| **Hosted-only** | Not required — but `LLM_BASE_URL` supports remote OpenAI-compatible APIs |

---

## Phase history

- **Shipped arcs:** 40–67 farmer UX · 68–81 SPA · 82–110 crop intelligence · 111–122 Guardian/Pi · 129–153 Guardian QA · **154–161** infra/trust + citation + a11y + ec-ph trim
- **Active / planned:** Insert Commons (opt-in); full `smoke-ec-ph` re-run on CPU (operator)
- **Archive:** [`plans/archive/`](plans/archive/) — closed plans (e.g. 88–92)
