# Backup, Restore, and Storage Operator Runbook

This runbook combines the storage cutover steps from Phase 12 Workstream 1 with the operator recovery guidance needed for Workstream 5. It covers:

- backing up PostgreSQL plus receipt blobs together
- restoring a deployment from those backups
- moving receipt storage from local `FILE_STORAGE_DIR` to object storage
- verifying cutovers and recoveries before deleting old data

## Scope

Use this when:

- Phase 11 or early Phase 12 deployments stored receipt attachments on local disk
- the target deployment should use `FILE_STORAGE_BACKEND=s3`
- operators need one place for backup, restore, cutover, and rollback guidance

This runbook assumes:

- `DATABASE_URL` points at the target database for the operation being run
- receipt attachment rows already exist in `gr33ncore.file_attachments`
- the target object store bucket/container already exists if using S3-compatible storage

## Storage Modes

Local development or single-node mode:

- `FILE_STORAGE_BACKEND=local`
- `FILE_STORAGE_DIR=./data/files`

Production object storage mode:

- `FILE_STORAGE_BACKEND=s3`
- `S3_BUCKET=<bucket>`
- `S3_REGION=<region>`
- optional: `S3_ENDPOINT=<custom endpoint>`
- optional: `S3_PREFIX=<key prefix>`
- optional: `S3_ACCESS_KEY_ID`
- optional: `S3_SECRET_ACCESS_KEY`
- optional: `S3_USE_PATH_STYLE=true`
- optional: `S3_DISABLE_HTTPS=true` for local/test endpoints only
- optional: `FILE_STORAGE_SIGNED_URL_TTL_SECONDS=300`

## Operator Principles

- Treat the database and blob storage as one recovery unit.
- Do not delete old blob storage until cutover or restore verification is complete.
- Prefer dry runs and reversible config changes before destructive cleanup.
- Keep at least one known-good backup from before major storage changes.

## Backup Runbook

### What to back up

- PostgreSQL data for the live `gr33n` deployment
- receipt blob storage:
  - local directory if `FILE_STORAGE_BACKEND=local`
  - object-store bucket/prefix if `FILE_STORAGE_BACKEND=s3`
- deployment config needed to reconnect the app:
  - `DATABASE_URL`
  - file storage env vars
  - JWT/API key secrets

### Pre-backup checklist

- Confirm the application is healthy before starting.
- Confirm the database is reachable.
- Confirm the blob store path or bucket is reachable.
- Confirm there is enough disk space in the backup destination.
- Record the active storage backend and storage location.

### Database backup example

```bash
pg_dump "$DATABASE_URL" > gr33n-db-$(date +%F-%H%M%S).sql
```

If you prefer PostgreSQL custom format for restore tooling:

```bash
pg_dump -Fc "$DATABASE_URL" > gr33n-db-$(date +%F-%H%M%S).dump
```

### Local blob backup example

```bash
tar -C /path/to/file_storage_parent -czf gr33n-files-$(date +%F-%H%M%S).tar.gz files
```

Replace `files` with the basename of the live `FILE_STORAGE_DIR`.

### Object storage backup note

If receipts already live in S3-compatible storage, use the provider's replication, versioning, or bucket-copy tooling to preserve the active bucket or prefix. The important requirement is that the blob backup corresponds to the same general point in time as the DB backup.

### Backup verification checklist

- Confirm the DB backup file exists and is non-empty.
- Confirm the blob backup archive or bucket snapshot completed successfully.
- Record where the backup artifacts are stored.
- Test that at least one recent backup can be opened or listed by operators.

## Restore Runbook

### Restore prerequisites

- A target PostgreSQL instance is available.
- A target blob location is available.
- The app is stopped or isolated from writes during restore validation.
- Operators know whether they are restoring to local storage or object storage.

### Database restore example

Plain SQL backup:

```bash
psql "$DATABASE_URL" -f gr33n-db-YYYY-MM-DD-HHMMSS.sql
```

