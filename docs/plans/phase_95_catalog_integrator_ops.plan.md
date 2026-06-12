---
name: Phase 95 — Catalog integrator ops cadence
overview: >
  Integrator playbook and Make targets for adding crops (YAML → seed SQL → parity →
  migrate → catalog_version → RAG re-ingest) — flowers, cacti, San Pedro, etc.
todos:
  - id: ws1-playbook
    content: "WS1: docs/catalog-integrator-playbook.md — full cadence + roles"
    status: pending
  - id: ws2-make
    content: "WS2: make add-crop-check / check-catalog-release checklist target"
    status: pending
  - id: ws3-version
    content: "WS3: catalog_version bump contract in seed SQL + commons manifest"
    status: pending
  - id: ws4-ci
    content: "WS4: CI gate — parity fails if YAML changed without seed regen"
    status: pending
  - id: ws5-smoke
    content: "WS5: smoke new crop appears in picker + commons API after migrate"
    status: pending
isProject: false
---

# Phase 95 — Catalog integrator ops cadence

## Status

**Planned.** Closes **blind spot #4** (catalog growth path stalls without process).

**Depends on:** [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) (shipped).

**Closure:** **OC-95**

---

## Blind spot #4

Adding crops is **migration-only** — correct for quality — but without a **repeatable cadence** the dropdown looks frozen while YAML grows in git.

---

## Integrator checklist (WS1)

1. Edit `data/crop_library.yaml` (+ field guide MD if supported crop)
2. `./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/YYYYMMDD_catalog_*.sql`
3. `make check-crop-catalog-parity`
4. Bump `catalog_version` in YAML + seed header
5. `make migrate`
6. `make check-crop-catalog-db`
7. `make rag-ingest-field-guides` (if guide body changed)
8. Restart API pods (`CROP_CATALOG_SOURCE=db`)
9. Smoke: picker count + `GET /commons/crop-catalog/{crop_key}`
10. Update agronomy seed pack manifest if enterprise bundle includes new crop

**Roles:** platform team owns YAML; integrator runs migrate on each site; operators never type new crops.

---

## Acceptance

- [ ] Playbook linked from phase-14, enterprise README, crop cutover runbook
- [ ] CI fails on YAML/seed drift
- [ ] Example PR template for “add San Pedro cactus” walkthrough

**Prompt loop:** **`phase 95`**.
