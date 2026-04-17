---
name: Phase 20.9b Terminology & Copy Pass
overview: >
  One focused pass over user-facing and API-facing text so we describe practices
  and modules respectfully and accurately: prefer the proper noun **JADAM** when
  referring to that method and its literature-backed inputs; use **natural farming**
  for generic descriptions of fermented inputs, soil drenches, and inventory.
  Avoid umbrella labels that tie a nationality to a whole practice (e.g. do not
  use "Korean Natural Farming" as the name of the API module). Database schema
  names (`gr33nnaturalfarming`) stay — this phase is copy and OpenAPI tags only.
todos:
  - id: ws1-openapi-and-readme
    content: "WS1: OpenAPI tag descriptions + README modularity line; grep for KNF/Korean in docs and UI"
    status: completed
  - id: ws2-ui-headings-and-helptips
    content: "WS2: Inventory / fertigation / farm onboarding UI strings — JADAM vs natural farming consistency; HelpTips where we explain the module"
    status: completed
  - id: ws3-seed-and-bootstrap-comments
    content: "WS3: Optional — SQL comment-only / seed description tweaks where 'Korean' appears; no data migration unless product asks for re-label of stored strings"
    status: completed
isProject: false
---

# Phase 20.9b — Terminology & copy pass

## Why this phase (product + respect)

- **JADAM** is a specific, named approach (and the name the starter pack already uses in data). Using it precisely is clearer for operators who read Cho Youngsang's materials.
- **Natural farming** (lowercase, generic) is a good umbrella in English for "fermented inputs, indigenous micro-organisms, soil drenches" without centering a country as the *definition* of the practice. That reduces the chance of sounding like we are stereotyping or claiming a whole nation's agriculture.
- Some audiences find **"Korean Natural Farming"** or **"KNF"** as a casual label for the whole module grating or imprecise. The codebase should not *require* that phrasing in the public API surface.

This phase does **not** rename SQL schemas or move tables — only human-readable strings, OpenAPI descriptions, and operator docs.

## Scope (explicit)

| In scope | Out of scope |
|----------|----------------|
| `openapi.yaml` tag `naturalfarming` description | Renaming `gr33nnaturalfarming` schema |
| README / workflow / playbook wording | Rewriting historical migration bodies unless we add a tiny follow-up comment migration |
| Vue headings, placeholders, HelpTips | Changing `jadam_indoor_photoperiod_v1` template key |

## Rationale note for docs

Short guideline to paste in `CONTRIBUTING.md` or `docs/` once (optional in this phase):

- Prefer **JADAM** when the content is specifically about that method, starter pack, or book-cited inputs.
- Prefer **natural farming** when describing the *category* of features (inputs, batches, recipes) for readers who do not use the JADAM label.
- Do not use **"Korean Natural Farming"** as the official product name for the module; do not use **KNF** alone in user-facing titles without expanding it once.

## After this phase

Operator-facing language is consistent with the above; OpenAPI-generated docs and client SDK descriptions read cleanly for international farms.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 20.9b per @docs/plans/phase_20_9b_terminology_and_copy_pass.plan.md.

Scope: grep for Korean|KNF in ui/, docs/, openapi.yaml, README; update user-facing strings per the plan's JADAM vs natural farming guideline; do not rename SQL schemas. Run npm run build and go test ./... when done.
```
