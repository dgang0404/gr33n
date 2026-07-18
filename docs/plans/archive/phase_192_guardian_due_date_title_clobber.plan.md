---
name: Phase 192 — Guardian create_task due-date revise must not clobber title
overview: >
  Live eval run of scenario-task-dialogue-pending (2026-07-16) completed all
  four turns and left a pending proposal at revision 4, but turn 4
  "make it due tomorrow" overwrote title to "due tomorrow" instead of
  keeping "Refill calcium nitrate". Root cause: reviseTitleCallPattern
  matches "make (it|the title) …" before due-date parsing runs, and
  capture group 1 greedily takes "due tomorrow" as the new title.
findings: >
  Pending proposal d02a4ae6… args: {"title":"due tomorrow","zone_id":1,
  "due_date":"2026-07-17"}. Eval archive notes:
  proposal title "due tomorrow" want "Refill calcium nitrate". due_date and
  zone_id are correct — only title is wrong.
todos:
  - id: ws1-title-pattern-guard
    content: "WS1: tighten reviseTitleCallPattern / parseTaskTitleRevision — reject captures that are only relative due-date phrases (due tomorrow, due in N days, due today, due next week); require explicit title cues (call it, rename to, title should be) when phrase contains due/deadline"
    status: completed
  - id: ws2-apply-order-priority
    content: "WS2: in applyRevisionDeltas create_task case, parse due date before title OR skip title revision when taskDueDateRevisionCue matches and no explicit title cue"
    status: completed
  - id: ws3-tests
    content: "WS3: Go tests — make it due tomorrow keeps prior title; call it X still works; make the title due tomorrow edge case; regression from live eval fixture"
    status: completed
  - id: ws4-scenario-closure
    content: "WS4: phase-192-closure.test.js + update scenario-task-dialogue-pending docs in ci-guardian-qa.md"
    status: completed
isProject: false
---

# Phase 192 — Guardian create_task due-date revise must not clobber title

**Status:** shipped · **Depends on:** [187](phase_187_guardian_relative_due_date_revise.plan.md) · **Blocks:** [198](phase_198_guardian_task_dialogue_eval_rerun.plan.md)

## The problem

Phase 187 shipped relative due-date revise (`make it due tomorrow`). The
`scenario-task-dialogue-pending` smoke expects after turn 4:

- `WantTitle`: **Refill calcium nitrate**
- `WantDueDateOffsetDays`: **1**
- `RequireTaskZone`: zone set
- `MinRevision`: **4**

Live run on phi3:mini (2026-07-16) left a pending row with correct
`due_date` and `zone_id`, but `title` became **`due tomorrow`** because
`reviseTitleCallPattern` includes `make (?:it|the title)` and matches
turn 4 before due-date logic can win.

## Root cause (code)

```go
reviseTitleCallPattern = `(?i)(?:call it|title(?:\s+should be)?|rename (?:it )?to|make (?:it|the title))\s+["']?([^"'\n.;]+?)...`
```

`make it due tomorrow` → capture group = `due tomorrow` → title overwritten.

## What to ship

### WS1 — Title matcher guard

- Add `looksLikeDueDatePhrase(s string) bool` for: `due tomorrow`, `due today`,
  `due next week`, `due in N days`, bare ISO dates after "due".
- `parseTaskTitleRevision` returns false when capture is only a due-date phrase.
- Optionally narrow `make it` title branch to require `make it <quoted title>`
  without `due` keyword unless `make the title` is explicit.

### WS2 — Revision priority in `applyRevisionDeltas`

For `create_task` / `create_task_from_alert`:

1. Parse due date first when `taskDueDateRevisionCue(question)`.
2. Only run `parseTaskTitleRevision` when due-date did not match OR explicit
   title cue present (`call it`, `rename`, `title should be`).

### WS3 — Tests

- `TestApplyRevisionDeltas_CreateTaskDueTomorrowPreservesTitle` — prior title
  `Refill calcium nitrate`, question `make it due tomorrow` → title unchanged,
  `due_date` set.
- `TestApplyRevisionDeltas_CreateTaskCallItStillWorks` — `call it Refill calcium nitrate instead`.
- Live eval regression string in test name/docs.

### WS4 — Closure & docs

- `ui/src/__tests__/phase-192-closure.test.js`
- Note in `docs/ci-guardian-qa.md` under `scenario-task-dialogue-pending`

## Acceptance criteria

- [ ] `applyRevisionDeltas("create_task", {title:"Refill calcium nitrate"}, "make it due tomorrow")` changes only `due_date`
- [ ] `make guardian-qa-change-requests-ui` scenario-task-dialogue-pending passes title assertion (after [198] re-run)
- [ ] Pending UI card shows **Create task: Refill calcium nitrate** with due tomorrow in args diff

## Operator verification

1. Open Pending → existing revision-4 proposal (or re-run eval after 198).
2. Confirm title reads **Refill calcium nitrate**, not **due tomorrow**.
3. Refine with "make it due in 3 days" — title must stay put, due_date updates.
