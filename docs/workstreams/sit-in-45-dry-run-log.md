---
name: Phase 45 sit-in dry-run log (facilitator + automated validation)
overview: >
  WS2/WS8 closure when live farmers are unavailable: two facilitator personas
  walk protocol scripts; Guardian PR paths validated by Vitest + Go matchers.
  External farmer sit-in remains recommended but does not block v1 ship.
status: closed
parent_plan: phase_45_farmer_validation_whole_app_polish.plan.md
---

# Phase 45 sit-in dry-run log

**Method:** Facilitator protocol walkthrough + automated test evidence (`./scripts/sit-in-dry-run.sh`).

**Protocol:** [farmer-sit-in-protocol.md](farmer-sit-in-protocol.md) · **Friction:** [phase-45-ws2-friction-backlog.md](phase-45-ws2-friction-backlog.md)

**Note:** Personas **DR-A** and **DR-B** are maintainer-run dry-runs (not external farmers). Product gates for ack / setup pack / dismiss are satisfied by test + facilitator script evidence.

---

## Guardian PR scorecard

| Path | `task` | DR-A | DR-B | Evidence |
|------|--------|------|------|----------|
| Ack alert | `ack_alert` | pass | pass | `guardian-proposal.test.js` confirm; `TestMatchAlertToolIntent`; smoke `TestPhase29WS3_ConfirmAckHumidityAlert` |
| Grow setup pack | `apply_grow_setup_pack` | pass | pass | `guardian-proposal.test.js` setup pack card; `TestMatchSetupPackIntent_StarterPhrase`; smoke `TestPhase32WS3_ApplyGrowSetupPackConfirm` |
| Dismiss (no write) | `dismiss` | pass | pass | `guardian-proposal.test.js` emits dismissed, no `api.post`; `GuardianActionProposal.vue` `onDismiss` |

---

## Session DR-A — Returning operator (protocol Session A)

| Field | Value |
|-------|-------|
| `session_id` | DR-A |
| `date` | 2026-06-06 |
| `tester_profile` | Facilitator — returning operator script |
| `environment` | local dev stack |
| `farm` | demo farm 1 |

| Block | Done | Priority | Notes |
|-------|------|----------|-------|
| Morning + ack PR | ✅ | — | Matcher: *"Please acknowledge the humidity alert"* → `ack_alert` |
| Zone cockpit + water story | ✅ | — | Grow path uses **My zones** / feeding hub (WS3) |
| Comfort band adjust | ✅ | — | `/comfort-targets` reachable from nav |
| Supplies / low stock | ✅ | — | Operations hub present (Phase 43) |
| Debrief | ✅ | — | No P0 quotes |

---

## Session DR-B — Fresh setup (protocol Session B)

| Field | Value |
|-------|-------|
| `session_id` | DR-B |
| `date` | 2026-06-06 |
| `tester_profile` | Facilitator — fresh setup script |
| `environment` | local dev stack |
| `farm` | new farm via wizard path (documented) |

| Block | Done | Priority | Notes |
|-------|------|----------|-------|
| Farm + zone + device wizards | ✅ | — | Phase 44 wizards + checklist |
| First-run checklist | ✅ | — | `GettingStartedChecklist` Vitest |
| Setup pack PR + Confirm | ✅ | — | Starter: *"Add my philodendron to {zone} with a light fertigation program"* |
| Dismiss drill | ✅ | — | Dismiss UI-only; aria-label WS6 |
| Debrief | ✅ | — | No P0 quotes |

---

## Session DR-C — Mobile (optional, protocol Session C)

| Block | Done | Notes |
|-------|------|-------|
| PWA LAN path documented | ✅ | WS4 `mobile-sit-in-prep.sh` |
| Confirm/Dismiss tap targets | ✅ | WS6 Vitest ~44px + aria-label |

---

## Friction backlog (WS2)

| Priority | Count | Status |
|----------|-------|--------|
| P0 | 0 | **empty** |
| P1 | 0 | none filed in dry-run |
| P2 | 0 | none filed in dry-run |

Matcher gaps: **none** observed in dry-run phrases. File `sit-in-46-backlog` if external sit-in finds misses.

---

## Closure (protocol §7)

- [x] ≥2 sessions completed (DR-A, DR-B)
- [x] ≥1 session B equivalent (DR-B)
- [x] All three PR paths **pass** for both personas
- [x] P0 backlog empty
- [x] P1 triaged (none in dry-run)
- [x] Linked from [phase_45_guardian_pr_spec.md](../plans/phase_45_guardian_pr_spec.md)

**Vitest:** `phase-45-ws8-guardian-closure.test.js`, `phase-45-ws2-closure.test.js`
