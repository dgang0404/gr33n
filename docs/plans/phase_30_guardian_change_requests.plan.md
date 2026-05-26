---
name: Phase 30 — Guardian change requests (PR queue)
overview: >
  Evolve Phase 29 propose→confirm into a farm-configuration agent that is never
  autonomous: a pending-request inbox (like pull requests) for schedules, programs,
  tasks, and Pi actuator commands. Optional zone photos + vision model for agronomic
  feedback—all changes require explicit operator Confirm; alerts/automation rules
  remain the autonomous safety layer.
todos:
  - id: ws1-pr-inbox-ui
    content: "WS1: Pending requests inbox — list open Guardian proposals (pending/expired); drawer tab + optional /guardian/requests; badge count in TopBar"
    status: pending
  - id: ws2-risk-tiers
    content: "WS2: Risk tiers on proposals — low/medium/high; high-tier (actuator, delete) requires Operate + optional stricter cap; extend guardian_action_proposals metadata"
    status: pending
  - id: ws3-config-tools
    content: "WS3: Configuration tools — create_task_from_alert, update_cycle_stage, patch schedule/program/rule (scoped patches); rule-assisted + LLM proposals"
    status: pending
  - id: ws4-actuator-pr-tool
    content: "WS4: Actuator PR tool — enqueue_actuator_command → device pending_command JSON; frozen args; audit; no auto-execute without Confirm"
    status: pending
  - id: ws5-zone-images
    content: "WS5: Zone images — attach photos to zones (file storage + zone meta_data or link table); show in Zone UI + pass URLs into Guardian snapshot"
    status: pending
  - id: ws6-vision-chat
    content: "WS6: Vision chat (optional) — multimodal LLM env (e.g. llava via Ollama); attach zone photo to chat turn; proposals still confirm-only; agronomic disclaimer"
    status: pending
  - id: ws7-operator-expectations-doc
    content: "WS7: Operator expectations — what Guardian is/isn't at ship; copilot vs actor; human tasks (defoliation, plumbing); alerts vs PRs"
    status: pending
  - id: ws8-openapi-tests
    content: "WS8: OpenAPI + smokes — list proposals API; confirm actuator PR; Vitest inbox; vision skipped in CI unless env set"
    status: pending
isProject: false
---

# Phase 30 — Guardian change requests (PR queue)

## Status

**Not started.** Depends on Phase 29 **WS3–WS5** (proposal store + confirm + card UI) being stable. Phase 31 validates that confirmed **actuator PRs** reach real Pi GPIO.

---

## Why this phase

Phase 29 answered: *"Can Guardian ack an alert with my OK?"*

Phase 30 answers: *"Can Guardian help me **run the farm** — programs, schedules, tasks, and eventually Pi commands — without ever going autonomous?"*

**Design answer:** treat every write like a **pull request**:

1. Guardian (chat + snapshot + optional photo) **opens a change request**.
2. Request appears in a **pending inbox** (and inline in chat).
3. Operator **Confirm** or **Dismiss** — same as Phase 29, expanded scope.
4. **Alerts and automation rules** stay the always-on layer; Guardian PRs are **intentional, reviewed changes**.

This is a **configuration agent**, not an autopilot.

---

## What Guardian is / isn't at phase ship

| Guardian **is** | Guardian **is not** |
|-----------------|---------------------|
| Copilot: explain, suggest, walkthrough setup | Autonomous scheduler ("do this every morning" without Confirm) |
| PR author: propose DB + Pi-enqueue changes | Silent writer |
| Vision assistant (WS6): comment on zone photos | Certified agronomist or IPM diagnosis guarantee |
| Task creator, program/schedule patch proposer | Replacement for human defoliation, plumbing, harvest |
| Pointer to UI when it lacks data | Pi firmware flasher |

**Human / humanoid work** (defoliation, plumbing, cleaning, harvest) stays **guidance in chat** — optionally a **`create_task`** PR so it appears on the task board.

---

## Architecture (PR queue)

```
Operator chat (+ optional zone photo)
        │
        ▼
  Guardian proposes change(s)
        │
        ▼
  gr33ncore.guardian_action_proposals  (status: pending)
        │
        ├──► Inbox UI (all pending for farm/user)
        └──► Chat transcript cards (Phase 29)
        │
        ▼
  Operator Confirm ──► tools.Execute ──► REST-equivalent DB / pending_command
        │
        ▼
  user_activity_log (guardian_tool_executed) + side effects
        │
        ▼  (Phase 31)
  Pi polls pending_command → GPIO → actuator_events
```

