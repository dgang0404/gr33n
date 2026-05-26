---
name: Phase 29 — Guardian Agent Layer
overview: >
  Evolve Farm Guardian from read-only Q&A (Phase 27–28) into a confirmed-action
  agent that can operate the farm through existing APIs, plus a global slide-out
  panel on every page, contextual "Ask Guardian" entry points, and a demo-ready
  bootstrap (seed sample alerts + optional RAG ingest).
todos:
  - id: ws1-slide-out-panel
    content: "WS1: Global Guardian slide-out — Pinia store + GuardianDrawer.vue, toggle from SideNav/TopBar on any route, preserve farm context + session_id, mobile bottom sheet"
    status: completed
  - id: ws2-tool-registry
    content: "WS2: Tool registry + executor — v1 tools (ack/read alert, create task, patch cycle stage, read-only list/summarize); map to existing REST handlers via farmauthz Operate cap"
    status: completed
  - id: ws3-propose-confirm-flow
    content: "WS3: Propose → confirm backend — extend SSE done payload with proposals[], POST /v1/chat/confirm, server-side proposal store with TTL + idempotency"
    status: completed
  - id: ws4-proposal-card-ui
    content: "WS4: Proposal card UI — GuardianActionProposal.vue in chat transcript; Confirm/Dismiss; viewer role shows disabled Confirm; optimistic refresh of alerts/tasks after success"
    status: completed
  - id: ws5-audit-and-rbac
    content: "WS5: Audit + RBAC — every confirmed action writes user_activity_log via internal/auditlog; RequireFarmCaps Operate gate; cost guard counts confirm round"
    status: completed
  - id: ws6-contextual-entry-points
    content: "WS6: Contextual Ask Guardian — prefill + open drawer from Alerts, CropCycleSummary, zone cards; pass contextRef in store for richer prompts"
    status: completed
  - id: ws7-demo-bootstrap
    content: "WS7: Guardian demo bootstrap — 2–3 realistic unread seed alerts; verify make dev-stack-fresh-rag path; Guardian demo in 3 commands section in bootstrap doc"
    status: completed
  - id: ws8-openapi-and-tests
    content: "WS8: OpenAPI + tests — openapi.yaml 0.4.0 (confirm endpoint + proposal shapes); smoke propose→confirm ack_alert; Vitest drawer + proposal card"
    status: completed
  - id: ws9-operator-docs
    content: "WS9: Operator + architecture docs — agent flow diagram in farm-guardian-architecture.md; README Phase 29 link; operator-tour confirmed-actions demo script"
    status: completed
isProject: false
---

# Phase 29 — Guardian Agent Layer

## Status

**Shipped (WS1–WS9).** Guardian propose→confirm path is end-to-end with audit + RBAC, OpenAPI 0.4.0, and operator docs. Phase 30 expands the tool registry and PR inbox.

**Preconditions (already met):**

- `POST /v1/chat` with streaming SSE, multi-turn sessions, farm grounding, live snapshot (zones, cycles, alerts, cycle analytics)
- Alert + crop-cycle REST endpoints Guardian will call as tools (`PATCH /alerts/{id}/read`, `POST /alerts/{id}/create-task`, `PATCH /crop-cycles/{id}/stage`, …)
- `internal/farmauthz.FarmCaps.Operate` for write gating; `internal/auditlog` for `gr33ncore.user_activity_log`
- `make dev-stack-fresh-rag` / `make rag-ingest-demo` exist from local-dev polish (rag-ingest is **not** Phase 29 scope — only the **seed alerts** piece lives here in WS7)

---

## Why this phase

Phases 27–28 shipped Farm Guardian as a **read-only** conversational layer:

- Answers questions using Llama + RAG corpus + live snapshot
- Explains alerts and crop-cycle metrics (Phase 28)
- Surfaces token usage and budget warnings

Operators asked the natural next question: *"Can it actually **do** things — acknowledge an alert, create a task, change a stage?"*

