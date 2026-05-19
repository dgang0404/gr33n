# Sit-in workstream: operator experience, observability, tasks-first

**Sit-in** means this backlog **stays named as-is** even if calendar phases (e.g. Phase 25 RAG) advance. New work that does not belong here should show up as **scope creep** if it lands in this file — keep product phases and this stream separate.

**Goal:** Make the live product **understandable** (docs + UI cues), **debuggable** (logging), and **usable day-to-day** with **tasks** as the spine — before leaning harder into RAG or net-new features.

**Where effort goes:** When **no sit-in items** are actively in flight (nothing queued here beyond standing maintenance), treat the team as **back on the current calendar phase** (e.g. Phase 25 RAG operations)—this stream does **not** block phase work. **Reopen** sit-in any time operator pain shows up (broken flows, unclear dashboards, logging gaps, onboarding gaps): add bullets under the right section or link a bugfix plan, same as the Fertigation tab sync fix.

**Phase 26:** fuller **operator tutorial + glossary** (WS1 v1 **Guide**), **observability** (WS2 runbook + optional **Loki** overlay), and explicit **RAG vs education vs logs** (WS3 v1 **`rag-scope-and-threat-model.md` §9**) — see **[Phase 26 plan](../plans/phase_26_operator_tutorial_observability_rag.plan.md)**.

---

## 1. Documentation / onboarding

| Item | Notes |
|------|--------|
| **Single-page operator tour** | **Done (v1):** [`docs/operator-tour.md`](../operator-tour.md) — narrative walk: Farm → Zones → Sensors/Controls → Schedules/Rules → Tasks → Fertigation; **mermaid data-flow** diagram; links to bootstrap + schema. Revise as nav/copy changes. **In-app (Phase 26 WS1 v1):** **System → Guide** — glossary + suggested route order (`/operator-guide`). |
| **“Why empty?” UX** | Per major UI area, future inline hints (telemetry vs setpoints vs automation inactive). Track implementation as **separate UX tickets**; tour §4 points here — **implementation still open**. |

**Artifact:** [`docs/operator-tour.md`](../operator-tour.md).

---

## 2. Logging / observability (“logging phase” — can align with Phase 26 docs)

| Item | Notes |
|------|--------|
| **API structured logs** | **Done:** `log/slog` per request — `request_id` (also **`X-Request-ID`**), `method`, `path`, `status`, `duration_ms`, `auth` (`jwt` / `api_key` / `public` / `jwt_or_pi`), `farm_id` (from `/farms/{id}/` paths), `user_id` (JWT). **`LOG_FORMAT=json`** for JSON lines. See `cmd/api/request_log.go` + `routes.go` wiring. |
| **Auth debug** | **Done:** **`AUTH_DEBUG_LOG=true`** — `auth_rejected` with **`reason`** (`missing_x_api_key`, `jwt_invalid`, …); **never** logs token or API key value. `cmd/api/auth.go`. |
| **Automation worker** | **Done:** `slog.Warn` on tick **list** failures (`phase` = `list_schedules` / `list_rules` / `list_programs`); **`automation schedule run`** / **`automation rule run`** on outcomes (`schedule_id` / `rule_id`, `status`); **Warn** when schedule `status=failed` or rule `status=failed`. |
| **Runbook doc** | **Done:** [`docs/operator-troubleshooting.md`](../operator-troubleshooting.md) — login / 401 / empty farms / reading logs; linked from [local-operator-bootstrap.md](../local-operator-bootstrap.md). |
| **Log aggregation / archival** | **Done (v1):** [`docs/operator-logging-runbook.md`](../operator-logging-runbook.md) — slog baseline, **`LOG_FORMAT`**, Docker **json-file** rotation (Compose), optional **`docker-compose.logging.yml`** (**Loki + Promtail + Grafana** overlay), journald, archival exports; **DB retention ≠ log retention**. Phase 26 WS2. |

