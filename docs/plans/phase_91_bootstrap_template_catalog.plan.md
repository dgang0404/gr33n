---
name: Phase 91 — Bootstrap template catalog
overview: >
  List farm bootstrap templates from DB/API with labels and summaries; remove
  constants/bootstrapTemplates.js as source of truth.
todos:
  - id: ws1-schema
    content: "WS1: bootstrap_templates table or commons_catalog_entries kind=bootstrap"
    status: pending
  - id: ws2-api
    content: "WS2: GET /platform/bootstrap-templates — key, label, summary bullets, module hints"
    status: pending
  - id: ws3-ui
    content: "WS3: Settings, farm setup wizard, moduleEmptyShell — fetch list"
    status: pending
  - id: ws4-enterprise
    content: "WS4: Link agronomy seed pack + jadam templates in catalog; site-manifest refs"
    status: pending
  - id: ws5-guardian
    content: "WS5: apply_grow_setup_pack / bootstrap proposals use catalog keys only"
    status: pending
isProject: false
---

# Phase 91 — Bootstrap template catalog

## Status

**Planned.** New farms and **Settings → Apply template** should not depend on a static JS file.

**Closure:** **OC-91**

---

## The one job

> **Bootstrap starter packs** (JADAM indoor, chicken coop, greenhouse, …) are listed from **Postgres** with accurate labels and “what’s included” summaries — keys still match `apply_farm_bootstrap_template`.

---

## Gap today

`ui/src/constants/bootstrapTemplates.js`:

- `BOOTSTRAP_TEMPLATE_KEYS`, `BOOTSTRAP_STARTER_OPTIONS`
- Long `*_SUMMARY` bullet lists per template (5+ templates)
- Duplicated in `farmSetupWizard.js`, `Settings.vue`, `moduleEmptyShell.js`

New template in DB/SQL → requires UI edit in **3+ files**; summaries can drift from actual SQL template.

---

## Target

| Approach | Notes |
|----------|--------|
| **A — commons catalog** | `kind=bootstrap_template` rows (like recipe/agronomy packs) |
| **B — dedicated table** | `gr33ncore.bootstrap_templates(template_key, label, summary_md, …)` |

Either way: **`GET /platform/bootstrap-templates`** returns list for pickers.

Apply still: `POST /farms/{id}/bootstrap` with `template_key` (existing).

---

## UI migration

| Surface | Change |
|---------|--------|
| Farm create / org default | Fetch starter options |
| Settings apply template | Fetch + show summary from API |
| Help / empty states | Template descriptions from API |

Keep `bootstrapTemplates.js` as **deprecated fallback** until API guaranteed on all deployments.

---

## Guardian (WS5)

`apply_grow_setup_pack` and bootstrap write tools must only propose keys present in catalog API response.

---

## Acceptance

- [ ] Add template row in migration → appears in UI without JS edit
- [ ] Summary bullets match operator-visible zones/programs after apply
- [ ] Phase 83 agronomy pack cross-linked where relevant

**Prompt loop:** **`phase 91`**.
