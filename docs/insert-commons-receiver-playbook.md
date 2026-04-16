# Insert Commons receiver — operator playbook

The gr33n **farm API** can POST coarse aggregate payloads to an external URL when Insert Commons sharing is enabled and environment variables are set. This document is for **operators and integrators** implementing or hosting that **receiver** (ingest endpoint). The farm-side sender behavior lives in `internal/handler/farm/insert_commons.go`.

## Farm-side configuration (sender)

Operators configure the **sending** API with:

| Variable | Role |
|----------|------|
| `INSERT_COMMONS_INGEST_URL` | HTTPS (or HTTP in lab) URL of the receiver; if empty, sync completes as `skipped_no_receiver` and no outbound POST is made |
| `INSERT_COMMONS_SHARED_SECRET` | Optional shared secret; when set, the farm API sends `Authorization: Bearer <secret>` on the POST |
| `INSERT_COMMONS_PSEUDONYM_KEY` | Optional dedicated key for HMAC farm pseudonym; if unset, the implementation may fall back to `INSERT_COMMONS_SHARED_SECRET` (see code comments in the sender) |

Farm users still must **opt in** per farm (`PATCH /farms/{id}/insert-commons/opt-in`); manual push uses `POST /farms/{id}/insert-commons/sync`.

## HTTP contract (receiver must implement)

### Request

| Item | Value |
|------|--------|
| Method | `POST` |
| URL | Exactly the value of `INSERT_COMMONS_INGEST_URL` |
| Header `Content-Type` | `application/json` |
| Header `Authorization` | Optional `Bearer <INSERT_COMMONS_SHARED_SECRET>` when the farm is configured with a secret |
| Body | JSON object (UTF-8), see **Payload** below |
| Body size | Farm API limits read snippets on error; keep responses concise. Sender uses a bounded client read size for error bodies (order of 1 MiB). |

### Success and idempotency

- Respond with **2xx** when the payload is accepted and stored (or is a **duplicate** of an already accepted payload for the same logical sync).
- **Farm-side idempotency:** the dashboard sync call may send `Idempotency-Key` to the **farm API**; the farm stores outcomes in `gr33ncore.insert_commons_sync_events`. That key is **not** forwarded on the outbound POST to the receiver today, so the receiver should dedupe using a **stable fingerprint of the JSON body** (for example a hash of the canonical serialized payload, or of `farm_pseudonym` plus `generated_at` if that matches your risk model) or treat duplicate deliveries as idempotent when the body bytes match. The **in-repo pilot receiver** (below) dedupes on **SHA-256 of the raw request body**.
- **429** or **5xx** from the receiver triggers **retryable** handling on the farm (backoff and consecutive failure tracking). **4xx** (except 429) is treated as a **client** failure and is not retried the same way.

### Response body

No strict schema is required on success. For failures, a short plain or JSON error body helps operators; the farm API may surface a truncated excerpt in sync history metadata for support.

## Payload (`gr33n.insert_commons.v1`)

Top-level keys (sender today):

| Field | Type | Description |
|-------|------|-------------|
| `schema_version` | string | Must be `gr33n.insert_commons.v1` for this revision |
| `generated_at` | string | RFC3339Nano UTC timestamp when the payload was built |
| `farm_pseudonym` | string | Opaque token derived from farm id + operator key; **not** reversible to farm id without the same key material |
| `farm_profile` | object | Coarse descriptors only |
| `aggregates` | object | Numeric rollups (see below) |
| `privacy` | object | Declarations and human-readable revocation hint |

### `farm_profile`

| Field | Example / enum | Notes |
|-------|----------------|--------|
| `scale_tier` | `micro`, `small`, … | From farm record |
| `timezone_bucket` | `UTC`, `IANA_REGIONAL`, `OTHER` | No raw IANA string in v1 |
| `currency` | `USD` | Trimmed string |
| `operational_status` | enum string | From farm record |

### `aggregates`