**Related:** [INSTALL.md](../../INSTALL.md) § Optional: observability (`LOG_FORMAT`, `AUTH_DEBUG_LOG`). Production capture and retention: **[operator-logging-runbook.md](../operator-logging-runbook.md)**.

---

## 3. Tasks-first / “tasks domination”

| Item | Notes |
|------|--------|
| **Primary journey** | **Done (v1):** [`docs/tasks-first-operator-guide.md`](../tasks-first-operator-guide.md) §1 — **Morning ops** path (Dashboard → Tasks → Alerts → Schedules). |
| **Tasks ↔ automation** | **Done (v1):** same doc §2 — rule actions **`create_task`** / **`control_actuator`** / **`send_notification`**; schedules vs rules; copy-gap stub §4. Tasks HelpTip updated for rules + offline. |
| **Offline / queue** | **Done (v1):** same doc §3 — `localStorage` key **`gr33n_offline_write_queue_v2`**, queue item types, retryable vs failed (`isRetryableTaskQueueError`), flush + **`online`** event; points to `farm.js` / `offline/taskQueue.js`. |

**Artifact:** [`docs/tasks-first-operator-guide.md`](../tasks-first-operator-guide.md).

---

## 4. Multi-device hardening

| Item | Notes |
|------|--------|
| **Machine checklist** | **Done (v1):** [`docs/machine-setup-checklist.md`](../machine-setup-checklist.md) — extended with **second machine / browser profile** (CORS, Vite port, offline queue per device). Re-run on every new laptop or VM. |
| **Troubleshooting link** | [operator-troubleshooting.md](../operator-troubleshooting.md) §3 — `localStorage` queue boundary across devices. |

**Bugfix (Fertigation tabs):** [`docs/plans/bugfix_fertigation_tab_router_sync.plan.md`](../plans/bugfix_fertigation_tab_router_sync.plan.md) — closed; router **`?tab=`** sync, loading/retry UX, and **`trigger_source`** display (nullable enum JSON from the API) in `Fertigation.vue`. **Tests:** no updates required for **`cmd/api/smoke_pi_contract_test.go`**, **`smoke_fertigation_test.go`**, or existing Vitest files — behaviour was front-end only. **Optional later:** a short Vitest case if **`formatTriggerSource`** is moved to a small **`ui/src/utils`** helper; optional smoke assertion on **`GET /farms/…/fertigation/events`** JSON shape if we want to pin nullable enum serialization.

---

## 5. Relationship to Phase 25 (RAG)

Phase 25 plans should **assume** this sit-in stream has at least **operator tour + troubleshooting doc + minimal API/worker logging** underway; avoid stacking RAG UX on top of an opaque dashboard.

---

## 6. Phase 27 — Farm Guardian (calendar phase)

**Phase 26** remains the operator-tutorial / logging / RAG-education boundary track — **done enough for v1**.

**Phase 27** (Farm Guardian AI layer) continues from **[phase_27_farm_guardian_ai_layer.md](../plans/phase_27_farm_guardian_ai_layer.md)** — e.g. **`AI_ENABLED`**, **`GET /capabilities`**, **`POST /v1/chat`** (stub → full RAG-backed chat). This section links the calendar phase without moving Phase 27 tasks into the sit-in backlog unless operator UX explicitly needs it.

---

## Changelog

