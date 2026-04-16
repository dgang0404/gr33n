# Phase 13 — operator documentation index

Phase 13 extends Phase 12 with receiver-side Insert Commons, farm audit APIs, deeper finance and offline behavior, multi-farm organizations (usage metering hooks), and optional Capacitor packaging. This page is the **single map** for operators and integrators; implementation detail lives in linked documents and in [`openapi.yaml`](../openapi.yaml). **Current engineering focus** is Phase 14 — see [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md) and the Phase 14 operator index [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md).

## API contract (machine-readable)

- **[`openapi.yaml`](../openapi.yaml)** — All dashboard and Pi routes, request/response schemas, and auth modes. Phase 13 surfaces include **organizations** (`/organizations`, usage summary), **Insert Commons** (`/farms/{id}/insert-commons/*`), **audit events** (`GET /farms/{id}/audit-events`), and existing **costs / finance** paths (COA mappings, exports, receipt attachments). Phase 14 adds **commons** (`/commons/catalog`, farm catalog imports).

## Runbooks and playbooks

| Topic | Document |
|--------|-----------|
| MQTT / edge bridges, topic conventions, batch ingest, device tasking with `PI_API_KEY` | [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) |
| Farm audit trail API, who can read it, event kinds, retention | [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md) |
| Insert Commons **farm pipeline** (preview, sync, validation, approval, export) | [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) |
| Insert Commons **receiver** contract, env vars, pilot service, DB migration | [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md) |
| Commons catalog (**gr33n_inserts** browse + farm import audit; Phase 14 WS3) | [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md) |
| Receipt blob storage cutover, backups, sensitive-action evidence | [`receipt-storage-cutover-runbook.md`](receipt-storage-cutover-runbook.md) |
| Optional **native wrapper** (same Vue `dist/`; PWA remains primary) | [`mobile-distribution.md`](mobile-distribution.md) |

## Planning and scope

- **[`plans/phase_13_platform_evolution.plan.md`](plans/phase_13_platform_evolution.plan.md)** — Workstreams WS1–WS7 and explicit out-of-scope items.

## Quick environment pointers

- **Farm API (sender):** `INSERT_COMMONS_INGEST_URL`, `INSERT_COMMONS_SHARED_SECRET`, `INSERT_COMMONS_PSEUDONYM_KEY` (see receiver playbook).
- **Pilot receiver:** `cmd/insert-commons-receiver`, `make run-receiver`, migration `db/migrations/20260417_phase13_insert_commons_receiver.sql`.
- **Organizations:** JWT routes under `/organizations`; farm linkage via `PATCH /farms/{id}/organization` (see OpenAPI).

## WS7 checklist (documentation)

- README Phase 13 banner links this index and the phase plan.
- OpenAPI description references operator docs and Phase 13 areas.
- Audit and Insert Commons receiver playbooks are the primary compliance/federation operator guides.
- Mobile distribution is documented without replacing the PWA workflow.
