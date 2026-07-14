---
name: Phase 184 — Guardian PR multi-turn conversation smoke
overview: >
  Extends the Phase 153/162 change-request smoke gate with multi-turn
  dialogues instead of single-shot prompts: a session that proposes, gets
  refined ("revise to 0.3L instead"), and is confirmed via the API end to
  end, plus several dialogue-then-leave-pending scenarios so an operator
  can exercise Confirm / Refine / Dismiss for real on the Pending tab.
todos:
  - id: ws1-scenario-types
    content: "WS1: Scenario/ScenarioTurn types + RunScenarioSuite with shared session_id per scenario (eval.RunQuestionInSession)"
    status: completed
  - id: ws2-fixtures
    content: "WS2: change-requests-ui fixture set — 1 revise-then-confirm, 4 dialogue-then-leave-pending scenarios; -quick subset"
    status: completed
  - id: ws3-cli-make
    content: "WS3: -suite change-requests-ui[-quick] in cmd/guardian-eval + make guardian-qa-change-requests-ui[-quick] targets"
    status: completed
  - id: ws4-tests-docs
    content: "WS4: Go tests (scenario counts, filter, suite detection) + docs/ci-guardian-qa.md scenario table"
    status: completed
  - id: ws5-live-verification
    content: "WS5: Run guardian-qa-change-requests-ui[-quick] live against a local stack; confirm 4 pending cards + 1 confirmed row match expectations in the UI"
    status: pending
isProject: false
---

# Phase 184 — Guardian PR multi-turn conversation smoke

**Status:** shipped (code) · pending live verification · **Depends on:** [153](phase_153_guardian_pr_smoke_gate.plan.md) · [162](phase_162_guardian_confirm_db_smoke.plan.md)

## The problem

Phases 153/162 smoke the change-request ("PR") queue with **single-shot**
write-intent prompts — one message in, one proposal out, either checked in
the pending queue or confirmed immediately. That doesn't exercise:

- **Refine** — a real back-and-forth where the operator corrects a proposal
  ("use 0.3L instead of 0.5") and the pending row should show a bumped
  `Revision` before Confirm.
- Leaving a **realistic mix** of pending cards for manual UI testing — not
  just N copies of the same fixture, but different tools/dialogues so the
  Pending tab has varied cards to click Confirm / Refine / Dismiss on.

Operator ask (2026-07-13): *"a convo in the PR's for the guardian — some PR
should get submitted and other leave unsubmitted so I can see them in the
UI and test the confirm button."*

## What shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `internal/farmguardian/eval/scenario.go` — `Scenario`, `ScenarioTurn`, `RunScenarioSuite`; each scenario keeps one `session_id` across turns via new `APIClient.RunQuestionInSession` (chat's `session_id` echoed back and reused). |
| **WS2** | `fixtures_change_requests_ui.go` — 5 scenarios (see table below); `ChangeRequestUIScenariosQuick()` for a 2-scenario fast subset. |
| **WS3** | `-suite change-requests-ui` / `change-requests-ui-quick` in `cmd/guardian-eval`; `make guardian-qa-change-requests-ui[-quick]`. |
| **WS4** | `fixtures_change_requests_ui_test.go`; `docs/ci-guardian-qa.md` scenario table + subset command. |
| **WS5** | Not yet run against a live local stack — needs Ollama + `DATABASE_URL` + a logged-in JWT (see Verification). |

## Scenario → outcome

| Scenario | Turns | End state |
|----------|-------|-----------|
| `scenario-feed-revise-confirm` | "Set feed to 0.5L for Veg Tent" → "revise — use 0.3L instead" | **Confirmed via API**, DB-verified (`write-feed` side-effect check, ≈0.3L) |
| `scenario-feed-revise-pending` | same dialogue | **Left pending** (revision ≥2, 0.3L) — test **Confirm** in UI |
| `scenario-task-dialogue-pending` | "Create a task to refill calcium nitrate" → "which zone should this refer to?" | **Left pending** — test **Refine** / **Confirm** |
| `scenario-schedule-pending` | "Pause the lights schedule for Veg Tent until tomorrow" | **Left pending** |
| `scenario-ack-pending` | "Acknowledge the highest severity unread alert" | **Left pending** |

Result: **1 confirmed + verified** row, **4 pending** cards of different
tools (feed, task, schedule, ack) for manual Confirm/Refine/Dismiss testing
— exactly the mix requested instead of 4 copies of one fixture.

## How proposal resolution works mid-dialogue

Each scenario's turns share a `session_id`. After the final turn,
`resolveScenarioProposal` fetches the pending queue and picks the
highest-`Revision` proposal for that `session_id` (falling back to the
response's inline `proposal_id`s if the session filter finds nothing) —
so a scenario that revises via chat text (not raw JSON args) still resolves
to the *corrected* pending row before Confirm/leave-pending, and
`WantVolumeLiters`/`MinRevision` assert the revise actually happened rather
than just that *a* proposal exists.

## Commands

```bash
# Full mix: 1 confirmed + 4 pending, different tools (~2-3h CPU)
make guardian-qa-change-requests-ui MODEL=phi3:mini FARM_ID=1

# Fast subset: feed-revise-pending + task-dialogue-pending (~50 min)
make guardian-qa-change-requests-ui-quick MODEL=phi3:mini FARM_ID=1

# Single scenario
go run ./cmd/guardian-eval -suite change-requests-ui \
  -prompt-ids scenario-task-dialogue-pending -models phi3:mini -farm-id 1
```

Open `http://localhost:5173/chat?tab=pending` once the run finishes.

## Verification

```bash
go test ./internal/farmguardian/eval/... ./cmd/guardian-eval/... -count=1
```

Passing as of this phase (unit-level: scenario construction, filtering,
suite detection, session-threading wiring). **Still needed (WS5):** a live
run against a running API + Ollama + seeded demo farm to confirm the 4
pending cards actually render correctly in the Pending tab and the
confirmed row's DB side effect matches (0.3L on the Veg Tent program).

## Non-goals

- Browser/Playwright automation of the actual Confirm click — this phase
  gets the *data* into the right state; clicking Confirm/Refine/Dismiss in
  the browser stays a manual step (that's the point — it's UI testing
  prep, not UI test replacement).
- GitHub CI automation — stays script-only like Phase 153/162.
- Extending revise matchers for `create_task` titles (tracked separately
  as [Phase 183](phase_183_guardian_knowledge_and_revise_followups.plan.md)
  WS3 — `scenario-task-dialogue-pending` today only asks a clarifying
  question, not a hard correction, until that matcher work lands).

## Acceptance

- [x] Multi-turn scenarios share one `session_id` per scenario via chat's
      existing `session_id` field.
- [x] At least one scenario proposes → refines → **confirms via API** with
      DB verification.
- [x] At least one scenario is left pending after a dialogue for manual UI
      testing; full suite leaves 4 different tools pending.
- [x] `make guardian-qa-change-requests-ui[-quick]` targets exist and are
      documented.
- [ ] Live run confirms the Pending tab shows all 4 cards distinctly and
      Confirm/Refine/Dismiss all work against them (WS5).
