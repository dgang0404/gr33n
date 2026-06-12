---
name: Phases 84–100 — Master roadmap (plants, platform data, blind spots)
overview: >
  Locked execution order for the full gr33n knowledge-base and platform-data arc.
  Every blind spot from the 84–87 planning review has an explicit phase — no random
  follow-up tasks during testing.
todos:
  - id: arc-84-87
    content: "Arc A (84–87): Plants, grows, Guardian crop chain — DO FIRST"
    status: pending
  - id: arc-88-92
    content: "Arc B (88–92): UI static → DB/API enums, devices, bootstrap"
    status: pending
  - id: arc-93-100
    content: "Arc C (93–100): Blind spots — identity, genetics, ops, CI, offline"
    status: pending
isProject: false
---

# Phases 84–100 — Master roadmap

## Big-dawg rule

> **Every gap gets a phase.** If testing finds a hole, add **Phase 101+** — don't ad-hoc fix without updating plans.

---

## Locked execution order (blind spot #12)

```
Arc A — PLANTS FIRST (highest operator pain)
  84 ✅ shipped
  85 → 86 → 87 → 93
       │              │
       │         (identity cleanup right after 85)

Arc B — PLATFORM DATA (after 87 or parallel 89 only)
  89 (quick win) → 88 → 90 → 91 → 92
  99 (CI guards — start early, finish after 88)

Arc C — BLIND SPOTS & ENTERPRISE
  95 (integrator cadence — can doc early)
  96 (feeding validation — after 86)
  97 (RAG governance — with 87)
  94 (genetics EC — after 87 docs v1)
  98 (enterprise promotion — anytime after 83/84)
  100 (offline cache — after 85 + 88)
```

**Do NOT** run 88/89 before 85 unless you explicitly want enum fixes while plants still flood the DB.

---

## Blind spot → phase map (complete)

| # | Blind spot | Phase | Notes |
|---|------------|-------|-------|
| 1 | Identity vs label fuzzy in UI | **85** WS2/WS3 + **93** | Server `display_name`; remove label field |
| 2 | Guardian alias vs picker diverge | **86** WS5 + **87** WS4 | Cycle → plant → crop_key required |
| 3 | Farm override vs per-genetics EC | **87** runbook v1 + **94** | Document before build genetics |
| 4 | Catalog growth cadence | **95** | Integrator playbook + CI |
| 5 | Picker 404 fallback hides broken deploy | **85** WS6 | Upgrade banner vs offline cache |
| 6 | `strain_or_variety` / `tab=strains` | **93** | `batch_label`, `tab=plants` |
| 7 | Feeding program ↔ stage mismatch | **96** | Warn UI + Guardian |
| 8 | RAG vs structured targets | **97** | Persona + re-ingest triggers |
| 9 | Multi-farm / commons promotion | **98** | Promote vs local matrix |
| 10 | CI enum drift (SetpointRow bug) | **99** + **88** | `check-ui-domain-parity` |
| 11 | Mobile / offline picker | **100** | IndexedDB cache |
| 12 | Execution order risk | **This doc** | 85 before 88 |

---

## Arc summaries

### [84–87 Plants & crop knowledge](phase_84_87_crop_identity_roadmap.plan.md)

Catalog in Postgres → catalog-bound plants → grow/Guardian chain → operator closure.

### [88–92 UI static → DB/API](phase_88_92_platform_data_gaps_roadmap.plan.md)

Domain enums, lighting presets API, device taxonomy, bootstrap catalog, zone vocabulary.

### 93–100 Blind spots & hardening

| Phase | Plan |
|-------|------|
| **93** | [Plant identity vocabulary cleanup](phase_93_plant_identity_vocabulary_cleanup.plan.md) |
| **94** | [Genetics & batch EC profiles](phase_94_genetics_batch_ec_profiles.plan.md) |
| **95** | [Catalog integrator ops](phase_95_catalog_integrator_ops.plan.md) |
| **96** | [Grow feeding program validation](phase_96_grow_feeding_program_validation.plan.md) |
| **97** | [RAG vs structured truth](phase_97_rag_structured_truth_governance.plan.md) |
| **98** | [Enterprise catalog promotion](phase_98_enterprise_catalog_promotion.plan.md) |
| **99** | [CI domain parity guards](phase_99_ci_domain_parity_guards.plan.md) |
| **100** | [Offline catalog cache](phase_100_offline_catalog_cache.plan.md) |

---

## Prompt loop

`phase 85 ws1` … `phase 100` — one phase per chat session or WS per prompt.

**Index:** [phase-14-operator-documentation.md](../phase-14-operator-documentation.md)

---

## Adding Phase 101+

Template for new gaps found in testing:

1. Add row to blind spot table above
2. Create `docs/plans/phase_NNN_<slug>.plan.md` with frontmatter todos
3. Link from this master roadmap
4. One line in phase-14 index