- **`aggregates.costs.totals`** — `income`, `expenses`, `net` (floats).
- **`aggregates.costs.by_category`** — array of objects: `category`, `currency`, `income`, `expense`, `tx_count`.
- **`aggregates.tasks.by_status`** — map of task status string to count.
- **`aggregates.devices.by_status`** — map of device status string to count.

### `privacy`

Sender sets `includes_pii: false`, `includes_raw_location_text: false`, and a static `revocation` message. Receivers must still honor **farm opt-out**: after opt-out, the farm stops sending new payloads.

## Receiver responsibilities (recommended)

1. **Authenticate** the request (Bearer secret, mTLS, or network allowlist — at least one).
2. **Validate** `schema_version` and required keys; reject unknown schema versions with **4xx** so the farm does not treat them as transient server errors.
3. **Persist** payload with **idempotency** at the receiver (see above; outbound POST does not yet include the farm API idempotency key).
4. **Apply retention** to received rows in the receiver’s store (cold storage, TTL, or aggregate-only downstream tables).
5. **Do not** treat `farm_pseudonym` as globally unique without coordination; it is unique given the same key material as the sender.

## Security notes

- Treat `INSERT_COMMONS_SHARED_SECRET` and `INSERT_COMMONS_PSEUDONYM_KEY` as **high-value secrets** (secret manager, rotation policy, separate keys per environment).
- Prefer **TLS** for `INSERT_COMMONS_INGEST_URL` in production.
- Log receiver-side accept/reject decisions with **correlation ids** (for example hash of idempotency key) rather than raw payloads in shared logs if policy requires minimization.

## In-repo pilot receiver (`cmd/insert-commons-receiver`)

This repository ships a **minimal optional service** that implements the contract above, persists accepted payloads into PostgreSQL (`gr33ncore.insert_commons_received_payloads`), and returns JSON `{ "ok", "accepted", "duplicate", "storage_id", "schema" }`.

### Apply migration

On the database that will store ingested rows (often the **same** database as the farm API):

```bash
psql "$DATABASE_URL" -f db/migrations/20260417_phase13_insert_commons_receiver.sql
```

### Configure the farm API (sender)

Point the farm process at the receiver URL (include the path):

```bash
export INSERT_COMMONS_INGEST_URL=http://127.0.0.1:8765/v1/ingest
export INSERT_COMMONS_SHARED_SECRET=your-long-random-secret
```

Use the **same** value for `INSERT_COMMONS_SHARED_SECRET` on both the farm API and the receiver.

### Run the receiver

| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | (see `cmd/insert-commons-receiver/main.go`) | Postgres for `insert_commons_received_payloads` |
| `INSERT_COMMONS_RECEIVER_LISTEN` | `:8765` | Listen address |
| `INSERT_COMMONS_SHARED_SECRET` | (empty) | Bearer token; must match farm unless insecure flag is set |
| `INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH` | unset | If `true`, allows empty secret (**local pilots only**) |
| `INSERT_COMMONS_RECEIVER_RETENTION_DAYS` | `90` | After each accepted ingest, deletes rows older than this many days (`0` disables cleanup) |

```bash
# Example: same DB as the API, authenticated ingest
export DATABASE_URL=postgres://user@/gr33n?host=/var/run/postgresql
export INSERT_COMMONS_SHARED_SECRET=your-long-random-secret
go run ./cmd/insert-commons-receiver/
```

Or use `make run-receiver` from the repository `Makefile`.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| POST | `/v1/ingest` | Accept one payload (`Content-Type: application/json`, optional `Authorization: Bearer …`) |

Validation rejects unknown `schema_version`, malformed JSON, missing required keys, and timestamps more than **10 minutes** in the future or older than **365 days** (abuse guard).

## Related documents

- [`docs/phase-13-operator-documentation.md`](phase-13-operator-documentation.md) — Phase 13 operator doc index
- [`README.md`](../README.md) — Core principles and env vars overview
- [`docs/audit-events-operator-playbook.md`](audit-events-operator-playbook.md) — Farm audit API (includes Insert Commons opt-in and sync attempts)
- [`openapi.yaml`](../openapi.yaml) — Dashboard routes for opt-in, sync, and sync history
