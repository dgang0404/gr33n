---
name: Phase 149 — Guardian alert citation ordering & enumeration discipline
overview: >
  Phase 148 detects citation-claim mismatches after the fact. Phase 149 attacks
  the same run #6 failure architecturally: alert_notification chunks arrive in
  semantic-similarity order, not severity order, so a model that instinctively
  lists "most urgent first" ends up out of sync with the citation numbers it
  was given. Sorting alerts by severity before numbering, plus an explicit
  numbering-discipline instruction, removes the mismatch by construction for
  the common case instead of only flagging it.
todos:
  - id: ws1-severity-sort
    content: "WS1: PrioritizeAlertChunks — sort alert_notification chunks severity-desc before numbering"
    status: completed
  - id: ws2-wire-retrieval
    content: "WS2: Wire into retrieveChunks after FilterRAGChunks"
    status: completed
  - id: ws3-numbering-instruction
    content: "WS3: alertCitationDiscipline system-prompt addendum when 2+ alert chunks present"
    status: completed
isProject: false
---

# Phase 149 — Guardian alert citation ordering & enumeration discipline

**Status:** **Shipped.** · **Depends on:** [148](phase_148_guardian_citation_claim_accuracy.plan.md) · [145](phase_145_guardian_topic_drift_and_grounding.plan.md)

---

## Why this phase

Run #6's `smoke-unread-alerts` answer cited `[3]` for the humidity alert when `[3]` was actually the light-schedule chunk (humidity was `[5]`). The RAG search returned alert chunks in embedding-similarity order — not severity order — so the model's own "most urgent first" writing instinct and the citation numbers it was handed disagreed. Phase 148 can only flag this after generation; Phase 149 removes the disagreement before generation for the common alert-listing case.

## Workstreams

### WS1 — Severity-first alert ordering ✅

**Shipped:** `internal/farmguardian/alert_chunk_order.go` — `PrioritizeAlertChunks` moves `alert_notification` chunks to the front of the retrieved list, sorted `critical` → `high` → `medium` → `low` (stable sort; non-alert chunks and single-alert results are left untouched).

### WS2 — Wired into retrieval ✅

**Shipped:** `internal/handler/chat/handler.go` `retrieveChunks` calls `PrioritizeAlertChunks` right after `FilterRAGChunks`, so citation numbers `[1]`, `[2]`, `[3]` for a 3-alert answer are now most-severe-first by construction.

### WS3 — Numbering discipline instruction ✅

**Shipped:** `internal/rag/synthesis/guardian.go` — `HasMultipleAlertChunks` + `alertCitationDiscipline` block appended to `GuardianRAGInstructions` when 2+ alert sources are present: *"use exactly that order: your list item 1 must cite [1] … do not repeat the same alert under a second number."*

---

## Acceptance

- [x] `PrioritizeAlertChunks` unit tests: severity-desc reordering, non-alert chunks untouched, single-alert no-op.
- [x] `GuardianRAGInstructions` includes the discipline instruction only when 2+ alert chunks are retrieved.
- [x] No change to citation numbering for non-alert or single-alert grounded answers.

## Non-goals

- Fully deterministic (non-LLM) alert list rendering — the model still composes prose; this phase only fixes ordering and adds an explicit instruction.
- Reordering non-alert operational chunks (tasks, inventory) — scoped to alerts, the concrete run #6 failure.