**No autonomous path:** the automation **worker** evaluates **rules** on sensor readings; it does **not** consume Guardian PRs. Conversely, Guardian does **not** fire rules directly.

---

## Design principles

1. **Human in the loop always** — zero silent mutations.
2. **Frozen args at propose time** — confirm replays server store, not client JSON.
3. **Risk tiers** — actuator and destructive ops labeled **high** in UI; optional manager-only cap later.
4. **Reuse handlers** — tools call same logic as dashboard (`internal/farmguardian/tools/` in-process).
5. **Alerts ≠ PRs** — system/automation generates **alerts**; Guardian generates **change requests** for operator-approved fixes.
6. **Vision is advisory** — image-derived proposals get the same Confirm gate; model can be wrong.
7. **Accountability by default** — every confirm writes **who**, **when**, **what**, and **frozen args** to durable tables for humans and future AI reporting (see below).

---

## Accountability & reporting (who changed what)

Operators and future analytics need to distinguish **Guardian-suggested** changes from **manual dashboard** edits and **automation** events. Phase 30 builds on Phase 29 storage — no silent writes.

### Already shipped (Phase 29)

| Layer | What is recorded | Who |
|-------|------------------|-----|
| **`guardian_action_proposals`** | Full proposal lifecycle: `user_id`, `farm_id`, `session_id`, `tool_id`, frozen `args`, `summary`, `status`, `created_at`, `expires_at`, `confirmed_at`, `result` JSON | User who chatted (proposer); only they can Confirm today |
| **`user_activity_log`** | On Confirm: `action_type = guardian_tool_executed`, `user_id` from JWT, `details`: `proposal_id`, `tool_id`, `args`, `result` | **Human who tapped Confirm** |
| **Chat history** | `conversation_turns` per `session_id` + user | Full Q&A transcript (what Guardian said vs what operator asked) |
| **Target rows** | e.g. alert ack stores acknowledging user on the alert row | Same Confirm actor |

Query farm audit: **`GET /farms/{id}/audit-events`** — see [`audit-events-operator-playbook.md`](../audit-events-operator-playbook.md) (`guardian_tool_executed` row).

### Phase 30 additions (planned)

| Gap today | Phase 30 fix |
|-----------|----------------|
| Dismiss is mostly UI-local | Persist `dismissed` on server + optional audit row (`guardian_proposal_dismissed`) |
| Inbox / list API | `GET /v1/chat/proposals` — filter by farm, user, status, date for reporting exports |
| Config tool targets | Audit `target_table` / `target_record_id` for schedules, programs, tasks — not only alerts |
| Actuator PR | Audit + `actuator_events` + `pending_command` JSON includes **`proposal_id`** in device config or event meta for traceability |
| Risk tier | Stored on proposal for compliance reports ("who approved high-tier Pi commands") |
| Guardian vs manual | Dashboard PATCH continues `update_record` audit; Guardian path always has **`proposal_id`** link in `details` |

### What automation records (separate trail)

**Rules / worker / alerts** are **not** Guardian PRs — they log to `automation_events`, `alerts_notifications`, etc. Reporting should join:

- **Human-approved:** `guardian_action_proposals` + `guardian_tool_executed`
- **Autonomous:** rule firings + system alerts
- **Manual UI:** existing `user_activity_log` `update_record` / `create_record`

Future AI ("what changed last week?") can RAG-ingest audit rows + proposal table — already aligned with Phase 25 ingest boundaries (operational logs policy in [`rag-scope-and-threat-model.md`](../rag-scope-and-threat-model.md)).

**Acceptance (Phase 30 WS8):** Smoke asserts audit row contains `user_id`, `proposal_id`, `tool_id`; confirmed proposal row has `confirmed_at` and `result`.

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Pending inbox UI | `GuardianRequestsInbox.vue`, store, `GET /v1/chat/proposals` (new) |
| **WS2** | Risk tiers | `risk_tier` on proposals; UI badges; confirm gate |
| **WS3** | Config tools | schedules, programs, rules, tasks, cycle stage |
| **WS4** | Actuator PR | `enqueue_actuator_command` → `devices.config.pending_command` |
| **WS5** | Zone images | zone photo upload + snapshot links |
| **WS6** | Vision chat | multimodal LLM path; optional |
| **WS7** | Docs | operator expectations, architecture diagram |
| **WS8** | OpenAPI + tests | smokes, Vitest |

