---
name: Phase 214 — Device IP history surface (API + UI)
overview: >
  Phase 213-adjacent follow-up. The device heartbeat now auto-tracks IP changes
  into gr33ncore.device_ip_events (append-only, hypertable), but there is no
  read path yet — operators can only see current/former device IPs via psql.
  This phase adds a read-only history endpoint and a small UI surface so a
  "Pi went dark" incident (like the 192.168.6.x → 192.168.1.x drift on the LED
  rig) is diagnosable from the dashboard, not a database script.
todos:
  - id: ws1-ip-history-endpoint
    content: "WS1: GET /devices/{id}/ip-history — reuse ListDeviceIPEventsByDevice sqlc query, same auth as GET /devices/{id} (farmauthz.RequireFarmMember), document in openapi.yaml per Phase 213 rule"
    status: completed
  - id: ws2-ui-ip-badge
    content: "WS2: ActuatorCard.vue — read-only current-IP line + expandable short history toggle (mirrors DeviceApiKeyPanel's show/hide pattern), no edit field by design"
    status: completed
  - id: ws3-runbook-tie-in
    content: "WS3: Note in docs (connectivity-requirements.md or a short runbook) — check device IP history in UI/API before reaching for psql when a device goes unreachable"
    status: completed
  - id: ws4-verify-ci
    content: "WS4: go test ./internal/handler/device/... + make audit-openapi + relevant ui vitest green; push; confirm Actions green (Phase 213 rule: gate updated in same push)"
    status: completed
isProject: false
---

# Phase 214 — Device IP history surface (API + UI)

**Status:** shipped · **Depends on:** Phase 214 IP tracking backend (already shipped: `device_ip_events` table, `recordDeviceIPIfChanged`, sqlc queries)

## Why this exists

The write path shipped already (commit `7c9e20b`): every heartbeat compares the observed request IP against `devices.ip_address` and, on change, logs a row and updates the device. But there's no read path — the only way to see *why* a device went dark (old IP vs. new IP, when it changed) is a manual `psql` query. That's exactly the ceiling this phase closes.

Current live example: the LED rig's Pi is unreachable right now (`ping 192.168.1.246` → Destination Host Unreachable). Once it's back on the network the heartbeat will self-correct `devices.ip_address` automatically — but an operator watching the dashboard has no way to *see* that happened, or that it drifted in the first place, without this phase.

## Design decisions

- **Read-only.** No "edit IP" field in the UI. A human overriding the IP reintroduces the stale-config problem the auto-tracking already solves; there's no legitimate case for manual entry.
- **No new auth surface.** `GET /devices/{id}/ip-history` reuses `farmauthz.RequireFarmMember` — identical gate to the existing `GET /devices/{id}`, which already returns `ip_address` in the device payload today. History is no more sensitive than current state.
- **ponytail:** short history only (e.g. `LIMIT 20` via the existing `ListDeviceIPEventsByDevice` query param) — no pagination UI yet; upgrade path is a `?limit=` query param if a farm's history grows past a screenful.

## Workstreams

### WS1 — API: `GET /devices/{id}/ip-history`
- Handler in `internal/handler/device/handler.go`, wired in `cmd/api/routes.go` next to the other `jwt(...)` device routes.
- Calls `h.q.ListDeviceIPEventsByDevice(ctx, id, limit)` (query already exists from the IP-tracking commit).
- Document the new path in `openapi.yaml` in the **same push** — Phase 213 rule: `make audit-openapi` must stay green.

### WS2 — UI: read-only IP + history toggle
- `ui/src/components/ActuatorCard.vue` already has a "show keys" toggle pattern (`DeviceApiKeyPanel`) — mirror it with a small `DeviceIPHistoryPanel.vue`:
  - Current IP line next to device status (only when `device.ip_address` is set).
  - Toggle button reveals last few `device_ip_events` rows (old → new, timestamp).
- No new store mutation needed — this is a fetch-on-toggle, not part of `farm.js` bulk load.

### WS3 — Runbook tie-in
- One or two lines in `docs/connectivity-requirements.md` (or wherever Pi-unreachable triage already lives): check device IP history in the dashboard/API before reaching for `psql`.

### WS4 — Verify & close
- `go test ./internal/handler/device/...`
- `make audit-openapi`
- Relevant UI vitest (ActuatorCard / device history panel)
- Push; confirm Actions green (both `go` and `ui` jobs) before starting further feature work — per Phase 213's rule.

## Out of scope

- Manual IP edit/override in UI (deliberately excluded — see Design decisions)
- Alerting on IP change (could be a future phase if false-positive rate from DHCP churn is low enough to justify it)
- Multi-device bulk IP history view / export

## Close when

- [x] `GET /devices/{id}/ip-history` live, documented in OpenAPI, audit-openapi green
- [x] ActuatorCard shows current IP + toggle-able short history, read-only
- [x] Runbook note added ([`operator-troubleshooting.md` §3b](../operator-troubleshooting.md#3b-device-went-dark--check-ip-history-before-psql-phase-214))
- [x] `go` + `ui` Actions jobs green (verified after push: Actions workflow for commits `9667eaf` + `17c16fd`)

## Related

- Phase 213: [`phase_213_ci_keep_green.plan.md`](phase_213_ci_keep_green.plan.md) — the "update the gate in the same push" rule this phase follows for openapi.yaml
- IP tracking backend: `db/migrations/20260811_device_ip_history.sql`, `internal/handler/device/handler.go` (`clientIP`, `recordDeviceIPIfChanged`)