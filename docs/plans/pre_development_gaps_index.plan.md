---
name: Pre-development gaps index (39–41 + backlog)
overview: >
  ARCHIVED historical index — UX/runtime gaps identified after Phases 35–38 and before
  Phase 39. All doc chunks (1–9) and Tier A/B gaps through Phase 67 are closed. Do not
  use this as an active tracker. For new work see phase_68_73_spa_workspace_roadmap.plan.md (68–77).
  Closure rollup through OC-67 is frozen in phase_35_37_operational_closure.plan.md.
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
  - id: doc-chunk-5-guardian-pr
    content: "Chunk 5 — guardian-change-requests-guide.md + guardian_pr_ux_through_farmer_phases.plan.md (PR triggers, starters, industry patterns)"
    status: done
  - id: doc-chunk-9-doc-hygiene
    content: "Chunk 9 — Archive index; point closure doc + phase-14 at 68–73 arc and plan lifecycle rules"
    status: done
isProject: false
---

# Pre-development gaps index

> ## ⛔ ARCHIVED — historical gap index (pre–Phase 40 context)
>
> **All Tier A/B gaps and doc chunks 1–9 are closed.** Phases 39–67 shipped. Do not treat
> this file as an active tracker or block new work on stale rows (e.g. C4 seed tasks, C5
> per-phase RAG gates).
>
> **Active planning:** [phase_68_73_spa_workspace_roadmap.plan.md](phase_68_73_spa_workspace_roadmap.plan.md)
> (SPA workspace refactor **68–77**). **Plan lifecycle rules** (shipped = conditions deprecated) live
> there. **Closure rollup (OC-35 … OC-67 only):**
> [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) — frozen.

**Purpose (historical):** One place to see what was **shipped**, **in flight (39–40)**, and **still gap** before the farmer UX arc. Kept for audit trail only.

**Canonical development order (historical → current):** [39](phase_39_edge_fertigation_execution.plan.md) ✅ → [39b](phase_39b_plain_irrigation.plan.md) ✅ → **[40 → 67 farmer + Guardian arcs](farmer_ux_roadmap_40_plus.plan.md)** ✅ → **[68 → 73 SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md)** (planned). Tier **B** backlog ✅ on `main`.

