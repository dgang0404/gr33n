---
name: Phase 159 — Guardian citation completeness + turn persistence
overview: >
  Finish Phase 152 WS2b (remaining citation deep-link source types), persist
  accuracy_note on conversation_turns so trust banners survive reload, and
  refresh current-state.md after the 154–158 arc.
todos:
  - id: ws1-schedule-routes
    content: "WS1: schedule citations → zone via fertigation_program.schedule_id or automation_rule (first match, farm-scoped)"
    status: completed
  - id: ws2-alert-routes
    content: "WS2: alert_notification citations → zone via triggering_event hop (sensor/rule → zone); route /zones/:id?tab=ops&ops=alerts"
    status: completed
  - id: ws3-doc-routes
    content: "WS3: field_guide/platform_doc → /farm-knowledge?source_id= or /symptom-guide with minimal deep-link read"
    status: completed
  - id: ws4-accuracy-persist
    content: "WS4: migration accuracy_note TEXT on conversation_turns; wire insert + session list + UI reload"
    status: completed
  - id: ws5-current-state
    content: "WS5: Update docs/current-state.md — 154–158 shipped, remove stale 115/158 rows"
    status: completed
  - id: ws6-closure
    content: "WS6: Go tests + phase-159-closure.test.js; update phase_152 plan WS2b → completed"
    status: completed
isProject: false
---

# Phase 159 — Guardian citation completeness + turn persistence

**Status:** shipped · **Hub:** [159–160 backlog](phase_159_160_post_158_gaps_backlog.plan.md) · **Continues:** [152](phase_152_guardian_live_accuracy_guardrails.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `schedule` → zone via program `schedule_id` or actuator executable action |
| **WS2** | `alert_notification` → `/zones/:id?tab=ops&ops=alerts` via sensor/rule/program hop |
| **WS3** | `field_guide` / `platform_doc` → symptom-guide, farm-knowledge, or operator-guide with `cited_doc` |
| **WS4** | `accuracy_note` column + persist/reload on session API |
| **WS5** | `current-state.md` refreshed (154–159) |
| **WS6** | `phase-159-closure.test.js` + Go unit tests |

## Close when

- [x] Citation chips link for schedule, alert, field_guide, platform_doc when resolvable
- [x] `accuracy_note` survives session reload
- [x] `current-state.md` reflects shipped arcs
- [x] Phase 152 WS2b marked complete
