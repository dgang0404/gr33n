---
name: Phase 161 — Guardian ec-ph smoke closure
overview: >
  Post-160 natural follow-up: smoke run #5/#6/#7 still fail smoke-ec-ph on
  uncited_tail, blueberry crop drift, and citation-number mismatch. Phase 161
  trims uncited tails live, expands crop-drift detection, and updates doc drift.
todos:
  - id: ws1-trim-uncited-tail
    content: "WS1: TrimUncitedTail in live finalize — drop paragraphs off cited excerpts"
    status: completed
  - id: ws2-crop-drift
    content: "WS2: EcphCropDriftNote — blueberry/strawberry drift on leafy-greens EC/pH prompts"
    status: completed
  - id: ws3-doc-hygiene
    content: "WS3: Fix phase-14 115 status + phase_152 WS2b acceptance stale text"
    status: completed
  - id: ws4-closure
    content: "WS4: Unit tests + phase-161-closure.test.js"
    status: completed
isProject: false
---

# Phase 161 — Guardian ec-ph smoke closure

**Status:** shipped · **Origin:** [159–160 backlog](phase_159_160_post_158_gaps_backlog.plan.md) § P2 ec-ph smoke drift

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `TrimUncitedTail` wired in chat finalize (streaming + non-streaming) |
| **WS2** | `EcphCropDriftNote` in smoke scorer + live trim trigger |
| **WS3** | Doc hygiene — phase-14 115 shipped; phase_152 WS2b closed |
| **WS4** | `answer_citation_align_trim_test.go` + `phase-161-closure.test.js` |

## Still open (not this phase)

- `citation_number_mismatch` — detection only (Phase 148); prevention needs retrieval/prompt tuning
- Full smoke re-run on CPU laptop (operator task)
- Insert Commons federation

## Verification

```bash
go test ./internal/farmguardian/... -run 'TrimUncited|EcphCrop|Blueberry' -count=1
cd ui && npm test -- --run src/__tests__/phase-161-closure.test.js
```
