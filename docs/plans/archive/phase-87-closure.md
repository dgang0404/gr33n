# Phase 87 — closure (OC-87)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_87_crop_knowledge_operator_closure.plan.md`](phase_87_crop_knowledge_operator_closure.plan.md)

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md), [Phase 86](phase_86_grow_ops_catalog_chain.plan.md), [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md).

---

## The one job (done)

> Operators and Guardian both trust **one farm knowledge base**: Postgres catalog, dropdown plants, Settings EC tweaks, structured targets in chat — no YAML at runtime, no typed crop names, no invented EC.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) | Operator steps 1–7 |
| **WS2** | [operator-tour §6m](../operator-tour.md#6m-plants--crop-knowledge-chain-phases-8587--shipped) | Catalog dropdown; no strain copy |
| **WS3** | [farm-guardian-architecture §7.0af](../farm-guardian-architecture.md#70af-plants--crop-knowledge-chain-phases-8587--shipped) | Chain table |
| **WS4** | `cmd/api/smoke_phase87_test.go` | Parity + compare + alias |
| **WS5** | This doc + phase-14 rows 84–87 **shipped** | Index updated |
| **WS6** | Runbook EC scope section | v1 farm-wide `crop_key` |
| **WS7** | Runbook → Phase 97 pointer | Structured wins on numbers |
| **WS8** | Runbook → Phase 98 pointer | Enterprise promotion |

---

## Arc 84–87 summary

| Phase | Shipped capability |
|-------|-------------------|
| **84** | Catalog + picker + commons API + field guides in Postgres |
| **85** | `plants.crop_key`; catalog-bound create/upsert; picker UI |
| **86** | Active grow → plant → profile; strip + Water/Light; Guardian chain |
| **87** | Runbook, architecture, smokes, closure |

**Next locked:** [Phase 93](phase_93_plant_identity_vocabulary_cleanup.plan.md) (vocabulary) · [Phase 101](phase_101_guardian_write_tools_crop_key.plan.md) (Guardian writes).

---

## Re-ingest after doc edits

```bash
make rag-ingest-platform-docs
```

Optional farm operational refresh: [`scripts/rag-ingest-farm-operational.sh`](../scripts/rag-ingest-farm-operational.sh).
