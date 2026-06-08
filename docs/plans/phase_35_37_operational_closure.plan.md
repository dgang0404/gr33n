---
name: Phase 35–37 operational closure (seed, bootstrap, docs, tests)
overview: >
  Cross-phase workstream for operator-facing parity after feature code lands: demo seed,
  bootstrap templates, OpenAPI, operator-tour, architecture notes, Go smokes, and Vitest.
  Each feature phase (35–37) keeps a final WS8 (or OC slice); this doc tracks what shipped
  vs deferred so nothing is marked "done" without docs/tests/seed alignment.
todos:
  - id: oc-35-seed-bootstrap
    content: "OC-35A: jadam_indoor_photoperiod_v1 bootstrap → lighting_program + paired schedules (Phase 35 WS6 remainder)"
    status: done
  - id: oc-35-docs-openapi
    content: "OC-35B: operator-tour 18/6 lighting walkthrough; OpenAPI LightingProgram + schedule-action paths; architecture grow-stack note (Phase 35 WS8)"
    status: done
  - id: oc-35-tests
    content: "OC-35C: cmd/api smoke preset apply + TZ cron; Vitest PhotoperiodClockEditor linked fields (Phase 35 WS8)"
    status: done
  - id: oc-36-bootstrap
    content: "OC-36A: greenhouse_climate_v1 bootstrap v2 → zone_type=greenhouse, typed actuators, meta profile, lux rules (20260603_phase36_greenhouse_climate_v2.sql, 0916aba)"
    status: done
  - id: oc-36-docs-openapi
    content: "OC-36B: operator-tour §5b greenhouse; OpenAPI GreenhouseClimate + POST actuators + rule-templates; architecture §7.0c cross-links (Phase 36 WS8)"
    status: done
  - id: oc-36-tests
    content: "OC-36C: cmd/api smokes — bootstrap apply, rule fire + cooldown, manual shade deploy via pending_command (Phase 36 WS8)"
    status: done
  - id: oc-37-closure
    content: "OC-37: Phase 37 WS8 — offline field walkthrough, procedure OpenAPI, field_guide corpus ingest smoke, safety-stop smokes"
    status: done
  - id: oc-37-final-sweep
    content: "OC-37E: End-of-37 sweep — verify OC-35A–C closed; platform-doc RAG manifest includes new operator-tour sections; README roadmap"
    status: done
  - id: oc-38-closure
    content: "OC-38: Phase 38 shipped — plant-needs zone tabs, nav Advanced group, pulse duration_seconds (no schema migration)"
    status: done
  - id: oc-39-closure
    content: "OC-39: Phase 39 WS8 — device_commands queue smokes, mix plan unit tests, pi-integration-guide queue+mix_batch, operator-tour automated mix"
    status: done
  - id: oc-39b-closure
    content: "OC-39b: irrigation_only programs — migration, worker skip mix, UI badge, smoke (commits with 39b)"
    status: done
  - id: backlog-operator-runtime
    content: "Product backlog B1–B4 — run-now, metadata.steps counter, create_lighting_program, mobile checklist (see product_backlog_operator_runtime.plan.md)"
    status: done
  - id: bug-guardian-nav
    content: "BUG (pre-40): Guardian edge tab + sidebar overlap — see phase_40 plan bug-guardian-nav"
    status: done
  - id: oc-40-closure
    content: "OC-40: Phase 40 WS8 — zone cockpit operator-tour §4b, architecture §7.0f, Vitest inline setpoints + Today strip (close when Phase 40 ships, not before)"
    status: done
  - id: oc-41-closure
    content: "OC-41: Phase 41 WS7 — farm hub operator-tour §3, architecture §7.0g, why-empty Vitest (close when Phase 41 ships, after 40)"
    status: completed
  - id: oc-42-closure
    content: "OC-42: Phase 42 WS7 — comfort targets + plain automation docs/tests (close when Phase 42 ships)"
    status: completed
  - id: oc-43-closure
    content: "OC-43: Phase 43 WS7 — operations hub docs/tests (close when Phase 43 ships)"
    status: completed
  - id: oc-44-closure
    content: "OC-44: Phase 44 WS6 — setup + edge wizard docs/tests (close when Phase 44 ships)"
    status: completed
  - id: oc-45-closure
    content: "OC-45: Phase 45 WS7 — farmer sit-in + farmer-ready v1 (close when Phase 45 ships)"
    status: completed
  - id: oc-46-closure
    content: "OC-46: Phase 46 WS6 — LLM tool proposals docs/tests (close when Phase 46 ships)"
    status: completed
  - id: oc-47
    content: "OC-47: Phase 47 WS7 — feeding & water plain language docs/tests (close when Phase 47 ships)"
    status: completed
  - id: oc-48-closure
    content: "OC-48: Phase 48 WS7 — dev seed profiles, idempotent seed, reset script, sanity report (close when Phase 48 ships)"
    status: completed
  - id: oc-49-closure
    content: "OC-49: Phase 49 WS4 — sidebar nav polish, Fertigation rename, related-route hover (close when Phase 49 ships)"
    status: completed
  - id: oc-50-closure
    content: "OC-50: Phase 50 WS6 — hardware wiring visibility, pi-config generator, sanity report, docs/tests (close when Phase 50 ships)"
    status: completed
  - id: oc-51-closure
    content: "OC-51: Phase 51 WS6 — Pi config platform sync, live reload, staleness badge, import script, docs/tests (close when Phase 51 ships)"
    status: completed
  - id: oc-52-closure
    content: "OC-52: Phase 52 — Guardian UI context, Pi setup guide, nav-hint wiggle chains (shipped)"
    status: completed
  - id: oc-53-closure
    content: "OC-53: Phase 53 WS6 — grow/stock/money closure, cross-links, Guardian starters, phase-53-closure.test.js"
    status: completed
  - id: oc-54-closure
    content: "OC-54: Phase 54 WS4 — zone connection pipeline, orphan link wiggles, phase-54-closure.test.js"
    status: completed
  - id: oc-55-closure
    content: "OC-55: Phase 55 WS5 — Guardian ops read tools, starters, phase_55_guardian_pr_spec.md"
    status: completed
  - id: oc-56-closure
    content: "OC-56: Phase 56 WS5 — plant_id migration, compare flow, phase-56-closure.test.js"
    status: completed
  - id: oc-57-closure
    content: "OC-57: Phase 57 WS5 — per-device API keys, pi guide, security smokes"
    status: completed
  - id: oc-58-closure
    content: "OC-58: Phase 58 WS4 — task consumptions UI, templates, phase-58-closure.test.js"
    status: completed
  - id: oc-59-closure
    content: "OC-59: Phase 59 WS4 — enterprise-tier-boundary.md, copy audit, index links"
    status: pending
  - id: oc-60-closure
    content: "OC-60: Phase 60 WS5 — morning walkthrough, walk_farm tool, operator-tour §6i, phase-60-closure.test.js"
    status: completed
  - id: oc-61-closure
    content: "OC-61: Phase 61 WS5 — proactive nudge dot, dismiss, operator-tour, phase-61-closure.test.js"
    status: completed
  - id: oc-62-closure
    content: "OC-62: Phase 62 WS5 — grow advisor, VPD starters, post-harvest, farm-guardian-architecture §7.0x"
    status: completed
  - id: oc-63-closure
    content: "OC-63: Phase 63 WS5 — session memory, topic tags, privacy note, delete, phase-63-closure.test.js"
    status: completed
  - id: oc-64-closure
    content: "OC-64: Phase 64 WS6 — crop knowledge base, 7 seeded profiles, lookup_crop_targets, grounding guard test"
    status: completed
  - id: oc-65-closure
    content: "OC-65: Phase 65 WS4 — Pi & hardware diagnostics, summarize_device_health, fieldGuideGrounding update"
    status: completed
  - id: oc-66-closure
    content: "OC-66: Phase 66 WS6 — weather & site, offline solar engine, ingestion tiers, supplemental-light starter"
    status: pending
  - id: oc-67-closure
    content: "OC-67: Phase 67 WS7 — hands-free field assistant, voice in/out, crop-grounded photo diagnosis"
    status: pending