Custom format backup:

```bash
pg_restore -d "$DATABASE_URL" gr33n-db-YYYY-MM-DD-HHMMSS.dump
```

### Local blob restore example

```bash
mkdir -p /path/to/file_storage_parent
tar -C /path/to/file_storage_parent -xzf gr33n-files-YYYY-MM-DD-HHMMSS.tar.gz
```

### Restore verification checklist

- Start the API against the restored DB and blob store.
- Check `GET /health`.
- Log into the UI successfully.
- Open several existing receipts from the Costs UI.
- Verify both an older and a recently uploaded receipt.
- Upload a new receipt and confirm it opens.
- Replace a receipt and confirm the new one opens.

### Restore rollback

If the restore target is invalid:

1. Stop traffic to the restored environment.
2. Discard the broken restore target.
3. Repeat the restore from the last known-good backup set.

Do not mix a DB backup from one point in time with unrelated blob storage from another point in time unless the mismatch is understood and accepted.

## Storage Cutover Runbook

Use this section when moving receipts from a local `FILE_STORAGE_DIR` to object storage without rewriting DB rows.

### Pre-cutover checklist

- Confirm the source `FILE_STORAGE_DIR` is readable from the machine running the backfill.
- Confirm the target bucket and credentials allow object writes.
- Confirm a recent database backup exists.
- Confirm operators know not to delete the old local storage until post-cutover verification is complete.

### Dry run

Run the backfill in dry-run mode first:

```bash
go run ./cmd/filebackfill --source-dir /path/to/old/files --dry-run
```

Optional receipt-only pass:

```bash
go run ./cmd/filebackfill --source-dir /path/to/old/files --file-type cost_receipt --dry-run
```

Expected result:

- the command connects to the DB successfully
- matching attachment rows are counted
- no target blobs are written
- there are no missing-file errors from the source directory

If dry-run fails:

- verify `DATABASE_URL`
- verify the source directory path
- verify the target backend env vars
- inspect whether some `storage_path` rows point to files that no longer exist on disk

### Backfill

Run the real copy:

```bash
go run ./cmd/filebackfill --source-dir /path/to/old/files
```

Optional receipt-only pass:

```bash
go run ./cmd/filebackfill --source-dir /path/to/old/files --file-type cost_receipt
```

Notes:

- the backfill preserves each row's existing `storage_path`
- DB rows are not updated during the copy
- this makes the operation reversible until application config is switched

### Cutover

After backfill completes:

1. Deploy the API with the new storage env vars.
2. Keep the old local storage mounted and unchanged during the verification window.
3. Restart the API or roll the deployment normally.

Recommended target env:

```bash
FILE_STORAGE_BACKEND=s3
S3_BUCKET=...
S3_REGION=...
S3_ENDPOINT=...
S3_PREFIX=...
S3_ACCESS_KEY_ID=...
S3_SECRET_ACCESS_KEY=...
FILE_STORAGE_SIGNED_URL_TTL_SECONDS=300
```

### Cutover verification checklist

- Open several existing receipts from the Costs UI.
- Verify both a recent receipt and an older receipt open successfully.
- Upload a new receipt after cutover.
- Replace an existing receipt and confirm the new one opens.
- Delete a cost with a receipt and confirm the app behavior is normal.
- Check API logs for storage errors or presign failures.

If you want a direct API-level spot check:

- call `GET /file-attachments/{id}/download`
- for local storage fallback, expect a proxied `/file-attachments/{id}/content` URL
- for S3 storage, expect a short-lived presigned URL

### Cutover rollback

If verification fails after cutover:

1. Switch the API config back to `FILE_STORAGE_BACKEND=local`.
2. Point `FILE_STORAGE_DIR` back at the original local path.
3. Restart or redeploy the API.

Because DB rows were not rewritten during backfill, rollback is config-only as long as the original local storage has been preserved.

### Post-cutover cleanup

