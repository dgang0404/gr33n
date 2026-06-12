# Phase 103 — closure (OC-103)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_103_legacy_plant_dedupe_backfill.plan.md`](phase_103_legacy_plant_dedupe_backfill.plan.md)

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md) catalog-bound plants schema.

**Closes:** Pre–Phase 85 typo plant rows merge into one `crop_key` slot per farm; cycles keep `plant_id` links.

---

## The one job (done)

> **Existing farms survive Phase 85** — Tomato / tomato / Romas merge to one `crop_key=tomato` slot; active and historical cycles relink to the canonical plant row.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | SQL audit report | `scripts/sql/legacy_plants_audit.sql` |
| **WS2** | Merge script | `scripts/merge-legacy-plants.sh`, `make merge-legacy-plants` |
| **WS3** | Backfill + ambiguous queue | `gr33ncrops.merge_legacy_plants()` — profile → catalog → alias match |
| **WS4** | Relink `crop_cycles.plant_id` | same function; preserves `batch_label` |
| **WS5** | Demo farm zero duplicates | `smoke_phase103_test.go` |

---

## Operator workflow

After `make migrate` on an existing deployment:

```bash
./scripts/merge-legacy-plants.sh              # audit only
./scripts/merge-legacy-plants.sh --apply --audit  # merge + re-audit
```

Documented in [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) upgrade section.

Rows still missing `crop_key` after merge need a manual catalog pick in **Zone → Plants**.

---

## Merge rules (implemented)

| Signal | Action |
|--------|--------|
| `crop_profile_id` → builtin profile | Set `crop_key` from profile |
| Display name matches catalog alias | Map via `crop_catalog_aliases` |
| Multiple rows same `crop_key` | Keep oldest id; soft-delete duplicates; relink cycles |
| No match | Left for operator review (audit SQL lists them) |

---

## Automated tests

| Test | Path |
|------|------|
| Typo merge + cycle relink | `cmd/api/smoke_phase103_test.go` — `TestPhase103_LegacyPlantMerge` |
| No duplicate crop_key on demo DB | same — `TestPhase103_AuditNoDuplicateCropKeyAfterMigrate` |

---

## OC-103

Phase 103 is **closed** when `merge_legacy_plants()` is idempotent, smokes pass, and the operator runbook documents the upgrade path.
