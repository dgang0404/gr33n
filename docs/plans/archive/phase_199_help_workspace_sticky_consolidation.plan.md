---
name: Phase 199 — Consolidate Help workspace sticky chrome
overview: >
  Phases 188 and 193 fix opacity on individual sticky bars, but Help → Library
  still stacks two independent sticky headers (WorkspaceShell subnav + Library
  section pills) with a magic top-[7.5rem] offset. This phase consolidates
  layout so one sticky chrome region handles tabs + Jump to + Library section
  pills with correct stacking and no gap bleed.
todos:
  - id: ws1-audit-sticky-heights
    content: "WS1: measure actual heights — WorkspaceShell header (scrolls away), subnav (tabs+jump), HelpLibraryHub pills; document scroll-mt-* on sections"
    status: completed
  - id: ws2-single-sticky-region
    content: "WS2: move Library section pills into WorkspaceShell slot OR shared HelpStickyChrome component; one sticky top-0 container with internal rows"
    status: completed
  - id: ws3-offset-scroll-margin
    content: "WS3: fix scroll-mt-36 / scrollIntoView for section deep links after chrome height changes"
    status: completed
  - id: ws4-tests-docs
    content: "WS4: phase-199-closure.test.js + visual regression notes in operator-tour"
    status: completed
isProject: false
---

# Phase 199 — Consolidate Help workspace stickies

**Status:** shipped · **Depends on:** [193](phase_193_help_library_sticky_bleed.plan.md)

## The problem

Current Help → Library scroll stack:

```
[scrolls away] WorkspaceShell header (Help title)
[sticky top-0]   WorkspaceShell subnav — Library | Pi+HAT, Jump to
[scrolls]        HelpKnowledgeSurfacesMap ("What lives where")
[sticky top-7.5rem] HelpLibraryHub pills — How-to | Search | Symptoms | Import
[scrolls]        Section content (field guides, symptoms, …)
```

Issues:

1. **Two sticky roots** — content can slip between them
2. **`top-[7.5rem]`** — brittle; breaks if subnav height changes
3. **Double borders** — visual seam between chrome layers

Phase 193 makes backgrounds opaque but does not fix architecture.

## What to ship

### WS1 — Layout audit

Measure rendered heights at `sm` and mobile breakpoints. Map which elements
should scroll away vs stick.

### WS2 — Unified chrome (preferred approach)

**Option A (recommended):** Extend `WorkspaceShell` with optional
`#subnav-extra` slot — Help workspace injects Library pills as second row
inside `workspace-shell__subnav`:

```
┌ Library | Pi+HAT setup ─────────────────┐
│ Jump to: Zones · Money · Feed & water   │
│ How-to · Search · Symptoms · Import    │  ← all one sticky block
└─────────────────────────────────────────┘
```

Remove separate sticky from `HelpLibraryHub.vue`.

**Option B:** Lift pills to `HelpWorkspace.vue` between shell and hub content.

### WS3 — Scroll targets

Update `scroll-mt-*` on `#help-section-*` to match unified chrome height.
Re-test `?section=symptoms` deep links and pill click scroll.

### WS4 — Tests

- Only one `sticky top-0` chrome in Help Library route
- Section jump still lands correctly

## Acceptance criteria

- [x] No content visible between/behind sticky rows when scrolling
- [x] Library / Pi+HAT tab switch still works
- [x] How-to / Search / Symptoms / Import jumps still work
- [x] Mobile layout acceptable (stacked rows or select)

## Out of scope

- Changing Help content structure (What lives where card)
- Guardian chat stickies (separate components)
