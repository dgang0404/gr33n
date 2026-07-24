---
name: Phase 211.02 — Recipe formula history & crop ops reporting
overview: >
  Immutable recipe revision snapshots so edits and component removes do not erase
  what was actually fed, watered, and lit during a crop cycle. Join program runs,
  mixing events, fertigation events, and lighting for AI-ready crop growth reports.
todos:
  - id: ws0-remove-switchover-ui
    content: "WS0 (211.01 follow-up): Remove Switchover guide tab and Jump to rail from Natural farming workspace"
    status: completed
  - id: ws1-revision-schema
    content: "WS1: application_recipe_revisions table + snapshot JSON (recipe row + recipe_input_components)"
    status: completed
  - id: ws2-revision-on-write
    content: "WS2: Create revision on recipe PATCH and component add/remove; programs pin revision_id optionally"
    status: completed
  - id: ws3-run-snapshots
    content: "WS3: Program tick + mixing event write revision_id / formula snapshot into automation_runs.details and mixing_events"
    status: completed
  - id: ws4-restore-ui
    content: "WS4: Recipes & apply — view revision history, restore components from revision (not blind pack re-import)"
    status: completed
  - id: ws5-crop-ops-report-api
    content: "WS5: GET crop-cycle ops timeline — feed, mix, light, stage events with formula at time"
    status: completed
  - id: ws6-guardian-read
    content: "WS6: Guardian read tool list_crop_cycle_ops + closure tests"
    status: completed
isProject: false
---

# Phase 211.02 — Recipe formula history & crop ops reporting

**Status:** Complete (WS0–WS6 shipped) · **Depends on:** [211.01](phase_211_01_nf_studio_declutter.plan.md) · **212 deferred** until happy main (UI declutter + 211.03 permissions)

## The one job

> When an operator removes FFJ from a recipe in March, a report for February still
> shows **FFJ 1:500 + WCA 0.5**. AI and Guardian can answer “what was this room
> getting?” from immutable history — not today’s live row.

## Why this matters (current gap)

| What exists today | What it records | Gap |
|-------------------|-----------------|-----|
| `application_recipes` | **Current** formula | Edits overwrite meaning of past program links |
| `recipe_input_components` | **Current** parts | DELETE is permanent; pack re-import only upserts listed rows |
| `programs.application_recipe_id` | Live recipe FK | No “which revision was active when program ran” |
| `automation_runs` (`program_id`) | **When** program fired | `details` has action_source, not formula |
| `mixing_events` + components | **Physical** tank mix | Good for hydro; `dilution_ratio` on component row helps |
| `fertigation_events` | Zone apply, `crop_cycle_id`, EC/pH | No `application_recipe_revision_id` |
| `crop_cycle_stage_events` | Stage transitions | Not joined to feed formula |
| `lighting_programs` + schedule runs | Photoperiod | Not joined to crop report API |

**Symptom:** “Crop grew better on the old formula” is unanswerable after a UI edit.

## Design (ponytail — one revision table, snapshot at write time)

### WS1 — `application_recipe_revisions`

```sql
CREATE TABLE gr33nnaturalfarming.application_recipe_revisions (
    id                      BIGSERIAL PRIMARY KEY,
    application_recipe_id   BIGINT NOT NULL REFERENCES ... ON DELETE CASCADE,
    revision_number         INTEGER NOT NULL,
    snapshot                JSONB NOT NULL,  -- { recipe: {...}, components: [...] }
    change_summary          TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id      UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    UNIQUE (application_recipe_id, revision_number)
);
```

`snapshot` is the audit truth — copy of dilution, stages, and component parts at commit time.

**Bootstrap:** one revision `revision_number = 1` per seeded recipe from `master_seed.sql` (migration backfill).

### WS2 — Revision on write

Create a new revision when:

- Recipe header updated (`PATCH /naturalfarming/recipes/{id}`)
- Component added or removed
- Optional: first revision on recipe create

Add nullable `application_recipe_revision_id` on `gr33nfertigation.programs`. New program links pin **latest** revision; existing programs keep working (NULL = resolve live recipe + log warning in worker until pinned).

**Not in scope:** versioning `input_definitions` or batch rows (batches already tracked by `input_batch_id` on mix components).

### WS3 — Run-time snapshots

On **program tick** (`runProgramTick`), before dispatch:

1. Resolve revision (program pin, else latest for recipe)
2. Append to `automation_runs.details`:

```json
{
  "application_recipe_id": 12,
  "application_recipe_revision_id": 45,
  "formula_snapshot": { "dilution_ratio": "...", "components": [...] }
}
```

On **mixing event create** (manual or mix job): store same `revision_id` on `mixing_events` (new column) or in `metadata` JSONB first (ponytail: column later if query-heavy).

On **fertigation_event** insert (when wired): copy revision snapshot into `metadata.formula_snapshot`.

### WS4 — Restore UI (Recipes & apply)

- **History** panel per recipe: list revisions with timestamp + dilution summary
- **Restore** = copy snapshot components into live row **as new revision** (never delete history)
- Distinct from Commons pack re-import (which skips existing recipe headers)

### WS5 — Crop ops report API

`GET /farms/{id}/crop-cycles/{cid}/ops-timeline?from=&to=`

Returns ordered events:

| `kind` | Source | Includes |
|--------|--------|----------|
| `stage` | `crop_cycle_stage_events` | growth_stage |
| `program_run` | `automation_runs` where `program_id` set | program name, zone, formula_snapshot |
| `mix` | `mixing_events` + components | reservoir, ml per input/batch |
| `apply` | `fertigation_events` | volume, EC/pH, zone |
| `light` | lighting schedule / `automation_runs` for light schedules | on/off or dim level |

Enables: “how much was watered + what formula + what light window” for a date range.

Phase 211.04 (optional follow-on): thin **Crop report** UI — see [phase_211_04_crop_ops_report_ui.plan.md](phase_211_04_crop_ops_report_ui.plan.md). 211.02 ships API + Guardian read only. Farm permissions: [211.03](phase_211_03_farm_permissions.plan.md).

### WS6 — Guardian

Read tool `list_crop_cycle_ops(farm_id, crop_cycle_id, from, to)` — no writes.

## WS0 — Natural farming UI trim (user request, ship with or just before WS1)

Remove from Natural farming workspace:

- **Switchover guide** tab (education duplicated by Make a batch step 3 + Recipe library)
- **Jump to** rail (Feed & water, Money, Zones, Help) — sidebar already navigates

Keep: Make a batch · Recipe library · Recipes & apply · On hand.

Bootstrap / switchover pack apply remains in **Settings → farm bootstrap** and Commons import on Recipes & apply.

## Acceptance

- Edit recipe components → new revision row; previous revision unchanged
- Program run after edit → `automation_runs.details.formula_snapshot` matches revision active at tick
- Restore revision 3 → live recipe matches rev 3; revision 4 created (not silent overwrite)
- Crop ops timeline returns stage + feed + mix rows for demo farm cycle
- `go test` + UI closure tests green

## Out of scope

- Cross-farm analytics (212)
- Full PDF report UI
- Versioning Guardian write proposals (read-only in 211.02)

## File touch list (implementation)

| Layer | Files |
|-------|--------|
| DB | `db/migrations/20260724_phase211_02_recipe_revisions.sql` |
| API | `internal/handler/naturalfarming/recipe.go`, `internal/automation/program_tick.go`, new `internal/handler/fertigation/crop_ops_timeline.go` |
| UI | `RecipesApplyPanel.vue` (history/restore), remove switchover mount |
| Guardian | `internal/farmguardian/tools/` read tool |
