---
name: Phase 151 — Guardian alert citation enforcement
overview: >
  Smoke run #8 completed unread-alerts with severity-correct prose but zero [N]
  citation markers (markdown links instead). Phase 149 only fixes ordering when
  the model uses bracket cites; Phase 151 forces, detects, filters, and
  backfills numbered alert citations end-to-end.
todos:
  - id: ws1-prompt-override
    content: "WS1: Strengthen alertCitationDiscipline — LIVE FARM STATE context-only, require [n], forbid markdown links"
    status: completed
  - id: ws2-smoke-heuristic
    content: "WS2: smoke-unread-alerts requires citation_count > 0 or [1]/[2] in answer"
    status: completed
  - id: ws3-missing-cites-detector
    content: "WS3: MissingNumberedCitationsNote + broaden hallucinated URL detection (gr33ncore.*)"
    status: completed
  - id: ws4-alert-only-retrieval
    content: "WS4: FilterChunksForAlertSummary — numbered sources = alert_notification only for summarize-alerts intent"
    status: completed
  - id: ws5-cite-injection
    content: "WS5: InjectAlertCitationRefs post-generation safety net wired into grounded finalize"
    status: completed
  - id: ws6-one-cite-per-item
    content: "WS6: One [n] per list item prompt + alert-only discipline + NormalizeAlertListCitations (run #9 stray [3])"
    status: completed
isProject: false
---

# Phase 151 — Guardian alert citation enforcement

**Status:** **Shipped.** · **Depends on:** [149](phase_149_guardian_alert_citation_ordering.plan.md) · [148](phase_148_guardian_citation_claim_accuracy.plan.md)

---

## Why this phase

Run #7 proved Phase 149's severity sort in unit tests but could not complete live. Run #8 completed under bumped timeouts with correct alert **order** but `citations=0` — the model used markdown links (`gr33ncore.sensor_alerts`) instead of `[1]`/`[2]`/`[3]`. `platformDocGrounding` told the model to rely on LIVE FARM STATE for unread alerts, conflicting with the base synthesis prompt. Phase 148's mismatch detector never fired without bracket numbers.

## Workstreams

### WS1 — Alert citation prompt override ✅

**Shipped:** `internal/rag/synthesis/guardian.go` — `alertCitationDiscipline` clarifies LIVE FARM STATE / `list_unread_alerts` are context only; each list item must end with matching `[n]`; no markdown links or invented URLs.

### WS2 — Smoke heuristic ✅

**Shipped:** `internal/farmguardian/eval/score.go` — `smoke-unread-alerts` requires `CitationCount > 0` or `[1]`/`[2]` in answer (run #8 archive would fail).

### WS3 — Missing-cite detector + fake URLs ✅

**Shipped:** `internal/farmguardian/answer_accuracy.go` — `MissingNumberedCitationsNote`; `answer_citation.go` — `gr33ncore` / `sensor_alerts` in `isHallucinatedCitationURL`.

### WS4 — Alert-only numbered sources ✅

**Shipped:** `internal/farmguardian/alert_summary.go` — `MatchAlertSummaryIntent` (includes summarize-alerts), `FilterChunksForAlertSummary`; wired in `retrieveChunks` after `PrioritizeAlertChunks`.

### WS5 — Post-generation cite injection ✅

**Shipped:** `internal/farmguardian/alert_cite_inject.go` — `InjectAlertCitationRefs` appends `[1]`…`[n]` to numbered list lines when model omits markers; wired in `answer_finalize.go` before `BuildCitations`.

### WS6 — One cite per list item (run #9 follow-up) ✅

**Shipped:** `guardian.go` — one `[n]` per item + `alertOnlyCitationDiscipline` when Sources are alert-only; `alert_cite_normalize.go` — `NormalizeAlertListCitations` strips stray extra `[n]` on the same list item; `MultipleCitationsPerListItemNote` eval detector.

---

## Acceptance

- [x] `GuardianRAGInstructions` includes LIVE FARM STATE / no-markdown-link discipline when 2+ alert chunks.
- [x] Run #8-style uncited alert list fails `smoke-unread-alerts` heuristic and `MissingNumberedCitationsNote`.
- [x] `gr33ncore.sensor_alerts` URLs sanitized and flagged as fake citation links.
- [x] Summarize-alerts questions get alert-only numbered source list when 2+ alert chunks retrieved.
- [x] Cite injection adds `[n]` to numbered list items when model omits them.
- [x] `make guardian-qa-smoke-unread-alerts` Make target exists (Phase 147 run #8).

## Non-goals

- Fully deterministic (non-LLM) alert answer rendering.
- Requiring citations on single-alert turns (injection and filter activate at 2+ alerts only).
