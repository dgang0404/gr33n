---
name: Phase 106 — Deficiency & pest symptom catalog
overview: >
  Structured symptom/deficiency catalog in DB + RAG; Guardian intents and vision
  hypotheses grounded per crop_key — extends Phase 82 WS9 after crop chain stable.
todos:
  - id: ws0-deps
    content: "WS0: Phases 87 + 97 shipped — structured EC wins over RAG numbers"
    status: completed
  - id: ws1-schema
    content: "WS1: agronomy_symptom_entries + links to crop_key / category"
    status: completed
  - id: ws2-guides
    content: "WS2: crop-deficiency-patterns.md + per-crop symptom sections; DB seed like field guides"
    status: completed
  - id: ws3-rag
    content: "WS3: RAG source_type symptom_guide; ingest from DB"
    status: completed
  - id: ws4-guardian
    content: "WS4: lookup_symptoms read tool + vision synergy (Phase 67); hypothesis not diagnosis"
    status: completed
  - id: ws5-ui
    content: "WS5: Zone Plants — 'What's wrong?' starter chips per crop_key"
    status: completed
  - id: ws6-smokes
    content: "WS6: Guardian yellow leaves tomato — cites symptom guide + live EC"
    status: completed
isProject: false
---

# Phase 106 — Deficiency & pest symptom catalog

## Status

**Shipped** on `main`. Closure: [`phase-106-closure.md`](phase-106-closure.md) (**OC-106**).

Carries forward [Phase 82 WS9](phase_82_guardian_crop_grounding_hardening.plan.md) as a **DB-backed catalog** after the crop identity chain (85–87) is stable.

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md), [Phase 97](phase_97_rag_structured_truth_governance.plan.md).

**Closure:** **OC-106**

---

## The one job

> **"Yellow leaves on my tomato"** → Guardian pulls **symptom catalog** + **live EC/pH** + **crop_key profile** — hypothesis with checks, not a fake diagnosis.

---

## Scope

| In | Out |
|----|-----|
| Deficiency patterns by crop category | Medical / pesticide label advice |
| Pest **symptom** checklists (chewing, spotting) | Pest ID from photo alone as fact |
| Links to `crop_key` or category (fruiting, leafy) | Per-pesticide product recommendations |

---

## Schema (WS1)

```sql
-- gr33ncrops.agronomy_symptom_entries
-- symptom_key, display_name, crop_keys[], categories[], body_md, severity_hint
```

Seed from `docs/field-guides/crop-deficiency-patterns.md` + per-crop guide sections.

---

## Guardian (WS4)

- New read tool **`lookup_crop_symptoms`** — filter by `crop_key` + keyword (yellow, tip burn, …)
- Vision (Phase 67): output **hypothesis** + "confirm with EC/pH and symptom guide"
- Persona: never diagnose; always offer measurable checks

---

## Acceptance

- [x] Symptom guide ingested; Guardian cites on interveinal yellowing question
- [x] Structured EC from `lookup_crop_targets` shown alongside narrative
- [x] UI starter chip on Zone Plants tab

**Prompt loop:** **`phase 106`**.
