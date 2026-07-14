---
name: Phase 182 — Guardian quick UX wins
overview: >
  Three small, independent fixes surfaced from the same operator sit-in as
  Phases 179–181: the 401 nav-badge polling spam, Pending tab cards hidden
  below the fold, and no in-composer hint for how Refine expects follow-up
  text. Bundled as one phase since each is a small, self-contained diff.
todos:
  - id: ws1-401-backoff
    content: "WS1: On 401 from unread-count/proposals polling, stop polling + redirect to login once instead of hammering the endpoint every interval"
    status: pending
  - id: ws2-pending-tab-scroll
    content: "WS2: Pending tab — sticky 'N requests' count + scrollable list (newest first) so all N cards are reachable, not just the first 2-3"
    status: pending
  - id: ws3-refine-hint
    content: "WS3: One-line hint under the composer input during Refine: 'Type after Correction: or ask a question first — same session.'"
    status: pending
  - id: ws4-tests-docs
    content: "WS4: Vitest for 401 single-redirect + no re-poll; Pending tab scroll/order; refine hint visibility; phase-182-closure"
    status: pending
isProject: false
---

# Phase 182 — Guardian quick UX wins

**Status:** planned · **Follows:** [181](phase_181_guardian_composer_diet.plan.md)

## The problem

From the same 2026-07-13 sit-in, three small but real papercuts:

1. **401 polling spam** — once the browser JWT (`gr33n_token`) expires, the
   nav unread-count poll keeps hitting the API every interval and logging a
   fresh 401 to the console instead of stopping after the first failure.
2. **Pending tab overflow** — the tab label said "Pending (4)" but only
   three cards were visible without scrolling; the fourth (an `ack`
   proposal) was easy to miss during manual Confirm/Refine/Dismiss testing.
3. **Refine ambiguity** — after clicking Refine, it's unclear whether to
   type *after* the prefilled `Correction:` line, replace it, or ask a
   clarifying question first in the same session.

None of these need new endpoints — all are UI-only.

## Workstreams

### WS1 — 401 → stop polling, redirect once
Wherever the nav badge/unread-count poller lives (likely a `setInterval` in
a composable or store), catch 401 specifically: clear the interval, redirect
to `/login` (or show a "session expired" toast) once, and don't resume
polling until a fresh login. Distinguish from transient network errors
(those can keep their existing retry/backoff).

### WS2 — Pending tab scroll + count
- Tab label count stays live ("Pending (N)").
- List container becomes independently scrollable with a visible
  scrollbar/fade cue when it overflows the panel.
- Sort newest-first (most recently proposed on top) so fresh test proposals
  from a smoke run are immediately visible without scrolling.

### WS3 — Refine hint copy
Single line under the input, shown only while in an active Refine flow
(prefilled `Correction:` text present): *"Type after Correction: or ask a
question first — same session."* Dismisses once the user sends a message.

### WS4 — Tests + docs
- Vitest: simulated 401 stops the poll interval and navigates once (no
  repeated calls in fake-timer advance).
- Vitest: Pending list with 5+ mock proposals is scrollable and ordered
  newest-first; tab count matches list length.
- Vitest: Refine hint renders only when `Correction:` prefill is active.
- `phase-182-closure.test.js`.

## Non-goals

- No token refresh/silent-renew implementation — just stop polling on 401.
- No Pending tab redesign beyond scroll + order (bigger changes are Phase
  180/181 territory).

## Acceptance

- [ ] 401 on nav polling logs once, redirects once, does not repeat every
      interval.
- [ ] Pending tab: every pending card reachable via scroll; count matches
      visible list; newest-first order.
- [ ] Refine flow shows the one-line hint under the input.
- [ ] Vitest closure green.