Phase 29 adds **confirmed agent actions** (never silent writes) and **UX that follows the operator** across pages.

---

## Design principles

1. **Human in the loop** — Guardian *proposes* actions; the operator taps **Confirm**. No autonomous mutations.
2. **Reuse existing APIs** — tools invoke the same handler logic the dashboard uses. No parallel business rules.
3. **Farm RBAC** — writes require `FarmCaps.Operate` (operator/worker/agronomist/manager/owner). Viewers may chat and see proposals but **Confirm is disabled**.
4. **Audit everything** — confirmed actions append to `gr33ncore.user_activity_log` via `internal/auditlog` with `action_type` like `guardian_tool_executed` (add enum value if needed).
5. **Frozen proposals** — args are stored server-side at propose time; confirm replays the stored payload, not client-supplied args (prevents tampering).
6. **On-prem still** — tool loop runs locally; no new cloud dependencies.

---

## Scope

| WS | Focus | Primary files |
|----|-------|---------------|
| **WS1** | Global slide-out panel | `ui/src/components/GuardianDrawer.vue`, `ui/src/stores/guardianPanel.js`, `App.vue`, `SideNav.vue`, `TopBar.vue` |
| **WS2** | Tool registry + executor | `internal/farmguardian/tools/` (new), shared service layer callable from chat confirm |
| **WS3** | Propose → confirm backend | `internal/handler/chat/proposals.go`, `internal/handler/chat/confirm.go`, `cmd/api/routes.go` |
| **WS4** | Proposal card UI | `ui/src/components/GuardianActionProposal.vue`, extend `FarmGuardianChat.vue` SSE parser |
| **WS5** | Audit + RBAC | `internal/auditlog`, `internal/farmauthz`, tool executor |
| **WS6** | Contextual entry points | `Alerts.vue`, `CropCycleSummary.vue`, zone views, `guardianPanel` store |
| **WS7** | Demo bootstrap | `db/seeds/master_seed.sql`, `docs/local-operator-bootstrap.md` |
| **WS8** | OpenAPI + tests | `openapi.yaml`, `cmd/api/smoke_phase29_*`, Vitest |
| **WS9** | Operator docs | `docs/farm-guardian-architecture.md`, `README.md`, `docs/operator-tour.md` |

---

## Work-stream detail

### WS1 — Global slide-out panel

**Today:** Guardian lives at `/chat` — operator leaves whatever page they were on.

**Target:** A **drawer** (right rail on desktop, bottom sheet on mobile) available from every authenticated route:

- SideNav **Guardian** icon toggles drawer instead of navigating to `/chat` (keep `/chat` as deep-link / full-page fallback for now)
- TopBar optional sparkle/Guardian button for discoverability (especially mobile where SideNav is hidden)
- Drawer receives `farmContext.farmId` automatically; "Use farm context" defaults **on** when a farm is selected
- Extract chat body from `FarmGuardianChat.vue` into a shared `GuardianChatPanel.vue`; full page wraps panel + wide session sidebar; drawer uses compact session picker
- Pinia `useGuardianPanelStore`: `{ open, toggle, prefilledMessage, contextRef, activeSessionId }`
- z-index above main content; does not steal focus from modals already open

**Acceptance:** Open Zones → toggle Guardian → ask "what's the humidity trend?" → drawer streams answer → close drawer → still on Zones.

**Tests:** Vitest — store toggle, drawer mounts, farm id passed to chat POST body.

---

### WS2 — Tool registry (v1)

Start with **read-safe + low-risk writes** that already have UI buttons:

| Tool ID | Maps to | Cap required |
|---------|---------|--------------|
| `mark_alert_read` | `PATCH /alerts/{id}/read` | Operate |
| `ack_alert` | `PATCH /alerts/{id}/acknowledge` | Operate |
| `create_task_from_alert` | `POST /alerts/{id}/create-task` | Operate |
| `update_cycle_stage` | `PATCH /crop-cycles/{id}/stage` | Operate |
| `list_active_programs` | `GET /farms/{id}/fertigation/programs` | member (read) |
| `summarize_cycle` | `GET /crop-cycles/{id}/summary` | member (read) |

