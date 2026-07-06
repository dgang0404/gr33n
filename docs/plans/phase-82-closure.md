# Phase 82 — closure (OC-82)

**Status:** **Partially shipped** on `main` — catalog + grounding core delivered; **WS7 plant context bundle shipped (Phase 136)**; full target-vs-actual deltas remain partial.

**Canonical plan:** [`phase_82_guardian_crop_grounding_hardening.plan.md`](phase_82_guardian_crop_grounding_hardening.plan.md)

**Formal audit:** [Phase 110](phase_110_phase_82_formal_closure.plan.md) · **OC-110** includes this closure.

**Builds on:** Phase 64 crop profiles, Phase 67 vision, Phases 85–87 catalog chain, Phase 106 symptoms.

---

## The one job (partial)

> **Guardian answers plant questions using structured crop profiles and honest grounding** — no fake `[n]` citations at zero RAG chunks, multi-crop compare, substrate-aware watering, and unsupported crops handled plainly.

---

## Workstream checklist

| WS | Deliverable | Status | Verify / deferred |
|----|-------------|--------|-------------------|
| **WS0** | Ops — RAG ingest gate, model floor | **Partial** | [`local-operator-bootstrap.md`](../local-operator-bootstrap.md) |
| **WS1** | Zero-chunk guardrail | **Shipped** (110) | `synthesis_test.go`, `smoke_phase82_test.go` |
| **WS2** | UI honesty banner | **Shipped** (110) | `phase-82-closure.test.js` |
| **WS3** | Multi-crop `lookup_crop_targets` | **Shipped** | `readtools_crop_test.go` |
| **WS4a** | `crop_library.yaml` v4 (~50 crops) | **Shipped** | Phase 84 DB seed |
| **WS4b–c** | Tier A/B/C profiles | **Shipped** | `smoke_phase64_test.go` |
| **WS4d** | Per-crop field guides | **Partial** | `field-guide-manifest.yaml` |
| **WS4e** | Unsupported registry | **Shipped** | `smoke_phase82_test.go` |
| **WS4f** | Grouped searchable picker | **Shipped** | `crop-library-picker.test.js` |
| **WS5** | Follow-up chips | **Partial** | `guardianFollowUps.js` |
| **WS6** | Docs + smokes + OC-82 | **Shipped** (110) | This doc |
| **WS7** | Plant context bundle | **Shipped** ([Phase 136](phase_136_guardian_plant_context_bundle.plan.md)) | `readtools_plant_bundle_test.go`, `smoke_phase136_test.go` |
| **WS8** | Substrate watering | **Shipped** | `readtools_crop.go` |
| **WS9** | Symptom / deficiency | **→ Phase 106** | `smoke_phase106_test.go` |
| **WS10** | Stage transitions | **Partial** | `readtools_grow.go` |
| **WS11** | Target vs actual deltas | **Partial** | [Phase 97](phase_97_rag_structured_truth_governance.plan.md) |

---

## Automated tests

| Test | Path |
|------|------|
| Picker + zero-chunk | `cmd/api/smoke_phase82_test.go` |
| Multi-crop read tool | `internal/farmguardian/readtools_crop_test.go` |
| UI honesty | `ui/src/__tests__/phase-82-closure.test.js` |

---

## OC-82

Phase 82 **core catalog + grounding** is **closed**: ≥46 profiles, catalog chain, multi-crop lookup, unsupported handling, zero-chunk guardrail, picker UI.

**Remaining depth** (WS7, WS11, WS10 widening) → Phase 97+ — not blockers for OC-82.
