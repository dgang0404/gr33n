# gr33n — current state

> **Generated:** 2026-07-14 · Regenerate after major phase ship · **Canonical history:** [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md) · **Numbers hint:** `make docs-current-state-hint`

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
| **Crops** | Postgres catalog (~52 crops), `crop_key` on plants, Guardian `lookup_crop_targets` |
| **Edge** | Pi client, MQTT bridge, Virtual Pi wiring, `/pi-setup-wizard` |
| **Ops** | Costs/receipts, tasks, audit events, optional Insert Commons export |
| **Quality** | `make test-unit`, `make backup`, `make vuln-check`, `make guardian-qa-smoke` |

---

## Demo farm seed (`farm_id = 1`, Phase 164)

After `make seed` or `make dev-stack-fresh`, **gr33n Demo Farm** is a **living** operational snapshot (not cannabis-themed):

| Aspect | Demo contents |
|--------|----------------|
| **Zones** | 7 — Veg/Flower rooms, propagation, herbs, outdoor beds/patch |
| **Crops** | Chrysanthemum veg/bloom runs (`Anastasia Green`, `Zembla White`), basil, tomato, pepper, strawberry, etc. |
| **Cannabis** | **Not** in demo plant/cycle rows — catalog + field guides still include `cannabis` |
| **Sensors** | Wired sensors have recent `sensor_readings`; bed sensors stay unwired (“not set up yet”) |
| **Attention** | Flower Room humidity 72.4% RH matches a seeded alert |
| **Irrigation** | Herb & Greens **gravity-fed drip** (`Herb Room Gravity Drip`, plain water) |

Smoke assertions: `go test ./cmd/api/... -run Phase164` (needs seeded DB).

## Today farm canvas API (Phase 165)

Backend + store plumbing for the visual farm map (Phase 166 UI):

- **Zone layout** — `zones.meta_data.layout` `{x,y,w,h}` (normalized 0–1); validated on `PUT /zones/{id}`; server merges with existing meta keys (greenhouse climate, photos).
- **Farm background** — `POST/GET/DELETE /farms/{id}/layout-background`; attachment id at `farms.meta_data.layout_background_attachment_id`; image via `/file-attachments/{id}/content`.
- **Store** — `saveZoneLayout`, `zoneLayout`, `loadLayoutBackground`, `uploadLayoutBackground`, `clearLayoutBackground` in `ui/src/stores/farm.js`.

## Today visual farm canvas (Phase 166)

The **Today** tab (`/`) is a spatial farm map:

- **FarmSiteStrip** — sunrise/sunset/daylength, outdoor sensor rollup, water source hint
- **FarmCanvas** — draggable zone tiles over optional background photo; arrange mode persists layout via Phase 165 API
- **FarmCanvasZoneTile** — plants, light, water, sensor health per zone (healthy / needs attention / not set up yet)
- **Dashboard** — canvas is the hero; tasks/alerts/schedules/sensors/actuators live under collapsed “All the details”

## Mobile stack + quick actions (Phase 167)

On phones (`md` breakpoint), Today shows **stacked zone cards** instead of the spatial canvas. Tapping any zone opens a **quick-action sheet**:

- Water now (program `run-now` or actuator pulse)
- Light on/off, greenhouse vent/shade when applicable
- Complete today's tasks / acknowledge alerts inline
- Ask Guardian with zone-scoped prompts
- Open zone → `/zones/:id`

## Today cleanup + polish (Phase 168)

Phase 168 finishes the Today redesign arc (164–167):

- **Removed** IT-style getting-started checklist from Dashboard — growers see a farm, not sysadmin todos
- **Empty farm** — canvas/stack CTA + Guardian setup chips (when 0 zones or 0 devices)
- **Copy sweep** — farmer-facing zone type labels on tiles and quick-action sheet; vocabulary test covers new Today surfaces
- **Docs** — operator tour §7k; `phase-168-closure.test.js`

## Today attention cockpit (Phase 169)

When zones need care, Today surfaces them explicitly:

- **Attention strip** — compact chips above the farm map (warn/alert zones); tap opens quick actions
- **Canvas sort** — desktop tiles order attention-first (parity with mobile stack)
- **Guardian attention starters** — contextual chips when flagged zones exist (`buildTodayAttentionStarters`)

## Today Guardian one-tap counsel (Phase 170)

Today starters that need farm data (morning check, attention chips, zone quick actions) open the Guardian drawer in **Farm counsel** and **auto-send** — one tap, same as in-panel morning walkthrough. Setup starters still prefill only.

