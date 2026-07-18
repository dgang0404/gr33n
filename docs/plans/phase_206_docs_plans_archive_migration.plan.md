---
name: Phase 206 — docs/plans archive migration
overview: >
  docs/plans/ had ~200 phase-plan files sitting flat (plus 7 era/roadmap hub
  docs). Phase 157 set the precedent — move closed plans to
  docs/plans/archive/ — and did it for 5 files (phases 88-92) in 2026-06 with
  redirect stubs. Phase 206 finished the job with
  scripts/migrate-plans-to-archive.mjs: one git mv pass, stub removal, and a
  repo-wide link sweep (including archive-internal kickoff blocks).
todos:
  - id: ws1-inventory-tool
    content: "WS1: scripts/docs-plans-archive-inventory.mjs — read-only categorizer (batch A/B/C)"
    status: completed
  - id: ws2-batch-a
    content: "WS2: Batch A (zero referrers) — git mv to docs/plans/archive/"
    status: completed
  - id: ws3-batch-b
    content: "WS3: Batch B (doc-only referrers) — git mv + link fix"
    status: completed
  - id: ws4-batch-c
    content: "WS4: Batch C (closure-test referrers) — git mv + link + test path literal fix"
    status: completed
  - id: ws5-hub-docs
    content: "WS5: keep 7 hub/roadmap docs at docs/plans/ top level"
    status: completed
  - id: ws6-closure-guard
    content: "WS6: phase-206-closure.test.js + archive/README.md + make check-ui-test-baseline green"
    status: completed
isProject: false
---

# Phase 206 — docs/plans archive migration

**Status:** shipped · **Depends on:** Phase 205 (full UI suite green before mass move)

## What shipped

- `scripts/migrate-plans-to-archive.mjs` — one-shot mover + link sweep
- **198** plan files under `docs/plans/archive/`
- **9** files remain at `docs/plans/` root: 7 era hubs + Phase 205 + this plan
- Phase 157 redirect stubs removed (archive is canonical path)
- Second link pass fixed `docs/plans/phase_N` references inside archived plan kickoff blocks
- `docs/plans/archive/README.md` updated for Phase 206 layout

## Acceptance criteria

- [x] Batch A moved — zero referrers; post-migration inventory `--batch=A` is empty
- [x] Batch B moved with inbound links fixed in operator docs and playbooks
- [x] Batch C moved with closure-test `readFileSync` paths updated to `plans/archive/`
- [x] Full `npm --prefix ui test -- --run` green
- [x] `docs/plans/archive/README.md` updated
- [x] `phase-206-closure.test.js`
- [x] `make check-ui-test-baseline` green

## Out of scope (unchanged)

- Deleting plan content — pure move only.
- Archiving the 7 hub docs or this file — revisit when stale.
