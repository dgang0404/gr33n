# Backup & restore runbook

**Quick reference** for operators. Full storage-cutover steps (local disk → S3, verification, rollback) remain in [receipt-storage-cutover-runbook.md](receipt-storage-cutover-runbook.md).

Linked from: [README](../README.md) · [INSTALL.md](../INSTALL.md) · [upgrade-guide.md](upgrade-guide.md) · [operator-troubleshooting.md](operator-troubleshooting.md)

---

## Principle

Treat **PostgreSQL** and **receipt/file blob storage** as one recovery unit. Back them up at the same time; restore both before bringing the app online.

---

## What to back up

| Asset | Where |
|-------|--------|
| Database | `DATABASE_URL` — all gr33n schemas |
| Receipt blobs | `FILE_STORAGE_DIR` (local) or `S3_BUCKET` + prefix |
| Secrets | `JWT_SECRET`, `PI_API_KEY`, device keys, `S3_*` credentials — store separately from git |
| Config | `.env` (redact before sharing) |

---

## Automated backup (Phase 155)

```bash
make backup                              # pg_dump + local files tar + manifest
make verify-backup                       # latest dump → scratch DB spot-check
make verify-backup BACKUP=path/to.sql    # specific dump
```

**Cron example** (adjust paths):

```cron
0 3 * * * cd /opt/gr33n-platform && make backup >> /var/log/gr33n-backup.log 2>&1
```

**Off-box copy:** rsync `data/backups/` to NAS or `aws s3 sync` — keep secrets out of git; encrypt backup volume at rest.

**Monthly:** `make verify-backup` after a fresh `make backup`.

Env: `GR33N_BACKUP_DIR`, `GR33N_BACKUP_KEEP_DAILY` (default 7), `GR33N_BACKUP_KEEP_WEEKLY` (default 4).

---

## Backup (manual — same as scripts)

```bash
# Database
pg_dump "$DATABASE_URL" > gr33n-db-$(date +%F-%H%M%S).sql

# Local files (adjust path to your FILE_STORAGE_DIR)
tar -C "$(dirname ./data/files)" -czf gr33n-files-$(date +%F-%H%M%S).tar.gz files
```

For S3-backed storage, use provider snapshot/versioning on the bucket prefix documented in your `.env`.

**Verify:** non-empty dump; tarball lists expected receipt keys.

---

## Restore (typical)

1. Stop gr33n API (no writes during restore).
2. Restore database:

```bash
psql "$DATABASE_URL" -f gr33n-db-YYYY-MM-DD-HHMMSS.sql
```

3. Restore local files:

```bash
tar -C /path/to/parent -xzf gr33n-files-YYYY-MM-DD-HHMMSS.tar.gz
```

4. Confirm `FILE_STORAGE_*` env matches where blobs live.
5. Start API; spot-check Costs → receipt attachment download.

---

## When to read the full cutover runbook

Use [receipt-storage-cutover-runbook.md](receipt-storage-cutover-runbook.md) when:

- Moving from `FILE_STORAGE_BACKEND=local` to `s3`
- Validating dual-write or migration of existing `file_attachments` rows
- Planning rollback after a storage migration

---

## Related

- [upgrade-guide.md](upgrade-guide.md) — pull → migrate → restart
- [environment-variables.md](environment-variables.md) — `FILE_STORAGE_*`, `S3_*`
