---
name: Phase 187 — Guardian relative due_date revise
overview: >
  Operators refine pending tasks with plain-language deadlines ("due tomorrow",
  "due in 3 days") from the sit-in Refine examples. Extends Phase 186 ISO
  matchers with relative parsing and updates the task dialogue smoke to assert
  tomorrow via WantDueDateOffsetDays.
todos:
  - id: ws1-relative-matchers
    content: "WS1: parseTaskRelativeDueDateAt — tomorrow, today, next week, due in N days"
    status: completed
  - id: ws2-scenario-offset
    content: "WS2: scenario WantDueDateOffsetDays + turn 4 make it due tomorrow"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: Go tests with fixed now + phase-187-closure + docs"
    status: completed
isProject: false
---

# Phase 187 — Guardian relative due_date revise

**Status:** shipped · **Depends on:** [186](phase_186_guardian_task_due_date_revise.plan.md)

## The problem

Phase 186 shipped ISO due-date revise (`set the due date to 2026-07-20`), but
operators naturally say **"make it due tomorrow"** — the exact phrasing from
the sit-in Refine coaching examples. Without relative parsing, those turns
answer in chat without bumping `Revision`.

## What shipped

### WS1 — Relative matchers (`proposals_revise.go`)

- `parseTaskDueDateRevisionAt` tries ISO patterns first, then relative:
  - `due tomorrow` / `make it due tomorrow`
  - `due today`
  - `due next week` (+7 days UTC)
  - `due in N days`
- `taskDueDateRevisionCue` avoids spurious matches on clarifying questions.

### WS2 — Smoke scenario

`scenario-task-dialogue-pending` turn 4 is now `make it due tomorrow` with
`WantDueDateOffsetDays: 1` (computed at assertion time in UTC).

### WS3 — Tests & docs

- Go: fixed-`now` unit tests + dynamic tomorrow assertion
- Vitest: `phase-187-closure.test.js`
- Docs: `current-state.md`, `ci-guardian-qa.md`, operator tour §7s

## Operator verification

Re-run `make guardian-qa-change-requests-ui` and Refine a pending task with
"due tomorrow" in the UI — pending row should show tomorrow's date at rev 4+.