Only after the verification window is complete:

- keep one more DB backup on the new configuration
- confirm a fresh sample of receipt downloads still works
- archive or remove the old local blob directory according to operator retention policy

## Failure Modes to Watch

- missing source blobs for some DB rows
- bad object-store credentials or missing bucket permissions
- wrong `S3_ENDPOINT` / path-style config for MinIO or other S3-compatible services
- cutting over before verification is complete
- deleting the old local files before rollback risk is gone
- restoring DB state without the matching blob data
- restoring blob data into the wrong backend or prefix

## Audit and Observability

This section defines what operators should watch during backup, restore, storage cutover, and future federation/sync operations.

### Sensitive-action audit trail checklist

Treat the following as sensitive operational or governance actions and ensure each one is captured in operator records (ticket, changelog, incident timeline, or audit system):

- farm membership changes:
  - member added
  - member removed
  - member role updated
- farm governance changes:
  - Insert Commons opt-in toggled
  - manual Insert Commons sync triggered
- cost and receipt changes:
  - receipt uploaded
  - receipt replaced
  - receipt opened/downloaded for investigation or finance workflows
  - cost with receipt deleted
- destructive mutations:
  - farm deletion
  - large batch deletions in natural farming modules
- auth and secret changes:
  - JWT secret rotation
  - PI/API key rotation
  - object storage credential rotation
- storage mode changes:
  - `FILE_STORAGE_BACKEND` changed
  - `S3_BUCKET` or `S3_PREFIX` changed
  - rollback from object storage to local storage

For each sensitive event, record:

- who executed it
- what changed
- when it changed
- why it changed (ticket or incident link)
- validation result after the change

### Suggested audit evidence sources

Use these as evidence anchors when reviewing incidents or audits:

