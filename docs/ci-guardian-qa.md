# Guardian QA — optional nightly CI (self-hosted)

**Audience:** Operators and maintainers with a **self-hosted GitHub Actions runner** (or equivalent) that has **Ollama** and enough CPU/GPU to run `make guardian-qa-smoke`.

**Not for GitHub-hosted runners** — they have no Ollama and smoke runs take 30–90 minutes on a CPU laptop.

**Related:** [Phase 131 plan](plans/phase_131_guardian_qa_harness.plan.md) · [local-operator-bootstrap.md](local-operator-bootstrap.md) § Guardian QA · [Phase 129–139 closure](plans/phase-129-139-closure.md)

---

## What runs

| Target | When | Output |
|--------|------|--------|
| `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1` | After Guardian changes; optional nightly/weekly | `data/guardian_qa_runs/<timestamp>_smoke_phi3-mini.json` — always exits 0 (artifact only) |
| `make guardian-qa-smoke-strict MODEL=phi3:mini FARM_ID=1` | When you want a real pass/fail instead of a report to read | Same archive, but **exits non-zero if any fixture fails its heuristic** |
| **Opt-in PR check** `guardian-qa-pr` CI job | Self-hosted runner + Ollama; add `guardian-smoke` label to PR | Runs `make guardian-qa-smoke-strict` — **not mandatory** on every PR (standard label-gated pattern) |
| `make guardian-qa-change-requests MODEL=phi3:mini FARM_ID=1` | After touching proposal/change-request code (Phase 153) | Fires the 4 write-intent prompts; **verifies each proposal_id in the pending queue immediately after its prompt** (proposals expire after 5m; prompts take 20+ min) |
| `make guardian-qa-change-requests-pending MODEL=phi3:mini FARM_ID=1` | Leave proposals in UI Pending tab for manual review | Same 4 prompts; **bumps TTL to 24h** after each (needs `DATABASE_URL`) — open `/chat?tab=pending` when done |
| `make guardian-qa-change-requests-pending-quick MODEL=phi3:mini FARM_ID=1` | Fast single-proposal demo (~25 min) | `write-ack` only + leave pending for UI |
| `make guardian-qa-change-requests-confirm MODEL=phi3:mini FARM_ID=1` | Full propose→confirm→DB loop (Phase 162) | Same, plus **per-prompt Confirm** and side-effect GETs |
| `make guardian-qa-change-requests-ui MODEL=phi3:mini FARM_ID=1` | Multi-turn Pending-tab prep + one API confirm | 5 scenarios: feed revise (confirm + pending), task dialogue, schedule, ack — **shared session_id** per scenario; 4 left pending (24h TTL) |
| `make guardian-qa-change-requests-ui-quick MODEL=phi3:mini FARM_ID=1` | Fast multi-turn UI demo (~50 min) | Ack + schedule single-turn scenarios (reliable CPU path) |
| `make guardian-qa-regression MODEL=phi3:mini` | Pre-release (slow) | Same directory, regression suite |
| `make guardian-qa-manual` | Human UI parity | Prints checklist from same fixtures |

Set `GUARDIAN_EVAL_TOKEN` (JWT from dev login) and optionally `GUARDIAN_EVAL_LOG=/tmp/gr33n-api.log` for log correlation.

---

## Guardian's change-request ("PR") queue smoke check (Phase 153)

"Guardian PR" in this codebase means the propose→confirm change-request queue (`gr33ncore.guardian_action_proposals`) — the proposal cards a farmer clicks Confirm on in the UI, **not** a GitHub pull request. Script-only smoke (no GitHub automation):

```
make guardian-qa-change-requests MODEL=phi3:mini FARM_ID=1
make guardian-qa-change-requests-confirm MODEL=phi3:mini FARM_ID=1 # full Confirm→DB loop
```

It fires 4 preset write-intent prompts (or one with `-ack`), logs per-prompt progress, then **immediately after each passed write-intent prompt** calls `GET /v1/chat/proposals?status=pending` and verifies that prompt's `proposal_id`(s) are still pending (batch end-of-run check was removed — proposals expire after 5m while each prompt takes 20+ min). **Confirm → DB:** `make guardian-qa-change-requests-confirm` confirms each proposal right after its prompt (Phase 162). See [Phase 153](plans/phase_153_guardian_pr_smoke_gate.plan.md) · [Phase 162](plans/phase_162_guardian_confirm_db_smoke.plan.md).

### Multi-turn UI scenarios (`change-requests-ui`)

For testing **Refine**, **Confirm**, and **Dismiss** on real back-and-forth dialogues (not single-shot prompts):

```
make guardian-qa-change-requests-ui MODEL=phi3:mini FARM_ID=1
make guardian-qa-change-requests-ui-quick MODEL=phi3:mini FARM_ID=1   # ~50 min: ack + schedule (single-turn)
```

Each scenario reuses one `session_id` across turns. The full suite runs **5 scenarios**:

| Scenario | Turns | End state |
|----------|-------|-----------|
| `scenario-feed-revise-confirm` | propose 0.5L → revise to 0.3L | Confirmed via API (DB verified) |
| `scenario-feed-revise-pending` | same dialogue | Left pending (rev 2, 0.3L) — test **Confirm** in UI |
| `scenario-task-dialogue-pending` | create task → zone → title → due tomorrow | Left pending (rev ≥4, zone + title + relative due) — test **Refine** / **Confirm** |
| `scenario-schedule-pending` | pause schedule | Left pending |
| `scenario-ack-pending` | acknowledge alert | Left pending |

Requires `DATABASE_URL` for TTL bump on leave-pending scenarios. Open `/chat?tab=pending` when the run finishes.

Subset one scenario: `guardian-eval -suite change-requests-ui -prompt-ids scenario-ack-pending`

## Opt-in GitHub PR check (Guardian answer smoke — not change-request queue)

A **label-gated** CI job (`guardian-qa-pr` in `.github/workflows/ci.yml`) is **not bad or weird** — it's a standard pattern for slow, model-dependent tests. It was briefly reverted when scope wasn't consented to; it's back as **opt-in only**:

- Add label **`guardian-smoke`** to a GitHub PR, or run **workflow_dispatch** on the CI workflow
- Requires a **self-hosted** runner tagged `ollama` with Ollama + phi3:mini
- Runs `make guardian-qa-smoke-strict` (heuristic pass/fail on the 4-prompt smoke suite)
- **Not** a required check on every push — GitHub-hosted runners can't run this

This is separate from `guardian-qa-change-requests` (internal proposal queue), which stays script-only.

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

- Mandatory PR gate on every push (too slow, LLM-flaky on shared CI) — use **`guardian-smoke` label** instead
- GitHub-hosted runner without Ollama for the smoke gate
- GitHub Actions automation for Guardian's **change-request** smoke (`guardian-qa-change-requests`) — script only
- Automated LLM grading of answer quality (deferred to a future phase when GPU CI is stable)
- Historical-baseline regression tracking (pass/fail per fixture, not "did this get worse than last week")
