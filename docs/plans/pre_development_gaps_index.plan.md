---
name: Pre-development gaps index (39–41 + backlog)
overview: >
  Master tracker for UX/runtime gaps identified after Phases 35–38 and before starting
  Phase 39 development. Links every gap to a canonical plan or backlog section, priority,
  dependencies, and documentation chunk status. Use this file to resume multi-session doc work.
todos:
  - id: doc-chunk-1
    content: "Chunk 1 — Index + Phase 41 + Phase 39b + cross-links (README, phase-14, 39, 40, sit-in, closure)"
    status: done
  - id: doc-chunk-2
    content: "Chunk 2 — Product backlog plan (run now, metadata.steps, Guardian lighting propose, mobile pointer)"
    status: done
  - id: doc-chunk-3
    content: "Chunk 3 — Phase 40 plan sync (bug-guardian-nav done, Related → 41/39b/index; operator-tour §4b stub)"
    status: done
  - id: doc-chunk-4
    content: "Chunk 4 — workflow-guide post-39/40 notes; platform-doc-manifest rows for §4b/§3; RAG manifest checklist"
    status: done
isProject: false
---

# Pre-development gaps index

**Purpose:** One place to see what is **shipped**, **in flight (39–40)**, and **still gap** before writing feature code. Prevents rediscovering the same disconnects (zone vs farm-wide pages, empty states, RO-only farms) mid-sprint.

**Canonical development order:** [Phase 39](phase_39_edge_fertigation_execution.plan.md) → [Phase 40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) → [Phase 41](phase_41_farm_hub_coherence.plan.md) (after 40 or partial parallel) → [Phase 39b](phase_39b_plain_irrigation.plan.md) (after 39 WS1 queue).

---

## Documentation chunks (multi-prompt)

| Chunk | Scope | Status |
|-------|--------|--------|
| **1** | This index, Phase 41 plan, Phase 39b plan, README + phase-14 + 39/40/sit-in/closure cross-links | ✅ |
| **2** | [`product_backlog_operator_runtime.plan.md`](product_backlog_operator_runtime.plan.md) — run now, steps deprecation, Guardian lighting propose, mobile | ✅ |
| **3** | Phase 40 DoD + operator-tour §4b stub + architecture §7.0f/§7.0g stubs | ✅ |
| **4** | workflow-guide §4b; operator-tour §3b + §3/§4b manifest notes; RAG checklist below | ✅ |

**Documentation status:** Chunks 1–4 complete. Safe to start **Phase 39 WS1** implementation.

---

## RAG re-ingest checklist (after each phase ships)

Run on each farm with embedding configured: **`make rag-ingest-platform-docs`** (and field guides if procedures changed).

| Phase shipped | Update these docs first | Manifest reminder |
|---------------|-------------------------|-------------------|
| **39** | operator-tour §3, workflow §4b, pi-integration-guide, architecture §7.0d | [`platform-doc-manifest.yaml`](../rag/platform-doc-manifest.yaml) Phase 39 comment block |
| **40** | operator-tour §4b (full walk), workflow §4b, architecture §7.0f | Phase 40 comment block |
| **41** | operator-tour §3b, §4, workflow §4b, architecture §7.0g, tasks-first guide | Phase 41 comment block |
| **39b** | workflow §4b, operator-tour §4a | Phase 39b comment block |

Whole-file ingest: `operator-tour.md` and `workflow-guide.md` are already in `include:` — section edits still require a full re-ingest pass.

---

## Stack map (what fixes what)

| Operator pain | Primary owner | Also helps |
|---------------|---------------|------------|
| Commands overwrite each other | **39 WS1** queue | 35/36/38 writers migrate to enqueue |
| Mix is manual-only on Pi | **39 WS2–WS5** | 40 WS5 grow story |
| Zone says “go to Setpoints” | **40 WS2–WS4** | 41 WS2 why-empty on cards |
| Fertigation/events feel separate from zone | **39 WS7** + **40 WS5** | **41 WS1–WS3** farm hub |
| Dashboard doesn’t answer “morning?” | **41 WS1** | 40 WS1 Today strip (zone scope) |
| Empty lists, no explanation | **41 WS4** (why-empty) | sit-in §1 |
| RO/well feed without nutrients | **39b** | 39 queue + pulse |
| Guardian nav clutter | **40 bug-guardian-nav** | ✅ shipped |
| Ad-hoc program run | **backlog** run-now | 39 WS5 pipeline |
| Store / Capacitor polish | **backlog** mobile | `docs/mobile-distribution.md` |

