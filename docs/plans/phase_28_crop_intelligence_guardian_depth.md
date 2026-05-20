---
name: Phase 28 — Crop Intelligence & Guardian Depth
overview: >
  Two converging threads: (1) close the long-open Phase 21 crop-cycle analytics
  gap so operators finally see a per-cycle story (liters, EC, cost, yield,
  stage history); and (2) deepen Farm Guardian so it can answer questions
  *about* those analytics, respond to alerts intelligently, and surface its own
  cost activity to operators. Also includes an OpenAPI parity pass for every
  Phase 26–27 endpoint that was never added to openapi.yaml.
todos:
  - id: ws1-crop-analytics
    content: "WS1: Crop cycle analytics API — GET /crop-cycles/{id}/summary + GET /farms/{id}/crop-cycles/compare?ids=… + CSV exports (closes Phase 21 backlog)"
    status: completed
  - id: ws2-analytics-ui
    content: "WS2: CropCycleSummary.vue + CropCycleCompare.vue — metric cards, stage timeline, compare columns, SideNav entry, link from cycle list"
    status: completed
  - id: ws3-guardian-analytics
    content: "WS3: Guardian ↔ crop cycle integration — Guardian can answer questions about the current/historical cycles; snapshot includes active-cycle summary metrics when farm_id present"
    status: completed
  - id: ws4-guardian-alerts
    content: "WS4: Guardian alert integration — Guardian can explain unread alerts (rule triggered, sensor threshold, what to do); alert context injected into grounded snapshot"
    status: completed
  - id: ws5-cost-dashboard
    content: "WS5: Token-usage dashboard — operator-visible per-user/per-farm rolling totals in Settings; alert hook that fires a notification when >80% of budget is consumed"
    status: completed
  - id: ws6-openapi-parity
    status: completed
    content: "WS6: OpenAPI parity for Phase 26–27 endpoints — /capabilities, /v1/chat, /v1/chat/sessions, /farms/{id}/rag/search, /farms/{id}/rag/answer all documented in openapi.yaml"
    status: pending
isProject: false
---

# Phase 28 — Crop Intelligence & Guardian Depth

## Why this phase

Phase 27 shipped Farm Guardian as a capable AI layer. Phase 28 makes it *more useful for farm operators in practice* by closing two gaps:

1. **Crop analytics** (Phase 21 backlog) — data has been accumulating for months (fertigation events, cost tags, yield entries, EC readings) but there is no way to read a summary of a cycle without manually querying the DB. Operators need "how did this run compare to the last one?" answered in the UI.

2. **Guardian depth** — Guardian knows about zones and open alerts from the snapshot, but it can't yet answer questions like "what was my average EC in the last flower cycle?" or "what does this high-humidity alert mean for bud rot risk?". Wiring the crop analytics summary and alert detail into the grounded snapshot closes that.

There is also housekeeping: every Phase 26 and 27 endpoint is missing from `openapi.yaml`, which breaks the `make audit-openapi` gate that proves route registration matches the spec.

---

## Scope

| WS | Focus | Primary files |
|----|-------|---------------|
| **WS1** | Crop cycle analytics API | `db/queries/cropcycle.sql`, `internal/db/cropcycle.sql.go`, `internal/handler/cropcycle/handler.go`, `cmd/api/routes.go` |
| **WS2** | Analytics UI pages | `ui/src/views/CropCycleSummary.vue`, `ui/src/views/CropCycleCompare.vue`, `ui/src/router/index.js`, `ui/src/components/SideNav.vue` |
| **WS3** | Guardian ↔ crop cycles | `internal/farmguardian/snapshot.go` (extend), `internal/handler/chat/handler.go` |
| **WS4** | Guardian ↔ alerts | `internal/farmguardian/snapshot.go` (alert detail), operator runbook |
| **WS5** | Token-usage dashboard | `internal/db/conversation_turns.sql.go`, `cmd/api/routes.go`, `ui/src/views/Settings.vue` |
| **WS6** | OpenAPI parity | `openapi.yaml` |

---

## Work-stream detail

### WS1 — Crop cycle analytics API

These endpoints close the Phase 21 todos and were the direct predecessor motivation for building Farm Guardian (see [Phase 21 plan](./phase_21_crop_cycle_analytics.plan.md)).

