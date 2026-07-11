---
name: Phase 153 — Guardian PR smoke gate
overview: >
  guardian-eval always exited 0 regardless of heuristic pass/fail, so
  guardian-qa-smoke was an artifact generator, not a gate — nothing failed a
  build when a fixture regressed. Phase 153 adds a real -fail-on-regression
  exit code and an opt-in (label-triggered) CI job so a PR can actually be
  smoke-tested against live Ollama, without violating Phase 131's documented
  non-goal of a mandatory LLM gate on every push.
todos:
  - id: ws1-fail-on-regression
    content: "WS1: -fail-on-regression flag in cmd/guardian-eval; exits non-zero on any fixture failure"
    status: completed
  - id: ws2-make-target
    content: "WS2: make guardian-qa-pr-check wraps smoke suite with -fail-on-regression"
    status: completed
  - id: ws3-ci-job
    content: "WS3: guardian-qa-pr CI job — opt-in via `guardian-smoke` PR label or workflow_dispatch, self-hosted+ollama runner"
    status: completed
  - id: ws4-docs
    content: "WS4: ci-guardian-qa.md updated; phase-14 index; architecture doc"
    status: completed
isProject: false
---

# Phase 153 — Guardian PR smoke gate

**Status:** **Shipped** (CI job unverified end-to-end — no self-hosted `ollama`-labeled runner registered on this repo yet; see Verification below) · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md) · [146](phase_146_guardian_quality_loop_and_judge.plan.md) · [152](phase_152_guardian_live_accuracy_guardrails.plan.md)

---

## Why this phase

Every Guardian smoke/regression run this whole 143–152 arc was analyzed **by a human reading the report** — `cmd/guardian-eval` never inspected its own `EvalQuestionScore.Passed` results before exiting, so `make guardian-qa-smoke` always printed "Eval report written to ..." and exited 0, pass or fail. Phase 148–151's detectors are real regression tests for a specific bug, but they only run inside `guardian-eval`'s own process — nothing in CI or a PR check ever asked "did any of them fail?"

