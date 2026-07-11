---
name: Phase 152 — Guardian live accuracy guardrails + citation deep links
overview: >
  A live Farm Counsel run surfaced answer-quality bugs (garbled truncation,
  an uncited "Week 9" claim borrowed from an unrelated completed task, and
  invented per-plant dosing math) that Phase 148/151's accuracy detectors
  would have caught — except those detectors only ran in the offline
  guardian-eval smoke harness, never in the live chat path. Phase 152 WS1
  closes that gap. WS2 (citation deep links to the zone/cycle/schedule a
  citation came from) is planned but not yet shipped.
todos:
  - id: ws1-live-detectors
    content: "WS1: Wire AnswerAccuracyNote into the live chat path (both streaming and non-streaming) as a non-mutating flag on the response + turn debug"
    status: completed
  - id: ws1b-new-detectors
    content: "WS1b: Add TruncatedAnswerTailNote, UncitedTimelineClaimNote, InventedAssumptionMathNote detectors"
    status: completed
  - id: ws1c-ui-banner
    content: "WS1c: UI banner surfaces accuracy_note as a farmer-facing caveat on the flagged turn"
    status: completed
  - id: ws2-citation-deep-links
    content: "WS2: Citation chips route to the zone/crop-cycle the source came from (crop_cycle, fertigation_program, task)"
    status: completed
  - id: ws2b-remaining-source-types
    content: "WS2b: Resolve routes for schedule, alert_notification, field_guide, platform_doc — shipped Phase 159"
    status: completed
isProject: false
---

# Phase 152 — Guardian live accuracy guardrails + citation deep links

**Status:** WS1 + WS2 + WS2b shipped · **Depends on:** [148](phase_148_guardian_citation_claim_accuracy.plan.md) · [151](phase_151_guardian_alert_citation_enforcement.plan.md) · [159](phase_159_guardian_citation_completeness.plan.md)

A live "Flower run (12/12)" Farm Counsel answer (`phi3:mini`, grounded, 5 chunks) contained four real defects in one turn:

1. **Truncated generation** — the answer stopped mid-word: `"...receiving consistent and ade0:"`.
2. **Cross-source contamination** — it stated `"it's now Week 9"` about the *active* early_flower cycle, but "Week 9" only appears in citation `[5]`, a **different, completed** `Harvest Flower Room A` task — and the claim had no `[n]` next to it at all.
3. **Invented derived math** — `"~1.2 mL per plant if we assume an average yield density"` — a number the model computed from an assumption, not from any cited record, dressed up with citation brackets that made it look sourced.
4. **Wrong year** — `"started on June 20th of last year"` when the cited `started_at` is this year.

None of this tripped anything, because `AnswerAccuracyNote` / `SmokeTopicDriftNote` (Phase 148/151) are called **only** from `internal/farmguardian/eval/score.go` — the offline smoke harness — never from `internal/handler/chat/answer_finalize.go` or `handler.go`. Every real farmer conversation was unprotected by detectors that already existed and already worked.

## Workstreams

### WS1 — Live wiring ✅

**Shipped:** `internal/handler/chat/answer_finalize.go` — `applyAnswerAccuracyNote(answer, citations)` converts response-shaped `synthesis.Citation` to `farmguardian.CitationSummary` and calls `farmguardian.AnswerAccuracyNote`. Called from both `PostV1`'s non-streaming branch and `streamResponse`'s streaming branch, right after `Citations = synthesis.BuildCitations(...)`. Logged via `slog.Info("guardian: answer_accuracy_flagged", "note", ...)`.

Deliberately **non-mutating**: the detectors are heuristic (regex/word-overlap) and can false-positive, so a flagged answer is never silently rewritten or blocked — it's surfaced so the operator (and QA) can judge it, matching the "best-effort, never silent" pattern used elsewhere in this codebase (`trim_summary`, `field_degraded`, `zero_chunk` banners).

Exposed in two places:
- `postResponse.AccuracyNote` (top-level `accuracy_note` field, always present when non-empty — not gated behind `AUTH_MODE=dev/auth_test` like `Debug`, so it reaches real farm operators too).
- `TurnDebug.AccuracyNote` (dev turn inspector).

### WS1b — New detectors ✅

**Shipped:** `internal/farmguardian/answer_accuracy.go`, folded into `AnswerAccuracyNote`:

- `TruncatedAnswerTailNote` — flags an answer whose final token is a lowercase word-fragment glued to 1-2 trailing digits (`ade0:`), the generation-cutoff pattern from this run. Allowlists legitimate chemistry/unit tokens (`CO2`, `H2O`, `3D`, ...).
- `UncitedTimelineClaimNote` — flags a `Week N` / `Day N` progress claim with no `[n]` citation within a 60-char window either side. Crop-cycle chunks never carry a week/day-of-cycle field (only `started_at`/`stage`), so an uncited week claim can only be a borrow from some other chunk. Only runs when the turn has 1+ citations (meaningless for ungrounded quick chat).
- `InventedAssumptionMathNote` — flags a numeric claim justified by a hedge phrase (`assuming`, `if we assume`, `let's assume`, `hypothetically`, ...) within 80 chars of a digit. The model disclosing, in its own words, that a number was derived rather than sourced is itself the violation the base synthesis prompt forbids ("do not invent facts").

