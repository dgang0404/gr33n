---
name: Phase 191 — Guardian revise: question-phrased description additions
overview: >
  Part 4 of the Phase 188 audit. tryReviseActiveProposal only recognizes
  correction-style revisions ("call it X", "description should be X", "zone
  5", "due tomorrow"). Turn id=13 shows a farmer naturally phrasing a
  correction as a question instead — "Should this task mention checking
  stock in Veg Tent?" — which matches none of the revise patterns, so the
  turn fell through to a raw, ungrounded LLM chat call that produced an
  incoherent, unrelated answer (Chrysanthemum care / electrical safety) and
  left the pending create_task proposal completely unrevised, silently
  dropping the operator's correction.
findings: >
  Turn id=13 (session 0e4fb217…, revision turn on a pending create_task
  proposal): user asks "Please revise this change request — Create task:
  Follow up from Guardian chat. Correction: Should this task mention checking
  stock in Veg Tent?" — applyRevisionDeltas' create_task case checks title,
  description ("description should be X" / "details: X" only), zone, and due
  date patterns; none match a bare "should this (task) mention X" question,
  so changed=false and the turn falls through to open-ended chat.
todos:
  - id: ws1-append-pattern
    content: "WS1: reviseDescriptionAppendPattern + parseTaskDescriptionAppendRevision — matches 'should (this/it)(task)? (also)? mention/include/say/note/add X' and appends X as a new sentence onto the existing description (or sets it if empty) rather than replacing it outright"
    status: completed
  - id: ws2-wire
    content: "WS2: wire into applyRevisionDeltas create_task/create_task_from_alert case, after the replace-style parseTaskDescriptionRevision so an explicit 'description should be' still wins"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: Go tests (question-phrased append, empty-description case, non-match passthrough) + phase-191-closure.test.js + docs"
    status: completed
isProject: false
---
