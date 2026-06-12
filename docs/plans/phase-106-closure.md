# Phase 106 — closure (OC-106)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_106_deficiency_pest_symptom_catalog.plan.md`](phase_106_deficiency_pest_symptom_catalog.plan.md)

**Depends on:** [Phase 87](phase_87_crop_knowledge_operator_closure.plan.md) crop knowledge chain; [Phase 97](phase_97_rag_structured_truth_governance.plan.md) structured EC over RAG numbers.

**Re-homes:** [Phase 82 WS9](phase_82_guardian_crop_grounding_hardening.plan.md) symptom/deficiency grounding as a DB-backed catalog.

---

## The one job (done)

> **“Yellow leaves on my tomato”** → Guardian runs **`lookup_crop_symptoms`** + **`lookup_crop_targets`** — ranked hypotheses and measurable checks, not a diagnosis from photos or narrative alone.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Structured EC wins over RAG | Phase 97 `StructuredTruthGroundingRule` |
| **WS1** | Symptom catalog schema + seed | `20260625_phase106_agronomy_symptom_catalog.sql` |
| **WS2** | Deficiency patterns guide | `crop-deficiency-patterns.md` in DB seed |
| **WS3** | RAG ingest for symptom guides | `source_type` symptom chunks |
| **WS4** | `lookup_crop_symptoms` read tool | `readtools_symptoms.go`, `SymptomGroundingRule` |
| **WS5** | Zone Plants starter chips | `guardianStarters.js` — `buildSymptomGrowStarters` |
| **WS6** | Smokes | `smoke_phase106_test.go` |

---

## Guardian contract

- Filter symptoms by `crop_key` + keyword (yellow, tip burn, spots, …)
- Always pair with `lookup_crop_targets` for live EC/pH/VPD
- Vision output is **hypothesis + inspection steps** — never pesticide or medical advice

---

## Automated tests

| Test | Path |
|------|------|
| Tomato yellow + EC targets | `cmd/api/smoke_phase106_test.go` |
| Read tool registered | `readtools_symptoms_test.go` |
| UI starters | `ui/src/__tests__/crop-symptoms.test.js` |

---

## OC-106

Phase 106 is **closed** when symptom catalog is seeded, Guardian cites structured symptoms with live targets on deficiency questions, and Plants tab offers symptom starter chips.