**Not in v1:** starting fertigation programs, actuator control, deleting records — too much blast radius until confirm UX is proven.

Implementation sketch:

```go
// internal/farmguardian/tools/registry.go
type Tool struct {
    ID          string
    Description string // injected into LLM system prompt
    RequiresOperate bool
    Execute     func(ctx context.Context, deps ExecutorDeps, args map[string]any) (result any, err error)
}
```

Executor calls existing handler/service functions **in-process** (not HTTP loopback) so auth context + audit stay in one place. Each tool validates args (positive int64 ids, allowed stage enum values) before touching DB.

**LLM exposure:** Append a `## Available actions` block to the grounded system prompt listing tool IDs + one-line descriptions. Phase 29 v1 uses **structured JSON in the assistant turn** (see WS3) rather than native Ollama tool-calling — simpler, works with current Llama 3.1 setup, no model-specific function-call API dependency.

---

### WS3 — Propose → confirm backend

**Flow (recommended — two-step over existing chat):**

1. User: "Acknowledge alert 42"
2. Guardian streams natural-language answer; on `event: done`, payload includes:
   ```json
   {
     "proposals": [{
       "proposal_id": "550e8400-e29b-41d4-a716-446655440000",
       "tool": "ack_alert",
       "args": {"alert_id": 42},
       "summary": "Acknowledge humidity alert in Flower Room",
       "expires_at": "2026-05-20T15:05:00Z"
     }]
   }
   ```
3. Server persisted proposal in **`gr33ncore.guardian_action_proposals`** (new migration) or ephemeral table with TTL — stores `proposal_id`, `user_id`, `farm_id`, `session_id`, `tool`, frozen `args` JSON, `created_at`, `expires_at`, `status` (`pending`|`confirmed`|`dismissed`|`expired`).
4. User taps Confirm → `POST /v1/chat/confirm` `{ "proposal_id": "..." }` + JWT
5. Server loads proposal, checks: same user, not expired, status pending, farm membership + Operate cap for writes
6. Executes tool, writes audit row, marks proposal confirmed, returns `{ "result": {...}, "summary": "Alert acknowledged." }`
7. Optional: append a system message to the session transcript ("✓ Alert #42 acknowledged") so history reflects the action

**Proposal detection (v1 heuristic):**

- After LLM completes, run a lightweight parser on the assistant text for a fenced JSON block ` ```guardian_proposal\n{...}\n``` ` **or** use a second non-streaming "extract proposals" pass with a tiny prompt (only when user message matches action intent keywords — defer if too fragile; start with explicit JSON instruction in system prompt).
- Safer v1: **rule-assisted** — if snapshot includes unread alert IDs and user message mentions "acknowledge/read alert", template a proposal without relying on model JSON reliability for the first ship.

**Safety:**

| Rule | Behavior |
|------|----------|
| TTL | Default 5 minutes; expired → 410 Gone on confirm |
| Idempotency | Second confirm on already-confirmed proposal → 200 with cached result |
| Tampering | Confirm body only accepts `proposal_id`; args come from DB row |
| Farm scope | Proposal `farm_id` must match tool target farm |
| Cost guard | Confirm round counts like a chat turn (reuse `persistTurn` or lightweight token estimate) |

**Defer:** Separate `/v1/agent/*` namespace — only if chat payload becomes unwieldy.

---

### WS4 — Proposal card UI

