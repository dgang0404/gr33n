# Insert Commons — farm-side pipeline runbook (Phase 14 WS2)

This document describes the **farm API → optional HTTP receiver** path for coarse, pseudonymous aggregates: validation rules, approval queue, export formats, and how to evolve **`gr33n.insert_commons.v1`** without breaking pilots.

**Related:** receiver deploy and DB — [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md). OpenAPI paths under `/farms/{id}/insert-commons/*`. Canonical validator — `internal/insertcommonsschema` (`ValidatePayload`).

## Design goals

| Goal | Mechanism |
|------|-----------|
| **No surprise fields** | Ingest JSON may only include the **six** top-level keys listed below; unknown keys are rejected. |
| **Stable aggregate shape** | `aggregates.costs` / `tasks` / `devices` have required child keys so receivers and analytics can rely on structure. |
| **Human gate** | Optional `insert_commons_require_approval` on the farm; payloads queue as bundles until owner/manager approves. |
| **Inspect before send** | `GET .../preview` builds and validates the same body as sync would, without persisting or POSTing. |
| **Audit** | Sync, opt-in, bundle approve/reject/export produce farm-scoped audit events (see [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md)). |

## Ingest JSON shape (`gr33n.insert_commons.v1`)

Allowed **top-level** keys only:

1. `schema_version` — exactly `gr33n.insert_commons.v1` (until a new major is introduced).
2. `generated_at` — RFC3339 or RFC3339Nano; not more than ~10 minutes in the future; not older than one year (validator window).
3. `farm_pseudonym` — non-empty string (HMAC-derived on the API from farm id + server secret).
4. `farm_profile` — object with `scale_tier`, `timezone_bucket`, `currency`, `operational_status`.
5. `aggregates` — object with:
   - `costs`: object with **`totals`** (object) and **`by_category`** (JSON **array** of category rows).
   - `tasks`: object with **`by_status`** (object, map of status → count).
   - `devices`: object with **`by_status`** (object).
6. `privacy` — object with **`includes_pii`** (boolean; farm payloads use `false`).

**Scrubbing:** The API builds this object in code from aggregate queries only (no raw notes, names, or GPS). Do not merge arbitrary client JSON into the ingest body.

### Custom senders (integrators without the dashboard)

If you **POST ingest JSON yourself** (custom script, third-party integration, or lab tool) to the same URL the farm API uses, the body must match this document **exactly**:

- **No extra top-level keys** — only the six keys above. A stray field such as `notes`, `source`, or `metadata` causes validation to fail.
- **Complete `aggregates` sub-shapes** — e.g. `costs` must include both `totals` and `by_category`; `tasks` and `devices` must each include `by_status`. Partial objects are rejected.
- **`privacy.includes_pii` must be a JSON boolean** (`true` / `false`), not a string.

The supported operator path for farms using the product is **`GET .../preview`** (inspect) and **`POST .../sync`** (build from DB and send); those calls produce a valid payload. Copy-paste from preview if you need a golden example.

## Schema versioning policy

1. **Patch-level tightening** (stricter validation on the same `gr33n.insert_commons.v1`): ship in API + receiver together; document in changelog.
2. **New major (e.g. v2):** add `gr33n.insert_commons.v2` (new constant), extend `ValidatePayload` or add `ValidatePayloadV2`, and run receiver **dual-accept** during an announced window; then retire v1.
3. **Package export** (`package_v1`): wrapper version `gr33n.insert_commons.package.v1` is archival only; the inner `payload` must still satisfy the active ingest schema.

## API lifecycle (operator)

| Step | Method | Role | Notes |
|------|--------|------|--------|
| Opt in | `PATCH .../insert-commons/opt-in` | Farm admin | Sets `insert_commons_opt_in` and optional `insert_commons_require_approval`. |
| Preview | `GET .../insert-commons/preview` | Farm owner/manager | Read-only validated JSON. |
| Sync | `POST .../insert-commons/sync` | Admin or finance | Idempotency key; rate limit; backoff on delivery failures; may queue bundle if approval is on. |
| History | `GET .../insert-commons/sync-events` | Admin / finance / cost viewers | Attempt log. |
| Bundles | `GET .../insert-commons/bundles` | Same as history | Pending / approved / rejected / delivery states. |
| Approve / reject | `POST .../bundles/{id}/approve` or `/reject` | Farm admin | |
| Retry delivery | `POST .../bundles/{id}/deliver` | Farm admin | After `delivery_failed`. |
| Export | `GET .../bundles/{id}/export?format=ingest` or `package_v1` | Admin or finance | `ingest` = raw inner JSON; `package_v1` = wrapper with metadata for archives. |

Environment variables for outbound POST: `INSERT_COMMONS_INGEST_URL`, optional `INSERT_COMMONS_SHARED_SECRET` (Bearer), `INSERT_COMMONS_PSEUDONYM_KEY` (stable pseudonym).

**Receiver correlation:** when the farm sync uses an idempotency key (`Idempotency-Key` header or `idempotency_key` in the sync body), the outbound POST to `INSERT_COMMONS_INGEST_URL` includes **`Gr33n-Idempotency-Key`** with the same value (max 128 chars). The in-repo pilot receiver stores it as `source_idempotency_key` and treats duplicates per `(farm_pseudonym, key)`; see [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md).

## Testing

- Unit tests: `go test ./internal/insertcommonsschema/...`
- Smoke: `TestInsertCommonsPreview` and any sync tests in `cmd/api/smoke_test.go` (with DB).

## When something fails validation

- **400** from the farm API **preview** or **sync** when building/sending: almost always **strict schema** — extra top-level keys, missing `aggregates.*` pieces, wrong types (e.g. `includes_pii` as a string), or bad `generated_at` / `schema_version`. Compare your JSON to **Ingest JSON shape** above or use **preview** as the reference body.
- **400** from the **receiver**: same shape rules apply if the farm API successfully POSTed; if you POST manually, fix payload shape or clock skew; check receiver logs for the rejection reason.
- **500** from farm preview/sync with “validation failed”: API bug or DB aggregate shape drift — compare `buildInsertCommonsIngestPayloadBytes` with `ValidatePayload`.