isProject: false
---

# Phase 35–39 operational closure (seed, bootstrap, docs, tests)

## Why this doc exists

Feature phases often land **code first** (schema, API, UI, worker) while **seed data**, **bootstrap templates**, **operator-tour**, **OpenAPI**, and **integration smokes** trail behind. That leaves uncommitted or “invisible” files in git and a false sense of completion.

This plan is the **rollup tracker** for closure work across Phases **35 → 39**. Each feature phase keeps its own **WS8** (or WS5 bootstrap + WS8 docs in Phase 36). This doc says **what is done, what is deferred, and when to close it**.

---

## Closure checklist (every feature phase)

Use this table when marking a phase shipped:

| Layer | Artifacts | Done when |
|-------|-----------|-----------|
| **Demo seed** | `db/seeds/master_seed.sql` | Fresh seed demonstrates the new domain entity (not only legacy rows) |
| **Bootstrap** | `_bootstrap_*` in migrations / `apply_farm_bootstrap_template` | New farms from template get the new model, idempotent |
| **Unit tests** | `internal/**/**_test.go`, Vitest | Core logic + UI component behavior |
| **Integration smokes** | `cmd/api/smoke_phase*_test.go` | HTTP round-trip against real DB in CI |
| **OpenAPI** | `openapi.yaml` | Paths, request/response schemas, examples |
| **Operator docs** | `docs/operator-tour.md`, workflow/guide cross-links | Walkthrough an operator can follow without reading code |
| **Architecture / Guardian** | `farm-guardian-architecture.md`, persona mirror if tools changed | Grow stack + tool catalog accurate |
| **RAG manifest** (optional) | `docs/rag/platform-doc-manifest.yaml` | New operator sections ingested (Phase 32 WS8 pattern) |

**Git hygiene:** closure PR should include **all** new files (migrations, sqlc, handlers, UI, tests, docs) — not only modified tracked files.

---

## Phase 35 — Lighting domain

### Shipped (implementation)

| Area | Status | Notes |
|------|--------|-------|
| Schema + migration | ✅ | `20260603_phase35_lighting_programs.sql`, sqlc, CRUD handler |
| Presets + from-preset API | ✅ | peas_22_2, veg_18_6, flower_12_12, seedling_16_8 |
| Schedule-action API | ✅ | GET/POST `/schedules/{id}/actions` |
| TZ-aware worker | ✅ | `shouldTriggerNow(expr, tz, …)` + unit test |
| UI | ✅ | `PhotoperiodClockEditor.vue`, `LightingPrograms.vue`, `/lighting` route |
| Guardian read + propose | ✅ | `summarize_zone_lighting`; **`create_lighting_program`** (product backlog B3) |
| **Demo seed** | ✅ | `master_seed.sql` Section 3B wraps 18/6 in `lighting_programs` |
| **Bootstrap** | ✅ | `jadam_indoor_photoperiod_v1` → `lighting_programs` (OC-35A migration) |
| **Unit tests** | ✅ | `handler_test.go`, `worker_test.go` TZ case |
| **Smokes / Vitest** | ✅ | `smoke_phase35_lighting_test.go`; `photoperiod-clock-editor.test.js` |
| **OpenAPI / operator-tour** | ✅ | `LightingProgram` schemas; operator-tour §5; architecture §7.0b |

