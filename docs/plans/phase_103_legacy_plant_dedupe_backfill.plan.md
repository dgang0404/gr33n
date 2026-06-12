---
name: Phase 103 — Legacy plant dedupe & backfill
overview: >
  Migration tooling to merge typo plant rows into catalog slots, backfill crop_key,
  and relink cycles before/after Phase 85 on existing demo and production farms.
todos:
  - id: ws1-audit
    content: "WS1: SQL audit report — duplicate display_names, missing crop_profile_id"
    status: completed
  - id: ws2-merge
    content: "WS2: scripts/merge-legacy-plants.sh — fuzzy match to catalog aliases"
    status: completed
  - id: ws3-backfill
    content: "WS3: Phase 85 backfill + manual review queue for ambiguous rows"
    status: completed
  - id: ws4-cycles
    content: "WS4: Relink crop_cycles.plant_id after merge; preserve batch_label"
    status: completed
  - id: ws5-smoke
    content: "WS5: master_seed + demo farm — zero orphan plants after migrate"
    status: completed
isProject: false
---

# Phase 103 — Legacy plant dedupe & backfill

## Status

**Shipped** on `main`. Closure: [`phase-103-closure.md`](phase-103-closure.md) (**OC-103**).

Real farms (including **gr33n Demo Farm**) may already have typo plant rows — run merge after migrate.

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md) schema.

**Closure:** **OC-103**

---

## The one job

> **Existing farms survive Phase 85** — tomato/Tomato/Romas merge to one `crop_key=tomato` slot; cycles keep history.

---

## Merge rules

| Signal | Action |
|--------|--------|
| `crop_profile_id` → builtin profile | Set `crop_key` from profile |
| Display name matches catalog alias | Map via `crop_catalog_aliases` |
| Ambiguous (two builtins possible) | Operator review CSV export |
| No match | Flag for manual catalog pick in UI |

---

## Acceptance

- [x] Demo farm post-migrate: ≤1 plant row per crop_key
- [x] No active cycle loses plant link after merge
- [x] Document in crop-knowledge-operator-runbook upgrade section

**Prompt loop:** **`phase 103`** (run with **85 WS1** on existing deployments).
