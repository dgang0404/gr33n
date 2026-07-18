---
name: Phase 204 — Docs navigation cleanup (README + roadmap)
overview: >
  Operator feedback: README doesn't explain what the app does without
  reading through dozens of phase docs, and there was no single readable
  roadmap — phase history was scattered across the README, phase-14 index,
  and per-era plan hub docs. This phase makes the README product-first,
  creates one narrative roadmap doc, and retires a pile of duplicate
  Phase 60 closure docs found during the pass.
todos:
  - id: ws1-roadmap-doc
    content: "WS1: docs/roadmap/README.md — one page, every shipped era in plain language, links out to phase-14 index + docs/plans/ for detail"
    status: completed
  - id: ws2-readme-rewrite
    content: "WS2: README opening rewritten product-first (what it does, not phase ranges); Roadmap & history section points to docs/roadmap/README.md as the primary link; trimmed stray phase-number badges from feature bullets"
    status: completed
  - id: ws3-phase60-dedup
    content: "WS3: delete OC-60-CLOSURE.md, PHASE-60-BUILD-SUMMARY.md, PHASE-60-QUICK-REFERENCE.md, phase-60-implementation-checklist.md, PHASE-60-IMPLEMENTATION-COMPLETE.md — fully superseded by docs/plans/archive/phase_60_pi_setup_wizard_ux.plan.md; verified no code/test/doc references before deleting"
    status: completed
  - id: ws4-closure-guard
    content: "WS4: phase-204-closure.test.js — roadmap doc exists and covers current era, README points at it, deleted phase-60 docs stay gone, existing README closure tests (45/46/59/157) still pass"
    status: completed
isProject: false
---

# Phase 204 — Docs navigation cleanup

**Status:** shipped · **Depends on:** none (docs janitorial)

## The problem

Two operator complaints after Phase 202/203:

1. The README explained features by citing phase ranges ("Phases 173-177",
   "Phases 183-187", …) instead of just describing the product — a new
   reader had to already know the phase history to parse it.
2. There was no single "read this and you understand the roadmap" doc.
   History was split across the README status line, `phase-14-operator-documentation.md`
   (an exhaustive 800+ line index), and several era-specific `*_roadmap.plan.md`
   hub docs. Finding "what shipped and when" meant opening several files.

A repo-wide pass also turned up a fully duplicated Phase 60 doc pile:
`OC-60-CLOSURE.md`, `PHASE-60-BUILD-SUMMARY.md`, `PHASE-60-QUICK-REFERENCE.md`,
`phase-60-implementation-checklist.md`, `PHASE-60-IMPLEMENTATION-COMPLETE.md` —
five files (~1,700 lines) all narrating the same shipped feature already
covered by the canonical `docs/plans/archive/phase_60_pi_setup_wizard_ux.plan.md`.

## What shipped

- **`docs/roadmap/README.md`** — the new single roadmap: one row per era
  (Foundation, Farmer UX, SPA workspaces, crop intelligence, Guardian
  hardening, Virtual Pi, Today cockpit, sit-in arc, answer-quality audit,
  janitorial consolidation, …) written as plain-language summaries, not a
  phase-by-phase table. Links out to `phase-14-operator-documentation.md`
  and `docs/plans/` only for readers who want implementation detail.
- **README.md** — opening paragraph now states what gr33n *does* before
  any phase number appears. Status line and "Roadmap & history" section
  both point at `docs/roadmap/README.md` as the one link to follow. Kept
  the two test-guarded call-outs (Phase 45 farmer-ready v1, Phase 46
  Guardian LLM proposals) and the `current-state.md` / `enterprise-tier-boundary.md`
  links intact — closure tests for phases 45/46/59/157 still assert on
  those exact strings.
- **Deleted** the five duplicate Phase 60 docs. Confirmed first that no
  test, code, or other doc referenced them (`grep -rn` across the repo);
  the design doc and OpenAPI docs already point at the canonical plan.

## Explicitly out of scope (deferred)

`docs/plans/` has ~200 phase plan files, and roughly 60 UI closure tests
do a `readFileSync` + `toContain` check directly against specific
`docs/plans/phase_N_*.plan.md` paths. Bulk-archiving those files (like the
existing `docs/plans/archive/` pattern used for phases 88-92 in Phase 157)
would break every one of those tests unless done file-by-file with the
test updated in the same commit. That's real work with real risk and is
its own phase, not a docs-wording pass — left for a future phase if
navigating `docs/plans/` directly (rather than through the new roadmap
doc) is still a problem in practice.

## Acceptance criteria

- [x] `docs/roadmap/README.md` exists, covers eras through current phase
- [x] README explains the product before citing any phase number
- [x] README "Roadmap & history" section links `docs/roadmap/README.md` first
- [x] Phase 60 duplicate doc pile removed, zero dangling references
- [x] Existing README-dependent closure tests (45, 46, 59, 157) still pass
- [x] `phase-204-closure.test.js`
