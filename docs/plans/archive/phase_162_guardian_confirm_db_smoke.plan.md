---
name: Phase 162 — Guardian confirm→DB smoke
overview: >
  Closes the propose→confirm loop: after write-intent prompts land in the
  pending queue (Phase 153), POST /v1/chat/confirm and verify DB side effects
  via confirm result + farm list GETs.
todos:
  - id: ws1-confirm-api-client
    content: "WS1: eval.APIClient.ConfirmProposal(proposal_id) wrapper"
    status: completed
  - id: ws2-side-effect-assertions
    content: "WS2: Per-fixture post-confirm DB checks (ack status, task row, program volume)"
    status: completed
  - id: ws3-flag-and-make
    content: "WS3: -confirm-proposals flag + make guardian-qa-change-requests-confirm"
    status: completed
  - id: ws4-closure
    content: "WS4: phase-162-closure.test.js + runbook note"
    status: completed
isProject: false
---

# Phase 162 — Guardian confirm→DB smoke

**Status:** shipped · **Depends on:** [153](phase_153_guardian_pr_smoke_gate.plan.md) · [162a](phase_162_guardian_confirm_db_smoke.plan.md) (progress logs + proposal_id match)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `eval.ConfirmProposal` → `POST /v1/chat/confirm` |
| **WS2** | `VerifyConfirmSideEffect` per fixture + list GETs |
| **WS3** | `-confirm-proposals` + `make guardian-qa-change-requests-confirm` |
| **WS4** | `confirm_verify_test.go` + `phase-162-closure.test.js` |

## Fixture → verification

| Fixture | After confirm |
|---------|----------------|
| `write-ack` | `is_acknowledged` in result + alert row in `GET /farms/{id}/alerts` |
| `write-feed` | `total_volume_liters` ≈ 0.3 on program in `GET /farms/{id}/fertigation/programs` |
| `write-schedule` | `is_active=false` on schedule in `GET /farms/{id}/schedules` |
| `write-task` | `task_id` in result + row in `GET /farms/{id}/tasks` |

## Commands

```bash
# Full loop: 4 prompts → pending queue → Confirm → DB checks (~1–2h CPU)
make guardian-qa-change-requests-confirm MODEL=phi3:mini FARM_ID=1

# Queue only (Phase 153)
make guardian-qa-change-requests MODEL=phi3:mini FARM_ID=1

# Fast ack only (~25 min) — queue check, no confirm
make guardian-qa-change-requests-ack MODEL=phi3:mini FARM_ID=1
```

## Verification

```bash
go test ./internal/farmguardian/eval/... -run Confirm -count=1
cd ui && npm test -- --run src/__tests__/phase-162-closure.test.js
```

## Non-goals

- Browser/UI Playwright
- GitHub CI automation for change-requests
