---
name: Phase 179 — Guardian chat status consolidation
overview: >
  One answering turn currently surfaces three separate "Guardian is thinking"
  indicators (transcript status line, awakening panel box, composer block
  notice). Consolidate to a single progress affordance anchored to the
  transcript, demote the rest to disabled-state styling, and declutter the
  composer while a turn is in flight.
todos:
  - id: ws1-status-inventory
    content: "WS1: Inventory + test map — every streaming/busy indicator in GuardianChatPanel, GuardianAwakeningPanel, chatUsage strip, mode cards"
    status: completed
  - id: ws2-single-progress
    content: "WS2: Single progress row — streamingStatus + elapsed timer live only in the transcript streaming row; Stop stays beside it"
    status: completed
  - id: ws3-composer-quiet
    content: "WS3: Composer during streaming — disable Send + input with subtle hint; remove amber groundedModelBlockReason duplicate when the busy turn is OURS"
    status: completed
  - id: ws4-awakening-scope
    content: "WS4: Awakening panel shows readiness states only (dormant/stirring/awake); hide its busy box while this session's turn is streaming"
    status: completed
  - id: ws5-mode-cards-collapse
    content: "WS5: Collapse Quick chat / Farm counsel cards to a compact segmented control once a session has ≥1 turn (full cards only on empty session)"
    status: completed
  - id: ws6-tests-docs
    content: "WS6: Vitest — exactly one visible status element while streaming; update operator-tour chat section; phase-179-closure"
    status: completed
isProject: false
---

# Phase 179 — Guardian chat status consolidation

**Status:** shipped · **Follows:** [178](phase_178_online_weather_forecast.plan.md)

## The problem

During a single grounded turn the chat screen shows **three concurrent
"Guardian is busy" messages**, each added by a different phase for a
different reason:

| # | Element | Source | Copy |
|---|---------|--------|------|
| 1 | Transcript streaming row | SSE `status` event → `guardianChat.streamingStatus` | "Composing answer — running on CPU (no GPU). Grounded turns may take several minutes." |
| 2 | Awakening panel box | `guardianReadiness.awakening.state === 'busy'` | "Guardian is answering…" |
| 3 | Composer amber notice | `groundedModelBlockReason` in `GuardianChatPanel.vue` | "Guardian is answering — wait for the current reply before sending another farm counsel message." |

Plus the Stop button and disabled inputs. Individually each is honest;
together they read as noise and push the composer far down the page.
Operator feedback (sit-in, 2026-07-13): *"guardian is thinking is shown in
3 different spots … this screen has a lot going on."*

## North star

> While Guardian answers, the operator sees **one** live progress line — in
> the transcript, where the answer will appear — plus a Stop button. The
> composer quietly disables. Nothing else animates, warns, or repeats the
> same message.

## Non-goals

- No change to SSE protocol, `status` events, or readiness API.
- No change to the cross-session busy guard (a *different* user/session
  hitting `chat_busy` still gets the explicit error).
- Keep the awakening panel for its real job: dormant / stirring / awake
  transitions before a turn starts.

## Workstreams

### WS1 — Inventory + test map
Catalog every element that reacts to `streaming` / `busy` in
`GuardianChatPanel.vue`, `GuardianAwakeningPanel.vue`, `GuardianTabNav.vue`
(spinner), and the chat-usage strip. Record which Vitest files pin each so
WS2–WS5 don't break closure tests blindly.

### WS2 — Single progress row
The transcript streaming row becomes the only live status: guardian label,
`streamingStatus` line, streamed text, blinking caret, and Stop. Add a small
elapsed-time counter (mm:ss) so long CPU turns feel accounted for rather
than hung.

### WS3 — Composer quiet mode
While `streaming === true` for **this session**: textarea + Send disabled
with `title` hint only; suppress the amber `groundedModelBlockReason`
duplicate (it repeats indicator #1's message). The block reason still
renders when the busy state belongs to *another* session (server
`chat_busy` case) since then there is no local streaming row.

### WS4 — Awakening panel scope
`GuardianAwakeningPanel` shows dormant / stirring / sleeping states and the
stray-session warning. When `state === 'busy'` **and** the local chat store
is streaming, render nothing (the streaming row owns it). Busy without a
local stream (another tab/user) keeps the current box.

### WS5 — Mode cards collapse
Quick chat / Farm counsel selection cards are large and repeat cost hints on
every turn. Once the session has ≥1 completed turn, collapse to a compact
two-segment control beside the input; full cards render only on an empty
session (first-run education stays).

### WS6 — Tests + docs
- Vitest: with `streaming=true` exactly **one** element matching
  `/answering|composing/i` is visible in the panel.
- Vitest: `chat_busy` from another session still shows the composer notice.
- Update `operator-tour.md` chat walkthrough screenshots/copy.
- `phase-179-closure.test.js`.

## Acceptance

- [x] One visible busy/status element during a local streaming turn.
- [x] Stop button unchanged and adjacent to the progress row.
- [x] Cross-session busy (`chat_busy`) still explains itself in the composer.
- [x] Mode cards collapse after first turn; full cards on empty session.
- [x] All existing chat Vitest suites green; closure test added.