All three are pure string/regex checks — no added LLM calls, no measurable latency impact on the hot path.

### WS1c — UI banner ✅

**Shipped:** `ui/src/lib/guardianCitationLabels.js` — `accuracyNoteMessage(note)` maps the terse backend code (`citation_number_mismatch`, `truncated_answer_tail`, `uncited_timeline_claim`, `invented_assumption_math`, ...) to a short farmer-facing caveat, with a generic fallback for future codes. `GuardianChatPanel.vue` renders it as an amber banner (`data-test="chat-accuracy-banner"`) below the citation list, same visual treatment as the existing zero-chunk / trim-warning banners.

**Known limitation:** `accuracy_note` (like `trim_summary`) is response-only, not persisted to `gr33ncore.conversation_turns` — it won't reappear if you reload an old session's history. A follow-up could add a column so QA/feedback review can query flagged turns after the fact.

### WS2 — Citation deep links ✅

**Shipped:** `internal/farmguardian/citation_route.go` — `ResolveCitationRoute(ctx, q, farmID, sourceType, sourceID)` mirrors `BuildContextRefBlock`'s switch (`context_ref.go`) in the opposite direction: instead of mapping a UI route to a Guardian prompt block, it maps a citation to a UI route.

| source_type | Lookup | Route |
|---|---|---|
| `crop_cycle` | `GetCropCycleByID` (farm-scope checked) | `/crop-cycles/{id}/summary` |
| `fertigation_program` | `GetFertigationProgramByID` → `TargetZoneID` | `/zones/{zone_id}?tab=water` |
| `task` | `GetTaskByID` → `ZoneID` (nullable — unresolved when null) | `/zones/{zone_id}` |

Wired via `attachCitationRoutes` (`internal/handler/chat/answer_finalize.go`), called right after `synthesis.BuildCitations(...)` in both the streaming and non-streaming handlers. `synthesis.Citation` gained a `Route string \`json:"route,omitempty"\`` field. `GuardianChatPanel.vue` renders the citation chip as a `<router-link v-nav-hint="c.route">` when `route` is present (same sidebar-wiggle affordance used by every other in-app cross-link — `ui/src/directives/navHint.js`), plain text otherwise.

Every path re-checks farm ownership (`row.FarmID != farmID → unresolved`) even though the RAG retrieval that produced the citation was already farm-scoped — defense in depth so a citation can never route a click into another farm's data.

**Tested against the real seeded dev DB** (`internal/farmguardian/citation_route_db_test.go`, skips gracefully when `DATABASE_URL` is unreachable): creates ephemeral crop-cycle/program/task rows tied to a real zone under farm 1, asserts the resolved path, and asserts a cross-farm lookup and a zone-less task both fail closed.

### WS2b — Remaining source types (planned, not started)

Left unresolved on purpose — each needs its own join/scoping decision rather than being force-fit into WS2's shape:

- **`schedule`** — `gr33ncore.schedules` rows carry no `zone_id` of their own; the zone link only exists indirectly through whichever `fertigation_program`/`automation_rule` references the schedule. Needs a reverse lookup, and a schedule can plausibly serve more than one zone.
- **`alert_notification`** — only has `triggering_event_source_type` / `_id` (one more hop — e.g. sensor → zone). No dedicated `/alerts/:id` page exists in the router today either, so the *target* of that link needs a decision (zone page? a future alerts feed?) before it's worth resolving.
- **`field_guide` / `platform_doc`** — curated docs, not per-farm rows. A route would point at `/farm-knowledge` or `/symptom-guide`, but neither currently supports linking to a specific doc/anchor.

---

## Acceptance

- [x] `AnswerAccuracyNote` runs on every grounded and ungrounded chat turn (streaming and non-streaming), not just in `guardian-eval`.
- [x] The exact live-UI "Flower run (12/12)" answer (replayed verbatim in tests) trips `AnswerAccuracyNote`.
- [x] A clean, correctly-cited answer produces no note (no false-positive banner on every turn).
- [x] `accuracy_note` is a top-level response field, not gated behind dev/auth_test debug mode.
- [x] UI shows a farmer-facing caveat banner, not the raw detector code.
- [x] Citation chips link to their source zone/crop-cycle for `crop_cycle`, `fertigation_program`, and `task` sources.
- [x] Cross-farm and zone-less lookups fail closed (no route) rather than guessing.
- [x] `schedule`, `alert_notification`, `field_guide`, `platform_doc` routes (WS2b — shipped Phase 159).

## Non-goals

- Auto-regenerating or auto-rewriting a flagged answer (too risky without a second LLM pass; treated as a future option, not this phase).
- Persisting `accuracy_note` to the database (noted as a follow-up under WS1c).
- Resolving WS2b's remaining source types in this phase — each needs its own scoping decision (see WS2b above), not a mechanical extension of WS2's switch.
