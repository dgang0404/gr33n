# Changelog

All notable **operator-visible** changes to gr33n. Internal refactors and test-only work are omitted unless they affect upgrades or behavior you would notice in the field.

Format: coarse era blocks for early history, then per-phase entries from Phase 113 onward.

---

## Unreleased

### Bug fixes

- Zone **How it connects** pipeline links now navigate (sensors → comfort → automation → controls → device); hover still highlights the sidebar tab
- **GPIO board** and **Pi setup** pages no longer request `/farms/undefined` when farm context is missing

### Phases 119–123 — Virtual Pi wiring arc

- **119** — Read-only 40-pin board at `/virtual-pi` (Wiring sidebar)
- **120** — Tap-to-wire GPIO, relay stack view (Sequent 8-layer HAT), conflict highlights
- **121** — Driver hookup steps, print wiring sheet, `config.yaml` download, `config_sha256` drift badge
- **122** — `make guardian-eval` model quality scores; context budget guard; proposal JSON repair
- **123** — **Notify Pi to reload** (`POST /devices/{id}/push-config`) for platform-sync edge devices

### Phase 116 — Documentation refresh

- Central [environment variables reference](docs/environment-variables.md) with CI parity check (`make audit-env`)
- [Upgrade guide](docs/upgrade-guide.md), [backup & restore runbook](docs/backup-restore-runbook.md)
- [API quickstart](docs/api-quickstart.md) curl cookbook
- `/openapi` Redoc browser on dev builds (set `OPENAPI_UI=true` in production)
- README status through Phase 115; screenshots in `docs/images/`

### Phase 115 — Schema utilization

- Farm **module toggles** (Animals, Aquaponics, …) in Settings; disabled modules hide nav and return HTTP 403
- **Notification template** CRUD + picker in automation/fertigation forms
- **Diagnostics** panel (system logs) in Settings
- Symptom guide at `/symptom-guide`
- Task **estimated duration** + **actual start/end** on complete
- Alert **delivery status** on Alerts page
- Dropped unused `validation_rules` table

### Phase 114 — Pi / edge integrity

- Stale device heartbeats mark devices offline and raise alerts
- Pi telemetry on status patch; command queue cancel; mixing events with device keys
- Relay HAT smbus calibration workflow; command queue inspector

### Phase 113 — Security hardening

- Registration modes (`open` / `invite` / `closed`); login rate limit
- JWT no longer accepted in query strings; security headers
- Guardian chat cost guard defaults; password change endpoint
- Per-device Pi keys preferred over shared `PI_API_KEY`

---

## Phase 111–112 — Guardian model selection & Ollama hardening

- In-app **model selector** and **pull** workflow for Ollama installs
- Model discovery cache; health checks; smoke coverage for Ollama E2E

---

## Phase 82–110 — Crop intelligence & SPA maturity

- Crop catalog and agronomy field guides in Postgres
- Workspace-first navigation (Today, Zones, Comfort, Money)
- Grow hub: crop cycles, stage history, harvest compare, economics
- Guardian write tools, proposals, Confirm gate, session history
- RAG retrieval + synthesis; offline field assistant procedures
- Device command queue; automated mixing on Pi; lighting programs
- Pi setup wizard; natural farming inventory; cost ledger & receipts

**Breaking / migration notes:** run `make migrate` after each pull; several phases add columns and hypertables. See [upgrade guide](docs/upgrade-guide.md).

---

## Phase 68–81 — SPA workspaces

- Dashboard, zone cockpit, Feed & water hub, Money workspace
- Mobile PWA patterns; offline task queue
- Legacy routes redirect into workspaces (bookmarks still work)

---

## Phase 40–67 — Farmer UX & Guardian depth

- Unified zone tabs (Water / Light / Climate)
- Farm Guardian chat, grounded snapshot, change requests
- Push-to-talk field assistant; vision attachments (optional)
- Automation rules + schedules operator UI

---

## Phase 10–33 — Core platform

- PostgreSQL schema v2, Go API, Vue dashboard
- JWT + Pi API key auth; farm RBAC
- Sensors, actuators, schedules, tasks, alerts
- Fertigation, natural farming, animals, aquaponics modules
- RAG ingest pipeline; Insert Commons opt-in

---

## How to read older detail

Per-phase closure notes and plans live under [`docs/plans/`](docs/plans/) and the [operator documentation index](docs/phase-14-operator-documentation.md).
