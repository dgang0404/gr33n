#!/usr/bin/env bash
# Phase 155 — backup PostgreSQL + local file storage as one recovery unit.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

usage() {
  cat <<'EOF'
Usage: scripts/backup-gr33n.sh [options]

  (default)   pg_dump + local FILE_STORAGE_DIR tar + manifest.json
  --no-prune  Skip retention pruning after a successful backup
  -h, --help  This message

Environment (from .env / .env.local):
  DATABASE_URL              Required
  FILE_STORAGE_BACKEND      local (default) or s3
  FILE_STORAGE_DIR          Local blob root (default ./data/files)
  GR33N_BACKUP_DIR          Output directory (default ./data/backups)
  GR33N_BACKUP_KEEP_DAILY   Keep daily dumps this many days (default 7)
  GR33N_BACKUP_KEEP_WEEKLY  Keep weekly snapshots beyond daily window (default 4)

See docs/backup-restore-runbook.md
EOF
}

PRUNE=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-prune) PRUNE=0 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
  shift
done

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=1091
  . "$ROOT/.env"
  set +a
fi
if [[ -f "$ROOT/.env.local" ]]; then
  set -a
  # shellcheck disable=1091
  . "$ROOT/.env.local"
  set +a
fi

: "${DATABASE_URL:?DATABASE_URL is required — set in .env}"

BACKUP_DIR="${GR33N_BACKUP_DIR:-$ROOT/data/backups}"
FILE_BACKEND="${FILE_STORAGE_BACKEND:-local}"
FILES_DIR="${FILE_STORAGE_DIR:-$ROOT/data/files}"
KEEP_DAILY="${GR33N_BACKUP_KEEP_DAILY:-7}"
KEEP_WEEKLY="${GR33N_BACKUP_KEEP_WEEKLY:-4}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing command '$1'" >&2
    exit 1
  }
}

need pg_dump
need python3

docker_db_up() {
  docker compose ps --status running -q db 2>/dev/null | grep -q .
}

run_pg_dump() {
  local err
  err="$(mktemp)"
  if pg_dump "$DATABASE_URL" > "$DB_FILE" 2>"$err"; then
    rm -f "$err"
    return 0
  fi
  if docker_db_up && grep -qiE 'version mismatch|server version' "$err"; then
    echo "==> host pg_dump/client mismatch — using docker compose exec"
    rm -f "$err"
    docker compose exec -T db pg_dump -U gr33n gr33n > "$DB_FILE"
    return $?
  fi
  cat "$err" >&2
  rm -f "$err"
  return 1
}

TS="$(date -u +%Y-%m-%d-%H%M%S)"
RUN_DIR="$BACKUP_DIR/run-$TS"
mkdir -p "$RUN_DIR"

DB_FILE="$RUN_DIR/gr33n-db-$TS.sql"
FILES_FILE="$RUN_DIR/gr33n-files-$TS.tar.gz"
MANIFEST="$RUN_DIR/manifest.json"

echo "==> gr33n backup ($TS)"
echo "    output: $RUN_DIR"

echo "==> pg_dump"
run_pg_dump
DB_BYTES="$(wc -c < "$DB_FILE" | tr -d ' ')"
if [[ "$DB_BYTES" -lt 1024 ]]; then
  echo "error: database dump is suspiciously small ($DB_BYTES bytes)" >&2
  exit 1
fi

FILES_BYTES=0
FILES_NOTE=""
if [[ "$FILE_BACKEND" == "local" ]]; then
  if [[ ! -d "$FILES_DIR" ]]; then
    echo "error: FILE_STORAGE_DIR not found: $FILES_DIR" >&2
    exit 1
  fi
  echo "==> tar file storage ($FILES_DIR)"
  parent="$(cd "$(dirname "$FILES_DIR")" && pwd)"
  base="$(basename "$FILES_DIR")"
  tar -C "$parent" -czf "$FILES_FILE" "$base"
  FILES_BYTES="$(wc -c < "$FILES_FILE" | tr -d ' ')"
