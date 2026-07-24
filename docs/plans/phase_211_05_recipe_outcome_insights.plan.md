---
name: Phase 211.05 — Recipe & program outcome insights (Guardian history correlation)
overview: >
  Join the formula-at-time data shipped in 211.02 (which recipe/revision fed
  a cycle) with cost + yield already tracked per crop cycle, so Guardian can
  answer "how did this recipe actually perform?" from real history — grounded
  aggregates with sample size, never a fabricated forecast.
todos:
  - id: ws0-data-audit
    content: "WS0: Confirm join path — mixing_events/automation_runs.details.application_recipe_id+revision → crop_cycle (via zone+time window) → cost_transactions + yield_grams. No schema changes expected."
    status: completed
  - id: ws1-attribution-query
    content: "WS1: SQL — resolve the dominant application_recipe_id/revision_id used by a harvested cycle from its ops events; flag cycles that mixed multiple revisions"
    status: completed
  - id: ws2-outcomes-builder
    content: "WS2: Go package internal/cropcycle/recipeoutcomes — aggregate cycle_count, avg/median yield_grams, avg cost_per_gram, avg duration_days per (crop_key, application_recipe_id, revision_id); enforce min sample size before surfacing stats"
    status: completed
  - id: ws3-api
    content: "WS3: GET /farms/{id}/crop-analytics/recipe-outcomes?crop_key=&recipe_id= — gate cost fields behind money.costs.read (211.03 scopes)"
    status: completed
  - id: ws4-guardian-tool
    content: "WS4: Guardian read tool summarize_recipe_outcomes + grounding rule — always states N, never claims causation, never forecasts beyond 'last N cycles averaged X'"
    status: completed
  - id: ws5-ui-track-record
    content: "WS5: 'Track record' chip on Recipes & apply cards + Crop Cycle Summary ('vs recipe average'), scope-gated $ figures"
    status: completed
  - id: ws6-closure
    content: "WS6: Go + Vitest closure tests, answer-accuracy regression (no fabricated numbers), operator-tour + farm-guardian-architecture.md cross-links"
    status: completed
isProject: false
---

# Phase 211.05 — Recipe & program outcome insights

**Status:** Complete (WS0–WS6 shipped) · **Depends on:** [211.02 recipe formula history](phase_211_02_recipe_formula_history.plan.md) (formula-at-time attribution), [211.03 farm permissions](phase_211_03_farm_permissions.plan.md) (`money.costs.read` gate) · **After:** [211.04 crop ops report UI](phase_211_04_crop_ops_report_ui.plan.md)

## The one job

> "Did switching from JMS-only to JMS+FPJ actually help, or does it just feel that way?" — answer it from **this farm's own recorded history**: recipe/revision actually mixed, cost actually tagged, yield actually weighed. No guessing, no generic agronomy claims — every number traces back to a row.

## Why now

211.02 already stamps `application_recipe_id` + `application_recipe_revision_id` + `formula_snapshot` onto every `mixing_events.metadata` and `automation_runs.details` row. 211.04 surfaces that per-cycle. What's missing is the **rollup across cycles**: nobody has joined "which recipe fed this cycle" to "what did this cycle cost and yield" and aggregated it by recipe. That join is the whole point of a Guardian who has "seen everything" — cost, production, recipes, and inputs in one place.

## Non-negotiable framing (read before building)

This is **historical correlation, not prediction**. The existing grounding rules in this codebase (`CropTargetsGroundingRule`, `StructuredTruthGroundingRule` in [`internal/farmguardian/readtools_crop.go`](../../internal/farmguardian/readtools_crop.go)) exist specifically to stop Guardian from inventing numbers. This phase must follow the same discipline:

- Every stat surfaced **must** state its sample size (`N=3 cycles`).
- Guardian may say *"cycles using this recipe averaged 180g"* — never *"you will get 180g"* or *"this recipe is better"* (correlation ≠ causation; too many confounders — stage timing, zone, season).
- Below a minimum sample size (default **2**), the tool reports "not enough history yet" instead of a lonely single-cycle average dressed up as a trend.
- No new ML/statistics dependency. This is `AVG()`/`COUNT()` over rows you already trust, not a model.

