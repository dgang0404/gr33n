---
name: Phase 162 — Guardian confirm→DB smoke (planned)
overview: >
  Phase 153 proves write-intent prompts land in the pending change-request
  queue. Phase 162 closes the remaining loop: POST /v1/chat/confirm on each
  persisted proposal_id and assert the expected DB side effect (alert ack,
  schedule pause, task create, feed volume patch).
todos:
  - id: ws1-confirm-api-client
    content: "WS1: eval.APIClient.ConfirmProposal(proposal_id) wrapper"
    status: pending
  - id: ws2-side-effect-assertions
    content: "WS2: Per-fixture post-confirm DB checks (ack status, task row, program volume)"
    status: pending
  - id: ws3-flag-and-make
    content: "WS3: -confirm-proposals flag + make guardian-qa-change-requests-confirm"
    status: pending
  - id: ws4-closure
    content: "WS4: phase-162-closure.test.js + runbook note"
    status: pending
isProject: false
---

# Phase 162 — Guardian confirm→DB smoke (planned)

**Status:** planned · **Depends on:** [153](phase_153_guardian_pr_smoke_gate.plan.md) · [30](phase_30_guardian_change_requests.plan.md)

## Gap (from Phase 153 smoke review)

| Step | Covered today? | Where |
|------|----------------|-------|
| LLM prompt → proposal card in chat | ✅ | `guardian-eval` `ExpectProposal` heuristic |
| Row in `guardian_action_proposals` pending queue | ✅ | `make guardian-qa-change-requests` + `-check-pending-proposals` |
| User Confirm in UI | ❌ | Manual only |
| `POST /v1/chat/confirm` → DB mutation | ⚠️ Partial | `cmd/api/smoke_phase29/30/32` (programmatic proposals, not LLM path) |

## Shipped in Phase 162 follow-up (162a — eval ergonomics)

While waiting on full confirm smoke, **162a** landed:

- Per-prompt progress logging in `guardian-eval` (`eval: [1/4] starting write-feed…`)
- Pending-queue check matches **proposal_id from this run** (not stale row count)
- `make guardian-qa-change-requests-ack` — ~25 min single-prompt fast path

## Phase 162 workstreams (not started)

### WS1 — Confirm API client

`eval.APIClient.ConfirmProposal(ctx, proposalID)` → `POST /v1/chat/confirm`.

### WS2 — Side-effect assertions

| Fixture | After confirm, assert |
|---------|----------------------|
| `write-ack` | Alert `status=acknowledged` |
| `write-feed` | Fertigation program volume updated |
| `write-schedule` | Schedule paused / next_run adjusted |
| `write-task` | Task row exists |

### WS3 — Flag + Make target

`-confirm-proposals` after `-check-pending-proposals`; `make guardian-qa-change-requests-confirm`.

## Non-goals

- Browser/UI Playwright test
- GitHub PR CI automation for change-requests
