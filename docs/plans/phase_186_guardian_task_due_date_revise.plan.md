---
name: Phase 186 — Guardian create_task due_date revise
overview: >
  Completes the Phase 183 task-revise arc: wire due_date through Guardian
  create_task execution and add rule-based due-date corrections on pending
  proposals, extending the multi-turn task dialogue smoke to four turns.
todos:
  - id: ws1-tool-due-date
    content: "WS1: optionalDateFromArgs + execCreateTask/createTaskFromAlert pass due_date to CreateTask"
    status: completed
  - id: ws2-revise-matchers
    content: "WS2: parseTaskDueDateRevision in proposals_revise.go + LLM schema validation"
    status: completed
  - id: ws3-scenario-docs
    content: "WS3: scenario-task-dialogue-pending 4th turn + WantDueDate + phase-186-closure + docs"
    status: completed
isProject: false
---

# Phase 186 — Guardian create_task due_date revise

**Status:** shipped · **Depends on:** [183](phase_183_guardian_knowledge_and_revise_followups.plan.md) · [185](phase_185_guardian_task_zone_revise.plan.md)

## The problem

Phase 183 listed due-date corrections alongside title/description/zone revise,
but only title/description shipped in 183 and zone in 185. The REST task API
and DB already support `due_date`, yet Guardian's `create_task` tool ignored
it — so even a matched revise could not persist a deadline on Confirm.

## What shipped

### WS1 — Tool execution (`tools/tasks.go`, `tools/args.go`)

- `optionalDateFromArgs` parses `YYYY-MM-DD` proposal args.
- `execCreateTask` and `createTaskFromAlertRow` pass `DueDate` into
  `CreateTaskParams`.

### WS2 — Revise matchers (`proposals_revise.go`)

- `parseTaskDueDateRevision` handles `due date should be 2026-07-20` and
  `set the due date to 2026-07-20`.
- LLM proposal validation rejects malformed `due_date` strings.

### WS3 — Smoke + docs

`scenario-task-dialogue-pending` is now **4 turns** (create → zone → title →
due date), asserting `MinRevision: 4`, `RequireTaskZone`, `WantTitle`,
`WantDueDate`.

## Operator verification

Re-run `make guardian-qa-change-requests-ui` and confirm the task pending
card shows rev 4 with zone, title, and due date before manual Confirm.