- New **`GuardianActionProposal.vue`**: shows summary, tool label, target ids, **Confirm** / **Dismiss**
- Rendered inline in transcript when `done.proposals[]` present
- Confirm calls `POST /v1/chat/confirm`; show spinner; on success replace card with green "Done" chip + optional link ("View task #123")
- Dismiss → local state only + optional `PATCH` proposal status dismissed (or ignore server-side)
- **Viewer role:** detect via farm caps endpoint or local role cache — Confirm disabled + tooltip "Operators only"
- After `ack_alert` / `create_task_from_alert`, emit event bus or store refresh so Alerts/Tasks views update if open behind drawer

**Acceptance:** Seeded alert → ask Guardian to acknowledge → proposal card → Confirm → alert moves to read in DB + UI refreshes.

---

### WS5 — Audit + RBAC

Every confirmed tool call:

```go
auditlog.Submit(ctx, q, r, auditlog.Event{
    FarmID: auditlog.FarmIDPtr(farmID),
    Action: db.Gr33ncoreUserActionTypeEnumGuardianToolExecuted, // add if missing
    TargetSchema: strPtr("gr33ncore"),
    TargetTable:  strPtr("alerts_notifications"), // per tool
    TargetRecordID: strPtr(strconv.FormatInt(alertID, 10)),
    Status: "success",
    Details: map[string]any{
        "tool_id": "ack_alert",
        "proposal_id": proposalID,
        "args": args,
    },
})
```

- Reject confirm when `RequireFarmCaps(..., Operate)` fails → **403** with message Guardian UI can display
- Read-only tools (`summarize_cycle`) may execute without Operate but still log `guardian_tool_read` at info level (optional — or skip audit for pure reads in v1)

Farm admins already have `GET /farms/{id}/audit-events` — no new list UI required in v1; document the new action types in [`audit-events-operator-playbook.md`](../audit-events-operator-playbook.md).

---

### WS6 — Contextual "Ask Guardian" entry points

High-value shortcuts that make the slide-out feel native:

| Source page | Trigger | Prefill / context |
|-------------|---------|-------------------|
| **Alerts** list / detail | "Ask Guardian" on unread row | `Explain alert #{{id}} and suggest next steps` + `contextRef: { type: 'alert', id }` |
| **CropCycleSummary** | Button in header | `Summarize this cycle and compare to typical flower targets` + `contextRef: { type: 'crop_cycle', id }` |
| **Zones** card | "Ask about this zone" | `What's the current status of {{zoneName}}?` |
| **Tasks** (optional) | On overdue task | `Help me prioritize this task` |

Store opens drawer + sets prefilled message in input (operator can edit before send). Backend may use `contextRef` in WS3+ to inject extra snapshot detail without operator typing ids.

**Acceptance:** Alerts page → Ask Guardian on humidity alert → drawer opens with alert-specific question prefilled.

---

### WS7 — Guardian demo bootstrap

Close gaps from local dev triage ([`local_dev_bugfix_todo.md`](./local_dev_bugfix_todo.md)):

1. **`make dev-stack-fresh`** — clean demo DB (already works)
2. **`make dev-stack-fresh-rag`** — optional RAG corpus (scripts exist; verify end-to-end once with `EMBEDDING_API_KEY`)
3. **Sample seed alerts** — insert 2–3 **unread** `alerts_notifications` rows for farm 1, e.g.:
   - `medium` — inventory low (OHN batch below threshold) — exercises inventory explainers
   - `high` — sensor threshold (humidity > 72% in Flower Room) — ties to seeded zone names
   - `low` — schedule reminder (light transition in 48h) — benign, good for dismiss/ack demos
   Use realistic `subject_rendered` / `message_text_rendered`; set `triggering_event_source_type` + `source_id` to seeded sensor or rule ids where possible.
4. **Bootstrap doc** — add a **"Guardian agent demo in 3 commands"** box:
   ```bash
   make dev-stack-fresh-rag    # or dev-stack-fresh if no embedding key
   make restart-local-serve
   # open dashboard → toggle Guardian → "acknowledge the humidity alert"
   ```

