# Operator troubleshooting — auth, empty farms, logs

Quick checklist when the dashboard misbehaves during local or LAN runs. Deeper setup lives in [local-operator-bootstrap.md](local-operator-bootstrap.md); architecture context in [operator-tour.md](operator-tour.md).

---

## 1. Login loops or **401** on farm routes

| Symptom | What to check |
|--------|----------------|
| Never get past login | `DATABASE_URL`, API reachable (`curl http://localhost:8080/health`), credentials vs `ADMIN_USERNAME` / `~/.gr33n/admin.hash` / `ADMIN_PASSWORD_HASH`. |
| Login succeeds but **every** `/farms/...` returns 401 | JWT is missing **`user_id`** for env-admin, or `ADMIN_BIND_USER_ID` does not match a real `gr33ncore.users` row. Set **`ADMIN_BIND_USER_ID`** and **`ADMIN_BIND_EMAIL`** in `.env` to a user that exists and is a **member of the farm** (see [master seed](../db/seeds/master_seed.sql) demo defaults). |
| `invalid or expired token` | `JWT_SECRET` changed between login and request, or token actually expired; re-login. |
| Pi or gateway rejected | **`PI_API_KEY`** must match on client and server when `AUTH_MODE` enforces auth. |

### **`AUTH_DEBUG_LOG`**

Set **`AUTH_DEBUG_LOG=true`** in `.env` (API restart required). The API will emit **`auth_rejected`** log lines with a **`reason`** field (`missing_x_api_key`, `invalid_x_api_key`, `missing_bearer_or_query_token`, `jwt_invalid`, …). **Tokens and API keys are never printed.**

---

## 2. Farm lists or dashboards feel **empty**

| Symptom | What to check |
|--------|----------------|
| No farms at all | Seed data (`./scripts/bootstrap-local.sh --seed`) or create a farm in-app; confirm DB connection. |
| Farm exists but no sensors / readings | Pi URL and **`PI_API_KEY`**; readings POST path; time skew on `reading_time`. |
| Automation never fires | Schedules **active**, cron expression, worker running (same API process — check startup logs); **`AUTOMATION_SIMULATION_MODE`** if hardware not wired. |
| Fertigation tab (e.g. **Events**) looks stuck or out of sync | Tab state is synced to **`?tab=`** in the URL (`selectTab` + router). Use `/fertigation?tab=events` deep links; hard-refresh should match the tab highlight. See [bugfix plan](plans/archive/bugfix_fertigation_tab_router_sync.plan.md). |

---

## 3. Multi-device browser profiles

The offline **task/cost write queue** lives in **`localStorage`** per browser profile — it does **not** follow you across machines automatically. Second laptop or tablet: run through [machine-setup-checklist.md](machine-setup-checklist.md) again; expect a separate queue until each device syncs.

---

## 4. Reading API logs

Every HTTP request emits one structured **`request`** line (`log/slog`) after the response completes:

- **`request_id`** — matches **`X-Request-ID`** on the response (client may send its own header).
- **`method`**, **`path`**, **`status`**, **`duration_ms`**
- **`auth`** — `public` \| `jwt` \| `api_key` \| `jwt_or_pi`
- **`farm_id`** — parsed from `/farms/{id}/...` when present
- **`user_id`** — when logged in with JWT (dashboard routes)

Set **`LOG_FORMAT=json`** for JSON lines (log aggregation–friendly).

**Deeper:** retention vs Postgres time-series pruning, Docker/systemd capture, Loki sketch, archival — **[operator-logging-runbook.md](operator-logging-runbook.md)**.

Automation worker:

- **`automation worker tick failed`** — `phase` is `list_schedules`, `list_rules`, or `list_programs`; includes **`err`**.
- **`automation schedule run`** / **`automation rule run`** — outcome after a schedule or rule execution (`schedule_id` / `rule_id`, `farm_id`, `status`). **`Warn`** is used for **`failed`** outcomes on schedules; rules use **`Warn`** only when **`status`** is **`failed`**.

---

## Related code

- JWT / API key middleware: `cmd/api/auth.go`
- Access logger: `cmd/api/request_log.go`
- Route wiring (logging wraps each handler): `cmd/api/routes.go`
- Fertigation tab ↔ URL sync: `ui/src/views/Fertigation.vue` (`selectTab`, `watch route.query.tab`)

## Operational logging (production)

For Docker/systemd log rotation, aggregation (e.g. Loki), separating **app logs** from **DB retention**, and archival exports, see **[operator-logging-runbook.md](operator-logging-runbook.md)**.

---

## 5. Edge actuator safety (Phase 31 WS3)

gr33n **v1 does not cut GPIO when comms drop** — the Pi client stops receiving new commands, but relays stay in their last state. **Operator wiring** must fail-safe (de-energize pumps/solenoids on power loss where flood or overheat risk exists).

| Risk | Mitigation (operator responsibility) |
|------|--------------------------------------|
| **Flood / over-watering** | Fail-safe **closed** solenoids; manual E-stop on irrigation circuits; test **`off`** command before leaving bench |
| **Heat / fire** | Do not drive mains loads from breadboard; listed enclosures, fuses, thermal limits on lighting |
| **Runaway on comms loss** | Assume last GPIO state persists; use hardware timers or contactors that drop out without hold power |
| **Unauthorized Confirm** | Restrict **Operate** role; high-tier Guardian actuator PRs show warning banner — only trusted operators Confirm |
| **Wrong device_id in config** | Run [`print-demo-actuator-ids.sh`](../scripts/print-demo-actuator-ids.sh); **`device_id`** in `pi_client` must match the DB row that receives **`pending_command`** |

**Bench checklist:** one LED or 5 V fan first → [`pi-integration-guide.md` §8–§9](pi-integration-guide.md#8-field-checklist--first-pi-on-a-real-bench-phase-31-ws2) → `./scripts/run-edge-actuator-smoke.sh --direct` before mains loads.

**E-stop story:** keep a physical power cut for the load under test; document who may Confirm actuator PRs on production farms. Software audit: confirmed PR → `pending_command` → `actuator_events` with `meta_data.proposal_id` when sourced from Guardian.
