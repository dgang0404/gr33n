---
name: Phase 185 — Guardian create_task zone revise
overview: >
  Completes Phase 183 WS3 follow-up: rule-based zone assignment on pending
  create_task proposals (name resolution + numeric zone_id), and extends the
  multi-turn task dialogue smoke scenario to assert zone + title revise chain.
todos:
  - id: ws1-zone-matchers
    content: "WS1: taskZoneRevisionCue + applyTaskZoneRevision (resolveZoneIDForIntent) + numeric zone_id in applyRevisionDeltas"
    status: completed
  - id: ws2-scenario
    content: "WS2: scenario-task-dialogue-pending turn 2 zone assign, MinRevision 3, RequireTaskZone"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: proposals_revise_test + fixture test + phase-185-closure + current-state/ci-guardian-qa/operator-tour"
    status: completed
isProject: false
---

# Phase 185 — Guardian create_task zone revise

**Status:** shipped · **Depends on:** [183](phase_183_guardian_knowledge_and_revise_followups.plan.md) · [184](phase_184_guardian_pr_conversation_smoke.plan.md)

## The problem

Phase 183 shipped title/description revise matchers for `create_task` /
`create_task_from_alert`, but zone assignment still required a fresh proposal
or model re-run. Operators refining a pending task often say things like
"put it in Veg Room" — that should bump `Revision` the same way volume/title
corrections do.

## What shipped

### WS1 — Zone revise matchers (`proposals_revise.go`)

- **Name-based:** `taskZoneRevisionCue` + `applyTaskZoneRevision` calls
  `resolveZoneIDForIntent` when the turn looks like an assignment (not a
  clarifying "which zone?" question).
- **Numeric:** `parseTaskZoneIDNumeric` handles `zone 3` / `zone id 12` in
  `applyRevisionDeltas` without a DB round-trip.

### WS2 — Multi-turn smoke extension

`scenario-task-dialogue-pending` is now **3 turns**:

1. Create task (no zone)
2. `Put it in Veg Room — that is the zone for this task.` (rev 2, `zone_id` set)
3. `call it Refill calcium nitrate instead` (rev 3, title)

Eval asserts `MinRevision: 3`, `RequireTaskZone: true`, `WantTitle`.

### WS3 — Tests & docs

- Go: `proposals_revise_test.go` (cue + numeric), `fixtures_change_requests_ui_test.go`
- Vitest: `phase-185-closure.test.js`
- Docs: `current-state.md`, `ci-guardian-qa.md`, operator tour §7q

## Operator verification

Re-run `make guardian-qa-change-requests-ui` and confirm the task pending card
shows **rev 3** with zone + title after the dialogue (manual Pending-tab check;
same as Phase 184 WS5).
