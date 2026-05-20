---
name: Phase 29 — Guardian Agent Layer
overview: >
  Evolve Farm Guardian from read-only Q&A (Phase 27–28) into a confirmed-action
  agent that can operate the farm through existing APIs, plus a global slide-out
  panel on every page and a one-command Guardian-ready bootstrap (seed + RAG).
todos:
  - id: ws1-slide-out-panel
    content: "WS1: Global Guardian slide-out — Pinia store + composable, open from SideNav on any route, preserve farm context + session_id, mobile-friendly drawer"
    status: pending
  - id: ws2-tool-registry
    content: "WS2: Tool registry + executor — define allowed actions (ack alert, create task from alert, patch cycle stage, list programs); map to existing REST handlers with farmauthz"
    status: pending
  - id: ws3-llm-tool-calling
    content: "WS3: LLM tool-calling loop — extend chat handler with tool schema in prompt, parse structured action proposals, require explicit operator Confirm before mutating"
    status: pending
  - id: ws4-audit-and-rbac
    content: "WS4: Audit + RBAC — every confirmed action writes farm audit event; viewer role cannot confirm writes; cost guard applies to tool rounds"
    status: pending
  - id: ws5-guardian-bootstrap
    content: "WS5: Guardian-ready bootstrap — dev-stack-fresh-rag + rag-ingest-demo (done); optional seed alert samples (open); bootstrap doc updated"
    status: in_progress
  - id: ws6-openapi-and-tests
    content: "WS6: OpenAPI + tests — document POST /v1/chat tool proposal/confirm shapes; Vitest for slide-out; smoke for one confirmed action path"
    status: pending
isProject: false
---

# Phase 29 — Guardian Agent Layer

## Why this phase

Phases 27–28 shipped Farm Guardian as a **read-only** conversational layer:

- Answers questions using Llama + RAG corpus + live snapshot
- Explains alerts and crop-cycle metrics
- Surfaces token usage and budget warnings

Operators asked the natural next question: *"Can it actually **do** things — start a program, acknowledge an alert, change a stage?"*

Phase 29 adds **confirmed agent actions** (never silent writes) and **UX that follows the operator** across pages.

---

## Design principles

1. **Human in the loop** — Guardian *proposes* actions; the operator taps **Confirm**. No autonomous mutations.
2. **Reuse existing APIs** — tools call the same handlers the UI uses (`PATCH /alerts/{id}/read`, `PATCH /crop-cycles/{id}/stage`, …). No parallel business logic.
3. **Farm RBAC** — tool execution respects `internal/farmauthz` (viewer cannot confirm writes).
4. **Audit everything** — confirmed actions append to `gr33ncore.audit_events` with `source=guardian_agent`.
5. **On-prem still** — tool loop runs locally; no new cloud dependencies.

---

## Scope

| WS | Focus | Primary files |
|----|-------|---------------|
| **WS1** | Global slide-out panel | `ui/src/components/GuardianDrawer.vue`, `ui/src/stores/guardianPanel.js`, `App.vue` or layout shell |
| **WS2** | Tool registry + executor | `internal/farmguardian/tools/` (new), wire into `internal/handler/chat/` |
| **WS3** | LLM tool-calling loop | Extend `POST /v1/chat` with optional `tools` mode or separate `POST /v1/chat/act` confirm endpoint |
| **WS4** | Audit + RBAC | `internal/farmauthz`, `internal/handler/audit`, confirm-gate middleware |
| **WS5** | Guardian-ready bootstrap | `Makefile`, `scripts/dev-stack.sh`, `master_seed.sql` (sample alerts), `rag-ingest` hook |
| **WS6** | OpenAPI + tests | `openapi.yaml`, `cmd/api/smoke_phase29_*`, Vitest |

---

## WS1 — Global slide-out panel

**Today:** Guardian lives at `/chat` — operator leaves whatever page they were on.

**Target:** A **drawer** (right rail on desktop, bottom sheet on mobile) available from every authenticated route:

- SideNav icon toggles drawer (does not navigate away)
- Drawer receives `farmContext.farmId` automatically
- Reuses existing `FarmGuardianChat.vue` message list + streaming
- Session sidebar collapses inside drawer (or full-screen on small viewports)

**Acceptance:** Open Zones page → ask Guardian "what's the humidity trend?" without route change.