## Demo farm zone layouts (Phase 171)

After `make seed`, farm-1 zones include `meta_data.layout` positions matching the Today canvas defaults — the spatial map renders correctly on first open without manual arrange.

## Field guide expansion (Phase 172)

Demo-farm crop guides (chrysanthemum, basil, pepper, strawberry) expanded for Guardian RAG. **Marigold** and **geranium** added to the crop catalog with new field guides. Regenerate after edits: `./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql` then `make rag-ingest-field-guides`.

## Today excellence arc (Phases 173–177 — shipped)

Locked roadmap after Phase 172: [`phase_173_177_today_excellence_roadmap.plan.md`](plans/phase_173_177_today_excellence_roadmap.plan.md) · Operator tour [§7l](operator-tour.md#7l-today-excellence-phases-173177--shipped)

| Phase | Focus |
|-------|--------|
| **173** ✅ | Large farms — filter chips, mobile paging (8/page), desktop Map/List toggle beyond 13 zones |
| **174** ✅ | Visual hierarchy — **Today** naming, `FarmTodayHeader` health pills, taller canvas, tile polish |
| **175** ✅ | Farm-first — action bar; ≤2 Ask gr33n chips; full starters in details |
| **176** ✅ | Farm pulse — next water, growing runs, devices in Site Strip (same card) |
| **177** ✅ | First impression — demo seed polish, `TodayCoachMarks`, perf/a11y closure |

**North star:** `/` reads as a grower cockpit, not an AI chat launcher. Fresh `make dev-stack-fresh` opens to sun, pulse, zones, and at most two Ask chips — Guardian stays in the sidebar and details.

## Today visual hierarchy (Phase 174)

- **FarmTodayHeader** — farm name, time greeting, health rollup pills (`N healthy`, `N need attention`, tasks, alerts)
- TopBar and browser tab say **Today** (not Dashboard)
- Attention pill applies the Phase 173 **Needs attention** filter
- Canvas min-height increased; zone tiles get hover lift and attention glow

## Today farm-first actions (Phase 175)

- **FarmTodayActionBar** below the map — Feed & water, New task, What runs when, My zones
- **FarmTodayAskGr33n** — at most two curated chips (morning check + ask about your farm)
- Full Guardian starter set (attention, weather, ops) moved under **All the details → Ask gr33n**
- Zone-scoped Guardian remains in the quick-action sheet (Phase 170 one-tap counsel unchanged)

## Today farm pulse (Phase 176)

`FarmSiteStrip` now includes operational pulse cells in the **same card** — no extra row:

- **Next water** — earliest active feeding plan + zone name
- **Lights** — zones on now, or next light schedule
- **Growing** — active crop runs and bloom count
- **Devices** — online count and command queue depth

## Today first impression (Phase 177)

- **Demo seed** — propagation room gets 24h T5 light so ≥5/7 zones show plants plus water or light on tiles
- **TodayCoachMarks** — three-step first visit (farm map → tap zone → attention or pulse); session dismiss via `gr33n_today_coach_done`; no Guardian step
- **Perf** — `refreshAll()` paints cached zones immediately; weather, layout background, and queue depth load in background
- **A11y** — attention strip `aria-live="polite"`; coach controls meet 44px touch targets

## Online weather forecast (Phase 178)

Optional Tier 3 forecast on top of Phase 66 offline solar math:

- **API** — `WEATHER_PROVIDER=openmeteo` (free, no key); farm opt-in via `meta_data.weather_forecast_enabled` + **Settings → Farm site**
- **`GET /farms/{id}/site-weather`** — `online_forecast` block with status (`connected`, `cached`, `cached_stale`, `offline`, `disabled`, …)
- **Today** — `FarmSiteStrip` forecast cell + `● Forecast live` / `cached (offline)` badge (sun dial unchanged when WAN drops)
- **Guardian** — `site_weather` read tool cites tonight low + frost when forecast tier is present

Plan: [`phase_178_online_weather_forecast.plan.md`](plans/phase_178_online_weather_forecast.plan.md) · Operator tour [§7n](operator-tour.md#7n-online-weather-forecast-phase-178--shipped)

## Guardian chat polish (Phases 179–182)

Sit-in follow-ups on full-page `/chat` and nav polling:

| Phase | Focus |
|-------|--------|
| **179** ✅ | One streaming status row; awakening panel quiet during local stream; mode cards collapse after first turn |
| **181** ✅ | Full-page composer diet — `+ Attach photos, starters, mode` toggle after turn 1; pending badge **only on TopBar** (sidebar shows readiness dot) |
| **182** ✅ | 401 → stop nav poll + single login redirect; Pending tab scroll + newest-first; Refine hint under composer |

Vitest: `phase-179-closure.test.js`, `phase-181-closure.test.js`, `phase-182-closure.test.js`

## Help Library + contextual knowledge (Phase 183)

- **Library hub** — Help opens on one **Library** page (`tab=library`) with scrollable sections (guide, knowledge, symptoms, catalog); legacy `?tab=knowledge` etc. still resolve
- **Contextual links** — **Symptoms for this crop** from Plants, zone grow strip, and alert cards → `/symptom-guide?crop_key=…`
- **Task revise** — rule-based `create_task` title/description corrections bump `Revision` (parity with feed volume revise)

Plan: [`phase_183_guardian_knowledge_and_revise_followups.plan.md`](plans/phase_183_guardian_knowledge_and_revise_followups.plan.md) · Operator tour [§7m](operator-tour.md#7m-help-knowledge-surfaces-phase-180--shipped)

## Task zone revise (Phase 185)

- **Zone assignment** — pending `create_task` / `create_task_from_alert` proposals accept name-based zone turns (`Put it in Veg Room — that is the zone…`) and numeric `zone N` / `zone id N` patterns; each bumps `Revision` like title/description revise
- **Smoke** — `scenario-task-dialogue-pending` now runs create → zone assign → title revise (`MinRevision: 3`, `RequireTaskZone`)

Plan: [`phase_185_guardian_task_zone_revise.plan.md`](plans/phase_185_guardian_task_zone_revise.plan.md) · Vitest: `phase-185-closure.test.js`

## Task due_date revise (Phase 186)

- **Due date on create** — Guardian `create_task` / `create_task_from_alert` now persist optional `due_date` (`YYYY-MM-DD`) on Confirm
- **Due date revise** — pending task proposals accept `set the due date to …` / `due date should be …` turns; bumps `Revision` like other task fields
- **Smoke** — `scenario-task-dialogue-pending` runs create → zone → title → due date (`MinRevision: 4`, `WantDueDate`)

Plan: [`phase_186_guardian_task_due_date_revise.plan.md`](plans/phase_186_guardian_task_due_date_revise.plan.md) · Vitest: `phase-186-closure.test.js`

## Multi-turn PR smoke (Phase 184)

`make guardian-qa-change-requests-ui` runs **5 multi-turn scenarios** (shared `session_id` per dialogue): 1 feed revise **confirmed via API** + 4 left pending (feed, task, schedule, ack) for manual Confirm/Refine/Dismiss on `/chat?tab=pending`. Quick subset: ack + schedule (`change-requests-ui-quick`).

Plan: [`phase_184_guardian_pr_conversation_smoke.plan.md`](plans/phase_184_guardian_pr_conversation_smoke.plan.md) · [`ci-guardian-qa.md`](ci-guardian-qa.md)

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
| `/operator-guide` | Help — **Library** hub (guide, knowledge, symptoms, catalog sections) + Pi setup ([§7m](operator-tour.md#7m-help-knowledge-surfaces-phase-180--shipped)) |
| `/catalog`, `/farm-knowledge`, `/symptom-guide` | Redirect into Help Library sections (`tab=library&section=…`) |
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
make guardian-qa-change-requests-ui       # multi-turn: 1 confirm + 4 pending (Phase 184)
make guardian-qa-change-requests-ui-quick # fast subset: ack + schedule pending
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
| `WEATHER_PROVIDER` | Online forecast: `off` (default), `openmeteo`, … — see Phase 178 |
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

- **Shipped arcs:** 40–67 farmer UX · 68–81 SPA · 82–110 crop intelligence · 111–122 Guardian/Pi · 129–153 Guardian QA · **154–161** infra/trust + citation + a11y + ec-ph trim · **164–177** visual Today farm cockpit + excellence arc · **178** online weather forecast · **179–186** Guardian UX polish, Help Library, multi-turn PR smoke, task revise chain
- **Active / planned:** Insert Commons (opt-in); Phase 184 WS5 live Pending-tab verification (operator); full `smoke-ec-ph` re-run on CPU (operator)
- **Archive:** [`plans/archive/`](plans/archive/) — closed plans (e.g. 88–92)
