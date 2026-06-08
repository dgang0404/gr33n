---
name: Phase 60 — Guardian morning walkthrough
overview: >
  A structured daily-check flow: one button press puts Guardian into a
  guided sequence — alerts, feed schedule, offline devices, comfort bands,
  low stock — delivered in farmer language without the operator knowing
  what to ask. The walkthrough adapts to what's actually wrong today.
todos:
  - id: ws1-walkthrough-flow
    content: "WS1: Go — walk_farm endpoint: aggregate alerts + feed schedule + offline + low-stock into ordered findings"
    status: pending
  - id: ws2-ui-starter
    content: "WS2: Dashboard + Guardian panel 'Morning check' chip; progress steps UI"
    status: pending
  - id: ws3-smart-skip
    content: "WS3: Skip categories with nothing to report; surface only actionable findings"
    status: pending
  - id: ws4-guardian-context
    content: "WS4: context_ref framing for walkthrough route; Guardian persona morning-check copy"
    status: pending
  - id: ws5-docs-tests
    content: "WS5: operator-tour § morning walkthrough; phase-60-closure.test.js; OC-60"
    status: pending
isProject: false
---

# Phase 60 — Guardian morning walkthrough

## Status

**Planned.** After [Phase 55](phase_55_guardian_ops_grow_money.plan.md) read tools ship (feeds walkthrough with real data).

**Arc:** [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md)

---

## The one job

> **One tap, and Guardian tells you what actually needs attention today — you don't have to know what to ask.**

---

## What the walkthrough covers (in priority order)

| # | Category | Skip if… | Example finding |
|---|----------|----------|-----------------|
| 1 | Unacknowledged alerts | No alerts | "Humidity alert in Flower Room since 6am — outside your 65–75% band" |
| 2 | Feed schedule today | No active feed zone | "Flower Room fertigation runs at 9am — reservoir at 68%, good" |
| 3 | Offline / stale devices | All devices online | "Pi in Veg Tent last seen 4h ago — may need reconnect" |
| 4 | Comfort bands out of range | All bands OK | "Temp in Clone Room is 29°C, above your 25° max" |
| 5 | Low stock | All batches above threshold | "CalMag batch B-12 has 0.8L left — below 1L minimum" |
| 6 | Summary + top action | Always | "Two things need attention. Start with the alert." |

---

## WS1 — Backend: `walk_farm` read tool

New read tool (not a change tool — read-only):

```go
func WalkFarm(ctx, farmID) WalkFarmResult {
    alerts     := unreadUnackedAlerts(farmID)
    feeds      := todayFeedSchedules(farmID)
    offline    := offlineDevices(farmID)
    bands      := comfortBandsOutOfRange(farmID)
    lowStock   := lowStockBatches(farmID)
    return rank(alerts, feeds, offline, bands, lowStock)
}
```

Returns ordered `findings[]` with `category`, `severity` (warn/ok), `plain_text`, `action_route`.

---

## WS2 — UI

**Dashboard strip:**
- "Morning check" chip → opens Guardian panel + sends walkthrough message

**Guardian panel (full page):**
- "Morning walkthrough" starter always visible on `/chat`
- Progress indicator: "Checking 5 areas…" → findings list with severity badges

**Compact panel:**
- Single "Run morning check" chip when no active session; findings render as chat turn

---

## WS3 — Smart skip

- Category with zero findings → Guardian skips, doesn't say "Alerts: none. Feed: none…"
- If everything is OK → single positive summary: "Farm looks good this morning — one schedule runs at 9am"

---

## WS4 — Guardian context + persona

`context_ref.go` route hint for `/chat` with `guardian_mode: 'morning_walkthrough'`:

```
You are doing a morning walkthrough for [farm name].
Report only what needs attention — skip categories with nothing to flag.
Use plain language. No schema terms. Cite zone name, not zone_id.
```

---

## WS5 — Docs, tests, OC-60

- operator-tour "Morning walkthrough" paragraph (§6i)
- `phase-60-closure.test.js` — walkthrough starter present; findings render
- Guardian PR spec note: walkthrough never proposes changes (read-only)

---

## Definition of done

- [ ] "Morning check" chip on Dashboard
- [ ] Guardian finds and ranks real farm issues in one message
- [ ] Empty farm returns positive summary, not empty bullets
- [ ] OC-60 closed
