---
name: Phase 98 — Enterprise catalog promotion model
overview: >
  Document what promotes platform-wide (catalog, seed packs) vs stays farm-local
  (overrides, plant slots) for multi-site orgs — site-manifest and commons clarity.
todos:
  - id: ws1-doc
    content: "WS1: docs/enterprise-catalog-promotion-model.md — promote vs local matrix"
    status: completed
  - id: ws2-manifest
    content: "WS2: site-manifest.example.yaml — catalog_version pin + override pack path"
    status: completed
  - id: ws3-commons
    content: "WS3: Commons playbook — agronomy pack vs platform catalog vs farm overrides"
    status: completed
  - id: ws4-smoke
    content: "WS4: Two-farm smoke — Farm A override ≠ Farm B builtin"
    status: completed
  - id: ws5-architecture
    content: "WS5: hypothetical-enterprise-topology.md + architecture cross-link"
    status: completed
isProject: false
---

# Phase 98 — Enterprise catalog promotion model

## Status

**Shipped (OC-98).** Closure: [`phase-98-closure.md`](phase-98-closure.md).

**Depends on:** [Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md), [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md).

**Closure:** **OC-98**

---

## Blind spot #9

| Artifact | Scope | Promotes how |
|----------|-------|--------------|
| `crop_catalog_entries` + profiles | **Platform** | SQL migrate on every site |
| Commons agronomy seed pack | **Org optional import** | `import-agronomy-seed-pack.sh` |
| Farm EC override | **Single farm** | Settings or YAML; never auto-promoted |
| `plants.crop_key` slots | **Single farm** | Per site |
| Field guide RAG | **Per farm ingest** | `guardian-bootstrap-farm` |

Integrators must not copy Farm A override YAML to Farm B expecting platform update.

---

## Promote vs local matrix (WS1)

Document with diagrams: HQ publishes catalog migration → all sites migrate → each site applies **local** override pack optional.

---

## Acceptance

- [x] Enterprise README links promotion model
- [x] site-manifest documents `platform_catalog_version` expectation
- [x] Two-farm smoke in CI or documented manual checklist

**Prompt loop:** **`phase 98`**.