## Scope

- Read-only aggregation across **harvested** crop cycles for one farm.
- Attribution source: existing `mixing_events` / `automation_runs` rows already tagged in 211.02 — **no new write paths, no schema changes**.
- Cost figures scoped behind `money.costs.read` (fails closed like the rest of 211.03).

## Out of scope

- True predictive modeling / ML (explicitly rejected above — grounding discipline).
- Cross-farm or cross-install benchmarking (that's federation territory — [212](phase_212_dual_farm_federation_test.plan.md), still deferred).
- Guardian **write** proposals (e.g. "auto-switch to the better recipe") — read-only insight only in this phase.
- Recipe recommendation ranking beyond the raw aggregate (no "best recipe" verdict — the operator draws that conclusion, Guardian shows the receipts).

## Workstreams

### WS0 — Data audit (no code)

Confirm the join path end to end on a real seeded farm:

1. `crop_cycles` (harvested: `is_active=false`, `yield_grams` set, `plant_id` → `crop_key` via `plants`)
2. → its `mixing_events` / `automation_runs` in `[started_at, harvested_at]` window, filtered by `zone_id`
3. → `metadata`/`details` JSON → `application_recipe_id`, `application_recipe_revision_id`
4. → `cost_transactions` linked via `crop_cycle_id` for expense/income totals

Document edge cases: cycles with zero fertigation events (soil/manual watering only — exclude from recipe attribution, not from yield stats elsewhere), cycles that used more than one recipe or revision mid-grow (tag as `mixed`, still countable for yield-only questions but excluded from recipe-specific aggregates).

### WS1 — Attribution query

New SQL (`db/queries/crop_recipe_attribution.sql` → sqlc-generated) resolving, per harvested cycle: the **dominant** `application_recipe_id`/`revision_id` (most mixing/program-run events referencing it within the cycle window). Return `NULL`/`is_mixed=true` when no single recipe accounts for a clear majority (e.g. >60% of events) — don't force a false single-recipe attribution.

### WS2 — Outcomes builder

`internal/cropcycle/recipeoutcomes/outcomes.go`:

```go
type RecipeOutcome struct {
    CropKey            string
    ApplicationRecipeID int64
    RevisionID          *int64
    RecipeName          string
    CycleCount          int
    AvgYieldGrams       *float64
    MedianYieldGrams    *float64
    AvgCostPerGram      *float64 // only when cost data + single currency
    AvgDurationDays     *float64
    SampleCycleIDs      []int64  // capped, for citing specifics
}
```

Enforce `CycleCount >= minSample` (const, default 2) before including in results — cycles below threshold roll into an "insufficient history" bucket the caller can still list by name without stats.

### WS3 — API

`GET /farms/{id}/crop-analytics/recipe-outcomes?crop_key=&recipe_id=`

- Handler in `internal/handler/cropcycle/analytics.go` (alongside existing `FarmAnalytics`).
- Cost fields (`avg_cost_per_gram`) omitted from the JSON entirely (not zeroed) when the caller's farm scopes lack `money.costs.read` — mirrors the SuppliesHub/RecipesApplyPanel fail-closed pattern from 211.03.
- Response shape mirrors `RecipeOutcome` above, grouped by `crop_key`.

### WS4 — Guardian read tool

`summarize_recipe_outcomes` in `internal/farmguardian/tools/` + render function in `internal/farmguardian/readtools_crop.go` (sibling to `summarize_farm_crops_by_key`).

Intent regex: phrases like *"which recipe worked best"*, *"did switching recipes help"*, *"predict my yield"*, *"based on history"* — the last two are explicitly **redirected** into historical-average framing by the tool's own output, not refused.

Sample rendered output:

```
summarize_recipe_outcomes — tomato (crop_key=tomato)
JMS Foliar rev 3: 4 harvested cycles — avg yield 182g (range 140–210g), avg $0.21/g, avg 61 days
FPJ+JMS combo rev 1: 2 harvested cycles — avg yield 205g (range 190–220g), avg $0.24/g, avg 58 days
1 cycle used a mixed/unclear recipe — excluded from recipe-specific averages.
Correlational only — stage timing, season, and zone differ between cycles; not a controlled comparison.
```

New grounding rule (append to `internal/farmguardian/readtools_crop.go` alongside `StructuredTruthGroundingRule`):

> **Recipe outcome grounding (Phase 211.05):** `summarize_recipe_outcomes` numbers are historical averages over named past cycles, not predictions. Always state N. Never say a recipe "is better" or "will produce X" — say cycles "averaged X". Below the tool's minimum sample size, say so explicitly instead of citing a single-cycle number as a trend.

### WS5 — UI track record (may split to 211.06 if WS1–WS4 alone are a full PR)

- **Recipes & apply** (`RecipeLibraryPanel.vue` / `RecipesApplyPanel.vue`): small chip per recipe — `Used in 4 harvested cycles · avg 182g · $0.21/g` — sourced from WS3 endpoint, scope-gated (`$` hidden without `money.costs.read`).
- **Crop Cycle Summary**: below Yield card, a line — `This cycle's recipe (JMS Foliar rev 3) averaged 182g across 4 cycles` — only rendered when the cycle has a clear (non-mixed) recipe attribution and sample size clears the threshold.
- Empty/low-sample states use existing `EmptyStateHint`/inline text pattern, not a blocking error.

### WS6 — Closure

- Go: `internal/cropcycle/recipeoutcomes/outcomes_test.go` — sample-size threshold, mixed-recipe exclusion, single-currency cost gate, zero-yield cycle exclusion.
- Go: extend `internal/farmguardian/answer_accuracy_test.go` (or sibling) to assert the tool never emits an unqualified number without "avg"/"N cycles" nearby.
- Vitest: `phase-211-05-closure.test.js` — endpoint call shape, chip rendering, scope gating.
- Docs: cross-link from `docs/farm-guardian-architecture.md` (new `§7.0x` or next free letter) and `docs/operator-tour.md` §7u natural farming section; mark this plan **Complete** when shipped.

## Shipped (WS0–WS6)

| WS | Deliverable |
|----|-------------|
| WS0 | Join path confirmed — no migration; attribution from `mixing_events.metadata` + `automation_runs.details` in cycle window |
| WS1 | `db/queries/crop_recipe_attribution.sql` — harvested cycles + recipe attribution hits |
| WS2 | `internal/cropcycle/recipeoutcomes` — dominant recipe at 60%, min sample 2, cost/yield aggregates |
| WS3 | `GET /farms/{id}/crop-analytics/recipe-outcomes` — cost fields omitted without `money.costs.read` |
| WS4 | `summarize_recipe_outcomes` read tool + `RecipeOutcomeGroundingRule` in platform context |
| WS5 | `RecipeTrackRecordChip` on Recipes & apply; `CycleRecipeTrackRecord` on crop cycle summary — cost gated by `money.costs.read` |
| WS6 | `outcomes_test.go`, `RecipeOutcomeToolGroundingNote`, `phase-211-05-closure.test.js`, docs §7.0ah + operator-tour §7u cross-links |

## Acceptance criteria

- [x] `summarize_recipe_outcomes` never appears in a transcript with a bare number lacking "avg" or a cycle count nearby (regression-tested).
- [x] Sample size < 2 never renders a stat as if it were a trend.
- [x] Cost figures absent (not zero) for callers without `money.costs.read`.
- [x] No new database migration required — pure read/aggregate over existing 211.02 attribution data.
- [x] UI track-record chip only renders for recipes with a clear (non-mixed) attribution across ≥ threshold cycles.

## Related

- [211.02 recipe formula history](phase_211_02_recipe_formula_history.plan.md) — source of `application_recipe_id`/`revision_id`/`formula_snapshot` attribution this phase joins against.
- [211.03 farm permissions](phase_211_03_farm_permissions.plan.md) — `money.costs.read` scope gate for cost fields.
- [211.04 crop ops report UI](phase_211_04_crop_ops_report_ui.plan.md) — per-cycle formula-at-time UI this phase rolls up across cycles.
- [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md) — cross-farm benchmarking is explicitly deferred there, not here.
