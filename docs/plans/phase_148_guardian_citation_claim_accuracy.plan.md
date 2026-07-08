---
name: Phase 148 — Guardian citation-claim accuracy hardening
overview: >
  Smoke run #6 found real answer-accuracy failures that Phase 145's topical
  CitationAlignmentNote never sees: a claim's [n] pointing at the wrong source
  (humidity claim cited [3] — the light-schedule chunk — while [5] was the
  actual humidity alert), a duplicated OHN alert re-listed under a second
  number, a garbled digit/word merge ("0sourced"), and a blueberry pH range
  relabeled with EC's mS/cm units. Phase 148 adds detectors for each failure
  mode and wires them into the smoke/regression drift scorer.
todos:
  - id: ws1-citation-claim-mismatch
    content: "WS1: CitationClaimMismatchNote — flag [n] claims whose discriminating terms match a different cite"
    status: completed
  - id: ws2-duplicate-garbled
    content: "WS2: DuplicateListItemNote + GarbledTokenNote for repeated items and digit/word merges"
    status: completed
  - id: ws3-ec-ph-unit-confusion
    content: "WS3: ECPHUnitConfusionNote — pH value relabeled with mS/cm (EC) units"
    status: completed
  - id: ws4-wire-scorer
    content: "WS4: AnswerAccuracyNote wired into SmokeTopicDriftNote for all categories (not just field_guide)"
    status: completed
isProject: false
---

# Phase 148 — Guardian citation-claim accuracy hardening

**Status:** **Shipped.** · **Depends on:** [145](phase_145_guardian_topic_drift_and_grounding.plan.md) · [147](phase_147_guardian_smoke_run5_closure.plan.md)

---

## Why this phase

Phase 145's `CitationAlignmentNote` only checks whether cited excerpts are *topically* on-topic for `field_guide` questions. Run #6 (2026-07-08, archive `20260708T153829_smoke_phi3-mini.json`) showed a class of failure that check cannot see: the model's own arithmetic linking a claim to a citation number is wrong, even when every cited excerpt is on-topic.

| Run #6 symptom | Root cause | Phase 148 detector |
|---|---|---|
| "High humidity alert … [3]" but [3] is the light-schedule chunk (humidity is [5]) | Small-model citation-number/content mismatch | `CitationClaimMismatchNote` |
| OHN alert listed as item 2, then again as item 4 | Free-form enumeration drift | `DuplicateListItemNote` |
| "threshold of **0sourced** from FIELD GUIDE" | Dropped space / token merge | `GarbledTokenNote` |
| Blueberry **pH 4.5–5.5** relabeled "**4.5–5.5 mS/cm**" for kale | Unit confusion across crops in the same RAG batch | `ECPHUnitConfusionNote` |

## Workstreams

### WS1 — Citation-claim mismatch ✅

**Shipped:** `internal/farmguardian/answer_accuracy.go` — `CitationClaimMismatchNote` extracts terms near each `[n]` and only trusts terms that discriminate between cited excerpts (terms present in *every* excerpt, like a shared zone name, are ignored so they can't mask a real mismatch).

### WS2 — Duplicate items & garbled tokens ✅

**Shipped:** `DuplicateListItemNote` (Jaccard ≥0.4 over significant words between numbered list items); `GarbledTokenNote` (digit glued to a 5+ letter word, e.g. `0sourced`, with a short unit allowlist).

### WS3 — EC/pH unit confusion ✅

**Shipped:** `ECPHUnitConfusionNote` — flags a `mS/cm`-labeled range in the answer when that same number range appears in a cited excerpt as a `pH` value.

### WS4 — Scorer wiring ✅

**Shipped:** `topic_drift.go` calls `AnswerAccuracyNote` for **every** category (not gated to `field_guide`), so farm-state answers like `smoke-unread-alerts` are now checked too — this is what run #6's alert mismatch needed and Phase 145 didn't cover.

---

## Acceptance

- [x] Unit tests reproduce the exact run #6 answer/citation text and are caught by each detector.
- [x] `TestSmokeTopicDrift_runSixUnreadAlertsCitationMismatchNowCaught` (end-to-end via `SmokeTopicDriftNote`) fails the run #6 answer.
- [x] No regression in existing `topic_drift_test.go` / `answer_citation_align_test.go` suites.
- [x] Detectors run for `farm_state` category, not just `field_guide`.

## Non-goals

- Perfect citation-number correction (detection only; Phase 149 addresses prevention for alerts).
- LLM-based fact verification (kept to cheap string/set heuristics, CPU-safe).