## Phase 35 — status

**Shipped.** OC-35A–C closed; WS1–WS8 complete. Optional follow-up: RAG re-ingest of operator-tour §5 (part of OC-37E sweep).

---

## Historical note

Phase 35–36 implementation commits are on `main`; Phase 36 **WS4 UI, WS6 interlocks, OC-36B/C** are **closed**. Use the status tables below, not this note, for current state.

---

## Phase 36 — Greenhouse climate

Feature detail: [`phase_36_greenhouse_climate.plan.md`](phase_36_greenhouse_climate.plan.md). Closure maps to **OC-36A–C** (mirrors OC-35A–C).

### Shipped (implementation)

| Area | Status | Notes |
|------|--------|-------|
| Zone climate profile (WS1) | ✅ | `meta_data.greenhouse_climate`; validation on zone POST/PUT when `zone_type=greenhouse` |
| Actuator taxonomy (WS2) | ✅ | `shade_screen`, `ridge_vent`, fans; `POST/GET /farms/{id}/actuators`, `GET /actuators/{id}` + `valid_commands` |
| Automation templates (WS3) | ✅ | `POST /farms/{id}/automation/rule-templates/greenhouse`; bootstrap lux/temp/vent rules (inactive) |
| Bootstrap → core (WS5) | ✅ | **OC-36A** — [`20260603_phase36_greenhouse_climate_v2.sql`](../../db/migrations/20260603_phase36_greenhouse_climate_v2.sql) |
| Guardian read (WS7) | ✅ | `summarize_zone_greenhouse_climate`; `enqueue_actuator_command` deploy/retract/open/close/stop |
| **Greenhouse UI (WS4)** | ✅ | ZoneDetail Climate tab, typed command buttons, sensor strip |
| **Sensor interlocks (WS6)** | ✅ | Missing lux/PAR banner; template guard without override |
| **Demo seed** | ✅ partial | Bootstrap apply suffices; optional greenhouse row in master_seed |
| **Unit tests** | ✅ partial | `greenhouse_test.go`, `taxonomy_test.go` |
| **Smokes / Vitest** | ✅ | **OC-36C** — `smoke_phase36_oc36c_test.go` (+ WS4-prep pending_command) |
| **OpenAPI / operator-tour** | ✅ | **OC-36B** — operator-tour §5b; OpenAPI paths/schemas |
| **Architecture** | ✅ | §7.0c in `farm-guardian-architecture.md` + operator-tour cross-links |

### Phase 36 — status

**Shipped.** WS1–WS7, WS4 UI, **WS6** interlocks, **OC-36A–C** closed.

Apply migration `20260603_phase36_greenhouse_climate_v2.sql` before re-running `greenhouse_climate_v1` bootstrap on existing dev DBs.

**Do not** fold Phase 36 into Phase 35 closure — different domain, same checklist pattern.

---

## Phase 37 — Guardian offline field assistant

Closure is **WS8** in [`phase_37_guardian_offline_field_assistant.plan.md`](phase_37_guardian_offline_field_assistant.plan.md). Map to **OC-37**:

| When | Work |
|------|------|
| After WS2 (field corpus) | Ingest `field_guide` sources; extend platform-doc manifest |
| After WS3–WS4 (procedures + safety) | OpenAPI procedure endpoints; smokes for step flow + safety hard-stop |
| After WS9 (background chat) | Vitest `guardian-chat-background.test.js`; operator-tour note: stream continues while browsing farm pages |
| WS8 | operator-tour “first install with Guardian offline”; link Pi wiring procedure to Phase 35 actuator path |

**OC-37E — End-of-37 sweep** ✅ (2026-06-03)

1. **OC-35A–C** and **OC-36B–C** closed (lighting + greenhouse operator docs/smokes).
2. Re-run **`make rag-ingest-platform-docs`** and **`make rag-ingest-field-guides`** on each farm after operator-doc / field-guide edits (requires `EMBEDDING_API_KEY` / LAN embedding endpoint).
3. README + phase-14: Phases **35–37 shipped**; Phase 38/39 tracked separately.

---

## Recommended schedule across phases

```
Phase 35 code PR  ──► OC-35A bootstrap (same sprint or +1)
                   ──► OC-35B + OC-35C docs/tests (before Phase 36 UI references lighting)

Phase 36 WS1–3,5,7 ──► OC-36A + OC-36B ✅
                   ──► WS4 + WS6 + OC-36C (remaining ship)

Phase 37 WS9     ──► Pinia guardianChat (can land before WS1 — no backend dependency)
Phase 37 WS1–7   ──► OC-37 inline with WS8
                   ──► OC-37E final sweep (OC-35A–C + OC-36B–C verified; RAG ingest)
```

**Rule:** Feature WS8 stays in each phase plan; **this doc** is the cross-phase backlog so deferred items are not lost when a phase plan todo is marked `done` too early.

---

## Phase 38 — Plant-needs UI + pulse

Feature detail: [`phase_38_plant_needs_ui_and_pulse_commands.plan.md`](phase_38_plant_needs_ui_and_pulse_commands.plan.md). **OC-38: done.**

