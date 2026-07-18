# Phase 102 — closure (OC-102)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_102_fertigation_program_catalog_metadata.plan.md`](phase_102_fertigation_program_catalog_metadata.plan.md)

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md) grow context; [Phase 96](phase_96_grow_feeding_program_validation.plan.md) attach-time warnings consume these tags.

**Related:** [Phase 108](phase_108_commons_recipe_packs_crop_key.plan.md) commons import preserves the same metadata on promotion.

**Closes:** Data-driven feeding chain — `crop_key` → profile EC → fertigation program → application recipe.

---

## The one job (done)

> **Programs and recipes carry catalog tags** (`recommended_crop_keys`, `recommended_stages`, optional `profile_ec_source` / `ec_band_mscm`). Filters, Start grow, Phase 96 validation, and Guardian cite metadata — not name heuristics.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Grow context has `crop_key` + stage | Phase 86 chain |
| **WS1** | Program `metadata` schema + PATCH | `internal/fertigation/programmeta/`, `PATCH /fertigation/programs/{id}/metadata` |
| **WS2** | Recipe crop/stage tags | `target_crop_categories`, `target_growth_stages` on `application_recipes` |
| **WS3** | EC band on seeded programs | `20260624_phase102_program_catalog_metadata.sql` |
| **WS4** | Demo seed — veg JLF + flower FFJ | same migration |
| **WS5** | Filter programs by `crop_key` + `stage` | `GET /farms/{id}/fertigation/programs?crop_key=&stage=` |
| **WS6** | Start grow + Water tab fit UX | `StartGrowWizard.vue`, `ZoneWaterGrowStory.vue`, `programFit.js` |
| **WS7** | Phase 96 reads metadata | `programfit.ValidateProgramForGrow` → `programmeta.CheckFit` |
| **WS8** | Commons import preserves tags | `import-recipe-pack.sh`, `20260627_phase108_*.sql` |
| **WS9** | Guardian cites mismatch | `ProgramFitHintLine` in `context_ref.go`, `readtools.go` |

---

## Metadata contract

**Program** (`gr33nfertigation.programs.metadata`):

```json
{
  "recommended_crop_keys": ["cannabis"],
  "recommended_stages": ["early_flower", "mid_flower"],
  "profile_ec_source": { "crop_key": "cannabis", "stage": "early_flower" },
  "ec_band_mscm": { "min": 1.6, "max": 2.4 }
}
```

**Recipe** (`application_recipes` columns): `target_crop_categories[]`, `target_growth_stages[]`.

Programs without tags do not filter or warn (unknown fit).

---

## Operator behavior

| Surface | Behavior |
|---------|----------|
| **Start grow** | Programs sorted by fit; ⚠ suffix on mismatched options |
| **Water tab** | Amber banner when primary program tags ≠ grow crop/stage |
| **Program list API** | `?crop_key=cannabis&stage=early_flower` returns only matching tagged programs |
| **Commons import** | Pack `recommended_*` applied via metadata PATCH |

---

## Automated tests

| Test | Path |
|------|------|
| Metadata seed + filter + attach warnings | `cmd/api/smoke_phase102_test.go` |
| Phase 96 metadata path | `cmd/api/smoke_phase96_test.go` |
| Commons pack tags | `cmd/api/smoke_phase108_test.go` |
| UI fit helpers | `ui/src/__tests__/program-fit.test.js` |
| Programmeta unit tests | `internal/fertigation/programmeta/meta_test.go` |

---

## OC-102

Phase 102 is **closed** when demo programs/recipes are tagged, API filters work, Phase 96 smokes use metadata, and Start grow / Guardian surface program crop fit.