**Acceptance:** Fresh volume → login → `/chat` or drawer sees 3 unread alerts in snapshot; propose→confirm works without manual SQL.

---

### WS8 — OpenAPI + tests

- Bump **`openapi.yaml` to 0.4.0**
- Document:
  - Extended SSE `done` event `proposals[]` shape on `POST /v1/chat`
  - New **`POST /v1/chat/confirm`** request/response
  - Proposal object schema (`proposal_id`, `tool`, `args`, `summary`, `expires_at`)
- **`cmd/api/smoke_phase29_test.go`**:
  - Seed or use WS7 alerts
  - POST chat (non-streaming test mode or parse SSE) → assert proposal returned
  - POST confirm → assert alert acknowledged + audit row + idempotent second confirm
  - Viewer JWT → confirm returns 403
- **Vitest:** `guardian-panel.test.js`, `guardian-proposal.test.js`
- Run gate: `make test`, `make audit-openapi`, `npm --prefix ui run test`

---

### WS9 — Operator + architecture docs

- **`docs/farm-guardian-architecture.md`** — new § "Agent actions (Phase 29)": propose→confirm diagram, tool list, audit trail pointer
- **`README.md`** — replace "Phase 29 plan doc TBD" with link to this plan; summarize confirmed-action scope
- **`docs/operator-tour.md`** — add **"Guardian can act (with your OK)"** stop: acknowledge alert, create task, slide-out from Alerts page
- **`docs/local-operator-bootstrap.md`** — cross-link WS7 demo commands (partially done; align after WS7 seed alerts land)
- **Guardian scope note** in Settings or drawer footer: "Guardian proposes changes; you confirm. All confirmed actions appear in the farm audit log."

---

## Out of scope (Phase 30+)

Phase 30 is the **Guardian change-request (PR) queue** — expanded confirm tools, config patches, actuator enqueue via `pending_command`, optional zone vision. See [`phase_30_guardian_change_requests.plan.md`](phase_30_guardian_change_requests.plan.md). Phase 31 is **field / Pi bench validation** — [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md).

**Not in Phase 29 (defer to Phase 30–31):**