| Area | Status | Notes |
|------|--------|-------|
| Zone Water/Light/Climate tabs | ✅ | All zones; connection cards |
| Nav Grow / Advanced | ✅ | [`navGroups.js`](../../ui/src/lib/navGroups.js) |
| `duration_seconds` on pending_command | ✅ | **No DB migration**; JSONB only |
| Command queue | ❌ | **Deferred to Phase 39 WS1** |

---

## Phase 39 — Edge fertigation execution

Feature detail: [`phase_39_edge_fertigation_execution.plan.md`](phase_39_edge_fertigation_execution.plan.md). **OC-39 complete** (WS8 docs, smokes, seed, OpenAPI 0.4.5).

**Phase 39b** (plain irrigation): [`phase_39b_plain_irrigation.plan.md`](phase_39b_plain_irrigation.plan.md) — **OC-39b done**.

**Product backlog** (run-now, steps counter, lighting propose, mobile checklist): [`product_backlog_operator_runtime.plan.md`](product_backlog_operator_runtime.plan.md) — **done** (OpenAPI 0.4.6); commit on `main` before Phase 40 kickoff.

---

## Pre–Phase 40 gate (start feature work only when these are green)

| Gate | Required before Phase 40 WS1? | Status |
|------|-------------------------------|--------|
| Phases **35–37** OC slices (seed, bootstrap, docs, smokes) | **Yes** | ✅ OC-35A–C, OC-36A–C, OC-37 + OC-37E |
| Phase **38** plant-needs + pulse | **Yes** | ✅ OC-38 |
| Phase **39** + **39b** runtime (queue, mix, irrigation_only) | **Yes** | ✅ OC-39, OC-39b |
| **bug-guardian-nav** hotfix | **Yes** (UX baseline) | ✅ |
| **Product backlog** B1–B4 | **Yes** (operator day-2; small diff) | ✅ code — ensure **committed** on `main` |
| **OC-40-closure** (operator-tour §4b, Vitest cockpit, arch §7.0f) | **No** — this *is* Phase 40 **WS8** | ✅ OC-40 |
| **OC-41-closure** (farm hub tour, why-empty Vitest) | **Yes** — Phase **41** WS7 | ✅ closed |
| Phase **41** feature work (dashboard hub, `?zone_id=`) | **No** | Planned after 40 |

**Pending rows in this plan’s todo list:** only **OC-41** (and OC-42+) are intentionally open — they track *future* closure. Phase 40 **OC-40** is closed.

**Documentation gate (before Phase 40 code):** [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) + [guardian_pr_ux_through_farmer_phases.plan.md](guardian_pr_ux_through_farmer_phases.plan.md) + per-phase Guardian specs **42–46** + [phase_47](phase_47_feeding_water_plain_language.plan.md) + [farmer-vocabulary.md](../farmer-vocabulary.md) + [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md). **Green to start Phase 40.**

**Optional hygiene (not blocking):** `make rag-ingest-platform-docs` after doc edits; push `main` to origin when ready.

---

## Phase 40 — Unified farmer UX (zone cockpit)

Feature detail: [`phase_40_unified_farmer_ux_zone_cockpit.plan.md`](phase_40_unified_farmer_ux_zone_cockpit.plan.md). **OC-40 complete** (WS1–WS8 shipped).

| Area | Status | Notes |
|------|--------|-------|
| Guardian nav hotfix | ✅ | `bug-guardian-nav` — Ask gr33n top of sidebar; full-page chat under System |
| Today strip + inline comfort targets | ✅ | WS1–WS2 — `ZoneTodayStrip`, `ZoneComfortTargets` |
| Zone rules/schedules/alerts | ✅ | WS3–WS4 — `ZoneAutomationPanel`, `ZoneAlertsPanel` |
| Water grow story + zone tasks | ✅ | WS5–WS6 — `ZoneWaterGrowStory`, `ZoneTasksPanel` |
| Nav IA + Guardian starters | ✅ | WS7–WS7b — `navGroups.js`, `GuardianStarterChips` |
| OC-40 docs/tests | ✅ | WS8 — operator-tour §4b, architecture §7.0f, `zone-cockpit.test.js` |

## Phase 41 — Farm hub coherence

Feature detail: [`phase_41_farm_hub_coherence.plan.md`](phase_41_farm_hub_coherence.plan.md). **OC-41** closed (WS7).

| Area | Status | Notes |
|------|--------|-------|
| Dashboard morning cockpit | ✅ | WS1 — `FarmMorningStrip`, `farmGrowSummary.js` |
| Fertigation `?zone_id=` | ✅ | WS2 — `ZoneContextBanner`, program highlight |
| Cross-page zone filter | ✅ | WS3 — Tasks, Schedules, Alerts, Automation |
| Why-empty hints | ✅ | WS4 — `EmptyStateHint.vue` |
| Seed tasks `zone_id` | ✅ | WS5 — `master_seed.sql` comment + demo rows |
| Lighting ↔ zone Light | ✅ | WS6 — `/lighting?zone_id=`, shared copy |
| OC-41 docs/tests | ✅ | WS7 — operator-tour §3b, architecture §7.0g, Vitest |

## Phase 43 — Operations hub (stock, feeding admin, money)

Feature detail: [`phase_43_operations_stock_feeding_finance.plan.md`](phase_43_operations_stock_feeding_finance.plan.md). **OC-43** closed (WS7). **WS8** Guardian read + starters shipped — [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md).