- Farm audit API: [`docs/audit-events-operator-playbook.md`](audit-events-operator-playbook.md) (`GET /farms/{id}/audit-events`).
- Insert Commons receiver contract: [`docs/insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md).
- API deploy/change history
- operator command history for backup/restore/backfill runs
- API application logs around the change window
- database backups and restore job output
- object storage access logs (when available)

### Operator events to record

Record these actions in the team's operator log, ticket, or deployment notes:

- when a DB backup was started and completed
- when a blob backup or bucket snapshot was started and completed
- when a restore was initiated, validated, and accepted or rejected
- when storage backfill dry-run was executed
- when storage backfill copy was executed
- when `FILE_STORAGE_BACKEND` was changed in production
- when rollback to local storage was performed
- when receipt verification checks passed or failed
- when future sync/federation jobs are paused, resumed, disabled, or revoked

For each event, capture:

- operator name
- environment
- timestamp
- command or deployment reference
- outcome
- follow-up action if something failed

### Logs to watch now

During backup, restore, and storage cutover, watch API and operator-command output for:

- startup logs showing the active `FILE_STORAGE_BACKEND`
- startup logs showing `FILE_STORAGE_DIR` or S3 target config
- receipt cleanup errors after replacement or deletion
- receipt download failures
- presign failures for object-storage-backed downloads
- backfill command failures while reading source blobs or writing target blobs
- health check failures after restore or cutover

Examples of high-signal failure patterns:

- `stored file missing`
- `file storage init:`
- `receipt cleanup attachment`
- `receipt cleanup old attachment`
- `backfill failed`
- `load s3 config:`
- `normalize S3_ENDPOINT:`

### Recommended counters and checks

Even before a dedicated metrics stack exists, operators should track these counts per run or deployment:

- total DB backup attempts
- total DB backup failures
- total blob backup attempts
- total blob backup failures
- total restore attempts
- total restore verification failures
- total backfill attachments scanned
- total backfill attachments copied
- total backfill failures
- total receipt open failures after cutover
- total storage rollback events

If you already have metrics collection, convert these into counters and alerts.

### Alert conditions

Operators should treat these as page-worthy or release-blocking conditions:

- backup job fails or produces an empty artifact
- restore validation fails
- cutover verification fails for existing receipts
- a newly uploaded receipt cannot be opened after cutover
- repeated `stored file missing` errors occur after restore or cutover
- repeated object storage auth or endpoint errors occur at startup or during downloads
- backfill fails before all expected attachments are copied

### Post-change verification window

For backup, restore, and storage cutover changes, keep an explicit verification window where operators:

- monitor logs for at least the first few successful receipt opens
- confirm no new storage errors appear after a fresh upload
- confirm no spike in support issues or user-reported receipt failures
- avoid deleting the old storage location until the window closes cleanly

### Future federation and sync observability

Phase 12 also expects observability for commons/federation sync flows. When that work lands, extend this runbook to track:

- sync attempts started
- sync attempts succeeded
- sync attempts failed
- last successful sync timestamp
- last failed sync timestamp
- rate-limit responses
- auth failures
- backoff/retry state
- opt-in, disable, revoke, and re-enable events

Recommended alerts for future sync work:

- repeated sync failures for the same farm
- no successful sync for longer than the expected interval
- auth failures after a credential rotation
- sudden increase in rate-limit or validation failures

## Production Deployment Hardening Checklist

Use this checklist for production deploys that touch backup, restore, storage, or receipt-serving behavior.

### Configuration and secrets

- Confirm `AUTH_MODE=production`.
- Confirm `JWT_SECRET` is set and not default.
- Confirm `PI_API_KEY` is set and not default.
- Confirm storage config matches the intended backend (`local` or `s3`).
- Confirm object-store credentials are loaded from secure secret storage.
- Confirm no storage credentials or JWT secrets are logged in plaintext.

### Change safety

- Link every storage or restore deploy to a change ticket.
- Ensure at least one operator can execute rollback quickly.
- Confirm pre-change DB backup completed successfully.
- Confirm pre-change blob backup or snapshot completed successfully.
- For storage cutovers, run backfill dry-run before the real copy.

### Post-deploy validation

- Check `GET /health`.
- Verify receipt download for a known existing file.
- Verify upload + open for a newly uploaded receipt.
- Verify receipt replacement flow works.
- Verify no new storage errors appear in logs.

### Rollback readiness

- Keep previous deployment artifact/version available.
- Keep previous storage config values available.
- Keep old local blob directory or prior bucket snapshot intact during the verification window.
- Define who can authorize rollback and where that decision is recorded.

### Minimum release gate (storage/recovery changes)

Do not close a storage/recovery change unless all are true:

- backup validated
- deployment validated
- receipt read/write flows validated
- rollback path confirmed
- audit record for the change is complete

## Operator Notes: PWA Offline Queue

Phase 12 mobile/offline behavior introduces queued task writes in the PWA so field operators can keep working with intermittent connectivity.

### What operators should expect

- task creates and status updates may be queued locally while offline
- queued items show up in the Tasks board with queue state
- users can trigger per-item `Retry` or `Discard`
- queue replay runs automatically when connectivity returns
- `Sync now` is available for manual replay

### Field shift checklist

- Before going offline:
  - ensure users have the installed PWA and recent task data loaded
  - confirm at least one successful online sync
- During offline work:
  - allow queued writes to accumulate
  - instruct users to avoid duplicate manual re-entry unless conflicts appear
- After reconnect:
  - confirm queue count returns to zero
  - review and resolve any stale/conflict items
  - validate a sample of synced tasks in the main task list

### Incident handling for offline queue issues

If queued items do not drain:

1. Check client connectivity and API health.
2. Use `Retry` on failed queue items.
3. If a specific queued change is invalid or obsolete, use `Discard`.
4. If conflicts persist, compare with latest server task state and reapply manually.

Record incident details in operator logs:

- affected farm/user
- queued item count
- conflict or error message shown
- resolution path (`retry`, `discard`, manual re-entry)
