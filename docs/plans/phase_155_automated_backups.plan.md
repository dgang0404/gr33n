---
name: Phase 155 — automated backups
overview: >
  Turn docs/backup-restore-runbook.md from a manual pg_dump recipe into an
  operator-ready backup script with retention, integrity verification, and
  Makefile targets.
todos:
  - id: ws1-backup-script
    content: "WS1: scripts/backup-gr33n.sh — pg_dump + FILE_STORAGE_DIR tar + timestamped output dir"
    status: completed
  - id: ws2-retention
    content: "WS2: Rotation — keep last N daily + last M weekly; configurable via env or flags"
    status: completed
  - id: ws3-verify
    content: "WS3: scripts/verify-backup-gr33n.sh — restore dump to scratch DB, spot-check row counts"
    status: completed
  - id: ws4-makefile
    content: "WS4: make backup / make verify-backup targets; source .env for DATABASE_URL and FILE_STORAGE_DIR"
    status: completed
  - id: ws5-runbook
    content: "WS5: Update backup-restore-runbook.md with cron example, off-box copy note, verify step"
    status: completed
isProject: false
---

# Phase 155 — automated backups

**Status:** shipped · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | [`scripts/backup-gr33n.sh`](../../scripts/backup-gr33n.sh) — `pg_dump`, local files tar, `manifest.json` |
| **WS2** | Retention via `GR33N_BACKUP_KEEP_DAILY` / `GR33N_BACKUP_KEEP_WEEKLY` |
| **WS3** | [`scripts/verify-backup-gr33n.sh`](../../scripts/verify-backup-gr33n.sh) — scratch DB restore + row-count spot-checks |
| **WS4** | `make backup` / `make verify-backup` |
| **WS5** | [`backup-restore-runbook.md`](../backup-restore-runbook.md) — cron, off-box copy, verify cadence |

## Operator commands

```bash
make backup
make verify-backup
make verify-backup BACKUP=data/backups/run-…/gr33n-db-….sql
```

## Close when

- [x] `make backup` produces non-empty SQL + manifest
- [x] `make verify-backup` uses scratch DB only
- [x] Runbook documents automation
- [x] `ui/src/__tests__/phase-155-closure.test.js`
