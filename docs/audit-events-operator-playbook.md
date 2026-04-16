# Farm audit events — operator playbook

This playbook covers the audit trail in `gr33ncore.user_activity_log`: **farm-scoped** listing (`GET /farms/{id}/audit-events`) and **organization-scoped** listing (`GET /organizations/{id}/audit-events`). It complements the qualitative checklist in [`receipt-storage-cutover-runbook.md`](receipt-storage-cutover-runbook.md) under **Audit and Observability**.

## Purpose

- Give owners and managers a **defensible, queryable history** of sensitive actions (governance, finance exports, receipt access, federation sync, membership, destructive changes).
- Support incident review without granting broad database access to every dashboard user.

## Who can read audit events

- **Farm trail:** users with **farm administration** capability (**owner** and **manager** roles; same gate as `GET /farms/{id}/members`).
- **Organization trail:** **organization owner or org admin** only (`GET /organizations/{id}/audit-events`).
- Other roles receive `403 Forbidden`.

The dashboard **Settings** page lists recent events for **Organization audit** (org owner/admin) and **Farm audit** (farm owner/manager on the selected farm), using these same endpoints.

## Scope and limitations

- **`GET /farms/{id}/audit-events`** returns rows where **`farm_id` equals that farm** (farm-scoped governance, finance, federation, etc.).
- **Organization-only events** (create/update org, add org member) use **`farm_id` NULL** (not linked to a farm row). They **do not** appear on farm audit; use **org audit** or SQL.
- **`GET /organizations/{id}/audit-events`** returns: (1) all farm-scoped events for farms currently linked to that org, and (2) org-only rows for that org (`details.organization_id`, or legacy rows with `farm_id` 0 / NULL matching org targets).
- Older rows may still have **`farm_id` = 0** from before NULL was used; org audit and direct SQL both match those where applicable.

## API

### Farm audit

- **Path:** `GET /farms/{id}/audit-events`
- **Auth:** `Authorization: Bearer <JWT>` (or dev auth mode per deployment).
- **Query:** `limit` (default `50`, max `200`), `offset` (default `0`, newest-first).

### Organization audit

- **Path:** `GET /organizations/{id}/audit-events`
- **Auth:** JWT; **org owner or org admin** only.
- **Query:** same `limit` / `offset` as farm audit.

Canonical contract and response shape: [`openapi.yaml`](../openapi.yaml) (schema `AuditActivityEvent`).

### Example (curl)

```bash
# Replace HOST, FARM_ID, ORG_ID, and JWT with real values.
curl -sS -H "Authorization: Bearer $JWT" \
  "https://HOST/farms/FARM_ID/audit-events?limit=50&offset=0"

curl -sS -H "Authorization: Bearer $JWT" \
  "https://HOST/organizations/ORG_ID/audit-events?limit=50&offset=0"
```

## Response fields (summary)

Each row includes at least:

| Field | Meaning |
|--------|---------|
| `id` | Monotonic id within the hypertable chunk |
| `activity_time` | When the action occurred (server clock) |
| `action_type` | One of `gr33ncore.user_action_type_enum` values (for example `change_setting`, `export_data`, `delete_record`, `execute_action`, `create_record`, `update_record`) |
| `details` | JSON object; always includes a stable `kind` string for machine filtering where implemented |
| `user_id` | Actor when the request was authenticated with a user JWT |
| `target_module_schema` / `target_table_name` / `target_record_id` | Optional pointer to the primary record touched |

`details` is the right place to look for **fine-grained semantics** (for example `kind: cost_receipt_access` with `endpoint: content` or `download`).

## Event kinds (current implementation)

These `details.kind` values are written by the API today (list may grow in later releases):

| `details.kind` | Typical `action_type` | Notes |
|----------------|----------------------|--------|
| `farm_member_added` | `create_record` | Member invite or add |
| `farm_member_role_changed` | `update_record` | RBAC change |
| `farm_member_removed` | `delete_record` | Membership removed |
| `farm_soft_deleted` | `delete_record` | Farm soft-delete |
| `insert_commons_opt_in` | `change_setting` | Opt-in toggled |
| `insert_commons_sync` | `execute_action` | Manual sync (includes idempotent replay with `duplicate: true` when applicable) |
| `cost_export` | `export_data` | CSV or GL CSV export |
| `finance_coa_mappings_upsert` | `change_setting` | COA mapping batch save |
| `finance_coa_mapping_reset` | `change_setting` | Single category reset |
| `finance_coa_mappings_reset_all` | `change_setting` | All overrides cleared |
| `cost_transaction_deleted` | `delete_record` | Cost row removed |
| `cost_receipt_uploaded` | `create_record` | Receipt file attached |
| `cost_receipt_access` | `export_data` | Receipt bytes or download URL issued |
| `organization_created` | `change_setting` | New org record; `farm_id` **NULL**; `details.organization_id` set |
| `organization_updated` | `change_setting` | Org name / plan / billing fields; `details.organization_id` |
| `organization_member_added` | `change_setting` | User added to org; `details.organization_id`; `details.role` is org role |
| `farm_organization_changed` | `update_record` | Farm linked or unlinked from an org (`PATCH /farms/{id}/organization`); `details.organization_id` / `previous_organization_id` |

Operations that are **not** yet mirrored into this log (for example JWT secret rotation, Pi API key rotation, storage env changes) should continue to use **external** operator evidence as described in the receipt storage runbook.

## Storage and retention

- Physical table: `gr33ncore.user_activity_log` (hypertable on `activity_time` when TimescaleDB is enabled in your deployment).
- **Retention is not enforced by the application.** Operators should align database retention, compression, and archival with organizational policy and the sensitive-action checklist in the receipt storage runbook.
- For long-term compliance, plan **periodic export** of audit rows (SQL copy, logical backup slice, or downstream SIEM) before aggressive chunk drop policies.

## Operational tips

- After security incidents, pull audit rows for the affected farm over the incident window and correlate with application logs and object-store access logs where available.
- When onboarding finance staff, confirm they understand that **opening or downloading a receipt** generates an audit row (`cost_receipt_access`).

## Related documents

- [`docs/phase-13-operator-documentation.md`](phase-13-operator-documentation.md) — Phase 13 operator doc index
- [`docs/plans/phase_13_platform_evolution.plan.md`](plans/phase_13_platform_evolution.plan.md) — Phase 13 scope and workstreams
- [`docs/receipt-storage-cutover-runbook.md`](receipt-storage-cutover-runbook.md) — Sensitive-action checklist and evidence sources
- [`docs/insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md) — Receiver contract for the farm-side sender
