---
name: Phase 194 — Pending proposal "View conversation"
overview: >
  Pending tab shows proposal cards only — no chat transcript. Operators
  cannot see the multi-turn refine dialogue that produced revision N without
  knowing to click Refine (which prefills a correction) or hunt the session
  in the Chat sidebar. Proposals carry session_id; conversation_turns has
  the full history.
todos:
  - id: ws1-view-conversation-button
    content: "WS1: GuardianActionProposal — add View conversation button when proposal.session_id present; data-test guardian-proposal-view-conversation"
    status: completed
  - id: ws2-wire-full-page-and-drawer
    content: "WS2: onViewConversation — FarmGuardianChat switches to Chat tab + loads session; GuardianDrawer same via guardianPanel (reuse requestRefine session load without prefill, or new requestViewConversation action)"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: Vitest on proposal card + FarmGuardianChat tab switch; phase-194-closure.test.js; operator-tour Pending section"
    status: completed
isProject: false
---

# Phase 194 — Pending proposal "View conversation"

**Status:** shipped · **Depends on:** [184](phase_184_guardian_pr_conversation_smoke.plan.md)

## The problem

User on `/chat?tab=pending` sees:

- Summary: `Create task: due tomorrow`
- Args diff, Confirm / Refine / Dismiss

They **do not** see the four operator turns that built revision 4. The data
exists (`GET /v1/chat/sessions/{session_id}` returns all turns) but there is
no affordance on the Pending card.

**Refine today** switches to Chat and loads the session *and* prefills a
correction message — not the same as "just show me what we said."

## What to ship

### WS1 — Button on proposal card

In `GuardianActionProposal.vue` (pending state only):

- **View conversation** — secondary/ghost style, left of Refine or below summary
- Disabled when `!proposal.session_id` or expired
- `aria-label`: "View chat history for this change request"

### WS2 — Navigation wiring

**`guardianPanel.js`** — new action:

```js
requestViewConversation(proposal) {
  if (proposal?.session_id) this.activeSessionId = proposal.session_id
  this.prefilledMessage = ''  // do not prefill
  this.drawerTab = 'chat'
  this.viewConversationTick += 1  // mirror refineTick pattern
}
```

**`FarmGuardianChat.vue`** — `onViewConversation` → `activeTab = 'chat'`

**`GuardianChatPanel.vue`** — watch `viewConversationTick`, `loadSession(sid)`

Works in both full-page `/chat` and slide-out drawer Pending tab.

### WS3 — Tests & docs

- Component test: button visible when `session_id` set
- Click emits / calls panel store
- `phase-194-closure.test.js`
- Operator tour: "Pending → View conversation opens the linked chat thread"

## Acceptance criteria

- [x] From Pending tab, one click shows full transcript in Chat (all turns for that session)
- [x] Composer is empty (no Refine prefill)
- [x] Session sidebar highlights the linked session
- [x] Works for eval-generated `scenario-task-dialogue-pending` proposal

## Out of scope

- Embedding transcript inline on Pending card (see [196](phase_196_pending_revision_timeline.plan.md))
