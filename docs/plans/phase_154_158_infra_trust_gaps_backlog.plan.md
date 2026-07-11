---
name: Phases 154–158 — infra & trust gaps backlog (candidate, unscoped)
overview: >
  Gap-analysis pass requested directly by the operator (post-153): "look at the whole
  app, tell me what it actually needs, push back if something was unnecessary."
  These are candidate phases, not committed work — pick which ones earn a real
  phase number. None of this is Guardian-answer-quality work (148–153 covers that);
  this is app-level trust/durability infrastructure that has no owner yet.
todos:
  - id: verdict-pr-ci-gate
    content: "Verdict on the reverted GitHub PR CI gate — see verdict section below. Not bad practice, just not-yet-consented-to scope creep."
    status: completed
  - id: phase-154-test-suite-health
    content: "Phase 154 — fix `go test ./...` red/build-broken packages at repo root; add `make test-unit` gate"
    status: completed
  - id: phase-155-automated-backups
    content: "Phase 155 (candidate) — turn backup-restore-runbook.md from a manual doc into a cron script + integrity check + retention"
    status: pending
  - id: phase-156-dependency-scanning
    content: "Phase 156 (candidate) — govulncheck + npm audit + Dependabot, wired as an opt-in/non-blocking CI job (same label-gated pattern as the reverted PR gate, done with explicit sign-off this time)"
    status: pending
  - id: phase-157-docs-consolidation
    content: "Phase 157 (candidate) — snapshot a single current-state doc distinct from the 150+ phase-history trail; archive folder for closed phases"
    status: pending
  - id: phase-158-accessibility-pass
    content: "Phase 158 (candidate, lower priority) — keyboard nav + screen-reader pass on Guardian chat and core workspaces"
    status: pending
isProject: false
---

# Phases 154–158 — infra & trust gaps backlog

**Ask that produced this doc:** operator asked, deliberately open-ended, "is the app missing something, push back if I'm wrong, don't code anything — just tell me / make a phase doc." This is that doc. Nothing below has been implemented.

---

## Verdict: was the reverted GitHub PR CI gate "bad"?

No. Pushing back as asked: a label-gated (or path-gated), non-blocking CI job that runs slow/model-dependent smoke tests only when explicitly triggered is a **standard, common pattern** — plenty of repos that need GPU/self-hosted runners or expensive integration tests gate them behind a label instead of running on every push. It was reasonable engineering.

It got reverted for one reason only: the operator didn't ask for it and hadn't consented to CI scope changing that day. That's a consent/scope problem, not a quality problem. Worth knowing so it isn't dismissed as "atypical" — it's fine to bring back later as **Phase 156** below, this time proposed explicitly instead of bundled into an unrelated ask.

---

## Phase 154 (candidate) — Test suite health at repo root

**Evidence, not speculation** — ran `go test ./...` from repo root just now:

```
FAIL  gr33n-api/cmd/api
FAIL  gr33n-api/internal/cropcycle          (panic — DefaultCatalog() nil deref)
FAIL  gr33n-api/internal/croplibrary
FAIL  gr33n-api/internal/farmguardian
FAIL  gr33n-api/internal/handler/chat
FAIL  gr33n-api/internal/handler/device
FAIL  gr33n-api/internal/handler/sensor     [build failed — mock querier doesn't
                                              implement db.Querier anymore]
FAIL  gr33n-api/internal/rag/ingest
```

Some of these are DB-required tests failing outside a live-DB harness (expected/known). But **`internal/handler/sensor` doesn't compile at all** — `wiringMockQuerier` drifted from the `db.Querier` interface (`UpdateSensorConfig` signature mismatch) — and `cropcycle` panics on a nil catalog load. Those two are real, not environmental.

This matters because `go test ./...` is the first thing any new contributor (or an agent picking up this repo cold) runs to sanity-check the tree, and right now it's red out of the box with no documented "these N packages need a DB / are known-flaky" note. At 150+ shipped phases, a bit of test rot is expected — but it compounds silently until it hides a real regression.

