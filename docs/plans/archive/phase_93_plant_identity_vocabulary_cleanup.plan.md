---
name: Phase 93 — Plant identity vocabulary & API cleanup
overview: >
  Kill dual identity systems: server-only display_name, rename strain_or_variety to
  batch_label, routes tab=plants, remove operator-facing "strain" from API surface.
todos:
  - id: ws1-display-name
    content: "WS1: Remove display_name from plant create UI — read-only catalog label in list"
    status: completed
  - id: ws2-batch-label
    content: "WS2: Migration crop_cycles.strain_or_variety → batch_label (+ OpenAPI alias period)"
    status: completed
  - id: ws3-routes
    content: "WS3: tab=strains → tab=plants; router redirects; growHub compare routes"
    status: completed
  - id: ws4-api-copy
    content: "WS4: OpenAPI descriptions — batch_label semantics; deprecate strain_or_variety"
    status: completed
  - id: ws5-guardian
    content: "WS5: Guardian tools/read tools use batch_label; never treat as crop_key"
    status: completed
  - id: ws6-smokes
    content: "WS6: Vitest closure + smoke — no strain in operator JSON responses"
    status: completed
isProject: false
---

# Phase 93 — Plant identity vocabulary & API cleanup

## Status

**Shipped.** Closes **blind spots #1** and **#6**.

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md) (`plants.crop_key` shipped).

**Closure:** [`phase-93-closure.md`](phase-93-closure.md) · **OC-93**

---

## Blind spots addressed

| # | Problem | Fix |
|---|---------|-----|
| **1** | Operators type “Flower Room Romas” into `display_name` → new crop identity | **Remove label field from create**; list shows `catalog.display_name` + optional `variety_or_cultivar` |
| **6** | `strain_or_variety` on cycles + `tab=strains` — two identity systems | Rename to **`batch_label`**; routes **`tab=plants`** |

---

## WS1 — No typed crop identity in UI

| Before | After |
|--------|-------|
| “Your label for this plant *” required | **Plant type** from dropdown only (required) |
| `display_name` user input | Server sets from catalog; UI shows catalog name |
| Optional room note | **`variety_or_cultivar`** only (genetics) |

List row: **Tomato** · Cherokee Purple (variety) — not “Veg Room Romas” as primary title.

---

## WS2 — Cycle batch label migration

```sql
ALTER TABLE gr33ncrops.crop_cycles
  RENAME COLUMN strain_or_variety TO batch_label;
```

OpenAPI: add `batch_label`; keep `strain_or_variety` as **deprecated** read/write alias for one release.

Guardian grow setup pack JSON: accept both keys during alias period.

---

## WS3 — Routes & workspace tab

| Old | New |
|-----|-----|
| `/zones?tab=strains` | `/zones?tab=plants` (301 redirect old) |
| `WORKSPACES.zones.tabs[].id === 'strains'` | `id: 'plants'` (redirect handler) |
| `buildCompareRoute` query `tab: 'strains'` | `tab: 'plants'` |

---

## Acceptance

- [ ] Cannot POST plant with client-supplied `display_name` as identity
- [ ] Cycle create uses `batch_label` in OpenAPI primary field
- [ ] Zero “strain” in operator-visible API field names (after alias period)
- [ ] Guardian cites `batch_label` for batch context, `crop_key` for targets

**Prompt loop:** **`phase 93`**.