| Area | Status | Notes |
|------|--------|-------|
| Operations nav group | ✅ | WS1 — Supplies, Feeding (details), Money |
| Supplies hub | ✅ | WS2 — `SuppliesHub.vue`, low-stock banner, `?zone_id=` |
| Feeding admin hub | ✅ | WS3 — `FeedingAdminHub.vue`, card tabs, mixing escape |
| Money hub | ✅ | WS4 — `MoneyHub.vue`, month summary, receipt form |
| Cross-links | ✅ | WS5 — zone Water, Dashboard **Supplies low** chip |
| Guardian persona + impact | ✅ | WS6 — ops vocabulary, refill task cites input name |
| OC-43 docs/tests | ✅ | WS7 — operator-tour §7 + §6f, architecture §7.0i, `phase-43-closure.test.js` |
| Guardian read + starters | ✅ | WS8 — `summarize_farm_low_stock`, ops starter chips, `guardian-ops-starters.test.js` |

## Phase 44 — Getting started & edge wizards

Feature detail: [`phase_44_getting_started_edge_wizard.plan.md`](phase_44_getting_started_edge_wizard.plan.md). **OC-44** closed (WS6). **WS8** Guardian PR slice partial — [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md).

| Area | Status | Notes |
|------|--------|-------|
| Farm setup wizard | ✅ | WS1 — `/farms/:id/setup`, template cards, bootstrap POST |
| Add zone wizard | ✅ | WS2 — greenhouse profile, lighting preset |
| Edge device wizard | ✅ | WS3 — Pi checklist, poll online, actuators |
| Guardian setup mode | ✅ | WS4 — `setup_mode.go`, starters, wizard footers |
| First-run checklist | ✅ | WS5 — `GettingStartedChecklist` on Dashboard |
| OC-44 docs/tests | ✅ | WS6 — operator-tour §8 + §6g, architecture §7.0j, Vitest + `TestPhase44WizardBootstrapApply` |
| Guardian empty-zone starters | ✅ | WS8 — `empty_zone_grow` on zone cockpit |

## Phase 45 — Farmer validation & whole-app polish

Feature detail: [`phase_45_farmer_validation_whole_app_polish.plan.md`](phase_45_farmer_validation_whole_app_polish.plan.md). **OC-45** closed (WS7). **Phase 45 shipped** — WS2/WS8 dry-run; [phase_45_guardian_pr_spec.md](phase_45_guardian_pr_spec.md).

| Area | Status | Notes |
|------|--------|-------|
| Sit-in protocol + scorecard | ✅ | WS1 — `farmer-sit-in-protocol.md`, `sit-in-45-session-log-template.md` |
| Vocabulary v2 (zones not rooms) | ✅ | WS3 — `farmerVocabulary.js`, grow-path Vitest |
| Module empty shells | ✅ | WS5 — Animals/Aquaponics `ModuleEmptyShell` |
| Light a11y | ✅ | WS6 — `farmerA11y.js`, Confirm/Dismiss/chips |
| OC-45 docs/tests | ✅ | WS7 — README Farmer-ready v1, operator-tour §9, architecture §7.0k, `phase-45-closure.test.js` |
| Live sit-in + friction backlog | ✅ | WS2 — dry-run DR-A/DR-B; P0 empty |
| Mobile sit-in path (PWA) | ✅ | WS4 — `phase-45-ws4-mobile-sit-in-path.md`, prep scripts; store deferred |
| Guardian PR path validation | ✅ | WS8 — ack, setup pack, dismiss; `sit-in-dry-run.sh` |

## Phase 46 — Guardian LLM tool proposals

Feature detail: [`phase_46_guardian_llm_tool_proposals.plan.md`](phase_46_guardian_llm_tool_proposals.plan.md). **OC-46** closed (WS6). **Phase 46 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Hybrid policy + allowlist | ✅ | WS1 — `proposals_llm.go`, `GUARDIAN_LLM_PROPOSALS` |
| Schema + farm ID binding | ✅ | WS2 — `proposals_llm_validate.go` |
| Chat handler hook | ✅ | WS3 — `attachProposals` + assistant text |
| Safety tests | ✅ | WS4 — `proposals_llm_safety_test.go`, `smoke_phase46_ws4_test.go` |
| Observability logs | ✅ | WS5 — `proposals_observability.go` |
| OC-46 docs/tests | ✅ | WS6 — guide §3.3, operator-tour §6h, architecture §7.0l, `phase-46-closure.test.js` |

## Phase 47 — Feeding & water plain language

Feature detail: [`phase_47_feeding_water_plain_language.plan.md`](phase_47_feeding_water_plain_language.plan.md). **OC-47** closed (WS7).

| Area | Status | Notes |
|------|--------|-------|
| Feeding plan view-model | ✅ | WS1 — `zoneFeedingPlan.js` |
| Zone Water primary | ✅ | WS2 — `ZoneWaterGrowStory`, Run now, advanced link |
| Inline plan editor + wizard | ✅ | WS3 — `ZoneFeedingPlanEditor`, `ZoneFeedingPlanWizard` |
| Feed & water hub | ✅ | WS4 — `/feeding`, `farmFeedingHub.js`, Dashboard links |
| Farmer vocabulary | ✅ | WS5 — `farmer-vocabulary.md`, `farmerVocabulary.js` Vitest gate |
| Guardian feeding | ✅ | WS6 — starters, `summarize_zone_fertigation` intents, patch matchers |
| OC-47 docs/tests | ✅ | WS7 — operator-tour §7b, architecture §7.0m, workflow §4c, Vitest |