---

## Tier A — Must plan before “polished site” (have plans)

| ID | Gap | Plan | Blocker for |
|----|-----|------|-------------|
| **A1** | Device command queue + automated mix | [phase_39](phase_39_edge_fertigation_execution.plan.md) | Honest Water tab, safe automation |
| **A2** | Zone cockpit (inline targets, alerts, today) | [phase_40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Zone ≠ DB admin |
| **A3** | Farm-wide hub (Dashboard, Fertigation context, deep links) | [phase_41](phase_41_farm_hub_coherence.plan.md) | Whole-app polish |
| **A4** | Plain-water / irrigation-only programs | [phase_39b](phase_39b_plain_irrigation.plan.md) | Non-mix farms |

---

## Tier B — Documented in Chunk 2 (product backlog plan)

| ID | Gap | Notes |
|----|-----|-------|
| **B1** | Program **run now** API + UI | Unscheduled ad-hoc program fire; README lists as open |
| **B2** | Deprecate `programs.metadata.steps` | After zero fallback warnings; harden `action_source` |
| **B3** | Guardian **`create_lighting_program`** propose tool | Optional; read `summarize_zone_lighting` exists |
| **B4** | **Mobile distribution** polish | Capacitor / store checklist — see `docs/mobile-distribution.md` |

---

## Tier C — Hygiene / closure (no new domain)

| ID | Gap | Canonical link / action | Doc status |
|----|-----|-------------------------|------------|
| **C1** | README Guardian nav hotfix wording | [README.md](../../README.md) current-focus line | ✅ Chunk 1 |
| **C2** | Project Roadmap through Phase 38 only | [README.md](../../README.md) roadmap + in-flight list | ✅ Chunk 1 |
| **C3** | Closure doc stale Phase 36 snapshot | [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | ✅ Chunk 1 |
| **C4** | Demo **tasks** missing `zone_id` | [phase_40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) WS6 · [phase_41](phase_41_farm_hub_coherence.plan.md) WS5 | ⏳ implement |
| **C5** | RAG re-ingest after operator-doc edits | [OC-37E](phase_35_37_operational_closure.plan.md) · checklist above | ⏳ per phase ship |

---

## Tier A–C plan link audit (2026-06)

Every Tier **A** and **B** gap has a dedicated plan file. Tier **C** items are doc hygiene or implementation notes inside phase plans.

| Tier | ID | Plan or doc |
|------|-----|-------------|
| A | A1 | [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) |
| A | A2 | [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) |
| A | A3 | [phase_41_farm_hub_coherence.plan.md](phase_41_farm_hub_coherence.plan.md) |
| A | A4 | [phase_39b_plain_irrigation.plan.md](phase_39b_plain_irrigation.plan.md) |
| B | B1 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b1--program-run-now) |
| B | B2 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b2--deprecate-programsmetadatasteps) |
| B | B3 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b3--guardian-create_lighting_program-propose) |
| B | B4 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b4--mobile-distribution-polish) · [mobile-distribution.md](../mobile-distribution.md) |
| C | C1–C3 | README + [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) |
| C | C4 | phase_40 WS6 + phase_41 WS5 |
| C | C5 | RAG checklist (this doc) + OC-37E |

---

## Tier D — Explicitly out of scope (v1 grow stack)

- Replacing farm-wide Advanced CRUD pages (40 out of scope; 41 links in, does not merge schemas)
- Closed-loop EC dosing with inline sensor (39 v2 note)
- CO₂ / weather API / Modbus peristaltic vendors
- LM Studio insert-sharing scaffolds (README roadmap item, separate product line)

---

## Phase 40 WS6 prerequisite check

**Tasks in zone (40 WS6)** needs `gr33ncore.tasks.zone_id` populated in demo/seed.

| Check | Where |
|-------|--------|
| Column exists | `db/queries/tasks.sql` |
| Seed assigns zone | Verify `master_seed.sql` task rows — **41 WS5** documents seed gap if empty |

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | OC-39, OC-40, future OC-41 |
| [phase-14-operator-documentation.md](../phase-14-operator-documentation.md) | Operator index |
| [sit-in-operator-experience.md](../workstreams/sit-in-operator-experience.md) | Why-empty → 41 WS4 |
| [operator-tour.md](../operator-tour.md) | §4 conceptual; §4b after 40 |
