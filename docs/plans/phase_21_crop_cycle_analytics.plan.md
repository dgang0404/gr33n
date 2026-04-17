---
name: Phase 21 Crop Cycle Analytics & Yield
overview: >
  Reporting-only phase that turns the data gr33n has been collecting into the operator
  story that sells the platform. Cycle summary (liters delivered, EC drift, cost tagged
  to cycle, yield per gram, days active), side-by-side cycle comparison, UI pages, CSV
  export. Low risk — no new automation, no new devices, just SQL + UI. Target: ~1 week.
todos:
  - id: ws1-cycle-summary-api
    content: "WS1: GET /crop-cycles/{id}/summary — liters delivered, avg/min/max EC, cost total tagged to cycle, yield grams, yield/L, days active, stage history"
    status: pending
  - id: ws2-cycle-compare-api
    content: "WS2: GET /farms/{id}/crop-cycles/compare?ids=1,2,3 — side-by-side metrics for 2+ cycles, same shape repeated"
    status: pending
  - id: ws3-ui-pages
    content: "WS3: Cycle Summary page + Compare view in ui/src/views; SideNav entry; HelpTips on each metric; link from CropCycles list"
    status: pending
  - id: ws4-csv-export
    content: "WS4: CSV export — GET /crop-cycles/{id}/summary.csv and /crop-cycles/compare.csv; reuse the cost export pattern"
    status: pending
isProject: false
---

# Phase 21 — Crop Cycle Analytics & Yield

## Why this phase

Phases 19 and 20 make the platform **safe** and **reactive**. Phase 21 makes it **provable**. The data is already there — fertigation events are linked to crop cycles, costs can be tagged to zones/cycles, yield grams is a column on the cycle — but nothing aggregates it into the story an operator tells at the end of a run: *"how much water, how much EC drift, how much did it cost me, how many grams did I get, and how does this compare to last time?"*

This phase is intentionally small, read-only, and independent of Phase 20. If 20 runs long, 21 can still ship.

## Scope

| WS | Focus | Location in repo |
|----|--------|------------------|
| **WS1** | Cycle summary API | `db/queries/cropcycle.sql` (new summary query), `internal/db/cropcycle.sql.go` (generated), `internal/handler/cropcycle/handler.go`, `cmd/api/routes.go`, `openapi.yaml` |
| **WS2** | Cycle compare API | same files; new `Compare` handler |
| **WS3** | UI pages | `ui/src/views/CropCycleSummary.vue` (new), `ui/src/views/CropCycleCompare.vue` (new), router, SideNav, link from existing cycle list |
| **WS4** | CSV export | reuse `GET /farms/{id}/costs/export` helper pattern |

## Work-stream detail

### WS1 — Cycle summary API

- **Endpoint:** `GET /crop-cycles/{id}/summary`
- **Response shape:**
  ```jsonc
  {
    "cycle": { /* existing CropCycle fields */ },
    "duration_days": 68,
    "fertigation": {
      "event_count": 142,
      "total_liters": 980.5,
      "avg_ec_mscm": 1.62,
      "min_ec_mscm": 1.12,
      "max_ec_mscm": 2.05,
      "avg_ph": 6.1
    },
    "cost": {
      "total_amount": 312.40,
      "currency": "USD",
      "by_category": [{ "category": "nutrients", "total": 180.10 }, ...]
    },
    "yield": {
      "grams": 412.0,
      "grams_per_liter": 0.42,
      "grams_per_day": 6.06,
      "cost_per_gram": 0.76
    },
    "stages": [
      { "stage": "seedling", "entered_at": "2026-03-01" },
      { "stage": "early_veg", "entered_at": "2026-03-14" },
      ...
    ]
  }
  ```
- **Auth:** JWT + farm-member via `farmauthz.RequireFarmMember` (resolve farm_id from cycle).
- **Query strategy:**
  - One SQL `SELECT` per block (cycle, fertigation aggregates, cost aggregates, stage history) — simpler to read and maintain than a single super-query. Shared db.Tx optional, not required.
  - Stage history may need a dedicated table or can be derived from audit events; confirm the schema and pick the simpler path.
