---
name: Phase 153 — Guardian change-request smoke fetcher
overview: >
  "Guardian PR" in this repo means the propose→confirm change-request queue
  (gr33ncore.guardian_action_proposals) — the cards a farmer confirms in the
  UI, not a GitHub pull request. Phase 153 adds a script-based smoke check
  (same style/tooling as the existing guardian-qa-smoke suite, no UI) that
  fires a few preset write-intent prompts at Guardian and then fetches
  GET /v1/chat/proposals?status=pending to confirm each one actually landed
  a row in the pending queue — not just that the chat response echoed a
  proposal object inline.
todos:
  - id: ws1-fixtures
    content: "WS1: change-requests fixture set (write-feed/write-ack/write-schedule/write-task) + -suite change-requests"
    status: completed
  - id: ws2-pending-fetch
    content: "WS2: APIClient.FetchPendingProposals hits GET /v1/chat/proposals?status=pending"
    status: completed
  - id: ws3-check-flag
    content: "WS3: -check-pending-proposals flag — fetch queue after the run, fail if fewer pending rows than passed write-intent fixtures"
    status: completed
  - id: ws4-make-target
    content: "WS4: make guardian-qa-change-requests wraps the suite + fetch check"
    status: completed
isProject: false
---

# Phase 153 — Guardian change-request smoke fetcher

**Status:** **Shipped** · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md) (existing smoke harness) · [30](phase_30_guardian_change_requests.plan.md) (the change-request queue this tests) · [152](phase_152_guardian_live_accuracy_guardrails.plan.md)

---

## Why this phase (and a correction)

An earlier pass at this phase misread "Guardian pull request" as a **GitHub pull request** and built a CI job (`guardian-qa-pr`) plus a `pull_request:` label trigger. That's not what was asked for and it's been fully removed — no GitHub Actions automation runs against Guardian, opt-in or otherwise. Nothing here touches `.github/workflows/`.

What was actually meant: Guardian's own **change-request ("PR") queue** — `gr33ncore.guardian_action_proposals`, surfaced in the UI as proposal cards a farmer clicks Confirm on (Phase 30). The ask was a **script** — same shape as `make guardian-qa-smoke` — that fires a few preset prompts designed to trigger a proposal, then checks the pending-request queue actually got populated. That's what this phase ships.

**Why the check matters (not just re-testing what already existed):** `guardian-eval`'s existing `write_intent` fixtures already check `ProposalCount > 0` on the raw chat response — but that only confirms the LLM's JSON *echoed back* in that one response looked like a proposal. It says nothing about whether a row was actually persisted to `guardian_action_proposals` and would show up for the farmer to confirm later. Those are two different code paths (`attachProposals` building the inline response object vs. `InsertGuardianProposal` writing the row) and a bug could break either independently.

## Workstreams

### WS1 — Change-request fixture set ✅

**Shipped:** [`fixtures_change_requests.go`](../../internal/farmguardian/eval/fixtures_change_requests.go) — `ChangeRequestFixtures()` pulls the 4 existing `write_intent` prompts out of the regression set (`write-feed`, `write-ack`, `write-schedule`, `write-task`) so a dedicated run stays short instead of dragging in the full ~24-prompt regression suite. Wired into `FixturesForSuite` as `-suite change-requests` (aliases: `proposals`, `pr`).

### WS2 — Fetch the actual pending queue ✅

**Shipped:** [`proposals.go`](../../internal/farmguardian/eval/proposals.go) — `APIClient.FetchPendingProposals(ctx)` calls `GET /v1/chat/proposals?status=pending&farm_id=...` — the exact endpoint [`GuardianSettingsFeedbackReviewCard.vue`](../../ui/src/components/GuardianSettingsFeedbackReviewCard.vue)/the UI's proposal inbox reads from (`internal/handler/chat/proposals.go` `ListProposals`). Returns `proposal_id`, `tool`, `summary`, `risk_tier` for each pending row.

### WS3 — `-check-pending-proposals` ✅

**Shipped:** `cmd/guardian-eval/main.go` — new flag. After the run, counts how many `ExpectProposal` fixtures **passed** their heuristic (`passedProposalFixtures`), then calls `FetchPendingProposals` and prints every pending row found. If the pending queue has fewer rows than that count, it prints an explanation and the process exits non-zero — the concrete signal that a prompt looked fine in the chat response but never actually reached the confirmable queue.

### WS4 — Make target ✅

**Shipped:** `make guardian-qa-change-requests` — `guardian-eval -suite change-requests -check-pending-proposals`, same JWT-refresh/`.env`-sourcing pattern as every other `guardian-qa-*` target. Run it exactly like `make guardian-qa-smoke`:

```
make guardian-qa-change-requests MODEL=phi3:mini FARM_ID=1
```

## What this is not

- **Not tied to GitHub in any way.** No workflow file changes, no PR labels, no required checks. `-fail-on-regression` (a separate, small exit-code fix for `guardian-eval` itself — it always used to exit 0 no matter what) is still there as a standalone flag because it's harmless and useful on its own, but nothing invokes it automatically.
- **Not a UI test.** This hits `/v1/chat` and `/v1/chat/proposals` directly over HTTP, exactly like `guardian-qa-smoke` already does — no browser, no Playwright.
- **Not confirming the proposal.** This only checks the propose step landed in the pending queue, not the Confirm→execute path (that's `internal/handler/chat/proposals.go` `PostConfirmProposal` + the tool's actual side effect, already covered elsewhere).

## Verification

Live end-to-end run (not just unit tests) against the real API + a real `phi3:mini` chat turn:

```
$ go run ./cmd/guardian-eval/ -models phi3:mini -farm-id 1 -suite change-requests -prompt-ids write-ack -check-pending-proposals ...
...
  phi3:mini: grounded cite 0% · decline 0% · proposal 100% · latency 1672923ms
...
Pending change-request queue: 1 row(s)
  - [91523e33-eb25-4b88-811e-754511c4b05e] ack_alert — Acknowledge: Humidity high — Flower Room (risk: low)
```

The `write-ack` prompt ("Acknowledge the highest severity unread alert") passed its heuristic, and `-check-pending-proposals` then fetched the real queue and found the actual persisted row waiting to be confirmed — proof the script exercises the real propose→pending-queue path end to end, not just the inline chat response.

## Acceptance

- [x] `make guardian-qa-change-requests` runs the 4 write-intent prompts and fetches the pending queue afterward — verified live against phi3:mini (see Verification).
- [x] Exits non-zero and prints an explanation if fewer pending rows exist than passed write-intent fixtures.
- [x] Prints every pending row found (`proposal_id`, `tool`, `summary`, `risk_tier`) so a human can eyeball the queue without opening the UI.
- [x] `passedProposalFixtures` / `reportPendingProposals` unit-tested against a local `httptest` stand-in for `/v1/chat/proposals` — no live LLM needed for the logic itself.
- [x] No GitHub Actions changes of any kind.

## Non-goals

- Automated Confirm-and-verify-side-effect testing (separate concern from "did the request get queued").
- Comparing proposal content/args against an expected shape — this is a presence/count smoke check, not a deep-equality assertion on proposal payloads.
