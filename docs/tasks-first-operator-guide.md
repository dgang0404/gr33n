# Tasks-first operator guide

This doc supports **[sit-in §3 — Tasks-first](workstreams/sit-in-operator-experience.md)**: a concrete **golden path**, how **automation** relates to **tasks**, and how **offline queuing** behaves in the dashboard.

**Related:** [operator-tour.md](operator-tour.md), [workflow-guide.md](workflow-guide.md) (workflows), `internal/automation/rules.go`, `ui/src/stores/farm.js`, `ui/src/offline/taskQueue.js`.

---

## 1. Golden path — “Morning ops”

Use this order until you know the farm cold; polish secondary pages afterward.

| Step | Where | Purpose |
|------|--------|---------|
| 1 | **`/`** Dashboard | Snapshot: today’s tasks shortcut, alerts glance if surfaced, jump-off links. |
| 2 | **`/tasks`** | **Primary spine:** what needs human attention today — board columns by status, zones, due dates. |
| 3 | **`/alerts`** | Anything the system surfaced (rules, stock, etc.); acknowledge or spawn follow-up. |
| 4 | **`/schedules`** | Confirm **cron schedules** are active and recently triggered if today’s run matters (irrigation, lights). |

Optional same session: **`/automation`** (rules + **run history** loaded with the page), **`/setpoints`** vs live readings if diagnosing climate.

Automation **runs** appear in context on **Schedules** and **Automation** views (`loadAutomationRuns`); full API: `GET /farms/{id}/automation/runs`.

---

## 2. Tasks ↔ automation — what creates what?

### Automation **rules** (condition → action)

Implemented action types in **`dispatchRuleAction`** (`internal/automation/rules.go`):

| Action type | Effect | Task list? |
|-------------|--------|------------|
| **`control_actuator`** | Inserts **actuator event**, updates simulated state, optionally **`pending_command`** on device for Pi pickup | No — hardware path |
| **`create_task`** | Inserts **`gr33ncore.tasks`** row with **`source_rule_id`** set | **Yes** — shows on Tasks with rule attribution in UI |
| **`send_notification`** | **`alerts_notifications`** row (+ optional push) | No — **Alerts** page |

**`create_task`** parameters (JSON body on the executable action) support `title`, `description`, `zone_id`, `task_type`, `priority`, `due_in_days`, `estimated_duration_minutes` — see code comments in `dispatchRuleCreateTask`.

### Automation **schedules** (cron → actions)

Schedule **`executable_actions`** use **`executeAction`** in `internal/automation/worker.go`:

| Action type | Effect |
|-------------|--------|
| **`control_actuator`** | Same idea as rules — actuator events + Pi **`pending_command`** when not in simulation |
| **`update_record_in_gr33n`** | Writes fertigation-related rows (bounded recipe in worker) |

Schedules do **not** use the rule **`create_task`** path in the worker — if you need **tasks** from time alone, model it with a **rule** whose conditions match “always” / time predicates, or create tasks manually / from **Alerts**.

### Alerts → tasks

The API supports **`POST /alerts/{id}/create-task`** (dashboard flow): converts an alert into a task — distinct from rule **`create_task`** actions.

---

## 3. Offline queue and sync (browser)

The Pinia store keeps a **persistent write queue** so task and cost creates survive refresh or flaky networks.

| Topic | Behavior |
|-------|-----------|
| **Storage** | **`localStorage`** key **`gr33n_offline_write_queue_v2`** (replaces legacy `gr33n_task_write_queue_v1` once). |
| **Queued operations** | **`create_task`**, **`update_task_status`**, **`create_cost`** (costs share the same queue machinery). |
| **When queue is used** | If **`createTask`** fails with a **retryable** error (see below), or **`navigator.onLine`** is false so the code paths queue immediately. |
| **Retryable errors** | No HTTP response (network) **or** status **≥ 500** **or** **429** → item stays **`pending`**. Other 4xx → **`failed`** with **`lastError`** (validation, auth, conflict). |
| **Flush** | **`flushTaskWriteQueue`** posts queued items in order; maps **`clientTaskId`** → server id for follow-up status patches; drops **`synced`** rows from the array and persists. |
| **UI** | **Tasks** / **Costs**: “Offline mode”, “N queued writes”, **Sync now**, queue details modal, conflict banner when **`_offline.stale`**. |
| **Reconnect** | **`window` `online`** event triggers **`syncNow()`** on Tasks (and Costs) to flush when the browser thinks connectivity returned. |

**Trust boundary:** the queue is **per browser profile**. It is not a server-side outbox; clearing site data clears the queue.

---

## 4. Product-copy gaps (for future tickets)

| Area | Gap |
|------|-----|
| Tasks HelpTip | Should mention **automation rules** that **`create_task`**, not only schedules (aligned in-app). |
| Automation empty states | When no rules/schedules, short hint linking **Schedules** vs **Rules** docs. |
| Dashboard | Optional dedicated “Morning ops” strip — product choice; this doc defines the path without requiring UI work now. |

---

*Introduced for sit-in §3. Refine when Automation UI exposes run history more prominently.*