**Scope if greenlit:** fix the sensor mock + cropcycle nil-catalog panic, audit the rest for DB-required vs. actually-broken, document the split (e.g. `make test-unit` vs `make test-integration`), add that split to CI so at least the unit lane is green on every push.

---

## Phase 155 (candidate) — Automated backups, not just a runbook

`docs/backup-restore-runbook.md` is good and accurate, but it's **entirely manual** — an operator has to remember to run `pg_dump` themselves. There is no cron example, no retention/rotation, no automated integrity check (does the dump actually restore?), no off-box copy guidance beyond "for S3 use provider snapshots."

This is the single highest-consequence gap for a self-hosted app: gr33n holds a farmer's sensor history, crop cycles, and cost/finance data, and there's currently no safety net if a disk dies between manual backups. This is exactly the kind of thing that's invisible until the day it isn't.

**Scope if greenlit:** a `scripts/backup-gr33n.sh` (pg_dump + file storage tar + rotation + optional restore-and-verify-in-a-scratch-db smoke), a cron example in the runbook, and a `make backup` / `make verify-backup` target.

---

## Phase 156 (candidate) — Dependency & vuln scanning

No `dependabot.yml`, no `govulncheck` in CI, no `npm audit`. Nothing currently tells the operator (or you) when a Go module or an npm package in the UI has a known CVE. For an app that handles auth, JWT, file uploads (receipts), and now a change-request/proposal execution path, that's worth closing.

**Scope if greenlit:** Dependabot config for `go.mod` + `ui/package.json`; a `govulncheck ./...` + `npm audit --audit-level=high` CI job. Given the PR-gate lesson above, this one should be proposed as blocking-on-`main`-only (not on every feature branch) unless you say otherwise — surfacing it here explicitly instead of just doing it.

---

## Phase 157 (candidate) — Docs consolidation

150+ phase-plan docs is a genuine asset (it's why an agent picking this repo up cold — including me, this session — can reconstruct exact intent). But there's no single "what does gr33n look like today" doc separate from the historical trail; `phase-14-operator-documentation.md` has become the de facto index and it's 185+ lines of links into the archive. New contributors and even Composer sessions pay a real "which of these 150 docs is still true" tax.

**Scope if greenlit:** one `docs/current-state.md` snapshot (features, routes, schemas, at a glance) regenerated periodically, and an `docs/plans/archive/` folder for phases whose "close when" conditions are fully met, so the live index shrinks.

---

## Phase 158 (candidate, lower priority) — Accessibility pass

Not evidence-checked as deeply as the above (no automated a11y audit run yet), but worth naming: a farm-ops tool used in bright sun / gloves / one-handed-in-the-field contexts benefits a lot from solid keyboard nav, focus states, and screen-reader labels — this hasn't had a dedicated pass the way Guardian answer quality has (143–153). Lower priority than 154–156, but flagged because "an app people will use" includes people who aren't you testing it on a laptop.

---

## What's NOT a gap (checked before writing this)

- **Hosted-LLM lock-in** — not an issue; `LLM_BASE_URL` is OpenAI-compatible, so a hosted provider works exactly like local Ollama. No forced local-hardware requirement.
- **Offline/PWA story** — already shipped and documented (README "Offline-First Mobile", connectivity-requirements.md).
- **RBAC** — Owner/Manager/Operator/Viewer already shipped (Phase 113-era hardening + earlier).
- **Guardian change-request smoke coverage** — closed by Phase 153 (`make guardian-qa-change-requests`).

---

## Recommendation if you want a starting point

154 and 155 first — one is "the test suite lies to you right now," the other is "there's no safety net under your own data." Both are small, both are unglamorous, both are the kind of thing that's cheap to fix now and expensive to discover the hard way later. 156–158 are real but lower urgency.