The user asked for "pull request added to smoke test for Guardian." Two things needed to be true for that to mean anything:
1. A failing fixture had to be able to **fail a command** (not just a JSON file nobody reads).
2. A PR had to be able to **trigger that command** without breaking [Phase 131's explicit non-goal](../ci-guardian-qa.md): *"Mandatory PR gate on every push (too slow, LLM-flaky on shared CI)."* GitHub-hosted runners have no Ollama, and a full smoke suite runs 30–90+ minutes on a CPU laptop — making it a *required* check on every PR would be exactly the flaky, slow gate Phase 131 explicitly rejected.

Phase 153 resolves the tension: real exit code, opt-in trigger.

## Workstreams

### WS1 — Real pass/fail exit code ✅

**Shipped:** `cmd/guardian-eval/main.go` — new `-fail-on-regression` flag. After the report is built and saved (so the JSON/QA archive is always written, pass or fail — you still get the evidence), `regressionFailures(rep.Details)` scans every model's scored fixtures for `Passed == false` and `os.Exit(1)` with a printed list of `<model>/<fixture-id>: <notes>` if any failed. Without the flag, behavior is unchanged (existing local `make guardian-qa-smoke` usage stays exit-0/artifact-only).

`regressionFailures` is a pure function over already-scored results — no LLM/DB — so it's fully unit-tested in `cmd/guardian-eval/main_test.go` without needing a live model.

### WS2 — Make target ✅

**Shipped:** `make guardian-qa-pr-check` (`Makefile`) — same shape as `guardian-qa-smoke` but adds `-fail-on-regression`. Defaults to `SUITE=smoke MODEL=phi3:mini FARM_ID=1`, overridable like every other `guardian-qa-*` target. This is the command a developer runs locally before opening/updating a Guardian-touching PR — same command the opt-in CI job runs.

### WS3 — Opt-in CI job ✅

**Shipped:** `.github/workflows/ci.yml` — `guardian-qa-pr` job, gated by:

```yaml
if: >
  (github.event_name == 'pull_request' && contains(github.event.pull_request.labels.*.name, 'guardian-smoke'))
  || github.event_name == 'workflow_dispatch'
runs-on: [self-hosted, ollama]
```

- **Not a required check** and **not run on every PR** — only when the `guardian-smoke` label is applied (or manually via Actions → Run workflow). Adding a label to an already-open PR needed the `pull_request:` trigger's `types:` widened to include `labeled` (previously implicit `[opened, synchronize, reopened]` only) — done at the top-level `on:` block.
- Same shape as the existing `ollama-smoke` job (Compose Postgres, bootstrap+seed, pull `phi3:mini`) but starts the real API (not just `cmd/api` in-process tests) so `/v1/chat` is live for `guardian-eval` to hit, then runs `make guardian-qa-pr-check` as the pass/fail step.
- Uploads `data/guardian_qa_runs/` as a build artifact either way (`if: always()`), so a failed gate still leaves the full transcript for review — same "artifact even on failure" pattern as `ollama-smoke`.

### WS4 — Docs ✅

**Shipped:** [`ci-guardian-qa.md`](../ci-guardian-qa.md) documents the label-triggered path alongside the existing self-hosted-nightly pattern; architecture doc and phase-14 index cross-link this phase.

## Verification

Unit-level (`regressionFailures`) is verified, plus a **live end-to-end run** of the exact command the CI job runs:

```
$ go run ./cmd/guardian-eval/ -models phi3:mini -farm-id 1 -suite smoke -prompt-ids smoke-ec-ph -fail-on-regression ...
...
Guardian eval regression — 1 fixture(s) failed their heuristic:
  - phi3:mini/smoke-ec-ph: uncited_tail
exit status 1
```

The machine was under heavy load from other work at the time (grounded warmup timed out, latency ~28 min, 0% citation rate) — a genuine degraded run, not a synthetic test — and `-fail-on-regression` correctly turned that into a process exit code of `1` while still writing the report and QA archive first. This is exactly the failure mode the gate exists to catch.

**The GitHub Actions job itself is still unverified end-to-end** — this repo has no `[self-hosted, ollama]`-labeled runner registered, so `guardian-qa-pr` will simply never pick up a runner and sit queued if triggered (same caveat as the existing `ollama-smoke`/`hardware-smoke` jobs, which are also unverified beyond their own `workflow_dispatch` runs by whoever owns a matching runner). To use this for real:

1. Register a self-hosted runner labeled `self-hosted, ollama` with Ollama installed and `phi3:mini` pullable.
2. Set the `guardian-smoke` label on a PR (or trigger `workflow_dispatch`) to run it.
3. No new secret needed for auth — `make guardian-qa-pr-check` refreshes its own dev-mode JWT via `scripts/source-local-env.sh --refresh-eval-token`, the same path the local Make target already uses.

## Acceptance

- [x] `guardian-eval -fail-on-regression` exits 1 when any fixture's heuristic fails, 0 when all pass — verified both by unit test and a live run against phi3:mini (see Verification).
- [x] `regressionFailures` unit-tested (all-pass, some-fail, multi-model sort) without a live LLM.
- [x] `make guardian-qa-pr-check` exists and dry-run (`make -n`) produces valid shell.
- [x] `guardian-qa-pr` CI job added, gated to label/workflow_dispatch + self-hosted runner — never blocks a default hosted-runner PR.
- [x] Report/QA archive is still written even when the gate fails (evidence survives a red check).
- [ ] End-to-end run against a real self-hosted `ollama` runner (needs runner registration outside this repo's control — see Verification).

## Non-goals

- Making `guardian-qa-pr` a required/default check on every PR — explicitly rejected by Phase 131 and reaffirmed here; hosted runners have no Ollama and the suite is too slow to gate merges by default.
- Auto-labeling PRs based on changed paths (e.g. auto-add `guardian-smoke` when `internal/farmguardian/**` changes) — a reasonable follow-up, not done here to keep the trigger surface simple and explicit.
- Comparing against a historical baseline report (regression *magnitude*, not just pass/fail) — `-fail-on-regression` is binary per-fixture; trend tracking would be its own phase.