- Full farm configuration agent (schedules, programs, Pi) — Phase 30 PR inbox
- Direct Pi / actuator control without proposal + Confirm — Phase 30 `enqueue_actuator_command` PR
- Zone photos + vision-based PRs — Phase 30 WS5–WS6
- Autonomous scheduling ("run this every morning without asking") — never; use **alerts + rules**
- Multi-farm agent routing in one chat thread
- Cloud LLM fallback
- Native Ollama `/api/chat` tool-calling API (revisit when model stack supports it reliably)
- Guardian-initiated push notifications
- **500-site enterprise fleet management** — hypothetical only: [`docs/hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md)

**Operator expectations at Phase 29 ship:** Guardian **advises** on defoliation, plumbing, setup (chat); **confirmed writes** are alert ack/read only. Copilot vs actor split is documented fully in Phase 30 WS7.

---

## Suggested implementation order

1. **WS7** — seed alerts + verify demo path (unblocks all smokes and manual QA)
2. **WS1** — slide-out shell (high operator value, no backend risk)
3. **WS2 + WS3** — one tool end-to-end (`mark_alert_read`) with proposal store + confirm endpoint
4. **WS4** — proposal card UI wired to confirm
5. **WS5** — audit + RBAC hardening on confirm path
6. **WS6** — contextual entry points (Alerts first)
7. **WS8** — OpenAPI + smokes (can start partial alongside WS3)
8. **WS9** — docs pass once one confirm path is green
9. Expand tool registry (`create_task_from_alert`, `update_cycle_stage`, read tools)

---

## Definition of done (phase ship)

- [x] Guardian drawer opens from any authenticated page; `/chat` still works
- [x] Operator can get a proposal + confirm **ack/read alert** without leaving Alerts context
- [x] Viewer cannot confirm writes (403 or disabled UI)
- [x] Confirmed actions appear in `GET /farms/{id}/audit-events`
- [x] `make dev-stack-fresh` + seed → 3 demo alerts → manual demo script works
- [x] `make audit-openapi` green at 0.4.0; smoke_phase29 passes
- [x] Architecture + operator docs updated

---

## Using this plan in a new chat

```text
Implement Phase 29 per @docs/plans/phase_29_guardian_agent_layer.md.

Start with WS7 (sample unread alerts in master_seed.sql + verify make dev-stack-fresh),
then WS1 (GuardianDrawer + guardianPanel store). Follow the implementation order in the plan.
Before WS3, read internal/handler/chat/handler.go and internal/auditlog/auditlog.go for patterns.
Add openapi.yaml entries as you add routes (partial WS8). Run go test ./cmd/api/... after each WS.
```

**Suggested first prompt (shorter):**

> "Start Phase 29 WS7 — seed 3 unread demo alerts for farm 1, then WS1 Guardian slide-out drawer. Plan: @docs/plans/phase_29_guardian_agent_layer.md"

---

## Risk register

| Risk | Mitigation |
|------|------------|
| LLM emits invalid proposal JSON | Rule-assisted proposals for v1 alert ack; JSON block parser as enhancement |
| Double-submit on Confirm | Idempotent confirm + UI disable while pending |
| Drawer + mobile nav z-index fights | Follow existing App.vue left-drawer patterns; test iPhone safe-area |
| Tool executor diverges from REST handlers | Executor calls same db.Queries methods handlers use — extract shared functions where needed |
| Proposal table migration on existing DBs | Standard forward migration; no backfill required |

---

## Shipped notes

### WS5 — Audit + RBAC (shipped 2026-05-20)

- **`guardian_tool_executed`** — new `user_action_type_enum` value + migration `20260522_phase29_guardian_audit_enum.sql`.
- **`POST /v1/chat/confirm`** — `RequireFarmOperate` gate (403 for viewer); `checkCostBudget` on confirm; audit success/failure with `details.kind: guardian_tool_executed`.
- **`docs/audit-events-operator-playbook.md`** — documents the new action type.
- **Smoke** — `smoke_phase29_ws5_test.go` (viewer 403, audit row on confirm).

### WS4 — Proposal card UI (shipped 2026-05-20)

- **`GuardianActionProposal.vue`** — inline card: summary, tool label, target id, Confirm/Dismiss; done chip + link to Alerts after success.
- **`GuardianChatPanel.vue`** — renders `proposals[]` from SSE `done`; refreshes unread count + alert list via `farmStore` after ack/read.
- **`useFarmOperate.js`** — disables Confirm for viewer/finance roles (tooltip: Operators only).
- **Vitest** — `guardian-proposal.test.js`, `guardian-chat-proposals.test.js`.

### WS2 + WS3 — Tool registry + propose→confirm (shipped 2026-05-20)

- **`internal/farmguardian/tools/`** — registry + in-process executor; v1 write tools `ack_alert`, `mark_alert_read` (farm-scoped args).
- **`internal/farmguardian/proposals.go`** — rule-assisted proposals when grounded chat mentions ack/read + snapshot has unread alerts; 5-minute TTL.
- **`gr33ncore.guardian_action_proposals`** — migration `20260521_phase29_guardian_proposals.sql` + schema + hand-written queries.
- **`POST /v1/chat/confirm`** — frozen-args replay, Operate cap, idempotent re-confirm, audit via `execute_action`.
- **Chat `done` payload** — `proposals[]` on streaming and non-streaming turns.
- **Tests** — `proposals_test.go`, `smoke_phase29_ws3_test.go` (confirm ack on seeded humidity alert).

### WS7 — Guardian demo bootstrap (shipped 2026-05-20)

- **`db/seeds/master_seed.sql` v1.006** — `SEED-OHN-001` batch (0.35 L remaining) plus three idempotent unread alerts for farm 1: medium inventory (OHN), high humidity (Flower Room / Air Humidity Indoor sensor), low schedule reminder (Light OFF 12/12 Flower).
- **`docs/local-operator-bootstrap.md`** — **Guardian agent demo in 3 commands** box (`dev-stack-fresh-rag` → `restart-local-serve` → drawer prompt).
- **`cmd/api/smoke_phase29_ws7_test.go`** — asserts all three seed subjects exist unread on farm 1.

### WS1 — Global slide-out panel (shipped 2026-05-20)

- **`ui/src/stores/guardianPanel.js`** — Pinia store: `open`, `toggle`, `openDrawer`, `prefilledMessage`, `contextRef`, `activeSessionId` (shared across drawer and `/chat`).
- **`ui/src/components/GuardianChatPanel.vue`** — extracted chat body; `layout="full"` (session sidebar) vs `layout="compact"` (dropdown picker). Farm context defaults **on** when a farm is selected.
- **`ui/src/components/GuardianDrawer.vue`** — Teleport to body; right rail on `md+`, bottom sheet on mobile; z-index 40; footer scope note; link to full `/chat` page.
- **`ui/src/views/FarmGuardianChat.vue`** — thin wrapper around `GuardianChatPanel` full layout.
- **`App.vue`**, **`SideNav.vue`** (Guardian button toggles drawer), **`TopBar.vue`** (✨ toggle when AI enabled).
- **Vitest** — `ui/src/__tests__/guardian-panel.test.js` (store, drawer teleported mount, farm_id in chat POST, route preserved on toggle).

### WS6 — Contextual Ask Guardian entry points (shipped 2026-05-20)

- **`ui/src/components/AskGuardianButton.vue`** — reusable ✨ trigger; opens drawer with prefilled prompt + `contextRef`.
- **Entry points** — unread rows on `Alerts.vue`; header on `CropCycleSummary.vue`; zone cards on `Zones.vue` + header on `ZoneDetail.vue`.
- **`internal/farmguardian/context_ref.go`** — `ContextRefPromptBlock` injects focused alert/cycle/zone detail into grounded system prompt when `context_ref` is POSTed.
- **`GuardianChatPanel.vue`** — sends `context_ref` on `/v1/chat`; clears prefill after send.
- **Vitest** — `ui/src/__tests__/guardian-context-entry.test.js`.

### WS8 — OpenAPI + tests (shipped 2026-05-20)

- **`openapi.yaml` 0.4.0** — `POST /v1/chat/confirm`, `GuardianActionProposal`, `GuardianContextRef`, `ChatConfirmRequest/Response`; `proposals[]` on chat responses and SSE `done`; `context_ref` on `ChatRequest`.
- **`cmd/api/smoke_phase29_test.go`** — confirm auth/validation smokes; full chat→proposal→confirm→audit E2E when LLM is configured (skips gracefully otherwise). Complements `smoke_phase29_ws3/ws5_test.go`.
- **Vitest** — `guardian-panel.test.js`, `guardian-proposal.test.js`, `guardian-chat-proposals.test.js` (WS4 + WS1 coverage gate).

### WS9 — Operator + architecture docs (shipped 2026-05-20)

- **`docs/farm-guardian-architecture.md`** — §7 Agent actions (propose→confirm diagram, v1 tool table, audit pointer); code map + phase ledger updated.
- **`README.md`** — Phase 29 shipped; Guardian section reflects confirmed ack/read scope; roadmap checkboxes.
- **`docs/operator-tour.md`** — §6 **Guardian can act (with your OK)** demo script (Alerts → Ask Guardian → Confirm).
- **`docs/local-operator-bootstrap.md`** — Guardian demo box includes ack confirm step; links to architecture §7.
- **`ui/src/views/Settings.vue`** + **`GuardianDrawer.vue`** — scope note: proposes / you confirm / audit log.
