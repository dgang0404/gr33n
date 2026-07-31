---
name: Phase 213 — CI keep green (fix gates when code changes)
overview: >
  GitHub Actions CI on main has been red for consecutive pushes since ~2026-07-29.
  Failures are mostly stale closure assertions and OpenAPI drift — not one product
  bug. This phase restores green CI and locks in a rule: when you change code that
  CI audits (routes, glossary links, phase plans, OpenAPI), update the matching
  gate in the same change (or the next commit before more feature work).
todos:
  - id: ws1-openapi-seed-pending
    content: "WS1: Document POST /v1/chat/proposals/seed-pending in openapi.yaml (211.07) so make audit-openapi passes"
    status: pending
  - id: ws2-phase76-feedwater
    content: "WS2: Update phase-76-closure (+ any leftover zone-water expectations) for feedWaterRoute → /feed-water"
    status: pending
  - id: ws3-closure-drift
    content: "WS3: Clear remaining red UI closures from the Jul 29–30 streak (phase-143/200 AccuracyNote / docs strings if still failing on main)"
    status: pending
  - id: ws4-verify-ci
    content: "WS4: Local make audit-openapi + npm --prefix ui test -- --run (or check-ui-test-baseline) green; push; confirm Actions go+ui jobs green"
    status: pending
  - id: ws5-process-link
    content: "WS5: Roadmap + CONTRIBUTING already point here — confirm agents/humans treat CI red as a ship blocker before the next feature phase"
    status: pending
isProject: false
---

# Phase 213 — CI keep green (fix gates when code changes)

**Status:** planned · **Priority:** before LED blink / auth hardening / next feature phase  
**Indexed on:** [`docs/roadmap/README.md`](../roadmap/README.md) (What's next)

## The rule (why this exists)

> **If you change code that CI checks, update the CI gate in the same PR/push.**

| You changed… | Also update… |
|--------------|--------------|
| `cmd/api/routes.go` new path | `openapi.yaml` + `make audit-openapi` |
| Dashboard / workspace deep links | Vitest that still asserts the old path (`phase-76-closure`, etc.) |
| New `docs/plans/phase_N_*.plan.md` at root | `ACTIVE_FEATURE_AT_TOP` in `phase-206-closure.test.js` |
| Guardian answer / eval JSON shape | `phase-143` / `phase-200` (and similar) string assertions |
| UI behavior shared helpers | Full `npm --prefix ui test -- --run` or at least `make check-ui-test-baseline` |

Running only the new phase's own closure test is how `main` went red for a week while every push still "felt fine" locally.

## Known reds as of 2026-07-31 (CI #379 / `1511a5a`)

| Job | Failure | Fix |
|-----|---------|-----|
| **go** | `make audit-openapi` — `POST /v1/chat/proposals/seed-pending` in routes, missing from OpenAPI | WS1 |
| **ui** | `phase-76-closure` expects `feedWaterRoute` → `/zones/3?tab=water`; product now uses `/feed-water` | WS2 |
| **earlier streak** | `phase-206` allowlist, then `phase-143` / `phase-200` AccuracyNote/docs drift | WS3 (verify still failing after WS1–2) |

Guardian / browser-e2e / hardware jobs stay **skipped** until go+ui are green — they are not the current blocker.

## Out of scope

- New product features, Guardian prompt rewrites, dual-farm re-runs
- Making Guardian LLM smokes mandatory on every push (stays label-gated — [`ci-guardian-qa.md`](../ci-guardian-qa.md))
- Replacing Phase 205's baseline mechanism — reuse it

## Close when

- [ ] `make audit-openapi` green locally and on Actions
- [ ] UI CI job green (no phase-76 / leftover closure failures from the streak)
- [ ] Roadmap "What's next" no longer lists CI as the active fire
- [ ] Next feature push starts from a green Actions baseline

## Related

- Phase 205 (shipped): [`phase_205_pre_existing_test_debt.plan.md`](phase_205_pre_existing_test_debt.plan.md) — baseline safety net
- CONTRIBUTING.md — full UI suite + `audit-openapi` before ship
- CI workflow: [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)
