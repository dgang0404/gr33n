---
name: Phase 155 — automated backups
overview: >
  Turn docs/backup-restore-runbook.md from a manual pg_dump recipe into an
  operator-ready backup script with retention, integrity verification, and
  Makefile targets. gr33n is self-hosted — sensor history, crop cycles, and
  cost data have no safety net if disk dies between manual backups.
todos:
  - id: ws1-backup-script
    content: "WS1: scripts/backup-gr33n.sh — pg_dump + FILE_STORAGE_DIR tar + timestamped output dir"
    status: pending
  - id: ws2-retention
    content: "WS2: Rotation — keep last N daily + last M weekly; configurable via env or flags"
    status: pending
  - id: ws3-verify
    content: "WS3: scripts/verify-backup-gr33n.sh — restore dump to scratch DB, spot-check row counts"
    status: pending
  - id: ws4-makefile
    content: "WS4: make backup / make verify-backup targets; source .env for DATABASE_URL and FILE_STORAGE_DIR"
    status: pending
  - id: ws5-runbook
    content: "WS5: Update backup-restore-runbook.md with cron example, off-box copy note, verify step"
    status: pending
isProject: false
---

# Phase 155 — automated backups

**Status:** planned · **Depends on:** [116 docs refresh](phase_116_docs_refresh.plan.md) (backup runbook exists) · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

---

## Why this phase

[`docs/backup-restore-runbook.md`](../backup-restore-runbook.md) is accurate but **entirely manual**. An operator must remember to run `pg_dump` and tar their file storage. There is no:

- cron example checked into the repo
- retention / rotation (old dumps pile up or get deleted by hand)
- automated integrity check ("does this dump actually restore?")
- single command that backs up **both** Postgres and receipt blobs as one recovery unit

For a self-hosted farm OS holding years of sensor readings, crop cycles, and finance receipts, this is the highest-consequence gap in the 154–158 arc. It's invisible until the day it isn't.

---

## Workstreams

### WS1 — Backup script

**Target:** `scripts/backup-gr33n.sh`

- Read `DATABASE_URL` and `FILE_STORAGE_DIR` from `.env` (same pattern as `scripts/source-local-env.sh`)
- Write to a configurable output root (default `data/backups/` or `$GR33N_BACKUP_DIR`)
- Artifacts per run:
  - `gr33n-db-YYYY-MM-DD-HHMMSS.sql` via `pg_dump`
  - `gr33n-files-YYYY-MM-DD-HHMMSS.tar.gz` when `FILE_STORAGE_BACKEND=local`
  - `manifest.json` — timestamp, sizes, schema version hint (`make migrate` revision), hostname
- For `FILE_STORAGE_BACKEND=s3`: skip tar, document in manifest that operator must rely on bucket versioning/snapshots (link to receipt cutover runbook)
- Exit non-zero on empty dump or missing storage dir when local backend expected
- **Non-goal:** encrypting backups in v1 — document "encrypt at rest on your backup volume" in runbook

### WS2 — Retention

- Flags or env: `GR33N_BACKUP_KEEP_DAILY=7`, `GR33N_BACKUP_KEEP_WEEKLY=4`
- After successful backup, prune older artifacts beyond retention
- Never delete the only remaining backup

### WS3 — Verify script

**Target:** `scripts/verify-backup-gr33n.sh`

- Takes path to a `.sql` dump (and optional files tarball)
- Creates a **scratch** database (`createdb gr33n_backup_verify_$$` or `psql` to temp DB) — never touches production `DATABASE_URL`
- Restores dump, runs cheap sanity queries:
  - `SELECT COUNT(*) FROM auth.users` (or skip if auth schema empty in minimal installs)
  - `SELECT COUNT(*) FROM gr33ncore.farms`
  - `SELECT COUNT(*) FROM gr33ncrops.crop_catalog_entries` when catalog seeded
- Drops scratch DB on exit (trap on ERR)
- Optional: list tarball contents and compare file count to `file_attachments` row count when local storage
- Exit 0 only when restore + spot-checks pass

### WS4 — Makefile targets

```makefile
backup:        ## Phase 155 — pg_dump + file storage backup
verify-backup: ## Phase 155 — restore latest dump to scratch DB and spot-check
```

- `verify-backup` accepts `BACKUP=path/to/dump.sql` or defaults to newest in backup dir

### WS5 — Runbook update

Extend [`backup-restore-runbook.md`](../backup-restore-runbook.md):

- Cron example (`0 3 * * * cd /opt/gr33n && make backup`)
- Off-box copy guidance (rsync to NAS, S3 sync for dump files — secrets stay out of git)
- "After backup, monthly: `make verify-backup`"
- Link from [INSTALL.md](../../INSTALL.md) § backup (if missing)

---

## Acceptance

- [ ] `make backup` produces non-empty SQL + files archive on a dev-auth-test install
- [ ] `make verify-backup BACKUP=…` restores to scratch DB without touching production data
- [ ] Retention prunes old backups but keeps at least one
- [ ] Runbook documents cron + verify cadence
- [ ] Script fails loudly when `DATABASE_URL` unset or `pg_dump` errors

## Non-goals

- Hosted/managed backup SaaS integration
- Point-in-time recovery / WAL archiving (Postgres PITR) — document as future enterprise item
- Backing up Ollama model weights or RAG index blobs (operational, regenerable from `make rag-ingest-*`)

## Operator commands (after ship)

```bash
make backup
make verify-backup                    # latest
make verify-backup BACKUP=data/backups/gr33n-db-2026-07-11.sql
```