---

## Work-stream detail

### WS1 — Pending requests inbox

**Today:** Proposals only visible inline in chat after a turn.

**Target:**

- **`GET /v1/chat/proposals?farm_id=&status=pending`** — JWT + farm member; paginated.
- Drawer **tab** or slide: "Pending (N)" listing open PRs with summary, tool, expiry, Confirm/Dismiss.
- TopBar badge when `N > 0` (optional).
- Deep link `/guardian/requests` for managers reviewing overnight queue.

**Acceptance:** Ack PR from Phase 29 appears in inbox; confirm from inbox matches confirm from chat card.

---

### WS2 — Risk tiers

Extend `guardian_action_proposals` (migration) with `risk_tier`: `low` | `medium` | `high`.

| Tier | Examples | UX |
|------|----------|-----|
| **low** | mark_alert_read | standard Confirm |
| **medium** | patch schedule time, update fertigation EC target, create_task | Confirm + short diff summary |
| **high** | enqueue_actuator_command, delete/disable rule | Confirm + warning copy; future: manager-only |

**Acceptance:** High-tier card shows warning; viewer role still cannot confirm writes.

---

### WS3 — Configuration tools (farm setup agent)

Expand registry beyond Phase 29 v1:

| Tool ID | Maps to | Tier |
|---------|---------|------|
| `create_task_from_alert` | `POST /alerts/{id}/create-task` | medium |
| `create_task` | `POST /farms/{id}/tasks` | medium |
| `update_cycle_stage` | `PATCH /crop-cycles/{id}/stage` | medium |
| `patch_fertigation_program` | `PATCH /fertigation/programs/{id}` (allowlisted fields) | medium |
| `patch_schedule` | `PATCH /schedules/{id}` (allowlisted fields) | medium |
| `patch_rule` | `PATCH /rules/{id}` (enable + threshold patches only v1) | medium |
| `apply_bootstrap_template` | existing bootstrap apply (if args frozen at propose) | high |

**Not v1:** bulk delete, cross-farm apply, run fertigation now without schedule.

**Rule-assisted proposals** (like Phase 29 ack): keyword + snapshot intent → template proposal when LLM JSON unreliable.

**Acceptance:** "Create a task to check Flower Room humidity" → PR → Confirm → task visible on Tasks page + audit row.

---

### WS4 — Actuator PR tool (Pi control via Confirm)

**This is the safe Pi path** — not chat → GPIO directly.

```json
{
  "tool": "enqueue_actuator_command",
  "args": {
    "device_id": 12,
    "actuator_id": 4,
    "command": "on",
    "reason": "Guardian: operator requested veg room lights on for inspection"
  },
  "risk_tier": "high"
}
```

**Execute:** same code path as dashboard/worker enqueue → `devices.config.pending_command` (see [`smoke_pi_contract_test.go`](../../cmd/api/smoke_pi_contract_test.go)).

**Pi client** unchanged — polls devices, executes, posts `actuator_events`, clears pending.

**Acceptance:** Confirm PR → `GET /farms/{id}/devices` shows pending_command → Phase 31 bench proves GPIO.

**Out of scope:** Guardian bypassing proposal store; Guardian calling Pi HTTP directly.

---

### WS5 — Zone images

**Goal:** Operators attach **reference / walkthrough photos** per zone; Guardian sees that photos exist in snapshot.

**Tasks:**

- Reuse [`internal/handler/fileattach`](../../internal/handler/fileattach) + [`filestorage`](../../internal/filestorage) (receipts pattern).
- Link attachment IDs in `zones.meta_data` (e.g. `photo_attachment_ids: []`) or small `zone_media` table if cleaner.
- Zone detail UI: upload + thumbnail gallery.
- Snapshot block: zone name + latest photo URL + "ask Guardian about this zone's photo."

**Acceptance:** Photo on Flower Room zone; chat with farm context mentions photo available; optional WS6 analysis.

