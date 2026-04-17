---
name: Phase 18 Platform polish & integration hardening
overview: >
  Solidify the platform before RAG/off-line AI: Part A shipped in-app guidance (HelpTip),
  collapsible sidebar + mobile nav drawer, and expanded API smoke tests. Part B syncs
  OpenAPI with routes, verifies the Pi client against the API, and adds a workflow
  guide for operators and future RAG training.
todos:
  - id: ws1-guidance
    content: "WS1: In-app guidance — HelpTip.vue; contextual help on Dashboard, Fertigation, Schedules, Tasks, Plants"
    status: completed
  - id: ws2-sidebar
    content: "WS2: Collapsible sidebar — hamburger, grouped nav (Operate/Grow/Monitor/System), icon rail, localStorage; TopBar mobile drawer"
    status: completed
  - id: ws3-smoke-tests
    content: "WS3: Smoke tests — AlertLifecycle (skip if no seed alerts), CropCycleFullCRUD, reservoir/program PATCH+DELETE, NF input+batch+recipe CRUD, profile GET+PUT, schedule active toggle + actuator-events"
    status: completed
  - id: ws4-openapi
    content: "WS4: OpenAPI sync — Plants, Task PUT/DELETE, Schedule paths; full audit routes.go vs openapi.yaml"
    status: completed
  - id: ws5-pi-integration
    content: "WS5: Pi client — post_readings_batch test in test_gr33n_client.py; docs/pi-integration-guide.md (Pi → API → UI)"
    status: completed
  - id: ws6-workflow-guide
    content: "WS6: docs/workflow-guide.md — farm→zones→sensors→schedules→fertigation→crops→tasks→alerts→costs (RAG-ready)"
    status: completed
isProject: false
---

# Phase 18 — Platform polish & integration hardening

## Why this phase

Phase 17 added **Commons Catalog**, **Plants CRUD**, and **mobile hardening** (push, PWA, responsive tweaks). Phase 18 closes the loop: **operators understand how features connect**, the **UI uses space better**, **automated tests** cover more API surface, and **spec + edge client + narrative docs** stay aligned before investing in **Llama RAG** trained on schemas and operator content.

## Split

| Part | Focus | Location in repo |
|------|--------|------------------|
| **A** | Guidance UI, sidebar/drawer, smoke tests | `ui/src/components/HelpTip.vue`, `SideNav.vue`, `TopBar.vue`, `App.vue`; `cmd/api/smoke_test.go` |
| **B** | OpenAPI, Pi verification, workflow guide | `openapi.yaml`, `pi_client/test_gr33n_client.py`, `docs/pi-integration-guide.md`, `docs/workflow-guide.md` |

Reference plan (duplicate / scratch): `.cursor/plans/phase_18_polish_hardening.plan.md` — this file under `docs/plans/` is the **canonical** plan for docs and cross-links.

## Part A — Completed (summary)

- **WS1:** `HelpTip.vue` with hover/click popover; integrated on Dashboard, Fertigation (page + tabs), Schedules, Tasks, Plants.
- **WS2:** Sidebar collapse (`gr33n_sidebar_collapsed`), grouped nav, mobile hamburger in TopBar + drawer in `App.vue`.
- **WS3:** New smoke tests for crop cycles, fertigation reservoir/program update+delete, NF inputs/batches/recipes, profile, schedule active toggle; alert test validates list + `unread_count` and skips if no seed alerts.

**Known unrelated smoke failures (data/template):** `TestFarmBootstrapOnCreate` and `TestOrgDefaultBootstrapOnFarmCreate` may expect more zones than the current template produces — fix separately if desired.

## Part B — To do

### WS4: OpenAPI sync

- Add **Plants**: `GET/POST /farms/{id}/plants`, `GET/PUT/DELETE /plants/{id}` (align path params and schemas with handlers).
- Add **Tasks**: `PUT /tasks/{id}`, `DELETE /tasks/{id}` (and any missing fields on create/list if drifted).
- Add **Schedules**: `PUT /schedules/{id}`, `DELETE /schedules/{id}`, `PATCH /schedules/{id}/active`, `GET /schedules/{id}/actuator-events` if missing.
- **Audit:** Walk `cmd/api/routes.go` `mux.Handle` registrations and ensure every JWT/Pi route exists in `openapi.yaml` with method, path, and request/response shapes (or explicitly marked out-of-spec with a short comment in the plan).

### WS5: Pi client integration verification

- Extend **`pi_client/test_gr33n_client.py`** with coverage for **`POST /sensors/readings/batch`** (same auth and error handling patterns as existing tests).
- Add **`docs/pi-integration-guide.md`**: how the Pi posts readings / device status / actuator events; how those show up in the API and UI; env vars (`GR33N_API_URL`, API key), and link to [`openapi.yaml`](../../openapi.yaml) after WS4.

### WS6: Workflow guide

- Add **`docs/workflow-guide.md`** in plain language: how **zones**, **sensors/actuators**, **schedules**, **automation runs**, **fertigation** (reservoirs, EC targets, programs, mixing, crop cycles, events), **tasks**, **alerts**, **costs**, and **catalog/plants** relate. This document is the primary **operator narrative** and a good **RAG chunk source** after WS4 tightens machine-readable spec.

## After Part B

- **Documentation audit:** Refresh README, operator docs, and any runbooks that reference routes or Pi flows once OpenAPI and the new guides exist.
- **RAG prep:** Chunk `workflow-guide.md`, schema summaries, and curated OpenAPI paths for offline help.

---

## Using this plan in a new chat (copy-paste prompt)

Use the block below as your **first message** in the next session (adjust paths if your clone differs):

```text
Implement Phase 18 Part B per @docs/plans/phase_18_platform_polish.plan.md.

Scope:
1) WS4 — Sync openapi.yaml with cmd/api/routes.go: add missing Plants, Task, Schedule (and related) paths; audit all mux routes vs spec; keep YAML valid and consistent with actual JSON bodies.
2) WS5 — Pi client: add tests for POST /sensors/readings/batch in pi_client/test_gr33n_client.py (mirror existing patterns); add docs/pi-integration-guide.md describing Pi → API → UI and env/API key usage.
3) WS6 — Add docs/workflow-guide.md as the operator-facing “how it all connects” narrative (farm, zones, sensors, schedules, fertigation, tasks, alerts, costs) suitable for future RAG.

Constraints: minimal unrelated refactors; run go test ./cmd/api/... and any Pi client tests you touch; update the plan frontmatter todo statuses when done.

Optional follow-up (only if quick): investigate smoke failures TestFarmBootstrapOnCreate / TestOrgDefaultBootstrapOnFarmCreate (zone count vs template) — do not block Part B on this unless trivial.
```
