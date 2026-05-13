# Sit-in workstream: operator experience, observability, tasks-first

**Sit-in** means this backlog **stays named as-is** even if calendar phases (e.g. Phase 25 RAG) advance. New work that does not belong here should show up as **scope creep** if it lands in this file — keep product phases and this stream separate.

**Goal:** Make the live product **understandable** (docs + UI cues), **debuggable** (logging), and **usable day-to-day** with **tasks** as the spine — before leaning harder into RAG or net-new features.

**Where effort goes:** When **no sit-in items** are actively in flight (nothing queued here beyond standing maintenance), treat the team as **back on the current calendar phase** (e.g. Phase 25 RAG operations)—this stream does **not** block phase work. **Reopen** sit-in any time operator pain shows up (broken flows, unclear dashboards, logging gaps, onboarding gaps): add bullets under the right section or link a bugfix plan, same as the Fertigation tab sync fix.

**Phase 26 (when scoped):** fuller **operator tutorial + glossary** + observability evolution (log aggregation/archival, boundary with DB retention) and an explicit **RAG education layer** (static help vs farm-grounded answers)—see **[Phase 26 plan](../plans/phase_26_operator_tutorial_observability_rag.plan.md)**.

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

**Related:** [INSTALL.md](../../INSTALL.md) § Optional: observability (`LOG_FORMAT`, `AUTH_DEBUG_LOG`).

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
