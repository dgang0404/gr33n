# Phase 95 — closure (OC-95)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_95_catalog_integrator_ops.plan.md`](phase_95_catalog_integrator_ops.plan.md)

**Depends on:** [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) (DB catalog shipped).

**Closes:** Blind spot **#4** — catalog growth path stalls without a repeatable integrator cadence.

---

## The one job (done)

> **Document and automate** the YAML → seed SQL → parity → migrate → `catalog_version` → RAG path so platform team and site integrators can add crops (flowers, cacti, San Pedro, …) without freezing the picker.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Integrator playbook | [`catalog-integrator-playbook.md`](../catalog-integrator-playbook.md) |
| **WS2** | Make / script targets | `add-crop-check`, `check-catalog-seed-drift`, `check-catalog-release` |
| **WS3** | `catalog_version` contract | Playbook § contract; YAML header + seed header |
| **WS4** | CI drift gate | `.github/workflows/ci.yml` + `check-catalog-seed-drift.sh` |
| **WS5** | Post-migrate smoke | `smoke_phase95_test.go`, `TestCatalogSeedMatchesCanonicalFile` |

---

## Integrator commands

| Command | DB | Purpose |
|---------|-----|---------|
| `make add-crop-check` | No | YAML validate + seed drift |
| `make check-catalog-seed-drift` | No | Regen vs `db/seed/crop_catalog_from_yaml.sql` |
| `make check-catalog-release` | Yes | Pre + post migrate full gate |

**PR template:** [`docs/templates/add-crop-pr-checklist.md`](../templates/add-crop-pr-checklist.md) (San Pedro walkthrough).

---

## Automated tests

| Test | Path |
|------|------|
| Canonical seed drift | `internal/croplibrary/catalog_seed_test.go` |
| Picker version + commons crop | `cmd/api/smoke_phase95_test.go` |

---

## OC-95

Phase 95 is **closed** when the playbook is linked, CI fails on YAML/seed drift, and smokes prove a catalog crop appears in picker + commons after migrate.
