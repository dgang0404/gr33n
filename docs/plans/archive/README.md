# Archived phase plans

Closed phase plans live here so [`../`](../) and [`phase-14-operator-documentation.md`](../../phase-14-operator-documentation.md) stay scannable. **Nothing is deleted** — Phase 206 moved ~198 shipped plans here in one scripted pass (replacing Phase 157's stub-at-old-path approach).

## Index

| Arc | Plans | Closure |
|-----|-------|---------|
| **88–92** UI static data → DB/API | [`phase_88_92_platform_data_gaps_roadmap.plan.md`](phase_88_92_platform_data_gaps_roadmap.plan.md) · [88](phase_88_domain_enums_api.plan.md) · [89](phase_89_lighting_presets_api_wiring.plan.md) · [90](phase_90_device_taxonomy_registry.plan.md) · [91](phase_91_bootstrap_template_catalog.plan.md) · [92](phase_92_zone_greenhouse_vocabulary.plan.md) | [`phase-84-110-closure.md`](../phase-84-110-closure.md) |

**Everything else:** use [`phase-14-operator-documentation.md`](../../phase-14-operator-documentation.md) or grep `docs/plans/archive/phase_N_*.plan.md` — the full set is flat in this directory (no nested folders).

## What stays at `docs/plans/` (not archived)

Era hub docs and active janitorial plans only:

- `product_backlog_operator_runtime.plan.md`
- `pre_development_gaps_index.plan.md`
- `phase_53_59_roadmap.plan.md`, `phase_68_73_spa_workspace_roadmap.plan.md`, `phase_84_100_master_roadmap.plan.md`, `phase_173_177_today_excellence_roadmap.plan.md`, `farmer_ux_roadmap_40_plus.plan.md`
- Recent meta plans (205, 206) until they age out

## Rules

- Move only when **close-when** boxes are checked and closure tests / rollups exist.
- Do **not** archive active arcs or hub navigation docs.
- Regenerate [`current-state.md`](../../current-state.md) after major phase ships; use `make docs-current-state-hint` for counts.