**`GET /crop-cycles/{id}/summary`**

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
    "totals": [{ "currency": "USD", "total": 312.40 }],
    "by_category": [{ "category": "nutrients", "total": 180.10 }]
  },
  "yield": {
    "grams": 412.0,
    "grams_per_liter": 0.42,
    "grams_per_day": 6.06,
    "cost_per_gram": 0.76
  },
  "stages": [
    { "stage": "seedling", "entered_at": "2026-03-01T00:00:00Z" }
  ]
}
```

Auth: JWT + farm member (resolve `farm_id` from the cycle row). Query per sub-object (fertigation aggregates, cost aggregates by category, stage history) rather than one super-query.

**`GET /farms/{id}/crop-cycles/compare?ids=1,2,3`**

Returns `{ "cycles": [CropCycleSummary, …] }` — same object shape repeated. All IDs must belong to the farm; cap at 5 per call.

**CSV exports**

- `GET /crop-cycles/{id}/summary.csv` — one row per metric block, useful for pasting into a spreadsheet.
- `GET /farms/{id}/crop-cycles/compare.csv?ids=…` — one row per cycle, one column per metric.

Reuse the existing `costhandler.Export` CSV helper and `Content-Disposition` pattern.

### WS2 — Analytics UI pages

**`CropCycleSummary.vue`** — routed from a "Summary" button on the CropCycles list row.
- Header strip: cycle name, strain, zone, stage badge, duration chip, active/harvested status.
- Four metric cards: **Fertigation**, **Cost**, **Yield**, **Stage timeline**.
- Each card has a `HelpTip` explaining what the number includes and what it misses (e.g. costs tagged to zone but not cycle are noted as "zone-level only").

**`CropCycleCompare.vue`** — multi-select from the cycle list, route to `/crop-cycles/compare?ids=…`.
- Columns = cycles, rows = metrics. Highlight best/worst column per metric row (CSS class, not hard-coded).
- Empty state: "Select two or more crop cycles from the same farm to compare."

**SideNav** — add "Analytics" entry in the Grow group (below CropCycles, above Tasks).

### WS3 — Guardian ↔ crop cycle integration

Extend `internal/farmguardian/snapshot.go` to pull the **active crop cycle summary** (WS1 data) when `farm_id` is present in the chat request and include it in the `PromptBlock`. Cap the summary at ~200 tokens to avoid blowing the context window on farms with long histories.

This lets Guardian answer questions like:
- "What's my current EC trend in the flower room?"
- "Am I on track vs my last cycle cost-wise?"
- "When did we move to the late-flower stage?"

Without WS3 those questions get vague "I don't have access to your live data" responses.

### WS4 — Guardian ↔ alert integration

The current snapshot already includes the *count* of unread alerts. Extend it to include the **top 3 unread alerts by severity** (rule_name, sensor_label, triggered_value, threshold, triggered_at) so Guardian can explain them.

This enables:
- "You have a high-humidity alert in the Flower Room triggered 2h ago (72% RH vs 65% threshold). High humidity at this stage increases bud-rot risk; consider increasing airflow or reducing foliar feeding."

Cap the alert detail block to the 3 most recent/severe to keep the prompt budget predictable.

### WS5 — Token-usage dashboard

Operators who run shared deployments (multiple staff using the same Guardian) need visibility into who is consuming token budget without diving into the DB.

**Backend:**

New `GET /v1/chat/usage` endpoint (JWT, returns the calling user's rolling-window totals + remaining budget):

```jsonc
{
  "window_hours": 1,
  "user": {
    "used_tokens": 3200,
    "max_tokens": 10000,
    "remaining_tokens": 6800
  },
  "farm": {               // only present if farm_id query param supplied
    "farm_id": 7,
    "used_tokens": 14000,
    "max_tokens": 50000,
    "remaining_tokens": 36000
  }
}
```

Also add a **notification hook**: when a chat turn pushes the user's rolling total past 80% of their budget, fire a `gr33ncore.alerts` row (type `chat_budget_warning`, severity `medium`) so the existing alert channel delivers it. Fired once per window (debounce by checking whether a warning alert already exists in the same window).

**UI:**

Add a "Guardian usage" card to `Settings.vue` that calls `GET /v1/chat/usage` and renders a compact progress bar (used / max, window label). Only shown when `AI_ENABLED=true` and at least one cap is configured.

### WS6 — OpenAPI parity

Every Phase 26–27 route is absent from `openapi.yaml`, breaking `make audit-openapi`. Routes to document:

| Route | Phase |
|-------|-------|
| `GET /capabilities` | 27 WS2 |
| `POST /v1/chat` | 27 WS5 |
| `GET /v1/chat/sessions` | 27 WS5 |
| `GET /v1/chat/sessions/{session_id}` | 27 WS5 |
| `PATCH /v1/chat/sessions/{session_id}` | 27 WS6 |
| `DELETE /v1/chat/sessions/{session_id}` | 27 WS6 |
| `POST /farms/{id}/rag/search` | 24–25 |
| `POST /farms/{id}/rag/answer` | 24–25 |
| `GET /v1/chat/usage` | 28 WS5 |

Add request/response schemas for each and make `make audit-openapi` green.

---

## After Phase 28

- **Phase 21 fully closed** — no more open backlog from that era.
- **Guardian is genuinely useful for day-to-day farm ops** — not just a chat demo.
- **OpenAPI is canonical again** — external tooling (Swagger UI, client generators) works on current routes.
- **Natural next themes:**
  - Agricultural reference corpus (static agronomic docs ingested into a separate RAG index alongside the farm-specific corpus — see `docs/rag-scope-and-threat-model.md §9` for the boundary).
  - Guardian-driven proactive alerts ("EC has been trending up for 3 days — schedule a reservoir flush?").
  - Multi-farm / cooperative data via Insert Commons (farm summaries shareable with the network with consent).

---

## Using this plan in a new chat

```text
Implement Phase 28 per @docs/plans/phase_28_crop_intelligence_guardian_depth.md.