**Master roadmap:** [`farmer_ux_roadmap_40_plus.plan.md`](farmer_ux_roadmap_40_plus.plan.md). Closure rows **OC-42 … OC-48** track each phase WS8/WS7 — not pre-40 work. Vocabulary: [`farmer-vocabulary.md`](../farmer-vocabulary.md). Guardian specs: [42](phase_42_guardian_pr_spec.md) · [43](phase_43_guardian_pr_spec.md) · [44](phase_44_guardian_pr_spec.md) · [45](phase_45_guardian_pr_spec.md) · [46](phase_46_guardian_llm_tool_proposals.plan.md).

| Phase | Focus (build order after 40–41) |
|-------|--------------------------------|
| [47](phase_47_feeding_water_plain_language.plan.md) | Feeding plan per room; zone Water primary |
| [42](phase_42_comfort_targets_automation_plain_language.plan.md) | Comfort bands; matchers + starters |
| [43](phase_43_operations_stock_feeding_finance.plan.md) | Supplies, feeding admin, money hubs |
| [44](phase_44_getting_started_edge_wizard.plan.md) | Farm + Pi wizards; setup starters second |
| [45](phase_45_farmer_validation_whole_app_polish.plan.md) | Sit-in + whole-app polish |
| [46](phase_46_guardian_llm_tool_proposals.plan.md) | LLM tool proposals (hybrid C) |
| [48](phase_48_dev_seed_and_small_farm_profiles.plan.md) | Dev seed hygiene; small farm profiles (parallel infra) |

## Phase 48 — Dev seed hygiene & small farm profiles

Feature detail: [`phase_48_dev_seed_and_small_farm_profiles.plan.md`](phase_48_dev_seed_and_small_farm_profiles.plan.md). **OC-48** closed (WS7). **Phase 48 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Profile spec (`small_indoor` / `demo_showcase`) | ✅ | WS1 — `docs/dev-farm-profiles.md`, `farms.meta_data` |
| Idempotent `master_seed` + unique sensors | ✅ | WS2 — migration `uq_sensors_farm_name_active`, `WHERE NOT EXISTS` |
| `dev-reset-farm.sh` | ✅ | WS3 — surgical reset without `--reset-volumes` |
| Bootstrap template alignment | ✅ | WS4 — `jadam_indoor` default for new farms; bootstrap idempotent |
| Timescale retention (dev-gated) | ✅ | WS5 — `apply-dev-retention.sh` + `TIMESCALE_RETENTION_DAYS` |
| `db-sanity-report` bloat metrics | ✅ | WS6 — sensors, profile, readings approx |
| OC-48 docs/smokes | ✅ | WS7 — local-operator-bootstrap, architecture §7.0n, `phase-48-closure.test.js`, `smoke_phase48_test.go` |

## Phase 49 — Sidebar nav polish

Feature detail: [`phase_49_sidebar_nav_polish.plan.md`](phase_49_sidebar_nav_polish.plan.md). **OC-49** closed (WS4). **Phase 49 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Fertigation rename + feeding disambiguation | ✅ | WS1 — Advanced **Fertigation**; Operations **Feeding admin** |
| `navRelations.js` map | ✅ | WS2 — zones ↔ feed & water ↔ targets; controls ↔ sensors |
| Related-route hover affordance | ✅ | WS3 — `SideNav.vue` wiggle + `prefers-reduced-motion` fallback |
| OC-49 docs/tests | ✅ | WS4 — `phase-49-closure.test.js`, `nav-relations.test.js`, operator-tour |

## Phase 50 — Hardware wiring visibility

Feature detail: [`phase_50_hardware_wiring_visibility.plan.md`](phase_50_hardware_wiring_visibility.plan.md). **OC-50** closed (WS6). **Phase 50 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Wiring model + demo backfill | ✅ | WS1 — `config.wiring` JSONB; migration by sensor/actuator name |
| PATCH wiring API + GET `wiring` | ✅ | WS2 — validated, conflict-checked |
| UI badges + sensor wiring editor | ✅ | WS3 — Sensors, Controls, sensor detail |
| Pi config generator | ✅ | WS4 — `GET /devices/{id}/pi-config`; device wizard download |
| Sanity report + inline conflicts | ✅ | WS5 — `db-sanity-report` exit on conflicts; wiring panel preview |
| OC-50 docs/tests | ✅ | WS6 — `pi-integration-guide` §2a, architecture §7.0o, `phase-50-closure.test.js` |
| **Extension (post-57)** | ✅ | Actuator wiring editor `ActuatorWiringPanel.vue` + `PATCH /actuators/{id}/assign` (HAT channel) / `/wiring` (GPIO). Docs only — closure test follow-up tracked below. |

## Phase 51 — Pi config platform sync

Feature detail: [`phase_51_pi_config_sync.plan.md`](phase_51_pi_config_sync.plan.md). **OC-51** closed (WS6). **Phase 51 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| API config by uid + version | ✅ | WS1 — `GET /devices/by-uid/{uid}/config`, `/config/version`; `config_version` + triggers |
| Pi bootstrap + cache | ✅ | WS2 — minimal YAML, `~/.gr33n/config-cache.json` |
| Live reload | ✅ | WS3 — `_poll_config_version`, `_reload_config` |
| Offline + staleness badge | ✅ | WS4 — cache fallback; `last_config_fetch_at`; `ActuatorCard` badge |
| Legacy opt-out + import | ✅ | WS5 — local YAML precedence; `import_config_to_platform.py` |
| OC-51 docs/tests | ✅ | WS6 — `pi-integration-guide` §2, architecture §7.0p, `phase-51-closure.test.js`, `smoke_phase51_test.go` |

## Phase 52 — Guardian UI context & Pi setup

