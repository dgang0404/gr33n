# Phase 108 — closure (OC-108)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_108_commons_recipe_packs_crop_key.plan.md`](phase_108_commons_recipe_packs_crop_key.plan.md)

**Depends on:** [Phase 102](phase_102_fertigation_program_catalog_metadata.plan.md) program metadata; [Phase 98](phase_98_enterprise_catalog_promotion.plan.md) enterprise promotion model.

**Closes:** Commons recipe pack import carries **`recommended_crop_keys`** / **`recommended_stages`** so Phase 96/102 validation works on promoted sites without hand-tagging.

---

## The one job (done)

> **Import recipe pack from commons** → fertigation programs get Phase 102 metadata at import; unknown `crop_key` values fail validation against the catalog.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Phase 102 recipe/program meta schema | `programmeta`, Phase 102 migration |
| **WS1** | Pack JSON manifest fields | `sample-recipe-pack-v7.body.json`, commons playbook |
| **WS2** | Import script validates + writes meta | `scripts/enterprise/import-recipe-pack.sh` |
| **WS3** | Demo pack tagged in DB | `20260627_phase108_commons_recipe_pack_crop_tags.sql` |
| **WS4** | Commons UI crop fit badges | `CommonsCatalog.vue` — `recommended_crop_keys` |
| **WS5** | Smokes | `cmd/api/smoke_phase108_test.go` |

---

## Import contract

`import-recipe-pack.sh`:

1. Records `POST /farms/{id}/commons/catalog-imports` (audit)
2. Creates programs idempotently by `name` (`is_active: false`)
3. Applies `recommended_crop_keys`, `recommended_stages`, `profile_ec_source`, `ec_band_mscm` via metadata PATCH
4. **Rejects** unknown `crop_key` / invalid `profile_ec_source.crop_key`

Imported programs participate in `GET /farms/{id}/fertigation/programs?crop_key=&stage=` (Phase 102 filter).

---

## Automated tests

| Test | Path |
|------|------|
| Pack tags + program filter | `cmd/api/smoke_phase108_test.go` |

Documented in [`commons-catalog-operator-playbook.md`](../commons-catalog-operator-playbook.md).

---

## OC-108

Phase 108 is **closed** when import fails on invalid crop_key, promoted programs appear in Phase 102 suggest/filter API, and commons playbook documents the manifest fields.
