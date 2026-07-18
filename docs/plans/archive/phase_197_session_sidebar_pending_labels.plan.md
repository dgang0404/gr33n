---
name: Phase 197 — Session sidebar labels for multi-turn pending threads
overview: >
  Guardian chat session sidebar shows truncated first-line titles and turn
  counts. Eval/manual multi-turn pending threads are hard to find — e.g. a
  session with 4 turns may show a weak auto-title or empty label. Link
  session list entries to pending proposal summary when a pending proposal
  exists for that session_id.
todos:
  - id: ws1-session-label-helper
    content: "WS1: sessionLabel() enhancement — when sessions list or proposals store has pending proposal for session_id, prefer proposal.summary as label prefix (e.g. 'Pending: Refill calcium nitrate')"
    status: completed
  - id: ws2-pending-badge
    content: "WS2: optional small 'pending' chip on session row when proposal still pending for that session"
    status: completed
  - id: ws3-backend-optional
    content: "WS3: (optional) extend GET /v1/chat/sessions list with pending_proposal_summary per session — only if client-side join is too heavy"
    status: cancelled
  - id: ws4-tests-docs
    content: "WS4: phase-197-closure.test.js + guardian-panel/session list tests"
    status: completed
isProject: false
---

# Phase 197 — Session sidebar labels for multi-turn pending

**Status:** shipped · **Depends on:** [194](phase_194_pending_view_conversation.plan.md)

## The problem

After `scenario-task-dialogue-pending`, session `a28a9684…` has **4 turns**
in DB but the sidebar may show:

- Empty or generic title
- Hard to distinguish from other single-turn eval sessions

Operators jumping to **Chat** tab (via Refine or View conversation) need to
recognize which session owns the pending card.

## What to ship

### WS1 — Label helper

In `GuardianChatPanel.vue` `sessionLabel(s)`:

1. If `guardianProposalsStore` has pending proposal with `session_id === s.session_id`,
   use `Pending: ${proposal.summary}` (trimmed).
2. Else existing logic (title, first user message snippet, etc.).

Refresh proposals when loading sessions list.

### WS2 — Pending chip

Reuse topic-chip styling from Phase 63:

```html
<span class="...">pending</span>
```

On session row when linked proposal is `status=pending`.

### WS3 — API (optional)

Only if client join fails:

- Extend session list response with `pending_proposal_summary` nullable field.

Prefer client-side join from already-fetched `GET /v1/chat/proposals`.

### WS4 — Tests

- Session with matching pending proposal shows summary in label
- Session without pending uses normal label

## Acceptance criteria

- [x] Eval session identifiable in sidebar without opening each session
- [x] Turn count still shows (e.g. `4 turns`)
- [x] Selecting session loads full transcript

## Out of scope

- Renaming sessions automatically in DB (title column update)
