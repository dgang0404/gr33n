---
name: Phase 61 — Guardian proactive nudges
overview: >
  Guardian surfaces contextual suggestions without the user asking — a gentle
  dot on the robot tab icon when something worth noting crossed a threshold.
  Not notifications spam; one nudge per category per session, cleared when read.
todos:
  - id: ws1-nudge-engine
    content: "WS1: Go nudge poller — compute nudge score per farm per session; categories"
    status: completed
  - id: ws2-dot-badge
    content: "WS2: Robot tab dot badge; Guardian panel nudge strip above starter chips"
    status: completed
  - id: ws3-dismiss
    content: "WS3: Dismiss / snooze nudge until next session; no re-nudge same category"
    status: completed
  - id: ws4-guardian-context
    content: "WS4: context_ref nudge framing; snooze state in guardianPanel store"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: operator-tour; phase-61-closure.test.js; OC-61"
    status: completed
isProject: false
---

# Phase 61 — Guardian proactive nudges

## Status

**Shipped.** Rule-based `GET /farms/{id}/guardian-nudge`, amber dot on edge tab + TopBar, dismissable nudge strip, `nudge_category` on `context_ref` when Review is tapped.

---

## The one job

> **Guardian taps you on the shoulder once — you decide if you care.**

---

## Nudge categories (one active at a time, priority order)

| Priority | Category | Trigger | Example nudge |
|----------|----------|---------|--------------|
| 1 | Critical alert unread | Severity ≥ warn, >15 min | "Humidity alert in Flower Room — tap to review" |
| 2 | Feed missed | Schedule time passed, no run logged | "9am feed in Veg Tent hasn't run — is that intentional?" |
| 3 | Comfort band breached | Sensor reading outside band >30 min | "Clone Room temp has been above max for 32 minutes" |
| 4 | Pi stale | `last_config_fetch_at` > 2h | "Veg Tent Pi hasn't checked in — worth a look" |
| 5 | Low stock | Batch below threshold | "CalMag is almost out — create a refill task?" |

---

## WS1 — Nudge engine (Go)

```go
GET /farms/{id}/guardian-nudge  // lightweight, called on panel open
```

Returns at most one `NudgePayload`:

```json
{
  "category": "critical_alert",
  "message": "Humidity alert in Flower Room — tap to review",
  "severity": "warn",
  "action_route": "/alerts",
  "nudge_id": "alert-99"
}
```

Stateless — client tracks dismiss in `guardianPanel` store.

---

## WS2 — UI: dot badge + nudge strip

**Robot tab icon:** amber dot when nudge available; clears when panel opened.

**Guardian panel top of form:**
```
⚠ Humidity alert in Flower Room — tap to review   [Review] [Dismiss]
```

- `Review` → sends a starter message to Guardian + clears dot
- `Dismiss` → snoozes category for session

---

## WS3 — Dismiss / snooze

- Snooze stored in `guardianPanel` Pinia store (session memory, no server round-trip)
- Re-appears next page load if issue persists
- At most one nudge per category per page load; never stacks

---

## WS4 — Guardian context

When user taps Review, `context_ref` includes `nudge_category` so Guardian frames response directly:

```
User is reviewing a Guardian nudge about: critical_alert (alert_id: 99).
Skip pleasantries — address the specific issue immediately.
```

---

## WS5 — Docs, tests, OC-61

- Dot badge renders when API returns nudge
- Dismiss clears dot until next reload
- operator-tour "Proactive nudges" note

---

## Design boundaries

- **Not push notifications** — page-load only, no browser notification API
- **Not AI-generated nudges** — rule-based only; LLM only responds *after* user engages
- **One nudge at a time** — no badges stacking; no anxiety loop

---

## Definition of done

- [x] Dot on robot icon when nudge present
- [x] One nudge strip, dismissable
- [x] OC-61 closed
