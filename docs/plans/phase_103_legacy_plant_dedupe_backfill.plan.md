---
name: Phase 103 — Legacy plant dedupe & backfill
overview: >
  Migration tooling to merge typo plant rows into catalog slots, backfill crop_key,
  and relink cycles before/after Phase 85 on existing demo and production farms.
todos:
  - id: ws1-audit
    content: "WS1: SQL audit report — duplicate display_names, missing crop_profile_id"
    status: pending
  - id: ws2-merge
    content: "WS2: scripts/merge-legacy-plants.sh — fuzzy match to catalog aliases"
    status: pending
  - id: ws3-backfill
    content: "WS3: Phase 85 backfill + manual review queue for ambiguous rows"
    status: pending
  - id: ws4-cycles
    content: "WS4: Relink crop_cycles.plant_id after merge; preserve batch_label"
    status: pending
  - id: ws5-smoke
    content: "WS5: master_seed + demo farm — zero orphan plants after migrate"
    status: pending
isProject: false
---

# Phase 103 — Legacy plant dedupe & backfill

## Status

**Planned.** Real farms (including **gr33n Demo Farm**) may already have typo plant rows.

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

- [ ] Demo farm post-migrate: ≤1 plant row per crop_key
- [ ] No active cycle loses plant link after merge
- [ ] Document in crop-knowledge-operator-runbook upgrade section

**Prompt loop:** **`phase 103`** (run with **85 WS1** on existing deployments).