- **OpenAPI:** new `CropCycleSummary` schema. Keep sub-objects named (`fertigation`, `cost`, `yield`, `stages`) so the UI can render them progressively.

### WS2 — Cycle compare API

- **Endpoint:** `GET /farms/{id}/crop-cycles/compare?ids=1,2,3`
- **Response shape:** `{ cycles: [CropCycleSummary, ...] }`, same object repeated; the UI does the side-by-side rendering.
- **Auth:** JWT + member on the farm; every cycle ID must belong to the farm or the request is 400.
- **Limits:** cap at 5 cycles per call. More than that is a dashboard, not a comparison.

### WS3 — UI pages

- **CropCycleSummary.vue** — consumed from the CropCycles list via a "View summary" button.
  - Header: cycle name, strain, zone, stage, duration, active/harvested badge.
  - Four metric cards: **Fertigation**, **Cost**, **Yield**, **Stages** (timeline).
  - Each card has a HelpTip describing what the metric includes and what it doesn't.
- **CropCycleCompare.vue** — multi-select cycles from the list, then route to `/crop-cycles/compare?ids=…`.
  - Columns = cycles, rows = metrics.
  - Highlight the best and worst column for each metric (e.g. highest yield/gram, lowest cost/gram).
  - Empty-state: "Pick two or more crop cycles from the same farm to compare."
- **SideNav:** add "Analytics" under "Grow", or fold entries under the existing CropCycles section — decide based on how crowded the nav looks after Phase 20.

### WS4 — CSV export

- **Endpoints:** `GET /crop-cycles/{id}/summary.csv` and `GET /farms/{id}/crop-cycles/compare.csv?ids=…`.
- **Shape:** one wide CSV with one row per cycle and a column per metric. Matches what operators paste into a spreadsheet for growers-group postings.
- **Impl:** reuse the CSV writing helper already in `costhandler.Export`; keep the same Content-Disposition pattern.

## After Phase 21

- **RAG-ready** — the summary objects become an excellent retrieval substrate: a cycle's full story in a single JSON. Phase 22 can chunk per-cycle summaries and answer "how did my last run go?" cleanly.
- **Product story** — the first time gr33n can be shown to someone with "here's what it actually did for my garden this year."
- **Unlocks billing discussions** — cost-per-gram is the number that sells a subscription.

## Risks / things to watch

- **Stage history schema** — check whether we record stage transitions with timestamps or only the *current* stage; if the latter, stage history for older cycles will be empty. If so, either backfill from audit events or document that stage history only works for cycles started after this phase.
- **Cost attribution** — costs are tagged to a zone, not always a cycle. Decide the attribution rule (zone + time window of the cycle) and document it; the number will always be approximate.
- **Currency mixing** — if a farm logs costs in multiple currencies, the summary should return one total per currency rather than summing blindly. Keep the shape `[{ "currency": "USD", "total": 312.40 }]` if needed.

---

## Using this plan in a new chat (copy-paste prompt)

```text
Implement Phase 21 per @docs/plans/phase_21_crop_cycle_analytics.plan.md.

Scope:
1) WS1 — GET /crop-cycles/{id}/summary with the shape documented in the plan (cycle, duration_days, fertigation aggregates, cost aggregates, yield metrics, stages). Add CropCycleSummary schema to openapi.yaml.
2) WS2 — GET /farms/{id}/crop-cycles/compare?ids=1,2,3 returning { cycles: [CropCycleSummary,...] }. Validate every id belongs to the farm; cap 5 ids.
3) WS3 — CropCycleSummary.vue + CropCycleCompare.vue under ui/src/views with HelpTips, SideNav entries, and a link from the existing cycle list.
4) WS4 — CSV variants for both endpoints; reuse the cost export helper.

Constraints: keep openapi.yaml 1:1 with routes.go, run `go test ./cmd/api/...` and `pytest pi_client/test_gr33n_client.py`, update this plan's YAML todo statuses when each WS lands. No new automation or worker changes in this phase — reporting-only.
```
