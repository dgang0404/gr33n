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
    status: pending
  - id: oc-37-closure
    content: "OC-37: Phase 37 WS8 — offline field walkthrough, procedure OpenAPI, field_guide corpus ingest smoke, safety-stop smokes (defer until WS1–WS7 land)"
    status: pending
  - id: oc-37-final-sweep
    content: "OC-37E: End-of-37 sweep — verify OC-35A–C closed; platform-doc RAG manifest includes new operator-tour sections; README roadmap"
    status: pending
isProject: false
---

# Phase 35–37 operational closure (seed, bootstrap, docs, tests)

## Why this doc exists

Feature phases often land **code first** (schema, API, UI, worker) while **seed data**, **bootstrap templates**, **operator-tour**, **OpenAPI**, and **integration smokes** trail behind. That leaves uncommitted or “invisible” files in git and a false sense of completion.

This plan is the **rollup tracker** for closure work across Phases **35 → 37**. Each feature phase keeps its own **WS8** (or WS5 bootstrap + WS8 docs in Phase 36). This doc says **what is done, what is deferred, and when to close it**.

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
| Guardian read | ✅ | `summarize_zone_lighting` (no `create_lighting_program` propose tool yet) |
| **Demo seed** | ✅ | `master_seed.sql` Section 3B wraps 18/6 in `lighting_programs` |
| **Bootstrap** | ✅ | `jadam_indoor_photoperiod_v1` → `lighting_programs` (OC-35A migration) |
| **Unit tests** | ✅ | `handler_test.go`, `worker_test.go` TZ case |
| **Smokes / Vitest** | ✅ | `smoke_phase35_lighting_test.go`; `photoperiod-clock-editor.test.js` |
| **OpenAPI / operator-tour** | ✅ | `LightingProgram` schemas; operator-tour §5; architecture §7.0b |

## Phase 35 — status

**Shipped.** OC-35A–C closed; WS1–WS8 complete. Optional follow-up: RAG re-ingest of operator-tour §5 (part of OC-37E sweep).

---

## Current git snapshot

Phase 35 shipped: `06e281d`, `362c0ac`, `9a19048`, `e09d4f4`. Phase 36 backend: `999bff1` (WS1+WS2), `0916aba` (WS3+WS5), `f686d76` (WS7), `46ecdbb` (plan status). **Open:** WS4 UI, WS6 interlocks, OC-36B/C.

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
| **Greenhouse UI (WS4)** | ⏳ | ZoneDetail tab, typed command buttons, sensor strip |
| **Sensor interlocks (WS6)** | ⏳ | Missing lux/PAR banner; template guard without override |
| **Demo seed** | ⏳ | Bootstrap apply suffices for new farms; `master_seed.sql` greenhouse row optional |
| **Unit tests** | ✅ partial | `greenhouse_test.go`, `taxonomy_test.go` |
| **Smokes / Vitest** | ⏳ | **OC-36C** |
| **OpenAPI / operator-tour** | ✅ | **OC-36B** — operator-tour §5b; OpenAPI paths/schemas |
| **Architecture** | ✅ | §7.0c in `farm-guardian-architecture.md` + operator-tour cross-links |

### Phase 36 — status

**In progress.** WS1–WS3, WS5, WS7 + **OC-36A** + **OC-36B** closed. **OC-36C** (smokes) remains with **WS4** (Greenhouse UI tab) and **WS6** (missing-sensor UX).

Apply migration `20260603_phase36_greenhouse_climate_v2.sql` before re-running `greenhouse_climate_v1` bootstrap on existing dev DBs.

**Do not** fold Phase 36 into Phase 35 closure — different domain, same checklist pattern.

---

## Phase 37 — Guardian offline field assistant

Closure is **WS8** in [`phase_37_guardian_offline_field_assistant.plan.md`](phase_37_guardian_offline_field_assistant.plan.md). Map to **OC-37**:

| When | Work |
|------|------|
| After WS2 (field corpus) | Ingest `field_guide` sources; extend platform-doc manifest |
| After WS3–WS4 (procedures + safety) | OpenAPI procedure endpoints; smokes for step flow + safety hard-stop |
| WS8 | operator-tour “first install with Guardian offline”; link Pi wiring procedure to Phase 35 actuator path |

**OC-37E — End-of-37 sweep**

Before marking the 35–37 arc complete:

1. Confirm **OC-35A–C** and **OC-36B–C** closed (lighting + greenhouse operator docs/smokes).
2. Re-run platform doc RAG ingest so new operator-tour sections are searchable (Phase 32 WS8 script).
3. README / phase-14 roadmap row: Phases 35–37 shipped with closure notes (Phase 36 marks shipped when WS4 + WS6 + OC-36B/C land).

---

## Recommended schedule across phases

```
Phase 35 code PR  ──► OC-35A bootstrap (same sprint or +1)
                   ──► OC-35B + OC-35C docs/tests (before Phase 36 UI references lighting)

Phase 36 WS1–3,5,7 ──► OC-36A + OC-36B ✅
                   ──► WS4 + WS6 + OC-36C (remaining ship)

Phase 37 WS1–7   ──► OC-37 inline with WS8
                   ──► OC-37E final sweep (OC-35A–C + OC-36B–C verified; RAG ingest)
```

**Rule:** Feature WS8 stays in each phase plan; **this doc** is the cross-phase backlog so deferred items are not lost when a phase plan todo is marked `done` too early.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Feature scope — **shipped** (WS1–WS8) |
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | **In progress** — WS8 = OC-36B + OC-36C; OC-36A done |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | WS8 = OC-37 + OC-37E sweep |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Reference closure pattern (WS7 OpenAPI + WS8 RAG) |