Start with WS1: GET /crop-cycles/{id}/summary and GET /farms/{id}/crop-cycles/compare. Follow the shapes in the plan doc. Auth: JWT + farm member. Add the two endpoints to openapi.yaml at the same time (partial WS6). Run `go test ./cmd/api/...` and check `make audit-openapi` is still clean before moving to WS2.
```

---

## Shipped notes

### WS1 — Crop cycle analytics API (shipped 2026-05-19)

- **`db/queries/crop_cycles.sql`** — `GetFertigationAggregatesByCropCycle` rolls up `fertigation_events` per cycle: `event_count`, `total_liters`, `avg/min/max ec_after_mscm`, `avg_ph` (blended pre + post). All COALESCEd to zero so empty cycles never serialise NULLs.
- **`internal/db/crop_cycle_analytics.sql.go`** — hand-written Go binding (same pattern Phase 27 used for `conversation_turns`) to avoid a repo-wide sqlc regen. Folds back cleanly next routine sqlc pass.
- **`internal/handler/cropcycle/analytics.go`** — `Summary` (GET `/crop-cycles/{id}/summary`) and `Compare` (GET `/farms/{id}/crop-cycles/compare?ids=…`) handlers, JWT + farm-member auth, one SQL call per sub-block (fertigation, costs, yield, stages). Reuses the existing `GetCostTotalsByCropCycle` query for costs. Stage history is a single-row stand-in (the schema only stores `current_stage`); `stage_history_supported: false` is exposed so the UI / OpenAPI consumers know the limitation.
- **Cost-per-gram** is only emitted when the cycle has costs in exactly one currency — mixing currencies blindly is misleading, and the JSON `cost.totals` already breaks it out per-currency anyway.
- **CSV exports** — `/summary.csv` and `/compare.csv` ride the same handlers via a `.csv` suffix check; emits a single wide row per cycle so spreadsheets work without nested JSON. `MIXED:` sentinel currency for multi-currency cycles directs operators back to the JSON view.
- **`MaxCompareCycles = 5`** caps the compare endpoint; deduplicates ids client-side so a duplicated id doesn't burn a slot.
- **Smoke tests** (`cmd/api/smoke_phase28_ws1_test.go`) — happy path JSON, 404, CSV header + USD row, 2-way compare, foreign-farm rejection, too-many ids, missing-ids. All 7 tests green against real Postgres.

### WS2 — Crop analytics UI pages (shipped 2026-05-19)

- **`ui/src/views/CropCycleSummary.vue`** — header strip (name, strain, stage, duration, status) + four metric cards (Fertigation, Cost, Yield, Stage timeline) + CSV download button + Compare ↔ deep-link. HelpTips on each card explain what the number includes (e.g. "EC averages use the after-feed reading" and "cost_per_gram only when one currency").
- **`ui/src/views/CropCycleCompare.vue`** — picker grid (checkbox per crop cycle, capped at 5 with disabled state past the limit), side-by-side compare table where columns are cycles and rows are metrics. Best column highlighted emerald, worst amber. `better: 'higher'` for yield/duration/liters, `better: 'lower'` for cost-per-gram and expenses. Rows where every cycle has `null` are filtered out so the table stays compact.
- **`ui/src/components/MetricChip.vue`** — small reusable labelled-value chip used by both views.
- **`ui/src/stores/farm.js`** — new `loadCropCycleSummary(id)` + `loadCropCycleCompare(farmId, ids)` methods. The compare helper short-circuits to `{cycles: []}` when `ids` is empty so the view doesn't make a 400-bound HTTP request on mount.
- **Router** — `/crop-cycles/:id/summary` (named `crop-cycle-summary`) + `/farms/:fid/crop-cycles/compare` (named `crop-cycle-compare`).
- **`ui/src/components/SideNav.vue`** — new **Analytics** entry under Monitor (📊). The route is computed from `farmContext.farmId` so it always lands on the current farm.
- **`ui/src/views/Fertigation.vue`** — added a **Summary →** router-link to every crop-cycle card (in the Crop Cycles tab) for one-click drill-in.
- **Tests** — `ui/src/__tests__/crop-cycle-analytics.test.js` (10 new tests): store methods (compare query-string join + empty-ids short-circuit), summary view (header + four cards + error path), compare view (empty state, table rendering after selection, 5-cycle cap, "select a farm" hint). All 40 UI tests pass.

### WS3 — Guardian ↔ crop cycle integration (shipped 2026-05-19)

- **`internal/farmguardian/cycle_analytics.go`** — new `CycleAnalytics` struct + `fetchCycleAnalytics(q, cycle)` helper that reuses the Phase 28 WS1 `GetFertigationAggregatesByCropCycle` + `GetCostTotalsByCropCycle` queries so the numbers Guardian sees match what the operator sees in `CropCycleSummary.vue`. Computes `liters/day`, `grams/day`, `grams/liter`, and `cost_per_gram` derived ratios in the same single-currency-only guard pattern the WS1 handler uses.
- **`internal/farmguardian/format.go`** — prompt-budget-aware number formatters (`formatLiters`, `formatEC`, `formatPH`, `formatMoney`, `formatGrams`, `formatGramsPerDay`). Whole-liter values render without a decimal so "980L" not "980.0L"; EC + pH at 2 decimals (operator-canonical resolution).
- **`internal/farmguardian/snapshot.go`** — extended `ActiveCycle` with `ID` + `Analytics CycleAnalytics`. `BuildSnapshot` now populates analytics for the first **`SnapshotMaxAnalyticsCycles` = 3** active cycles in `started_at DESC` order. Per-cycle failures log at WARN and continue (the cycle still appears in the snapshot without metrics). `Render` indents a `metrics:` line under each cycle bullet that has analytics.
- **Prompt output shape** — Guardian's system prompt now includes lines like `metrics: feed: 142 events / 980L (14.7L/d); EC 1.62 (1.12–2.05); pH 6.10; cost: 312.40 USD; yield: 412g (6.06g/d); cost/g: 0.76 USD` beneath each active cycle, so "how's my flower run going?" gets a concrete answer.
- **Bounded prompt cost** — `SnapshotMaxAnalyticsCycles = 3` keeps the analytics block ≤200 tokens even on farms with 10+ active cycles. Older cycles still render their basic name/strain/stage line.
- **Tests** — `internal/farmguardian/cycle_analytics_test.go` (7 unit tests: Empty/Render/cost-omit-on-mixed-currency/formatters) + `cmd/api/smoke_phase28_ws3_test.go` (2 real-DB smokes: attaches analytics to a seeded cycle with verifiable fertigation+cost data; budget cap is honoured when more than N cycles are active). All farmguardian + chat tests green.

### WS4 — Guardian ↔ alert integration (shipped 2026-05-19)

- **`db/queries/alerts.sql`** — new `ListRecentUnreadAlertsByFarm` query returns the top-N unread alerts ordered by `severity DESC NULLS LAST, created_at DESC`. Critical + recent wins both axes; the LIMIT is supplied by the caller (3 in the snapshot path).
- **`internal/db/alerts_guardian.sql.go`** — hand-written Go binding (same sqlc-bypass pattern Phase 27/28 has been using) projects the row into `RecentUnreadAlertSummary` with only the fields Guardian needs: `id`, `severity`, `subject_rendered`, `message_text_rendered`, `triggering_event_source_type`, `triggering_event_source_id`, `created_at`. The wide row (delivery_attempts, html bodies, etc.) is intentionally omitted to keep prompt budgets predictable.
- **`internal/farmguardian/snapshot.go`** — `Snapshot` gained `UnreadAlertDetails []UnreadAlertDetail`. New `UnreadAlertDetail` struct exposes the rendered fields plus a `TriggeredAt time.Time` for the age humaniser. `BuildSnapshot` only fires the list query when `UnreadAlerts > 0` (saves a round-trip on quiet farms); failures fall through to "just the count" so a transient alerts-table hiccup never strips the snapshot. `Render` adds a `- [severity] subject (source #id, Xh ago)` line per alert and an indented `detail: …` line for the message snippet. A `(+ N more unread alerts)` tail accurately reflects `UnreadAlerts - rendered` (not the trimmed slice length, which would always be ≤ cap).
- **Constants** — `SnapshotMaxAlertDetails = 3` caps how many alerts get full detail; `AlertMessageSnippetMax = 160` runes caps the message-text snippet so a 5KB markdown template can't blow the prompt. `humanizeAge` emits "just now / Xm ago / Xh ago / Xd ago" so the LLM doesn't have to do time math from raw timestamps.
- **Prompt output shape** — Guardian now sees lines like `[high] Humidity threshold breach — Flower Room (sensor_reading #4242, 4h ago)` with the rendered message body folded onto the indented `detail:` line beneath. The LLM can now answer "why is my humidity alert firing?" with real numbers instead of "I don't know which alerts you mean."
- **Tests** — 4 new unit tests in `internal/farmguardian/snapshot_test.go` (alert lines render, "+ N more" reflects total vs rendered, message snippet cap, `humanizeAge` table) + 2 real-DB smokes in `cmd/api/smoke_phase28_ws4_test.go` (attach path uses critical+now to guarantee ranking despite the smoke DB's accumulated test alerts; cap-budget test seeds 5 critical alerts and asserts only 3 details + the "+N more" marker). Existing chat/farmguardian tests untouched and green.

### WS5 — Token-usage dashboard (shipped 2026-05-20)

- **`GET /v1/chat/usage`** (`internal/handler/chat/usage.go`) — JWT-gated endpoint returning rolling-window token totals. User dimension is always present; per-farm dimension is opt-in via `?farm_id=N` and gated by farm-member auth so multi-farm deployments don't leak utilisation across operators. Response includes `pct_used` (rounded to 4 decimals), `remaining_tokens`, and `warning_threshold_pct: 0.80` so the UI can render the 80% threshold marker without hard-coding it. Returns **503** when `AI_ENABLED=false`. SUM-query failures fall back to zeros + WARN log (the endpoint never 500s on a transient hiccup).
- **80% warning hook** (`internal/farmguardian/budget_warning.go`) — `MaybeFireBudgetWarning` runs *after* every successful conversation-turn insert in `persistTurn`. Threshold = **`WarningThresholdPct = 0.80`** of the per-user cap. Inserts a `gr33ncore.alerts_notifications` row with `severity=medium`, `triggering_event_source_type='chat_budget_warning'`, `recipient_user_id` = the caller, `subject_rendered="Chat token budget at N%"`. Skipped entirely when: per-user cap is 0 (disabled), the just-finished turn is ungrounded (no `farm_id` → alerts table requires non-null), or an existing warning row in the current window says "already fired" (debounce). DB errors are best-effort — the chat turn keeps flowing even if the alert insert fails.
- **`db/queries/alerts.sql` + `internal/db/alerts_guardian.sql.go`** — new `GetRecentChatBudgetWarningForUser(recipient, since)` query is the debounce lookup, hand-written binding (sqlc-bypass pattern) folds back into `alerts.sql.go` cleanly at the next sqlc pass.
- **Settings.vue Guardian-usage card** — calls the new endpoint with the current farm_id, renders two compact progress bars (per-user + per-farm) with `bg-emerald-500` below 80%, `bg-amber-400` at 80–100%, `bg-red-500` over 100%. The card hides itself when `ai_enabled=false`, when neither cap is configured AND the user has zero usage, and when the API errors (an inline amber "Couldn't read usage" message renders instead). `chatUsage.load` is called on mount + every farm switch.
- **`ui/src/stores/chatUsage.js`** — Pinia store with `hasAnyCap` / `nearLimit` / `atLimit` derived getters. `load({ farmId })` short-circuits the query-string entirely when `farmId` is 0 / NaN / undefined so the card never sends a malformed request. Treats HTTP 503 specially → flips `aiEnabled` so the card hides on Lite-mode servers without flagging an error.
- **Tests** — 9 unit tests (`internal/farmguardian/budget_warning_test.go`) cover the threshold, debounce hit, debounce-lookup-error fail-closed, SUM-error fail-open, CreateAlert-error fail-open, nil-querier programmer-mistake. 8 real-DB smoke tests (`cmd/api/smoke_phase28_ws5_test.go`) cover the endpoint contract (user / farm / invalid id / foreign farm / no auth / shape) AND the warning hook (fires at 95% → exactly one alert row with the right shape, second call debounces → still exactly one row, below-threshold → zero rows). 11 new Vitest cases (`ui/src/__tests__/chat-usage.test.js`) cover the store loader, derived getters, NaN-ish farmId guard, 503-as-disabled, and 5xx-doesn't-flip-aiEnabled.

### WS6 — OpenAPI parity (shipped 2026-05-20)

- **`openapi.yaml` bumped to `0.3.0`** with a Phase 24–28 changelog block in `info.description` pointing at every new area (rag, crop-cycle-analytics, chat, capabilities) and a back-link to `docs/farm-guardian-architecture.md` for the request-flow primer.
- **New tags:** `crop-cycle-analytics`, `chat`, `capabilities`.
- **New path entries** (all matching the live `cmd/api/routes.go` registrations):
  - `GET /capabilities` (Phase 27 WS6)
  - `GET /crop-cycles/{id}/summary` + `GET /crop-cycles/{id}/summary.csv` (Phase 28 WS1)
  - `GET /farms/{id}/crop-cycles/compare` + `GET /farms/{id}/crop-cycles/compare.csv` (Phase 28 WS1)
  - `POST /v1/chat` (Phase 27 WS5 v1–v4 + Phase 28 WS5 cost guard) — documents both the JSON and `text/event-stream` response variants, the full 400/401/403/405/429/501/502/503/504 matrix, and references the `Retry-After` header.
  - `GET /v1/chat/sessions`, `GET|PATCH|DELETE /v1/chat/sessions/{session_id}` (Phase 27 WS4)
  - `GET /v1/chat/usage` (Phase 28 WS5)
- **New schema components** (24 in total): `Capabilities`, `ChatRequest`, `ChatCitation`, `ChatResponse`, `ChatStreamEvent`, `ChatCostGuardError`, `ChatSessionSummary`, `ChatSessionListResponse`, `ChatSessionTurn`, `ChatSessionDetailResponse`, `ChatSessionPatchRequest`, `ChatSessionPatchResponse`, `ChatUsageDimension`, `ChatUsageFarmDimension`, `ChatUsageResponse`, `CropCycleSummaryFertigation`, `CropCycleSummaryCostTotal`, `CropCycleSummaryCostCategory`, `CropCycleSummaryCost`, `CropCycleSummaryYield`, `CropCycleSummaryStage`, `CropCycleSummary`, `CropCycleCompareResponse`, plus a new reusable `SessionID` UUID path parameter.
- **Parity guard** (`cmd/api/openapi_parity_test.go`) — runs as part of the smoke suite. Scrapes every `mux.Handle("METHOD /path", …)` line out of `routes.go` and confirms each one has a matching `<path>:` entry plus the right verb block in `openapi.yaml`. **130 paths × 159 schemas** at WS6 ship; the test fails loudly on future drift. Allow-list (`routesIntentionallyUndocumented`) is empty — everything must be documented.
- **YAML parse** validated with PyYAML (130 paths, 159 schemas). Strict OpenAPI 3.0.3 validation reports the same pre-existing description-on-response gap in `/farms/{id}/bootstrap-template` that's been there since Phase 13; my new path entries follow the established house style (`{$ref}` style for shared 4xx responses, full inline blocks elsewhere).

### Still open

Phase 28 is **complete** — WS1 ✅ WS2 ✅ WS3 ✅ WS4 ✅ WS5 ✅ WS6 ✅. Suggested follow-ups for Phase 29 live in `docs/workstreams/sit-in-operator-experience.md`.
