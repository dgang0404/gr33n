---
name: Phase 97 — RAG vs structured truth governance
overview: >
  Persona rules, re-ingest triggers, and smokes so Guardian never cites stale RAG
  EC when farm override or catalog seed changed structured profiles.
todos:
  - id: ws1-persona
    content: "WS1: Persona block — structured lookup_crop_targets wins over RAG on numbers"
    status: completed
  - id: ws2-triggers
    content: "WS2: Re-ingest runbook triggers table — override vs catalog bump vs guide edit"
    status: completed
  - id: ws3-chunk-meta
    content: "WS3: RAG chunk metadata catalog_version + crop_key for stale detection"
    status: completed
  - id: ws4-guardian
    content: "WS4: Chat handler tag — if RAG chunk EC conflicts with read tool, drop chunk numbers"
    status: completed
  - id: ws5-smokes
    content: "WS5: Override cannabis EC → Guardian uses new mS/cm even if RAG chunk old"
    status: completed
isProject: false
---

# Phase 97 — RAG vs structured truth governance

## Status

**Shipped (OC-97).** Closure: [`phase-97-closure.md`](phase-97-closure.md).

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md).

**Closure:** **OC-97**

---

## Blind spot #8

| Source | Updates when |
|--------|--------------|
| **Structured** (`crop_profiles`, farm override) | Immediately on PUT Settings |
| **RAG** (field guides) | Only after `rag-ingest-field-guides` |

Guardian may cite old narrative EC unless governed.

---

## Operational rules (WS2)

| Event | Structured | RAG re-ingest |
|-------|------------|---------------|
| Farm EC override | ✅ immediate | ❌ not required for numbers |
| Platform catalog seed bump | ✅ after migrate | ✅ field guides if body changed |
| YAML EC edit + new migration | ✅ migrate | ✅ re-ingest |
| Operator chat question | Read tool first | Narrative supplement only |

Document in `crop-knowledge-operator-runbook.md` + Guardian persona.

---

## WS4 — Conflict resolution

Hard rule in `readtools_crop.go` / chat composer:

> If retrieved RAG chunk contains EC/pH/VPD numbers and `lookup_crop_targets` ran this turn, **strip numeric claims from RAG** before prompt injection.

---

## Acceptance

- [x] Persona + platform_context mirror structured-wins rule
- [x] Smoke: override EC → chat answer matches strip, not stale chunk
- [x] Runbook “when to re-ingest” table

**Prompt loop:** **`phase 97`**.
