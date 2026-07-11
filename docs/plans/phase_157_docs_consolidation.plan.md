---
name: Phase 157 — docs consolidation
overview: >
  Add a single "what gr33n looks like today" snapshot and archive closed phase
  plans so the live index shrinks without losing the 150+ phase history agents
  and operators rely on for context.
todos:
  - id: ws1-current-state
    content: "WS1: docs/current-state.md — features, routes, schemas, Guardian modes, at a glance"
    status: pending
  - id: ws2-archive-folder
    content: "WS2: docs/plans/archive/ — move fully-closed phase plans; leave stub links in index"
    status: pending
  - id: ws3-index-trim
    content: "WS3: Trim phase-14-operator-documentation.md — 'active' table vs 'archived phases' link"
    status: pending
  - id: ws4-regen-hint
    content: "WS4: Makefile or script hint to regenerate current-state sections from openapi.yaml + README"
    status: pending
isProject: false
---

# Phase 157 — docs consolidation

**Status:** planned · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

---

## Why this phase

150+ phase-plan docs are a **genuine asset** — they let an agent or new contributor reconstruct exact intent and closure criteria. But:

- [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) is 240+ lines of links and still growing
- There is **no** single doc that answers "what does gr33n look like *right now*?" without reading README + INSTALL + architecture + phase index
- Closed phases (e.g. 35–67 arc, 84–110 master roadmap) sit beside active work in the same folder

New contributors and Composer sessions pay a **"which of these 150 docs is still true?"** tax on every cold start.

---

## Workstreams

### WS1 — Current-state snapshot

**Target:** [`docs/current-state.md`](../current-state.md)

One scannable page, regenerated when major phases ship (not on every commit). Suggested sections:

| Section | Source of truth |
|---------|-----------------|
| **What it is** | README lede + AGPL/self-hosted positioning |
| **Shipped feature list** | README "What You Can Do" (trimmed) |
| **Workspaces & routes** | `ui/src/router` + operator-tour anchors |
| **API surface** | `openapi.yaml` tag list + `/v1/chat` Guardian endpoints |
| **Postgres schemas** | `db/migrations` schema list (`gr33ncore`, `gr33nfertigation`, …) |
| **Guardian** | Farm Counsel vs Quick Chat, proposals queue, smoke targets (`make guardian-qa-smoke`) |
| **Edge / Pi** | Virtual Pi, MQTT bridge, command queue vs `pending_command` |
| **Env knobs operators touch** | Link to `environment-variables.md` top 20 |
| **What's explicitly not shipped** | Phase 115 stubs, optional commons, etc. |

**Header block:**

```markdown
> Generated: YYYY-MM-DD · Regenerate after major phase ship · Canonical history: phase-14-operator-documentation.md
```

### WS2 — Archive folder

**Target:** `docs/plans/archive/`

Move plans whose **close-when** boxes are all checked and that appear in a closure rollup (e.g. [`phase-84-110-closure.md`](phase-84-110-closure.md), [`phase-129-139-closure.md`](phase-129-139-closure.md)).

**Rules:**

- Move file, don't delete
- Leave a one-line stub in the old path: `Moved to archive/phase_NN_….plan.md — see phase-84-110-closure.md`
- Or use git mv and fix links in one pass with `rg 'phase_NN_' docs/`

**Do not archive:** active Guardian quality arc (143+), infra arc (154–158), or any plan with open todos.

### WS3 — Index trim

Restructure [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md):

1. **Start here** — link `current-state.md`, operator tour, first session
2. **Active phases** — table of unshipped / recent (last ~20 rows)
3. **Shipped arcs** — one row per arc hub (68–81, 84–110, 129–139, 143–154) linking closure docs
4. **Archive** — `docs/plans/archive/README.md` index

### WS4 — Regeneration hint (optional)

Lightweight `make docs-current-state-hint` that prints:

- OpenAPI path count
- Latest migration filename
- Guardian smoke suite list from `guardian-eval -manual`

Full auto-generation is **non-goal** for v1 — human curates prose; script supplies numbers.

---

## Acceptance

- [ ] `docs/current-state.md` exists and is linked from README + phase-14 index
- [ ] At least one closed arc (e.g. phases 88–92 platform data gaps) lives under `docs/plans/archive/` with working redirects/stubs
- [ ] `phase-14-operator-documentation.md` active section is ≤ half its current phase-table row count
- [ ] [`INSTALL.md`](../../INSTALL.md) "Start here" points to `current-state.md` for "what's in the box"

## Non-goals

- Deleting phase history
- Auto-generating operator-tour or architecture doc bodies
- Wikifying every closure test into prose

## Operator path (after ship)

New clone: **README → current-state.md → operator-tour.md → first-session-after-clone.md**
