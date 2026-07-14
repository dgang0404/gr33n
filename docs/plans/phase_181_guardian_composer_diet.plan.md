---
name: Phase 181 — Guardian composer diet + single "Ask gr33n" entry
overview: >
  Beyond the three busy indicators (Phase 179), the full-page chat composer
  stacks vision attach, prompt starters, mode cards, and the usage strip
  every time — collapse to progressive disclosure. Also resolve "Ask gr33n"
  appearing with a live badge in both the sidebar and the top bar.
todos:
  - id: ws1-audit-entries
    content: "WS1: Inventory composer chrome (vision attach, starters, mode cards, usage strip) in full-page chat vs drawer; inventory both 'Ask gr33n' nav entries + badge source"
    status: completed
  - id: ws2-progressive-disclosure
    content: "WS2: Full-page chat — starters + vision attach + usage strip collapse behind a compact toggle/disclosure once session has ≥1 turn; drawer keeps richer always-on layout"
    status: completed
  - id: ws3-single-primary-entry
    content: "WS3: Pick one primary 'Ask gr33n' entry point (sidebar or top bar); demote the other to icon-only, no duplicate badge count"
    status: completed
  - id: ws4-tests-docs
    content: "WS4: Vitest for disclosure collapse + single-badge assertion; operator-tour nav update; phase-181-closure"
    status: completed
isProject: false
---

# Phase 181 — Guardian composer diet + single "Ask gr33n" entry

**Status:** shipped · **Follows:** [180](phase_180_knowledge_surfaces_discoverability.plan.md)

## The problem

Operator feedback (sit-in, 2026-07-13), on top of the three-status-indicator
issue already tracked in [Phase 179](phase_179_guardian_chat_status_consolidation.plan.md):

> *"Guardian composer diet — beyond status: collapse vision attach +
> starters + mode cards + usage strip into progressive disclosure on
> full-page chat (drawer can stay richer). 'Ask gr33n' in two places —
> sidebar + top bar both show badge 4 — fine if intentional; if not, pick
> one primary entry and demote the other to icon-only."*

The full-page `/chat` composer renders every affordance at once regardless
of session state, and the nav shows the same unread/pending badge count
twice — most operators read that as a bug, not a feature.

## North star

> Full-page chat composer starts minimal (input + send) and reveals
> starters/attach/usage on demand or after the first turn. Exactly one nav
> entry carries the "Ask gr33n" badge; the other, if kept, is a plain icon
> link with no count.

## Workstreams

### WS1 — Audit
List every composer element and its current always-visible vs conditional
state in `GuardianChatPanel.vue` / `FarmGuardianChat.vue` (full page) vs the
drawer variant. Separately, find both "Ask gr33n" render sites (sidebar nav
item + top bar item) and confirm they read the same badge count source
(likely unread proposals / unread chat count) — decide if duplication is
intentional (quick access) or accidental copy-paste.

### WS2 — Progressive disclosure on full-page chat
- Empty session: show starters + mode cards (first-run education, matches
  Phase 179 WS5 mode-card collapse).
- After first turn: starters and vision-attach controls collapse behind a
  small "+" / disclosure toggle next to the input; usage strip shrinks to a
  one-line summary (tap to expand for full token/cost breakdown).
- Drawer (`GuardianChatDrawer`-style, if present) keeps the current richer
  always-expanded layout — this is a full-page-only diet.

### WS3 — Single primary "Ask gr33n" entry
- Decide primary surface (recommend: whichever is visible from every
  authenticated route — likely top bar).
- Demote the other to an icon-only link, no badge, or remove it if truly
  redundant.
- Badge count logic stays centralized (one composable/store selector) so
  there is only one source of truth regardless of how many render sites
  remain.

### WS4 — Tests + docs
- Vitest: full-page chat composer shows only input+send on empty session
  init; starters appear on toggle; usage strip is collapsed by default
  after turn 1.
- Vitest: exactly one nav element renders a non-zero badge for the same
  underlying count.
- `operator-tour.md` nav section updated.
- `phase-181-closure.test.js`.

## Non-goals

- No change to what triggers the badge count (unread proposals/chat) —
  purely presentational dedupe.
- No new composer features (voice, slash commands, etc.).

## Acceptance

- [x] Full-page chat: starters/attach/usage strip are collapsed by default
      once a session has a turn; reachable via one toggle.
- [x] Drawer chat unchanged (richer, always-on).
- [x] Only one "Ask gr33n" entry shows the live badge; the other is
      icon-only or removed.
- [x] Vitest closure green.
