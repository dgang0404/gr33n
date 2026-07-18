# Phase 91 — closure (OC-91)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_91_bootstrap_template_catalog.plan.md`](phase_91_bootstrap_template_catalog.plan.md)

**Depends on:** Existing `POST /farms/{id}/bootstrap-template` / `apply_farm_bootstrap_template` SQL.

**Cross-link:** Phase 83 agronomy seed pack — `related_commons_slug` on catalog rows (e.g. `gr33n-cultivator-seed-pack-v1`).

---

## The one job (done)

> **Bootstrap starter packs** (JADAM indoor, chicken coop, greenhouse, …) are listed from **Postgres** with accurate labels and “what’s included” summaries — keys still match `apply_farm_bootstrap_template`.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `gr33ncore.bootstrap_templates` table + seed | `db/migrations/20260623_phase91_bootstrap_template_catalog.sql` |
| **WS2** | `GET /platform/bootstrap-templates` | `internal/platform/bootstraptemplates/catalog.go` |
| **WS3** | Settings + farm setup wizard fetch list | `Settings.vue`, `FarmSetupWizard.vue`, `bootstrapCatalog.js` |
| **WS4** | Agronomy / commons cross-links in seed | `related_commons_slug`, `playbook_section` columns |
| **WS5** | Guardian validates catalog keys only | `internal/farmguardian/tools/bootstrap.go` · `bootstraptemplates.Current().IsValid` |

---

## API shape

```
GET /platform/bootstrap-templates
```

Returns `{ templates: [{ template_key, label, short_label, tagline, summary_title, summary_bullets, module_hints, recommended, wizard_primary, related_commons_slug, … }] }`.

Apply unchanged: `POST /farms/{id}/bootstrap-template` with `template_key`.

UI caches via `loadBootstrapCatalog`; `constants/bootstrapTemplates.js` is **deprecated** — re-exports fallback for backward compatibility.

---

## Operator impact

| Before | After |
|--------|-------|
| Starter options in static JS (`bootstrapTemplates.js`) | Fetched from Postgres catalog |
| Summaries duplicated in 3+ Vue/JS files | `summary_bullets` from API |
| New SQL template required UI edits | Migration row → picker updates on deploy |

**Surfaces:** Settings create-farm picker, org default bootstrap, farm setup wizard primary choices.

---

## Guardian alignment

`apply_bootstrap_template` rejects keys not in the catalog with a message pointing operators to `GET /platform/bootstrap-templates`.

---

## Automated tests

| Test | Path |
|------|------|
| API contract (≥5 templates, JADAM bullets) | `cmd/api/smoke_phase91_test.go` |
| Loader cache + fallback keys | `ui/src/__tests__/bootstrap-catalog.test.js` |

---

## OC-91

Phase 91 is **closed** when smokes pass and farm bootstrap pickers load from **`GET /platform/bootstrap-templates`** (or bundled fallback).