---

## WS2 — Tool registry (v1 actions)

Start with **read-safe + low-risk writes** that already have UI buttons:

| Tool ID | Maps to | Min role |
|---------|---------|----------|
| `ack_alert` | `PATCH /alerts/{id}/acknowledge` | operator |
| `mark_alert_read` | `PATCH /alerts/{id}/read` | operator |
| `create_task_from_alert` | `POST /alerts/{id}/create-task` | operator |
| `update_cycle_stage` | `PATCH /crop-cycles/{id}/stage` | operator |
| `list_active_programs` | `GET /farms/{id}/fertigation/programs` | viewer (read) |
| `summarize_cycle` | `GET /crop-cycles/{id}/summary` | viewer (read) |

**Not in v1:** starting fertigation programs, actuator control, deleting records — too much blast radius until confirm UX is proven.

Implementation sketch:

```go
// internal/farmguardian/tools/registry.go
type Tool struct {
    ID          string
    Description string // for LLM system prompt
    MinRole     farmauthz.Role
    Execute     func(ctx, db, userID, farmID, args map[string]any) (result any, err error)
}
```

---

## WS3 — Propose → confirm flow

**Option A (recommended):** Two-step over existing chat:

1. User: "Acknowledge alert 42"
2. Guardian streams answer + embeds a **proposal card** in SSE `done` payload:
   ```json
   { "proposals": [{ "tool": "ack_alert", "args": {"alert_id": 42}, "summary": "Acknowledge humidity alert in Flower Room" }] }
   ```
3. UI renders **Confirm / Dismiss** buttons
4. User taps Confirm → `POST /v1/chat/confirm` with `proposal_id` + JWT
5. Server executes tool, returns result, Guardian summarizes outcome

**Option B:** Separate `/v1/agent/*` namespace — more surface area; defer unless chat payload gets too heavy.

Cost guard: each confirm round counts tokens like a normal turn.

---

## WS4 — Audit + RBAC

Every confirmed tool call:

```json
{
  "action": "guardian_tool_executed",
  "tool_id": "ack_alert",
  "farm_id": 1,
  "user_id": "...",
  "args": {"alert_id": 42},
  "result_status": "ok"
}
```

Reject confirm when JWT user lacks role — return 403 with plain message Guardian can relay.

---

## WS5 — Guardian-ready bootstrap

Close the gap discovered in local dev triage ([`local_dev_bugfix_todo.md`](./local_dev_bugfix_todo.md)):

1. **`make dev-stack-fresh`** — already ships clean demo DB (fixed 2026-05-20)
2. **`make dev-stack-fresh-rag`** — fresh stack + `rag-ingest` for farm 1 when `EMBEDDING_API_KEY` set (**done** 2026-05-20: `scripts/rag-ingest-demo.sh`, `make rag-ingest-demo`, `--rag-ingest` on dev-stack)
3. **Optional seed alerts** — 2–3 realistic unread alerts (sensor threshold, inventory low) so Guardian alert explainers have something to show without smoke pollution (**open**)
4. Bootstrap doc section: "Guardian demo in 3 commands" (**done** — see `docs/local-operator-bootstrap.md`)

---

## WS6 — OpenAPI + tests

- Document proposal + confirm request/response on `openapi.yaml` (0.4.0 bump)
- Smoke: propose ack_alert on seeded alert → confirm → assert `is_read` / audit row
- Vitest: drawer opens, farm context passed, confirm button calls API mock

---

## Out of scope (Phase 30+)

- Autonomous scheduling ("run this every morning without asking")
- Pi / actuator control via Guardian
- Multi-farm agent routing in one chat thread
- Cloud LLM fallback

---

## Suggested implementation order

1. **WS5** — bootstrap polish (quick win for demos)
2. **WS1** — slide-out (high operator value, no backend risk)
3. **WS2 + WS3** — one tool end-to-end (`mark_alert_read`)
4. **WS4** — audit + RBAC hardening
5. **WS6** — docs + tests
6. Expand tool registry based on operator feedback

---

## Suggested first prompt (new chat)

> "Start Phase 29 WS5 — Guardian-ready bootstrap: `make dev-stack-fresh-rag`, optional sample alerts in seed, update bootstrap doc. Then WS1 slide-out panel."
