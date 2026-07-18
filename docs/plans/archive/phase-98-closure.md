# Phase 98 — closure (OC-98)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_98_enterprise_catalog_promotion.plan.md`](phase_98_enterprise_catalog_promotion.plan.md)

**Depends on:** [Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md), [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md).

**Closes:** Blind spot **#9** — multi-farm / commons promotion confusion.

---

## The one job (done)

> **Document what promotes platform-wide** (catalog migration) **vs stays farm-local** (EC overrides, plants, RAG) so integrators do not copy Farm A YAML expecting a platform update.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Promotion model doc | [`enterprise-catalog-promotion-model.md`](../enterprise-catalog-promotion-model.md) |
| **WS2** | Site manifest pins | `platform.catalog_version_min` in `site-manifest.example.yaml` |
| **WS3** | Commons playbook cross-link | [`commons-catalog-operator-playbook.md`](../commons-catalog-operator-playbook.md) |
| **WS4** | Two-farm smoke | `smoke_phase98_test.go` |
| **WS5** | Topology + architecture links | `hypothetical-enterprise-topology.md`, phase-14 index |

---

## Integrator quick reference

| Promotes everywhere | Stays on one farm |
|---------------------|-------------------|
| `make migrate` (catalog seed) | Settings EC override |
| Same YAML on all sites | Genetics profile (94) |
| | `apply-agronomy-overrides.sh --farm-id N` |
| | RAG ingest per farm |

---

## Automated tests

| Test | Path |
|------|------|
| Farm A override ≠ Farm B builtin | `cmd/api/smoke_phase98_test.go` |

---

## OC-98

Phase 98 is **closed** when promotion model is linked from enterprise docs, site manifest pins catalog version, and two-farm smoke proves overrides do not promote.
