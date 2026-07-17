---
name: Phase 201 — Help knowledge surfaces unification
overview: >
  Phase 199 unified Help sticky chrome; Phase 201 (partial) moved Symptom guide
  to its own Help tab and removed the embedded duplicate from HelpLibraryHub.
  This phase finishes the job: Knowledge (search + field guides) and Catalog
  (Commons import) become first-class Help tabs — same UX as Symptom guide —
  with one component surface each, legacy redirects preserved, and Guardian
  citation routes aligned.
todos:
  - id: ws1-help-tabs
    content: "WS1: add Help workspace tabs `knowledge` and `catalog` (or `import`); remove embedded FarmKnowledge + CommonsCatalog from HelpLibraryHub scroll stack"
    status: pending
  - id: ws2-redirects-routes
    content: "WS2: remove shadow standalone routes for /farm-knowledge and /catalog from router/index.js; rely on absorbs → /operator-guide?tab=…; keep query params (cited_doc, crop_key, section legacy)"
    status: pending
  - id: ws3-surfaces-map-nav
    content: "WS3: update HelpKnowledgeSurfacesMap + What lives where copy; Library section pills become How-to | Search | Import only (or drop pills if tabs replace them); legacy ?section=knowledge|catalog → tab redirect in HelpWorkspace"
    status: pending
  - id: ws4-guardian-citations
    content: "WS4: align internal/farmguardian/citation_route.go landing URLs with Help tabs (field_guide → knowledge tab, platform_doc → library/how-to); verify Symptom guide tab links unchanged"
    status: pending
  - id: ws5-tests-docs
    content: "WS5: update phase-180/183/199 closure tests; operator-tour Help §7m; current-state.md one-liner"
    status: pending
isProject: false
---

# Phase 201 — Help knowledge surfaces unification

**Status:** planned (Symptom tab shipped in-session 2026-07-17) · **Depends on:** [199](phase_199_help_workspace_sticky_consolidation.plan.md), [180](phase_180_knowledge_surfaces_discoverability.plan.md)

## The problem

Help → Library is still a **long scroll** with four embedded apps:

| Surface | Current entry | Duplicate? |
|---------|---------------|--------------|
| How-to | Library scroll + pills | OperatorGuide only here — OK |
| Knowledge | Library scroll + `/farm-knowledge` standalone route | **Yes** — same FarmKnowledge.vue |
| Symptom guide | Help tab + `/symptom-guide` redirect | **Fixed** — one tab, redirect absorbs |
| Catalog | Library scroll + `/catalog` standalone route | **Yes** — same CommonsCatalog.vue |

Router registers `/farm-knowledge` and `/catalog` **before** `buildLegacyRedirectRoutes()`, so absorbs never run — operators and tests hit two URLs for the same UI.

## What to ship

### WS1 — Help tabs

Extend `WORKSPACES.help.tabs`:

```
Library | Pi + HAT setup | Search | Symptom guide | Import
```

(or shorter labels: Knowledge / Catalog — match farmer vocabulary)

`HelpWorkspace.vue` renders:

- `library` → HelpLibraryHub (OperatorGuide + surfaces map only)
- `knowledge` → FarmKnowledge (embedded, no duplicate header if unified chrome suffices)
- `symptoms` → SymptomGuide (already shipped)
- `catalog` → CommonsCatalog

Remove `#help-section-knowledge` and `#help-section-catalog` from HelpLibraryHub.

### WS2 — Redirects

Delete standalone routes in `ui/src/router/index.js` for `/farm-knowledge` and `/catalog`.

Update absorbs:

```js
'/farm-knowledge': { tab: 'knowledge' },
'/catalog': { tab: 'catalog' },
```

Preserve query params on redirect (`cited_doc`, `cited_chunk`, `cited_type`).

### WS3 — Navigation chrome

- **What lives where** cards → tab deep links (not `section=` scroll)
- **Library pills** — drop Search/Import if tabs replace them; keep How-to | Search-within-library only if How-to stays scroll-only
- **HelpLibrarySectionNav** — only How-to jumps inside Library tab

### WS4 — Guardian citations

Audit `citation_route.go` + UI citation links so field guides land on `?tab=knowledge&cited_doc=…`, not orphan paths.

### WS5 — Tests & docs

- No `HelpLibraryHub` assertions for knowledge/catalog sections
- Redirect tests: `/farm-knowledge` → operator-guide?tab=knowledge
- operator-tour: single Help story — tabs not scroll

## Acceptance criteria

- [ ] One FarmKnowledge surface (Help tab only; redirect from legacy URL)
- [ ] One CommonsCatalog surface (Help tab only; redirect from legacy URL)
- [ ] Library hub is shorter — how-to + surfaces map, not four apps stacked
- [ ] Guardian citations and SymptomCropLink still resolve correctly
- [ ] phase-201-closure.test.js

## Out of scope

- Merging FarmKnowledge search UI with Symptom guide (different data sources)
- RAG ingest / backend changes
- Removing OperatorGuide embedded glossary ( stays in Library )

## Ponytail note

Prefer **delete routes + move tab** over new wrapper components. Reuse `embedded` prop on existing views.