| Step | Phase | Focus |
|------|-------|--------|
| 1 | [40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Zone cockpit ✅ |
| 2 | [41](phase_41_farm_hub_coherence.plan.md) | Farm hub + why-empty ✅ |
| 3 | [47](phase_47_feeding_water_plain_language.plan.md) | Feeding & water plain language (room-first) ✅ |
| 4 | [42](phase_42_comfort_targets_automation_plain_language.plan.md) | Comfort bands; schedules/rules plain language ✅ |
| 5 | [43](phase_43_operations_stock_feeding_finance.plan.md) | Supplies, feeding admin, money ✅ |
| 6 | [44](phase_44_getting_started_edge_wizard.plan.md) | Farm + Pi wizards ✅ |
| 7 | [45](phase_45_farmer_validation_whole_app_polish.plan.md) | Sit-in + whole-app polish ✅ |
| 8 | [46](phase_46_guardian_llm_tool_proposals.plan.md) | Guardian LLM tool proposals ✅ |
| — | [48](phase_48_dev_seed_and_small_farm_profiles.plan.md) | Dev seed hygiene ✅ |
| 9+ | [53→59 arc](phase_53_59_roadmap.plan.md) | Grow/stock/money, nav, Guardian ops, schema, Pi keys ✅ |
| 10+ | [60→67 arc](phase_53_59_roadmap.plan.md) | Guardian intelligence + knowledge + field assistant ✅ |
| **Next** | [68→77 arc](phase_68_73_spa_workspace_roadmap.plan.md) | SPA workspace refactor (sidebar → workspaces; ops, comfort, Today, polish) |

**Post-52 shipped:** [49](phase_49_sidebar_nav_polish.plan.md) · [50](phase_50_hardware_wiring_visibility.plan.md) · [51](phase_51_pi_config_sync.plan.md) · [52](phase_52_guardian_ui_context.plan.md) · … through **67** ✅

---

## Documentation chunks (multi-prompt)

| Chunk | Scope | Status |
|-------|--------|--------|
| **1** | This index, Phase 41 plan, Phase 39b plan, README + phase-14 + 39/40/sit-in/closure cross-links | ✅ |
| **2** | [`product_backlog_operator_runtime.plan.md`](product_backlog_operator_runtime.plan.md) — run now, steps deprecation, Guardian lighting propose, mobile | ✅ |
| **3** | Phase 40 DoD + operator-tour §4b stub + architecture §7.0f/§7.0g stubs | ✅ |
| **4** | workflow-guide §4b; operator-tour §3b + §3/§4b manifest notes; RAG checklist below | ✅ |
| **5** | [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) + [guardian_pr_ux_through_farmer_phases.plan.md](guardian_pr_ux_through_farmer_phases.plan.md) | ✅ |
| **6** | Guardian PR specs [42](phase_42_guardian_pr_spec.md)–[46](phase_46_guardian_llm_tool_proposals.plan.md) + [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md) + operator-tour §5c–§9, §6e–§6h, arch §7.0h–§7.0l | ✅ |
| **7** | [phase_47](phase_47_feeding_water_plain_language.plan.md) + [farmer-vocabulary.md](../farmer-vocabulary.md) + tour §7b + arch §7.0m | ✅ |
| **8** | [phase_48](phase_48_dev_seed_and_small_farm_profiles.plan.md) + [local-operator-bootstrap.md](../local-operator-bootstrap.md) slow-dev section | ✅ |
| **9** | Archive index + [phase-14](../phase-14-operator-documentation.md) → 68–73 arc; closure doc frozen at OC-67 | ✅ |

**Documentation status:** Chunks 1–9 complete. Phases **39–67** and **product backlog** shipped on `main`. **Next:** [68–77 SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md).

---

## RAG re-ingest checklist (after each phase ships)

Run on each farm with embedding configured: **`make rag-ingest-platform-docs`** (and field guides if procedures changed).

| Phase shipped | Update these docs first | Manifest reminder |
|---------------|-------------------------|-------------------|
| **39** | operator-tour §3, workflow §4b, pi-integration-guide, architecture §7.0d | [`platform-doc-manifest.yaml`](../rag/platform-doc-manifest.yaml) Phase 39 comment block |
| **40** | operator-tour §4b (full walk), workflow §4b, architecture §7.0f | Phase 40 comment block |
| **41** | operator-tour §3b, §4, workflow §4b, architecture §7.0g, tasks-first guide | Phase 41 comment block |
| **39b** | workflow §4b, operator-tour §4a | Phase 39b comment block |
| **43** | operator-tour §7 + §6f, workflow §4b, architecture §7.0i, farmer-vocabulary | Phase 43 comment block |

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
| Ad-hoc program run | ✅ **B1** run-now | `POST …/programs/{rid}/run-now` |
| Store / Capacitor polish | **backlog** mobile | `docs/mobile-distribution.md` |

---

## Tier A — Farmer UX arc (plans 40–47)

**Master map:** [`farmer_ux_roadmap_40_plus.plan.md`](farmer_ux_roadmap_40_plus.plan.md)

| ID | Gap | Plan | Blocker for |
|----|-----|------|-------------|
| **A1** | Device command queue + automated mix | [phase_39](phase_39_edge_fertigation_execution.plan.md) ✅ | Honest Water tab |
| **A2** | Zone cockpit | [phase_40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) ✅ | Daily grow in the room |
| **A3** | Farm-wide hub | [phase_41](phase_41_farm_hub_coherence.plan.md) ✅ | Morning + `?zone_id=` |
| **A4** | Plain-water programs | [phase_39b](phase_39b_plain_irrigation.plan.md) ✅ | RO/well farms |
| **A5** | Feeding & water — room-first, no fertigation console | [phase_47](phase_47_feeding_water_plain_language.plan.md) ✅ | "How does this room get water?" |
| **A5b** | Setpoints / rules / schedules understandable | [phase_42](phase_42_comfort_targets_automation_plain_language.plan.md) ✅ | Not a DB console |
| **A6** | Inventory / fertigation admin / costs coherent | [phase_43](phase_43_operations_stock_feeding_finance.plan.md) ✅ | Stock & money jobs |
| **A7** | New farm + Pi setup in-app | [phase_44](phase_44_getting_started_edge_wizard.plan.md) | Onboarding without shell docs |
| **A8** | Non-technical validation | [phase_45](phase_45_farmer_validation_whole_app_polish.plan.md) | Farmer-ready v1 |
| **A9** | Guardian LLM tool proposals (PR from free-form ask) | [phase_46](phase_46_guardian_llm_tool_proposals.plan.md) | After 45; not starter chips |

---

## Tier B — Documented in Chunk 2 (product backlog plan)

| ID | Gap | Notes |
|----|-----|-------|
| **B1** | Program **run now** API + UI | ✅ Implemented (0.4.6) |
| **B2** | Deprecate `programs.metadata.steps` | After zero fallback warnings; harden `action_source` |
| **B3** | Guardian **`create_lighting_program`** propose tool | Optional; read `summarize_zone_lighting` exists |
| **B4** | **Mobile distribution** polish | Capacitor / store checklist — see `docs/mobile-distribution.md` |

---

## Tier C — Hygiene / closure (historical — all closed or superseded)

| ID | Gap | Canonical link / action | Doc status |
|----|-----|-------------------------|------------|
| **C1** | README Guardian nav hotfix wording | [README.md](../../README.md) current-focus line | ✅ Chunk 1 |
| **C2** | Project Roadmap through Phase 38 only | [README.md](../../README.md) roadmap + in-flight list | ✅ Chunk 1 (superseded by farmer_ux_roadmap) |
| **C3** | Closure doc stale snapshots | [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | ✅ **Archived** at OC-67 (Chunk 9) — do not extend |
| **C4** | Demo **tasks** missing `zone_id` | [phase_40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) WS6 · [phase_41](phase_41_farm_hub_coherence.plan.md) WS5 | ✅ shipped / historical |
| **C5** | RAG re-ingest after operator-doc edits | Checklist below + OC-37E (historical) | ✅ ongoing ops habit, not a gate |
| **C6** | Dev DB bloat / duplicate sensors from re-seed | [phase_48](phase_48_dev_seed_and_small_farm_profiles.plan.md) · [local-operator-bootstrap.md](../local-operator-bootstrap.md#slow-ui-and-dev-db-hygiene) | ✅ `dev-reset-farm.sh`, idempotent seed |

---

## Tier A–C plan link audit (2026-06)

Every Tier **A** and **B** gap has a dedicated plan file. Tier **C** items are doc hygiene or implementation notes inside phase plans.

| Tier | ID | Plan or doc |
|------|-----|-------------|
| A | A1 | [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) |
| A | A2 | [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) |
| A | A3 | [phase_41_farm_hub_coherence.plan.md](phase_41_farm_hub_coherence.plan.md) |
| A | A4 | [phase_39b_plain_irrigation.plan.md](phase_39b_plain_irrigation.plan.md) |
| A | A5–A9 | [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) → phases 40–47 |
| A | A10 | [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md) → phases 53–59 (grow/stock/money closure arc) |
| B | B1 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b1--program-run-now) |
| B | B2 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b2--deprecate-programsmetadatasteps) |
| B | B3 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b3--guardian-create_lighting_program-propose) |
| B | B4 | [product_backlog_operator_runtime.plan.md](product_backlog_operator_runtime.plan.md#b4--mobile-distribution-polish) · [mobile-distribution.md](../mobile-distribution.md) |
| C | C1–C3 | README + archived [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) |
| C | C4 | phase_40 WS6 + phase_41 WS5 (historical) |
| C | C5 | RAG checklist (this doc) — ops habit |

---

## Tier D — Explicitly out of scope (v1 grow stack)

- Replacing farm-wide Advanced CRUD pages (40 out of scope; 41 links in, does not merge schemas)
- Closed-loop EC dosing with inline sensor (39 v2 note)
- CO₂ / weather API / Modbus peristaltic vendors
- LM Studio insert-sharing scaffolds (README roadmap item, separate product line)
- **Enterprise tier** (POs, METRC, multi-entity GL) — see [enterprise-tier-boundary.md](../enterprise-tier-boundary.md) (Phase 59 ✅)

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
| [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | **Archived** OC-35 … OC-67 rollup — do not extend |
| [phase_68_73_spa_workspace_roadmap.plan.md](phase_68_73_spa_workspace_roadmap.plan.md) | **Active** — SPA workspace arc + plan lifecycle rules |
| [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md) | Farmer + Guardian arcs 53–67 (shipped) |
| [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) | Master farmer UX map 40–73 |
| [phase-14-operator-documentation.md](../phase-14-operator-documentation.md) | Operator doc index |
| [sit-in-operator-experience.md](../workstreams/sit-in-operator-experience.md) | Why-empty → 41 WS4 |
| [operator-tour.md](../operator-tour.md) | §4 conceptual; §4b after 40 |