---

### WS6 — Vision chat (optional — is it realistic?)

**Short answer: yes, with constraints** — not with default **text-only** `llama3.1:8b`.

| Piece | Realistic? | Notes |
|-------|------------|-------|
| Store zone photos (WS5) | ✅ | File storage already exists |
| Show image in chat UI | ✅ | Standard upload + preview |
| Model ** understands** crop photos | ⚠️ | Requires **multimodal** model: e.g. **LLaVA**, **llama3.2-vision**, or cloud vision API via `LLM_*` |
| Accurate IPM / deficiency diagnosis | ⚠️ | Good for **flags** ("possible wilting, check irrigation"); bad as sole authority |
| PR from image ("create task: inspect leaf spot") | ✅ | Same Confirm gate — **recommended** over auto rule changes |

**Implementation sketch:**

- Env: `LLM_VISION_MODEL`, `LLM_VISION_BASE_URL` (or reuse Ollama with vision tag).
- Chat request: optional `attachment_ids[]` or inline base64 (size cap).
- System prompt: vision outputs are **hypotheses**; destructive/high-tier PRs need human verification.
- **CI:** skip vision smokes unless `GR33N_VISION_TEST=1`.

**v1 ship option:** WS5 only (photos in UI + snapshot text "photo on file") and defer WS6 to Phase 30.1 if multimodal setup is heavy.

**Acceptance (full WS6):** Upload zone photo → ask "anything wrong with these leaves?" → Guardian describes observations in prose → optional medium-tier `create_task` PR.

---

### WS7 — Operator expectations doc

Add to [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) and operator tour:

- Copilot vs actor vs automation/alerts
- PR inbox workflow
- Vision limits disclaimer
- Link Phase 31 for Pi bench validation

---

### WS8 — OpenAPI + tests

- `GET /v1/chat/proposals`, extend proposal schema with `risk_tier`
- Smoke: config tool PR → confirm → row changed
- Smoke: actuator PR → pending_command present
- Vitest: inbox list, high-tier warning copy

---

## Out of scope (Phase 31+)

- **Autonomous Guardian** scheduling or actuation
- Guardian PRs that apply to **multiple farms** in one confirm
- Native Ollama function-calling (keep rule-assisted + optional JSON block)
- Replacing automation rules with Guardian
- Certified agricultural diagnosis
- Phase 31 items: breadboard validation, MQTT patterns (see [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md))

---

## Suggested implementation order

1. **WS1** — inbox (unblocks "PR queue" UX immediately)
2. **WS2** — risk tiers
3. **WS3** — medium config tools (tasks, cycle stage, then schedule/program)
4. **WS4** — actuator PR tool
5. **WS5** — zone photos
6. **WS6** — vision (optional / follow-up)
7. **WS8** — OpenAPI + smokes alongside WS3–WS4
8. **WS7** — doc pass at end

Phase 29 **WS6–WS9** should complete first.

---

## Definition of done (phase ship)

- [ ] Pending Guardian PRs listable outside chat thread
- [ ] Operator can confirm **task**, **schedule/program patch**, and **actuator enqueue** PRs with audit trail
- [ ] High-tier PRs show clear warning; viewers cannot confirm
- [ ] Zone photos attachable; snapshot references them
- [ ] Vision chat documented as optional; agronomic disclaimer in UI
- [ ] Docs: Guardian is not autonomous; alerts remain separate
- [ ] Phase 31 can bench-test actuator PR → Pi

---

## Using this plan in a new chat

```text
Implement Phase 30 per @docs/plans/phase_30_guardian_change_requests.plan.md.

Start with WS1 (GET proposals + inbox UI). Extend Phase 29 proposal store — do not
bypass Confirm. Actuator tool must write pending_command only. Read
@internal/farmguardian/tools/ and @cmd/api/smoke_pi_contract_test.go.
Vision (WS6) is optional — ship WS5 photos first if multimodal is not ready.
```

---

## Related

| Doc | Role |
|-----|------|
| [`phase_29_guardian_agent_layer.md`](phase_29_guardian_agent_layer.md) | Propose→confirm foundation |
| [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md) | Pi executes confirmed actuator PRs |
| [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) | Request flow |
| [`pi-integration-guide.md`](../pi-integration-guide.md) | pending_command contract |
