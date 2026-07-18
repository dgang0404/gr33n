# Phase 97 — closure (OC-97)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_97_rag_structured_truth_governance.plan.md`](phase_97_rag_structured_truth_governance.plan.md)

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md) crop chain.

**Closes:** Blind spot **#8** — field guides vs `crop_profiles` contradiction.

---

## The one job (done)

> **Structured `lookup_crop_targets` wins over stale RAG narrative EC** — farm overrides apply immediately; field guides supplement qualitative context only.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Persona structured-wins rule | `StructuredTruthGroundingRule` in `platform_context.go` |
| **WS2** | Re-ingest triggers table | [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) |
| **WS3** | RAG chunk `crop_key` + `catalog_version` metadata | `field_guides.go` ingest |
| **WS4** | Strip nutrient numbers from field_guide chunks when read tool ran | `structured_truth.go` + `chat/handler.go` |
| **WS5** | Override EC smoke | `smoke_phase97_test.go` |

---

## Operator behavior

| Change | Guardian EC source | Re-ingest RAG? |
|--------|-------------------|----------------|
| Settings farm override | Immediate via read tool | No |
| Genetics profile (94) | Immediate via read tool | No |
| Catalog seed bump | After migrate | Yes if guide body changed |
| Chat question | `lookup_crop_targets` first | Guides = narrative only |

When both RAG chunks and `lookup_crop_targets` appear on one turn, field-guide EC numbers are stripped before the LLM sees them.

---

## Automated tests

| Test | Path |
|------|------|
| Farm override EC in read tool | `cmd/api/smoke_phase97_test.go` |
| RAG nutrient strip | `internal/rag/synthesis/structured_truth_test.go` |
| Field guide metadata | `internal/rag/ingest/field_guides_test.go` |

---

## OC-97

Phase 97 is **closed** when persona + runbook document structured-wins, chat strips conflicting RAG numbers, and smokes prove farm override EC beats stale narrative.