else
  FILES_NOTE="s3 — rely on bucket versioning/snapshots; see docs/receipt-storage-cutover-runbook.md"
  echo "==> skipping file tar (FILE_STORAGE_BACKEND=$FILE_BACKEND)"
fi

MIGRATION_HINT=""
if [[ -d "$ROOT/db/migrations" ]]; then
  MIGRATION_HINT="$(ls -1 "$ROOT/db/migrations"/*.sql 2>/dev/null | tail -1 | xargs -r basename)"
fi

python3 - "$MANIFEST" "$TS" "$DB_FILE" "$DB_BYTES" "$FILES_FILE" "$FILES_BYTES" "$FILE_BACKEND" "$FILES_NOTE" "$MIGRATION_HINT" "$(hostname)" <<'PY'
import json, sys
path, ts, db_path, db_bytes, files_path, files_bytes, backend, files_note, mig, host = sys.argv[1:11]
manifest = {
    "timestamp_utc": ts,
    "hostname": host,
    "database_dump": db_path.split("/")[-1],
    "database_bytes": int(db_bytes),
    "file_storage_backend": backend,
    "files_archive": files_path.split("/")[-1] if int(files_bytes) > 0 else None,
    "files_bytes": int(files_bytes) if int(files_bytes) > 0 else None,
    "latest_migration": mig or None,
    "notes": files_note or None,
}
with open(path, "w", encoding="utf-8") as f:
    json.dump(manifest, f, indent=2)
    f.write("\n")
PY

# Symlink latest for verify-backup convenience
ln -sfn "$RUN_DIR" "$BACKUP_DIR/latest"

echo "==> manifest written"
cat "$MANIFEST"

prune_backups() {
  local dir="$1" keep_daily="$2" keep_weekly="$3"
  mapfile -t dumps < <(find "$dir" -maxdepth 2 -name 'gr33n-db-*.sql' -printf '%T@ %p\n' 2>/dev/null | sort -rn | cut -d' ' -f2-)
  local n="${#dumps[@]}"
  if [[ "$n" -le 1 ]]; then
    return 0
  fi
  local now epoch cutoff_daily cutoff_weekly
  now="$(date +%s)"
  cutoff_daily=$((now - keep_daily * 86400))
  cutoff_weekly=$((now - (keep_daily + keep_weekly * 7) * 86400))

  declare -A keep_week=()
  local to_delete=()

  for f in "${dumps[@]}"; do
    local mtime run
    mtime="$(stat -c %Y "$f" 2>/dev/null || stat -f %m "$f")"
    run="$(dirname "$f")"
    if [[ "$mtime" -ge "$cutoff_daily" ]]; then
      continue
    fi
    if [[ "$mtime" -ge "$cutoff_weekly" ]]; then
      local week
      week="$(date -u -d "@$mtime" +%G-W%V 2>/dev/null || date -u -r "$mtime" +%G-W%V)"
      if [[ -z "${keep_week[$week]:-}" ]]; then
        keep_week[$week]=1
        continue
      fi
    fi
    to_delete+=("$run")
  done

  # Never delete latest symlink target's run if it would leave zero backups
  local delete_count=0
  for run in "${to_delete[@]}"; do
    if [[ "$run" == "$(readlink -f "$BACKUP_DIR/latest" 2>/dev/null || true)" ]]; then
      continue
    fi
    echo "==> prune old backup $run"
    rm -rf "$run"
    delete_count=$((delete_count + 1))
  done
  if [[ "$delete_count" -gt 0 ]]; then
    echo "    pruned $delete_count run(s); kept daily=$keep_daily weekly=$keep_weekly"
  fi
}

if [[ "$PRUNE" -eq 1 ]]; then
  echo "==> retention (daily=$KEEP_DAILY weekly=$KEEP_WEEKLY)"
  prune_backups "$BACKUP_DIR" "$KEEP_DAILY" "$KEEP_WEEKLY"
fi

echo "==> backup complete: $RUN_DIR"
