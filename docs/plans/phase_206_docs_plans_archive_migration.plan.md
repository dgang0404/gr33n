---
name: Phase 206 — docs/plans archive migration
overview: >
  docs/plans/ has 199 phase-plan files sitting flat (plus 7 era/roadmap hub
  docs). Phase 157 set the precedent — move closed plans to
  docs/plans/archive/, leave nothing dangling — and did it for 5 files
  (phases 88-92) in 2026-06, then stalled. This phase finishes the job with
  a scripted, batch-verified migration instead of hand-editing 199 files.
  Batches are ordered by blast radius: zero-referrer files first (trivial),
  then doc-only referrers (mechanical link fix), then closure-test
  referrers last (mechanical path fix + must re-run each affected test).
todos:
  - id: ws1-inventory-tool
    content: "WS1: scripts/docs-plans-archive-inventory.mjs — read-only categorizer (batch A/B/C); ships in this phase, re-run before each batch since counts drift as other work lands"
    status: completed
  - id: ws2-batch-a
    content: "WS2: Batch A (33 files, zero referrers anywhere) — git mv to docs/plans/archive/, no other changes needed. Lowest risk, do first."
    status: pending
  - id: ws3-batch-b
    content: "WS3: Batch B (109 files, referenced only by non-test docs — mostly docs/phase-14-operator-documentation.md and playbooks) — git mv + sed-fix every inbound relative link from plans/X.plan.md to plans/archive/X.plan.md"
    status: pending
  - id: ws4-batch-c
    content: "WS4: Batch C (57 files touching 59 closure test files) — git mv + fix inbound doc links (same as WS3) + update each test's readFileSync path literal from plans/X.plan.md to plans/archive/X.plan.md; run npm --prefix ui test -- --run after every ~10 files, not just at the end"
    status: pending
  - id: ws5-hub-docs
    content: "WS5: decide fate of the 7 kept-at-top-level hub/roadmap docs (product_backlog_operator_runtime, pre_development_gaps_index, phase_84_100_master_roadmap, phase_68_73_spa_workspace_roadmap, farmer_ux_roadmap_40_plus, phase_53_59_roadmap, phase_173_177_today_excellence_roadmap) — likely stays put since docs/roadmap/README.md (Phase 204) already supersedes them for a first read; keep as deep-link targets"
    status: pending
  - id: ws6-closure-guard
    content: "WS6: phase-206-closure.test.js — docs/plans/ root has 0 (or a documented small number of) *.plan.md files left outside archive/ + the 7 hubs; archive/README.md index updated; make check-ui-test-baseline still green"
    status: pending
isProject: false
---

# Phase 206 — docs/plans archive migration

**Status:** planned · **Depends on:** Phase 205 landed first (don't run a
mechanical mass-move against a test suite that already has 24 unexplained
failures — fix or fully catalog those first so any new failure after a
Batch C move is unambiguously caused by the move)

## Why this is safer than it looks

The original worry (see Phase 204's "explicitly out of scope" note) was
that ~60 closure tests do a direct `readFileSync` against a specific
`docs/plans/phase_N_*.plan.md` path, so moving files would break all of
them. True, but the fix per file is **one line**: the test's path string.
The plan content itself does not change, so there is no risk of a test's
assertion becoming factually wrong — only the join() path argument moves.
`scripts/docs-plans-archive-inventory.mjs` (ships in WS1) makes the exact
list reproducible instead of guessed:

```
docs/plans/ total (excluding kept-at-top-level hubs): 199
  Batch A (zero referrers anywhere)        : 33
  Batch B (other docs/code, no tests)      : 109
  Batch C (referenced by closure tests)    : 57  (touching 59 test files)
  Kept at top level (hub/roadmap docs)      : 7
```

Re-run the script before starting each batch — counts will drift slightly
as other phases land (a plan gaining a new cross-reference moves it from A
to B, etc.) — treat the numbers above as "as of Phase 206 authoring," not
gospel.

## Per-batch procedure

**Batch A** (`--batch=A` to list):
```bash
node scripts/docs-plans-archive-inventory.mjs --batch=A | while read -r f; do
  git mv "docs/plans/$f" "docs/plans/archive/$f"
done
```
No follow-up needed — that's the point of Batch A.

**Batch B** (`--batch=B`): same `git mv`, then for every file that referenced
it (the script's default report doesn't print referrer paths for B; extend
it inline or reuse the `otherReferrers()` helper), replace the substring
`plans/<name>` with `plans/archive/<name>` in that referrer. Almost all
Batch B referrers are `docs/phase-14-operator-documentation.md` and a
handful of operator playbooks — expect this to be a handful of sed passes,
not 109 manual edits.

**Batch C** (`--batch=C`, prints the test file(s) per plan): same two steps
as Batch B, plus: in each listed test file, replace the literal
`'plans/<name>'` with `'plans/archive/<name>'` inside the `join(...)` call.
Run the affected test files immediately after each small group (suggest
groups of ~10) — a failure at this stage means the sed missed a variant
(different quote style, string built via template literal, etc.), not a
real regression, but confirm that before moving on.

## What does NOT move

The 7 hub/roadmap docs (WS5) — these are era summaries meant for
navigation, not closed single-phase implementation logs. `docs/roadmap/README.md`
(Phase 204) is now the first stop for "what shipped when" in prose; these
stay as the next layer down for anyone who wants the workstream-level
detail for one era. Re-evaluate in a future phase once `docs/roadmap/README.md`
has been the front door for a while — if nobody clicks past it, this list
can move too.

## Acceptance criteria

- [ ] Batch A moved (33 files or the then-current count), zero other changes
- [ ] Batch B moved with all inbound links fixed; no dead links in
      `docs/phase-14-operator-documentation.md` or playbooks (spot check
      with a markdown link checker or manual grep for `plans/<name>` minus
      `plans/archive/<name>`)
- [ ] Batch C moved with all inbound links + test path literals fixed;
      full `npm --prefix ui test -- --run` shows no new failures beyond
      Phase 205's (shrinking) baseline
- [ ] `docs/plans/archive/README.md` index extended to cover the new
      arrivals (or a generated list, so it doesn't go stale like the
      original Phase 157 attempt did)
- [ ] `phase-206-closure.test.js`
- [ ] `make check-ui-test-baseline` green

## Out of scope

- Deleting any plan content — this is a pure move, nothing is deleted.
- The 7 hub docs (see WS5).
- Any plan file created *during* this phase's own execution (i.e. this
  file and Phase 205's) — archive those in a later phase once they're old
  news, same as everything else.
