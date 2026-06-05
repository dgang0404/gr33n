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
    status: pending
  - id: oc-43-closure
    content: "OC-43: Phase 43 WS7 — operations hub docs/tests (close when Phase 43 ships)"
    status: pending
  - id: oc-44-closure
    content: "OC-44: Phase 44 WS6 — setup + edge wizard docs/tests (close when Phase 44 ships)"
    status: pending
  - id: oc-45-closure
    content: "OC-45: Phase 45 WS7 — farmer sit-in + farmer-ready v1 (close when Phase 45 ships)"
    status: pending
  - id: oc-46-closure
    content: "OC-46: Phase 46 WS6 — LLM tool proposals docs/tests (close when Phase 46 ships)"
    status: pending
  - id: oc-47
    content: "OC-47: Phase 47 WS7 — feeding & water plain language docs/tests (close when Phase 47 ships)"
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

**Master roadmap:** [`farmer_ux_roadmap_40_plus.plan.md`](farmer_ux_roadmap_40_plus.plan.md). Closure rows **OC-42 … OC-47** track each phase WS8/WS7 — not pre-40 work. Vocabulary: [`farmer-vocabulary.md`](../farmer-vocabulary.md). Guardian specs: [42](phase_42_guardian_pr_spec.md) · [43](phase_43_guardian_pr_spec.md) · [44](phase_44_guardian_pr_spec.md) · [45](phase_45_guardian_pr_spec.md) · [46](phase_46_guardian_llm_tool_proposals.plan.md).

| Phase | Focus (build order after 40–41) |
|-------|--------------------------------|
| [47](phase_47_feeding_water_plain_language.plan.md) | Feeding plan per room; zone Water primary |
| [42](phase_42_comfort_targets_automation_plain_language.plan.md) | Comfort bands; matchers + starters |
| [43](phase_43_operations_stock_feeding_finance.plan.md) | Supplies, feeding admin, money hubs |
| [44](phase_44_getting_started_edge_wizard.plan.md) | Farm + Pi wizards; setup starters second |
| [45](phase_45_farmer_validation_whole_app_polish.plan.md) | Sit-in + whole-app polish |
| [46](phase_46_guardian_llm_tool_proposals.plan.md) | LLM tool proposals (hybrid C) |

## Phases 42–47 — Farmer UX + Guardian PR completion
|------|------------|-------|
| `device_commands` queue | WS1 | **Fixes last-write-wins** for all actuators + mix |
| Mix calculator + `mix_batch` | WS2–WS3 | Recipe + base EC + target |
| Pi executor + program pipeline | WS4–WS5 | After queue |
| Schema migration | WS1 | First grow-stack migration since 38 (additive) |

**Stack rule:** 35/36/38 keep working during 39; migrate writers to queue with `pending_command` head mirror for one Pi release.

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
