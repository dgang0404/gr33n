# Guardian QA — optional nightly CI (self-hosted)

**Audience:** Operators and maintainers with a **self-hosted GitHub Actions runner** (or equivalent) that has **Ollama** and enough CPU/GPU to run `make guardian-qa-smoke`.

**Not for GitHub-hosted runners** — they have no Ollama and smoke runs take 30–90 minutes on a CPU laptop.

**Related:** [Phase 131 plan](plans/phase_131_guardian_qa_harness.plan.md) · [local-operator-bootstrap.md](local-operator-bootstrap.md) § Guardian QA · [Phase 129–139 closure](plans/phase-129-139-closure.md)

---

## What runs

| Target | When | Output |
|--------|------|--------|
| `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1` | After Guardian changes; optional nightly/weekly | `data/guardian_qa_runs/<timestamp>_smoke_phi3-mini.json` — always exits 0 (artifact only) |
| `make guardian-qa-pr-check MODEL=phi3:mini FARM_ID=1` | Before/during a PR that touches Guardian (Phase 153) | Same archive, but **exits non-zero if any fixture fails its heuristic** — run this locally before opening a Guardian PR |
| `make guardian-qa-regression MODEL=phi3:mini` | Pre-release (slow) | Same directory, regression suite |
| `make guardian-qa-manual` | Human UI parity | Prints checklist from same fixtures |

Set `GUARDIAN_EVAL_TOKEN` (JWT from dev login) and optionally `GUARDIAN_EVAL_LOG=/tmp/gr33n-api.log` for log correlation.

---

## PR gate (Phase 153 — opt-in, shipped in `ci.yml`)

`guardian-qa-pr` in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs `make guardian-qa-pr-check` (smoke suite + `-fail-on-regression`) on a self-hosted `[self-hosted, ollama]` runner. It is **not a required check and does not run on every PR** — it only fires when:

- the `guardian-smoke` label is added to a PR, or
- it's triggered manually via **Actions → CI → Run workflow**.

This keeps the Phase 131 non-goal intact (below) while making "smoke-test this PR" a real, one-click action once a self-hosted Ollama runner is registered. `cmd/guardian-eval -fail-on-regression` is what turns a fixture failure into an actual red check — see [`plans/phase_153_guardian_pr_smoke_gate.plan.md`](plans/phase_153_guardian_pr_smoke_gate.plan.md).

## Example workflow (documented pattern — not enabled in default repo CI)

Save as `.github/workflows/guardian-qa-nightly.yml` **only on a fork/site with a self-hosted runner** tagged `ollama`:

```yaml
name: guardian-qa-smoke

on:
  schedule:
    - cron: '0 6 * * 1'  # weekly Monday 06:00 UTC
  workflow_dispatch:

jobs:
  guardian-smoke:
    runs-on: [self-hosted, ollama]
    steps:
      - uses: actions/checkout@v4

      - name: Start stack
        run: |
          make compose-db-up
          make migrate
          # Start API + seed as your site requires; smoke needs live /v1/chat

      - name: Guardian QA smoke
        env:
          GUARDIAN_EVAL_TOKEN: ${{ secrets.GUARDIAN_EVAL_TOKEN }}
          GUARDIAN_EVAL_LOG: /tmp/gr33n-api.log
        run: make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: guardian-qa-${{ github.run_id }}
          path: data/guardian_qa_runs/
          retention-days: 30
```

---

## Pass criteria (v1 — heuristics + Phase 145/146 drift)

Smoke uses **recorded JSON + heuristic checks** (answer length, citation count, embed relevance, citation alignment, optional log scrape for `walk_farm`). **Heuristic pass ≠ operator-quality pass** — see [smoke report 2026-07-07](guardian-qa-smoke-report-20260707.md).

**Phase 146 judge policy:**

| Mode | Env | Behavior |
|------|-----|----------|
| **Default (CPU laptop)** | `GUARDIAN_ANSWER_CRITIQUE` unset or `0` | Embed + citation drift from Phase 145 only |
| **Optional GPU critique** | `GUARDIAN_ANSWER_CRITIQUE=1` | One short YES/NO LLM gate after finalize; eval fails on NO |

`make guardian-qa-smoke` refreshes `GUARDIAN_EVAL_TOKEN` via `scripts/source-local-env.sh` before running. Tune CPU warmup with `GUARDIAN_EVAL_WARMUP_TIMEOUT=90` (see `make guardian-laptop-tune`).

**After smoke:** Run the [smoke quality checklist](guardian-feedback-review-runbook.md#smoke-quality-checklist-phase-143) · Settings → Guardian feedback · promote down-votes with [feedback→fixture script](../scripts/guardian-feedback-to-fixture.sh).

Phase 131 deferred full LLM-as-judge — **Phase 146 supersedes that for GPU profile only** (binary critique, not rubric grading).

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Timeouts on CPU | `make guardian-laptop-tune ARGS="--apply"`; use smoke model `phi3:mini`; raise `LLM_TIMEOUT_SECONDS` |
| Missing `walk_farm` in logs | `./scripts/guardian-qa-scrape-logs.sh --expect walk_farm` |
| Warmup HTTP 503 before grounded smoke | Eval now sends `chat_model` matching `-models`; ensure `phi3:mini` is installed when `.env` `LLM_MODEL` is tinyllama |
| 401 on eval | `make guardian-qa-smoke` refreshes token via `source-local-env.sh`; or run manually before eval |
| 4th smoke prompt client timeout | Re-run `make guardian-qa-smoke-ec-ph` (Phase 147); raise `GUARDIAN_EVAL_TIMEOUT_SECONDS` or use eval client buffer (Phase 146 default +15m) |
| Warmup blocks 5m | Set `GUARDIAN_EVAL_WARMUP_TIMEOUT=90` on CPU; smoke uses async warmup (Phase 146) |

---

## Non-goals

- Mandatory PR gate on every push (too slow, LLM-flaky on shared CI) — Phase 153's `guardian-qa-pr` is opt-in (label/`workflow_dispatch`) precisely to avoid this
- GitHub-hosted runner without Ollama
- Automated LLM grading of answer quality (deferred to a future phase when GPU CI is stable)
- Historical-baseline regression tracking (Phase 153 is pass/fail per fixture, not "did this get worse than last week")
