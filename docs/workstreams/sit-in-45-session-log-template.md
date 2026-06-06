---
name: Phase 45 sit-in session log (facilitator copy-paste)
overview: >
  Blank scorecard for farmer-sit-in-protocol.md sessions A/B/C and the three
  required Guardian PR paths. Copy into a spreadsheet or GitHub issue.
status: shipped
parent_plan: phase_45_farmer_validation_whole_app_polish.plan.md
---

# Phase 45 sit-in session log (template)

**Protocol:** [farmer-sit-in-protocol.md](farmer-sit-in-protocol.md) · **Guardian PR detail:** [phase_45_guardian_pr_spec.md](../plans/phase_45_guardian_pr_spec.md)

**Labels:** `sit-in-45` (session notes) · `sit-in-46-backlog` (matcher gaps only)

---

## Session header

| Field | Value |
|-------|-------|
| `session_id` | A1 / B1 / C1 |
| `date` | |
| `tester_profile` | e.g. indoor veg, no SQL |
| `environment` | local / staging |
| `farm` | demo farm 1 / new farm id |
| `facilitator` | |
| `observer` | |

---

## Guardian PR scorecard (required)

Mark **pass / fail / skip** per tester per path. All three must **pass** for ≥2 testers before the **farmer-ready v1 product claim** (OC-45 docs closed in WS7).

| Path | `task` | Tester 1 | Tester 2 | Tester 3 | Notes |
|------|--------|----------|----------|----------|-------|
| Ack alert | `ack_alert` | | | | |
| Grow setup pack | `apply_grow_setup_pack` | | | | |
| Dismiss (no write) | `dismiss` | | | | |

### Path detail (copy per failure)

```text
session_id:
task: ack_alert | apply_grow_setup_pack | dismiss
result: pass | fail | skip
blocker: P0 | P1 | P2 | —
route: /dashboard | /zones/{id} | Guardian drawer
matcher_gap: (phrase that should have proposed) | —
quote: (verbatim)
time_sec: (optional)
fix_owner: WS2 | Phase 46 | WS3 | WS6
```

---

## Session A — Returning operator

| Block | Done | P0/P1/P2 | Quote |
|-------|------|----------|-------|
| Morning + ack PR | ☐ | | |
| Zone cockpit + water story | ☐ | | |
| Comfort band adjust | ☐ | | |
| Supplies / low stock | ☐ | | |
| Debrief | ☐ | | |

---

## Session B — Fresh setup

| Block | Done | P0/P1/P2 | Quote |
|-------|------|----------|-------|
| Farm + zone + device wizards | ☐ | | |
| First-run checklist | ☐ | | |
| Setup pack PR + Confirm | ☐ | | |
| Dismiss drill | ☐ | | |
| Debrief | ☐ | | |

---

## Session C — Mobile (optional)

| Block | Done | P0/P1/P2 | Quote |
|-------|------|----------|-------|
| Dashboard + ack on phone | ☐ | | |
| Confirm / Dismiss tap targets | ☐ | | |
| Debrief | ☐ | | |

---

## Friction backlog (WS2 triage)

| ID | Priority | Route | Summary | Owner |
|----|----------|-------|---------|-------|
| 1 | P0/P1/P2 | | | |
| 2 | | | | |

---

## Closure checklist (protocol §7)

- [ ] ≥2 sessions A completed
- [ ] ≥1 session B completed
- [ ] All three PR paths pass for ≥2 testers (or documented skip + fix)
- [ ] P0 backlog empty
- [ ] P1 triaged (fix or defer with reason)
- [ ] Matcher gaps filed (`sit-in-46-backlog`)