Feature detail: [`phase_52_guardian_ui_context.plan.md`](phase_52_guardian_ui_context.plan.md). **OC-52** closed. **Phase 52 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Nav history + POST `nav_history` | ✅ | WS1 — `guardianPanel`, `context_ref` breadcrumb |
| Pi setup guide `/pi-setup` | ✅ | WS2 — Sequent HAT visual guide |
| Sidebar wiggles + navRelations | ✅ | WS3 — wiring, offline, config stale chains |
| Starter cleanup | ✅ | WS4 — no redundant "I'm on…" prefixes |
| **Extension (post-57)** | ✅ | `/pi-setup` live "Your farm channels" view — actual wired actuators/sensors link to detail pages (`data-test="pi-setup-live-wiring"`). Docs only — closure test follow-up tracked below. |

## Phase 53 — Grow + stock + money closure

Feature detail: [`phase_53_grow_stock_money_closure.plan.md`](phase_53_grow_stock_money_closure.plan.md). **OC-53** closed (WS6). **Phase 53 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Grow strip + wizards | ✅ | WS1 — `ZoneCurrentGrowStrip`, start/harvest/post-harvest, `growHub.js` |
| Supplies restock actions | ✅ | WS2 — `+ Add qty`, new batch, unit cost, refill task |
| Money tagging + spend chip | ✅ | WS3 — cycle tag, autolog split, `ZoneGrowCostPeek`, energy nudge |
| Cross-links + checklist | ✅ | WS4 — `v-nav-hint`, `firstRunChecklist` optional rows, operator guide |
| Guardian starters | ✅ | WS5 — grow strip, supplies restock-first, money by category, harvest yield |
| OC-53 docs/tests | ✅ | WS6 — operator-tour §7c + §6i, architecture §7.0q, `phase-53-closure.test.js` |

## Phase 54 — Zone connection nav

Feature detail: [`phase_54_zone_connection_nav.plan.md`](phase_54_zone_connection_nav.plan.md). **OC-54** closed (WS4). **Phase 54 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Interactive pipeline | ✅ | WS1 — `ZoneConnectionPipeline` on zone tabs + overview |
| Orphan link hints | ✅ | WS2 — tasks, actuators, connection cards, water history, automation edit |
| navRelations expansion | ✅ | WS3 — tasks/alerts↔zones, fertigation↔feeding |
| Guardian zone water hint | ✅ | WS4 — `zoneTabConnectionPipelineHint` in `context_ref.go` |
| OC-54 docs/tests | ✅ | WS4 — operator-tour §7d, architecture §7.0r, `phase-54-closure.test.js` |

## Phase 55 — Guardian ops read depth

Feature detail: [`phase_55_guardian_ops_grow_money.plan.md`](phase_55_guardian_ops_grow_money.plan.md). **OC-55** closed (WS5). **Phase 55 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Read tools | ✅ | WS1 — `summarize_cycle_cost`, `summarize_farm_spending`, `restock_priority`, `summarize_active_grows` |
| Hub starters | ✅ | WS2 — Supplies, Money, grow strip, post-harvest, dashboard |
| Ops persona | ✅ | WS3 — `context_ref.go`, `platform_context.go` |
| Guardian PR spec | ✅ | WS4 — `phase_55_guardian_pr_spec.md` |
| OC-55 docs/tests | ✅ | WS5 — architecture §7.0s, `readtools_ops_test.go`, `phase-55-closure.test.js` |

## Phase 56 — Grow schema + harvest analytics

Feature detail: [`phase_56_grow_schema_harvest_analytics.plan.md`](phase_56_grow_schema_harvest_analytics.plan.md). **OC-56** closed (WS5). **Phase 56 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| plant_id FK | ✅ | WS1 — migration, wizard, Plants page, list filter |
| Stage history | ✅ | WS2 — `crop_cycle_stage_events`, summary timeline |
| Compare polish | ✅ | WS3 — pre-selected `?ids=`, Guardian `compare_ids` |
| Income rollup | ✅ | WS4 — harvest economics banner, Money `?cycle_id=` |
| OC-56 docs/tests | ✅ | WS5 — architecture §7.0t, operator-tour §6k, `phase-56-closure.test.js`, crop-cycle smokes |

## Phase 64 — Crop knowledge base

Feature detail: [`phase_64_crop_knowledge_base.plan.md`](phase_64_crop_knowledge_base.plan.md). **OC-64** closed (WS6). **Phase 64 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Schema + seed | ✅ | WS1–WS2 — `crop_profiles`, `crop_profile_stages`, `plants.crop_profile_id`; 7 built-in crops |
| API | ✅ | List/get/clone/export/import; plants carry `crop_profile_id` |
| Guardian | ✅ | WS3 — `lookup_crop_targets`; persona crop-target rule |
| UI | ✅ | WS4 — Start grow picker, grow strip EC chip, `/crop-profiles/:id` |
| OC-64 docs/tests | ✅ | WS6 — architecture §7.0w, crop field guides, `phase-64-closure.test.js`, `smoke_phase64_test.go` |

## Phase 57 — Per-device Pi API keys

Feature detail: [`phase_57_pi_device_api_keys.plan.md`](phase_57_pi_device_api_keys.plan.md). **OC-57** closed (WS5). **Phase 57 shipped.**

