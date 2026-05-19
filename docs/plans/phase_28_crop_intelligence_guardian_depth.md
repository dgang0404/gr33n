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
    status: pending
  - id: ws2-analytics-ui
    content: "WS2: CropCycleSummary.vue + CropCycleCompare.vue — metric cards, stage timeline, compare columns, SideNav entry, link from cycle list"
    status: pending
  - id: ws3-guardian-analytics
    content: "WS3: Guardian ↔ crop cycle integration — Guardian can answer questions about the current/historical cycles; snapshot includes active-cycle summary metrics when farm_id present"
    status: pending
  - id: ws4-guardian-alerts
    content: "WS4: Guardian alert integration — Guardian can explain unread alerts (rule triggered, sensor threshold, what to do); alert context injected into grounded snapshot"
    status: pending
  - id: ws5-cost-dashboard
    content: "WS5: Token-usage dashboard — operator-visible per-user/per-farm rolling totals in Settings; alert hook that fires a notification when >80% of budget is consumed"
    status: pending
  - id: ws6-openapi-parity
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
