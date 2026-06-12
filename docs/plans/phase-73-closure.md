# Phase 73 — closure (OC-73)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_73_guardian_pr_discoverability.plan.md`](phase_73_guardian_pr_discoverability.plan.md)

**Depends on:** Guardian proposal system (Phases 29–46, 55); [Phase 68](phase_68_workspace_shell_spa_nav.plan.md) workspace shell.

**Closes:** Last phase of the [68–73 SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md).

---

## The one job (done)

> **Guardian change requests are easy to find and act on**, empty zones get a proactive setup offer, dismiss persists on the server, and read tools fire on natural phrasing with clear “set this up” messaging when data is missing.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Global pending badge on nav + TopBar | `GuardianNavLaunch.vue`, `TopBar.vue`, `guardianProposals.js` |
| **WS2** | Empty-zone grow nudge | `EmptyZoneGrowNudge.vue`, `POST …/suggest-empty-zone` |
| **WS3** | Server-side dismiss | `POST …/proposals/{id}/dismiss`, `DismissGuardianProposal` SQL |
| **WS4** | Read-tool widening + missing-coords copy | `readtools_weather.go` |
| **WS5** | Settings farm-site copy for Guardian weather | `Settings.vue` |
| **WS6** | Tests + docs | `smoke_phase73_test.go`, `phase-73-closure.test.js` |

---

## API routes

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/v1/chat/proposals/{id}/dismiss` | Persist dismiss; badge stays accurate |
| `POST` | `/v1/chat/proposals/suggest-empty-zone` | Proactive `apply_grow_setup_pack` for empty zone |

---

## Automated tests

| Test | Path |
|------|------|
| Pending badge, dismiss API, empty-zone nudge | `ui/src/__tests__/phase-73-closure.test.js` |
| Dismiss calls API | `ui/src/__tests__/guardian-proposal.test.js` |
| Empty-zone propose + dismiss smoke | `cmd/api/smoke_phase73_test.go` |
| Weather intent phrasing | `internal/farmguardian/readtools_weather_test.go` |

---

## OC-73

Phase 73 is **closed** when the global pending count is visible, empty zones surface a Confirm-able setup proposal, dismiss persists server-side, read tools fire on natural supplemental-light phrasing, and missing coordinates produce actionable setup guidance.
