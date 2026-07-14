---
name: Phase 180 — Knowledge surfaces discoverability
overview: >
  Field guides, the symptom guide, Help→Knowledge semantic search, and the
  commons Catalog are four related-but-different surfaces that the UI never
  explains or cross-links. Symptom guide has no nav entry at all; Knowledge
  search looks like an exact-match developer tool; symptom filters are
  free-text exact fields. Make the knowledge layer browsable, self-explaining,
  and forgiving.
todos:
  - id: ws1-help-map
    content: "WS1: Help landing 'What lives where' map — Guide / Knowledge / Catalog / Symptom guide one-liners with links"
    status: pending
  - id: ws2-symptom-nav
    content: "WS2: Symptom guide gets a Help tab (or card link) + crop/category dropdowns fed from distinct values; keep deep-link query params"
    status: pending
  - id: ws3-knowledge-simplify
    content: "WS3: Knowledge search — semantic hint copy, big single search box, advanced filters (module/since/until/limit) behind a disclosure"
    status: pending
  - id: ws4-field-guide-browse
    content: "WS4: Browsable field-guide list (title + crop + ingested-at from manifest/chunks) so citations aren't the only door in"
    status: pending
  - id: ws5-citation-roundtrip
    content: "WS5: Citation chips land on a readable doc view (not raw chunk filter); back-link 'Ask Guardian about this' returns to chat"
    status: pending
  - id: ws6-tests-docs
    content: "WS6: Vitest for nav + dropdowns + disclosure; operator-tour §Help update; phase-180-closure"
    status: pending
isProject: false
---

# Phase 180 — Knowledge surfaces discoverability

**Status:** planned · **Follows:** [179](phase_179_guardian_chat_status_consolidation.plan.md)

## The problem

Operator feedback (sit-in, 2026-07-13), after Guardian cited *Field guide
#11* and linked to `/symptom-guide?crop_key=lettuce`:

> *"could I get there from navigating the ui or only from the guardian? what
> this knowledge base and catalog do they go along with the field guide or
> they something different? searching looks like it is just text box have to
> get the exactly right or will not find what user is looking for"*

All four complaints are accurate:

| Surface | Route | Today's reality |
|---------|-------|-----------------|
| **Field guides** | none (RAG only) | Curated crop-care docs in `docs/field-guides/`, ingested to `rag_embedding_chunks`. Cited by Guardian; **no browsable list**. |
| **Symptom guide** | `/symptom-guide` | Route exists, **zero nav entries** — reachable only from citation links or typed URL. Filters are free-text `crop_key` / `category` exact-match boxes. |
| **Help → Knowledge** | `/operator-guide?tab=knowledge` | Semantic (embedding) search — *not* exact-match — but the form (module filter, RFC3339 since/until, limit) reads like an internal debug tool, so operators assume keyword-exact. |
| **Help → Catalog** | `/operator-guide?tab=catalog` | Insert Commons import packs (recipes, seed packs). Unrelated to search; nothing says so. |

Nothing on any of these pages says how they relate, and the only reliable
path between them is a Guardian citation.

## North star

> From **Help**, an operator can see in one glance what each knowledge
> surface is for, browse field guides and symptoms without Guardian, and
> search in plain language without knowing module names or timestamp
> formats. Guardian citations remain a shortcut — never the only door.

## Workstreams

### WS1 — "What lives where" map on Help landing
Top of the **Guide** tab (or a strip above the tab bar): four cards —
Guide (how-to), Knowledge (search your farm + indexed docs), Catalog
(importable packs), Symptom guide (crop symptom lookup) — one sentence each,
linking to the tab/page. Kill the mystery for new operators.

### WS2 — Symptom guide becomes a first-class page
- Add a **Symptoms** tab to the Help workspace (`workspaces.js`) or, minimum,
  a linked card in WS1's map (decide during build; tab preferred).
- Replace free-text `crop_key` / `category` inputs with **dropdowns**
  populated from `GET /commons/agronomy-symptoms` distinct values (or a
  small `/commons/agronomy-symptom-filters` endpoint if cheaper).
- Case-insensitive matching server-side regardless.
- Keep `?crop_key=&category=` deep links working (Guardian citations rely
  on them).

### WS3 — Knowledge search that reads as semantic
- Hero copy: "Ask in plain language — search is by meaning, not exact
  words."
- One large search box + Search/Ask buttons; **module filter, since/until,
  limit collapse behind an "Advanced" disclosure** (defaults unchanged).
- Empty-state examples ("wilting in flower room", "when did feed volume
  change").

### WS4 — Browsable field-guide list
Field guides exist only as chunks today. Surface a simple list — title,
crop, last-ingested — sourced from the ingest manifest
(`docs/rag/field-guide-manifest.yaml`) or distinct `doc_path` values in
`rag_embedding_chunks`. Clicking one opens WS5's doc view (or, v1, a
pre-filtered Knowledge search for that doc). Lives as a section inside the
Knowledge tab.

### WS5 — Citation round-trip polish
- "Open Field guide #N" lands on a readable doc-scoped view: doc title,
  matched chunk highlighted, sibling chunks in order — not a bare filter
  form with a yellow banner.
- Doc view offers **"Ask Guardian about this"** (prefills chat with the doc
  as context ref) so the loop closes both directions.
- Symptom citations keep landing on `/symptom-guide` with filters applied
  (now dropdown-selected, WS2).

### WS6 — Tests + docs
- Vitest: Help map renders 4 surface cards; Symptoms tab routes; dropdowns
  populate from mocked API; Knowledge advanced fields hidden by default.
- `operator-tour.md` — rewrite the Help section around the four surfaces.
- `phase-180-closure.test.js`.

## Out of scope

- New embedding models or retrieval changes (search quality itself is fine).
- Editing/uploading field guides from the UI (ingest stays CLI:
  `make rag-ingest-field-guides`).
- Commons federation changes — Catalog import flows untouched.

## Acceptance

- [ ] Symptom guide reachable from Help without a Guardian citation.
- [ ] Symptom filters are dropdowns; wrong-case input can no longer zero-out results.
- [ ] Knowledge tab: advanced filters hidden until expanded; semantic hint visible.
- [ ] Field guides listed and openable without searching.
- [ ] Citation chip → doc view → "Ask Guardian" round-trip works.
- [ ] Help landing explains all four surfaces in one screen.
