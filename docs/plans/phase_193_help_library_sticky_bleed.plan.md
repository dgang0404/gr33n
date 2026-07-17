---
name: Phase 193 — Help Library sticky nav bleed (opaque backgrounds)
overview: >
  Operator sit-in screenshot on Help → Library shows field-guide markdown
  table rows ("| Sensor reads nothing |", "| Feed did not run |") bleeding
  through the sticky navigation chrome. Phase 188 fixed WorkspaceShell
  subnav opacity, but HelpLibraryHub's second sticky row (How-to / Search /
  Symptoms / Import) still uses bg-zinc-950/95 + backdrop-blur.
todos:
  - id: ws1-library-pills-opaque
    content: "WS1: HelpLibraryHub jump nav — replace bg-zinc-950/95 backdrop-blur with solid bg-zinc-950 (match WorkspaceShell fix)"
    status: completed
  - id: ws2-workspace-subnav-verify
    content: "WS2: confirm WorkspaceShell subnav stays opaque; add regression in phase-193-closure if not already covered by phase-188"
    status: completed
  - id: ws3-tests-docs
    content: "WS3: phase-193-closure.test.js + operator-tour note under Help Library"
    status: completed
isProject: false
---

# Phase 193 — Help Library sticky nav bleed

**Status:** shipped · **Depends on:** [188](phase_188_guardian_answer_quality_audit.plan.md) (partial fix) · **Superseded in part by:** [199](phase_199_help_workspace_sticky_consolidation.plan.md)

## The problem

Help → Library has **two sticky layers** inside `#main-content`:

1. `WorkspaceShell` `workspace-shell__subnav` — Library / Pi+HAT + Jump to
2. `HelpLibraryHub` pill nav — How-to / Search / Symptoms / Import

Layer 2 still uses semi-transparent styling:

```vue
class="sticky top-[7.5rem] z-10 ... bg-zinc-950/95 backdrop-blur ..."
```

When the How-to / field-guide content scrolls, monospace table text shows
through the pill bar (and sometimes through layer 1 if dev server not restarted).

## What to ship

### WS1 — Opaque Library pills

In `ui/src/views/HelpLibraryHub.vue`:

- `bg-zinc-950` solid background, remove `backdrop-blur`
- Keep `border-b border-zinc-800/80` for separation

### WS2 — WorkspaceShell regression

Verify `ui/src/components/WorkspaceShell.vue` has no `/95` or `backdrop-blur`
on `workspace-shell__subnav`.

### WS3 — Tests & docs

- `phase-193-closure.test.js` asserts HelpLibraryHub has solid bg, no blur
- Short note in `docs/operator-tour.md` §7m Help Library

## Acceptance criteria

- [ ] Scroll Help → Library through How-to field guides — no table text visible behind How-to/Search/Symptoms/Import bar
- [ ] No visual regression on mobile section jump nav

## Out of scope (Phase 199)

Sticky offset consolidation (`top-[7.5rem]` magic number) and merging the
two sticky bars into one chrome — deferred to Phase 199.
