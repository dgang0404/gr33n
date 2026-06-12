---
name: Phase 101 — Guardian write tools crop_key
overview: >
  create_plant, apply_grow_setup_pack, and grow PR proposals require crop_key from
  catalog — same rules as UI Phase 85; no Guardian-created typo plant rows.
todos:
  - id: ws1-create-plant
    content: "WS1: execCreatePlant requires crop_key; server display_name; reject unsupported"
    status: pending
  - id: ws2-grow-pack
    content: "WS2: apply_grow_setup_pack uses crop_key + plant_id chain"
    status: pending
  - id: ws3-proposals
    content: "WS3: Proposal schema + impact summary — crop_key not display_name"
    status: pending
  - id: ws4-openapi
    content: "WS4: OpenAPI Guardian tool docs mirror POST /plants contract"
    status: pending
  - id: ws5-smokes
    content: "WS5: grow_create_test + smoke — duplicate crop_key upsert; ramps block"
    status: pending
isProject: false
---

# Phase 101 — Guardian write tools crop_key

## Status

**Planned.** Guardian must not bypass catalog rules the UI enforces.

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md).

**Closure:** **OC-101**

---

## The one job

> **`create_plant` write tool** uses **`crop_key`** from the knowledge base — never free-text `display_name` as crop identity.

---

## Gap today

```10:31:internal/farmguardian/tools/plants.go
	displayName, err := stringFromArgs(args, "display_name")
	// ...
	row, err := deps.Q.CreatePlant(ctx, db.CreatePlantParams{
		FarmID:            deps.FarmID,
		DisplayName:       displayName,
```

Guardian can flood the DB with typed names while UI moves to catalog-only.

---

## Target contract

| Tool arg | Rule |
|----------|------|
| `crop_key` | **Required**; catalog supported |
| `display_name` | Ignored or rejected — server sets |
| `variety_or_cultivar` | Optional genetics |

Same upsert as `POST /farms/{id}/plants`.

---

## Acceptance

- [ ] Grow setup pack proposal cannot create plant without `crop_key`
- [ ] Persona lists crop_key in write tool examples
- [ ] Smokes mirror Phase 85 plant API

**Prompt loop:** **`phase 101`**.
