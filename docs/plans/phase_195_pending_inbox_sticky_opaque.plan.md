---
name: Phase 195 — Pending inbox sticky count bar opaque
overview: >
  GuardianRequestsInbox uses a sticky "N requests — newest first" header with
  bg-zinc-950/95 backdrop-blur-sm. When scrolling the proposal list on
  /chat?tab=pending, proposal card text can bleed through the count row —
  same class of bug as Phase 193/188 sticky bleed.
todos:
  - id: ws1-inbox-count-opaque
    content: "WS1: GuardianRequestsInbox count row — solid bg-zinc-950, remove backdrop-blur-sm and /95 opacity"
    status: completed
  - id: ws2-tests-docs
    content: "WS2: phase-195-closure.test.js + cross-ref in operator-tour Pending tab"
    status: completed
isProject: false
---

# Phase 195 — Pending inbox sticky count bar opaque

**Status:** shipped · **Depends on:** [193](phase_193_help_library_sticky_bleed.plan.md) (same pattern)

## The problem

```vue
<!-- GuardianRequestsInbox.vue -->
class="sticky top-0 z-10 ... bg-zinc-950/95 backdrop-blur-sm ..."
```

The count header sticks at `top-0` inside the Pending scroll area. Proposal
cards scrolling beneath it can show through the semi-transparent background.

## What to ship

### WS1 — Solid background

Replace with `bg-zinc-950 border-b border-zinc-800/80` (no blur, no `/95`).

Consider `z-20` if proposal cards overlap due to stacking context — match
WorkspaceShell subnav convention.

### WS2 — Tests

- `phase-195-closure.test.js` grep for solid bg on `guardian-inbox-count` row

## Acceptance criteria

- [x] Scroll Pending list with 1+ proposals — count bar fully obscures content beneath
- [x] No change to Confirm/Refine/Dismiss behavior

## Note

Farm Guardian drawer Pending tab uses the same `GuardianRequestsInbox`
component — fix applies everywhere.