| Date | Note |
|------|------|
| 2026-04-21 | Stream created from operator bootstrap learnings (Compose, auth_test, seed, env-admin JWT binding). |
| 2026-04-21 | Phase 26 hook: tutorial + glossary vs RAG; links to intranet doc from bootstrap + rag-scope. |
| 2026-04-21 | §1: Added [`operator-tour.md`](../operator-tour.md) (narrative + mermaid); “why empty” remains UX tickets. |
| 2026-04-21 | §2: Structured HTTP logs (`request_log.go`), `AUTH_DEBUG_LOG`, automation `slog` outcomes, [`operator-troubleshooting.md`](../operator-troubleshooting.md). |
| 2026-04-21 | Linked **[Phase 26 plan](../plans/phase_26_operator_tutorial_observability_rag.plan.md)** (tutorial, log management/archival vs DB retention, RAG boundary). |
| 2026-04-21 | §3: [`tasks-first-operator-guide.md`](../tasks-first-operator-guide.md) (golden path, automation×tasks, offline queue); Tasks.vue HelpTip. |
| 2026-04-21 | §4: Checklist + multi-device notes; Fertigation **tab↔URL** fix + [bugfix plan](../plans/bugfix_fertigation_tab_router_sync.plan.md). |
| 2026-04-21 | §4: Bugfix doc marked closed; noted **no mandatory Pi/API/UI test updates** (UI-only fix); optional Vitest/smoke follow-ups. |
| 2026-05-13 | §1: Added in-app **Guide** (`/operator-guide`) — glossary + walk (Phase 26 WS1 v1); complements operator-tour.md. |
| 2026-05-13 | §2: **[operator-logging-runbook.md](../operator-logging-runbook.md)** — Compose json-file rotation + runbook (Phase 26 WS2 v1). |
| 2026-05-13 | §2: **`docker-compose.logging.yml`** — Promtail + Loki + Grafana overlay + **`make compose-logging-*`** (Phase 26 WS2 follow-up). |
| 2026-05-13 | Phase **26 WS3** v1: **[rag-scope-and-threat-model.md §9](../rag-scope-and-threat-model.md)** — static Guide vs DB RAG vs ops logs; workflow §10.6 + Knowledge HelpTip. |
| 2026-05-18 | Phase **27 WS1 + WS3 stream + WS5 v2/v3 + WS6 chat panel**: **[farm-guardian-ollama-setup.md](../farm-guardian-ollama-setup.md)** runbook; SSE `ChatCompletionStream`; `/v1/chat` accepts `farm_id` + `stream` + `session_id` with citations; new **/chat** Guardian UI page (streaming + Lite banner). |
| 2026-05-18 | Phase **27 WS5 follow-up — multi-turn history**: `conversation_turns` migration + `InsertConversationTurn` / `ListConversationTurnsBySession` / `ListRecentConversationSessions`; `/v1/chat` persists every turn and replays up to 20 prior turns; new `GET /v1/chat/sessions[/{id}]` endpoints; `/chat` panel grew a sessions sidebar and multi-turn transcript. |
| 2026-05-18 | Phase **27 WS4 follow-up — live farm snapshot**: `internal/farmguardian/snapshot.go` (zones + active cycles + unread alerts) injected into the system message on grounded `/v1/chat` turns (capped 12 zones / 8 cycles), so Guardian answers can speak to the farm's current state without making it up. |
| 2026-05-19 | Phase **27 WS5/WS6 follow-up — session lifecycle + token usage**: new `conversation_sessions` metadata table, `PATCH` / `DELETE /v1/chat/sessions/{id}`, `prompt_tokens` / `completion_tokens` columns on `conversation_turns`, `UsageAwareChatCompleter` in the LLM client; `/chat` sidebar grew rename/delete controls and per-session + per-turn token chips. |
| 2026-05-19 | Phase **27 WS3 follow-up — retry / backoff**: `internal/rag/llm/retry.go` adds `IsTransientLLMError` + exponential-backoff retry loop (env knobs `LLM_RETRY_MAX_ATTEMPTS` / `LLM_RETRY_BACKOFF_MS`); non-streaming chat retries the full request, streaming retries only the connect phase (never duplicates already-forwarded deltas). Caller cancellation is never retried. Documented in INSTALL.md + `.env.example`. |
| 2026-05-19 | Phase **27 WS6 follow-up — inline rename modal**: `/chat` sidebar rename now opens an accessible in-page dialog (autofocus, Esc/click-outside to close, Enter to save, max 120 chars, empty clears title) instead of `window.prompt`. API errors render inside the modal so the operator can correct the title without losing context. Covered by `chat-rename-modal.test.js`. |
| 2026-05-19 | Phase **27 WS6 follow-up — bulk delete (closes WS6)**: `Select` button on the sessions sidebar enters a multi-select mode with per-row checkboxes + a toolbar (count, Select all, Cancel, Delete N). Confirm modal fires `Promise.allSettled` DELETEs; succeeded rows leave the sidebar and the transcript clears if the active session was among them; failed rows stay selected with an inline `Failed to delete N of M` so the operator can retry. Covered by `chat-bulk-delete.test.js`. |
| 2026-05-19 | Phase **27 Ollama runbook update**: `docs/farm-guardian-ollama-setup.md` §1.1/§1.2 — adds the on-farm-intranet data-flow diagram (Pi + UI → API → Postgres + Ollama, no external hops in Full mode) and the three-layer knowledge model (Llama weights + per-farm RAG corpus + live snapshot). Calls out the future-extension path for a static agricultural reference corpus alongside the boundary defined in `rag-scope-and-threat-model.md` §9. |
| 2026-05-19 | Phase **27 WS5 follow-up — cost guards (closes Phase 27 backend)**: rolling-window token caps on `POST /v1/chat`. New `CHAT_COST_WINDOW_HOURS` (default 1, clamp 1..168) + `CHAT_COST_MAX_TOKENS_PER_USER` + `CHAT_COST_MAX_TOKENS_PER_FARM` (both default 0 = disabled, clamp 0..100M). Over-budget requests return **429** with `Retry-After` and a JSON body `{reason, used_tokens, max_tokens, window_seconds, retry_after_seconds}`; per-user dimension takes precedence over per-farm. Guard short-circuits before any LLM work so rejected turns cost zero tokens; fails open (with WARN log) on DB hiccups so a transient outage never takes chat offline. Unit tests + real-DB smoke (`smoke_cost_guard_test.go`) cover SUM rollup across sessions, per-farm dimension, and rolling-window cutoff. Documented in INSTALL.md + `.env.example`. |
| 2026-05-19 | Phase **27 WS5 follow-up — TTL pruning**: `internal/farmguardian/prune.go` + `cmd/api/main.go` spawn a background loop that drops `conversation_turns` + `conversation_sessions` whose latest activity is older than `CHAT_SESSION_TTL_DAYS` (default 30, 0 disables). Loop only runs when `AI_ENABLED=true`; cadence + startup delay configurable. Unit tests + real-DB smoke test (`smoke_prune_test.go`) prove fresh sessions survive a 30-day pass. Documented in INSTALL.md + `.env.example`. |
| 2026-05-19 | Phase **27 WS5 follow-up — streaming token usage**: LLM client now sets `stream_options.include_usage: true` and `ChatCompletionStreamMessagesWithUsage` returns the OpenAI-style token block parsed from the terminal SSE chunk. Chat handler prefers the new `UsageAwareStreamingChatCompleter` and falls back to the legacy interface, so `prompt_tokens` / `completion_tokens` now flow into the streaming SSE `done` event **and** the persisted `conversation_turns` row — closing the gap where non-streaming turns recorded usage but streaming turns did not. Servers that ignore `include_usage` still work (row lands with zero tokens). |
| 2026-05-15 | §6: **[Phase 27](../plans/phase_27_farm_guardian_ai_layer.md)** pointer — Farm Guardian / `AI_ENABLED` / capabilities API (calendar phase; not sit-in backlog unless UX asks). |
| 2026-05-18 | §6: **Phase 27 WS4/WS5/WS6 v1** — `internal/farmguardian` persona, `POST /v1/chat` non-streaming with 503 Lite, `/capabilities` Pinia store, Settings Lite/Full label, FarmKnowledge Ask-LLM gating. |
