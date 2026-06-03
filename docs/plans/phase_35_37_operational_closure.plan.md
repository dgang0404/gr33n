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
  - id: oc-36-closure
    content: "OC-36: Phase 36 WS8 — greenhouse operator-tour, OpenAPI, bootstrap greenhouse_climate_v1 → core types, smokes (defer until WS1–WS7 land)"
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
| **Demo seed** | ⚠️ partial | `master_seed.sql` Section 3B wraps 18/6 in `lighting_programs` |
| **Bootstrap** | ✅ | `jadam_indoor_photoperiod_v1` → `lighting_programs` (OC-35A migration) |
| **Unit tests** | ✅ | `handler_test.go`, `worker_test.go` TZ case |
| **Smokes / Vitest** | ✅ | `smoke_phase35_lighting_test.go`; `photoperiod-clock-editor.test.js` |
| **OpenAPI / operator-tour** | ✅ | `LightingProgram` schemas; operator-tour §5; architecture §7.0b |

### OC-35 tasks (close before calling Phase 35 shipped)

**OC-35A — Bootstrap (Phase 35 WS6 remainder)**

- Update `gr33ncore._bootstrap_jadam_indoor_photoperiod_v1` to create one `lighting_program` + generated ON/OFF schedules + actions (same transactional pattern as handler).
- Keep idempotent `NOT EXISTS` guards; legacy farms with orphan schedules may coexist (document in operator-tour one-liner).

**OC-35B — Docs + OpenAPI (Phase 35 WS8)**

- `docs/operator-tour.md` — “Set up 18/6 lights” using preset + PhotoperiodClockEditor (`/lighting`).
- `openapi.yaml` — `LightingProgram`, preset list, `/farms/{id}/lighting-programs`, `/lighting-programs/from-preset`, schedule-action paths.
- `docs/farm-guardian-architecture.md` — lighting in grow environment stack; cite `summarize_zone_lighting`.

**OC-35C — Tests (Phase 35 WS8)**

- `cmd/api/smoke_phase35_lighting_test.go` — create from preset → list → deactivate; optional TZ assertion via worker unit test (already covered) or schedule metadata check.
- `ui/src/__tests__/photoperiod-clock-editor.test.js` — duration 18h updates end; preset chip sets 12/12.

**Suggested timing:** OC-35A can land immediately after Phase 35 code PR merges; OC-35B/C in the same PR or follow-up before Phase 36 starts (greenhouse docs will reference supplemental vs blocking light).

---

## Phase 36 — Greenhouse climate

Closure is already **WS8** in [`phase_36_greenhouse_climate.plan.md`](phase_36_greenhouse_climate.plan.md). Map to this doc as **OC-36**:

| When | Work |
|------|------|
| After WS1–WS4 (core + UI) | OpenAPI for zone climate profile + typed actuators |
| WS5 (bootstrap → core) | **`greenhouse_climate_v1` bootstrap** uses core types — same pattern as OC-35A |
| WS8 | operator-tour greenhouse section, architecture “block sun ≠ add light” cross-link to Phase 35, smokes for shade rule + manual command |

**Do not** fold Phase 36 bootstrap into Phase 35 closure — different domain, same checklist.

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

1. Confirm **OC-35A–C** closed (lighting bootstrap + docs + smokes).
2. Re-run platform doc RAG ingest so new operator-tour sections are searchable (Phase 32 WS8 script).
3. README / phase-14 roadmap row: Phases 35–37 shipped with closure notes.

---

## Recommended schedule across phases

```
Phase 35 code PR  ──► OC-35A bootstrap (same sprint or +1)
                   ──► OC-35B + OC-35C docs/tests (before Phase 36 UI references lighting)

Phase 36 WS1–7   ──► OC-36 inline with WS5 + WS8

Phase 37 WS1–7   ──► OC-37 inline with WS8
                   ──► OC-37E final sweep (all deferred 35 items verified)
```

**Rule:** Feature WS8 stays in each phase plan; **this doc** is the cross-phase backlog so deferred items are not lost when a phase plan todo is marked `done` too early.

---

## Current git snapshot (Phase 35 — illustrative)

Uncommitted work typically includes:

- **Untracked:** migration, `lighting_programs.sql`, sqlc output, `internal/handler/lighting/`, Guardian tool, UI components.
- **Modified, not staged:** routes, schema, `master_seed.sql`, worker, automation handler, SideNav, router, plan status.

Closure PR should `git add` the full set above plus OC-35B/C artifacts when ready.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Feature scope — **shipped** (WS1–WS8) |
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | WS8 = OC-36 |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | WS8 = OC-37 + OC-37E sweep |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Reference closure pattern (WS7 OpenAPI + WS8 RAG) |