| Area | Status | Notes |
|------|--------|-------|
| Schema | ✅ | WS1 — `device_api_keys`, bcrypt hash, last_used |
| Platform UI | ✅ | WS2 — wizard + Controls card issue/rotate/revoke |
| Pi agent | ✅ | WS3 — `GR33N_DEVICE_API_KEY`, `/etc/gr33n/device.key`, legacy fallback |
| Edge auth | ✅ | WS4 — `X-Device-Key`, scoped routes, rate limit, audit |
| OC-57 docs/tests | ✅ | WS5 — architecture §7.0u, operator-tour §6l, `phase-57-closure.test.js`, `smoke_phase57_test.go` |

## Phases 58–59 — Farmer closure arc (planned)

Hub: [`phase_53_59_roadmap.plan.md`](phase_53_59_roadmap.plan.md). Close **OC-58 … OC-59** when each phase WS docs/tests ship.

| Phase | OC | Plan | Close when |
|-------|-----|------|------------|
| ~~53 Grow + stock + money~~ | ~~OC-53~~ | [phase_53](phase_53_grow_stock_money_closure.plan.md) | ✅ Shipped |
| ~~54 Zone connection nav~~ | ~~OC-54~~ | [phase_54](phase_54_zone_connection_nav.plan.md) | ✅ Shipped |
| ~~55 Guardian ops~~ | ~~OC-55~~ | [phase_55](phase_55_guardian_ops_grow_money.plan.md) | ✅ Shipped |
| ~~56 Grow schema~~ | ~~OC-56~~ | [phase_56](phase_56_grow_schema_harvest_analytics.plan.md) | ✅ Shipped |
| ~~57 Device API keys~~ | ~~OC-57~~ | [phase_57](phase_57_pi_device_api_keys.plan.md) | ✅ Shipped |
| ~~58 Task consumptions~~ | ~~OC-58~~ | [phase_58](phase_58_task_consumptions_runtime.plan.md) | ✅ Shipped |
| 59 Enterprise boundary | OC-59 | [phase_59](phase_59_enterprise_tier_boundary.plan.md) | `enterprise-tier-boundary.md` |
| ~~60 Morning walkthrough~~ | ~~OC-60~~ | [phase_60](phase_60_guardian_morning_walkthrough.plan.md) | ✅ Shipped |
| ~~61 Proactive nudges~~ | ~~OC-61~~ | [phase_61](phase_61_guardian_proactive_nudges.plan.md) | ✅ Shipped |
| ~~62 Grow advisor~~ | ~~OC-62~~ | [phase_62](phase_62_guardian_grow_advisor.plan.md) | ✅ Shipped |
| ~~63 Session memory~~ | ~~OC-63~~ | [phase_63](phase_63_guardian_session_memory.plan.md) | ✅ Shipped |
| ~~64 Crop knowledge base~~ | ~~OC-64~~ | [phase_64](phase_64_crop_knowledge_base.plan.md) | ✅ Shipped |
| ~~65 Pi & hardware diagnostics~~ | ~~OC-65~~ | [phase_65](phase_65_guardian_pi_diagnostics.plan.md) | ✅ Shipped |
| 66 Weather & site | OC-66 | [phase_66](phase_66_weather_site_context.plan.md) | Offline solar + ingest tiers |
| 67 Field assistant | OC-67 | [phase_67](phase_67_guardian_field_assistant.plan.md) | Voice + grounded photo diagnosis |

**Note:** Phase 51 "Phase 52+ per-device API keys" → **[Phase 57](phase_57_pi_device_api_keys.plan.md)** (not Phase 52). **Phase 64 must precede Phase 62** (grow advisor reads real targets from the crop knowledge base).

### Deferred closure tests (shipped post-57 extensions)

These extensions shipped with code + docs; closure tests are a small follow-up (not blocking):

| Extension | Home phase | Suggested test |
|-----------|-----------|----------------|
| Actuator wiring editor (`ActuatorWiringPanel.vue`, `/actuators/{id}/assign`) | [Phase 50](phase_50_hardware_wiring_visibility.plan.md) | Vitest: HAT-channel vs GPIO mode toggle, save calls correct endpoint; Go smoke on `PATCH /assign` |
| `/pi-setup` live farm channels | [Phase 52](phase_52_guardian_ui_context.plan.md) | Vitest: wired actuator renders a link; empty state when no wiring |

---

**Master roadmap:** [`farmer_ux_roadmap_40_plus.plan.md`](farmer_ux_roadmap_40_plus.plan.md) · [`phase_53_59_roadmap.plan.md`](phase_53_59_roadmap.plan.md). Closure rows **OC-42 … OC-67** track each phase WS8/WS7 — not pre-40 work. Vocabulary: [`farmer-vocabulary.md`](../farmer-vocabulary.md). Guardian specs: [42](phase_42_guardian_pr_spec.md) · [43](phase_43_guardian_pr_spec.md) · [44](phase_44_guardian_pr_spec.md) · [45](phase_45_guardian_pr_spec.md) · [46](phase_46_guardian_llm_tool_proposals.plan.md) · [55](phase_55_guardian_ops_grow_money.plan.md).

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Feature scope — **shipped** (WS1–WS8) |
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | **Shipped** — WS6 + OC-36C done |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | WS8 = OC-37 + OC-37E sweep |
| [phase_38_plant_needs_ui_and_pulse_commands.plan.md](phase_38_plant_needs_ui_and_pulse_commands.plan.md) | **Shipped** — UI + pulse |
| [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) | **Next** — queue + automated mix |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Reference closure pattern (WS7 OpenAPI + WS8 RAG) |
| [phase_48_dev_seed_and_small_farm_profiles.plan.md](phase_48_dev_seed_and_small_farm_profiles.plan.md) | Dev seed hygiene — parallel to 43–46 |
