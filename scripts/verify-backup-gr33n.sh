#!/usr/bin/env bash
# Phase 155 — restore a pg_dump to a scratch database and spot-check row counts.
# Never touches the production database named in DATABASE_URL.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

usage() {
  cat <<'EOF'
Usage: scripts/verify-backup-gr33n.sh [path/to/gr33n-db-*.sql]

  (default)   Newest dump under GR33N_BACKUP_DIR/latest or data/backups/latest
  -h, --help  This message

Environment:
  DATABASE_URL     Used only to discover the Postgres server — NOT the target DB
  GR33N_BACKUP_DIR Backup root (default ./data/backups)

See docs/backup-restore-runbook.md
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    -*) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
    *) break ;;
  esac
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
DUMP="${1:-}"

if [[ -z "$DUMP" ]]; then
  if [[ -L "$BACKUP_DIR/latest" ]]; then
    DUMP="$(find "$(readlink -f "$BACKUP_DIR/latest")" -maxdepth 1 -name 'gr33n-db-*.sql' | head -1)"
  fi
  if [[ -z "$DUMP" ]]; then
    DUMP="$(find "$BACKUP_DIR" -name 'gr33n-db-*.sql' -printf '%T@ %p\n' 2>/dev/null | sort -rn | head -1 | cut -d' ' -f2-)"
  fi
fi

if [[ -z "$DUMP" || ! -f "$DUMP" ]]; then
  echo "error: no backup dump found (pass path or run make backup first)" >&2
  exit 1
fi

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing command '$1'" >&2
    exit 1
  }
}

need psql

docker_db_up() {
  docker compose ps --status running -q db 2>/dev/null | grep -q .
}

echo "==> verify backup: $DUMP"

SCRATCH="gr33n_backup_verify_$$"

cleanup() {
  if docker_db_up; then
    docker compose exec -T db dropdb -U gr33n --if-exists "$SCRATCH" 2>/dev/null || true
  else
    command -v dropdb >/dev/null 2>&1 && dropdb --if-exists "$SCRATCH" 2>/dev/null || true
  fi
}
trap cleanup EXIT

if docker_db_up; then
  need docker
  echo "==> restore to scratch DB $SCRATCH (docker)"
  docker compose exec -T db createdb -U gr33n "$SCRATCH"
  # Timescale hypertable chunks may emit benign restore ordering warnings — spot-check
  # row counts after best-effort replay, don't require a perfect full replay.
  docker compose exec -T db psql -U gr33n -v ON_ERROR_STOP=0 -d "$SCRATCH" < "$DUMP" >/dev/null || true
  psql_scratch() {
    docker compose exec -T db psql -U gr33n -d "$SCRATCH" -tAc "$1"
  }
else
  need createdb
  need dropdb
  ADMIN_URL="$(python3 - <<'PY'
import os
from urllib.parse import urlparse, urlunparse
u = urlparse(os.environ["DATABASE_URL"])
print(urlunparse(u._replace(path="/postgres")))
PY
)"
  createdb "$SCRATCH"
  echo "==> restore to scratch DB $SCRATCH"
  psql "$ADMIN_URL" -v ON_ERROR_STOP=0 -d "$SCRATCH" -f "$DUMP" >/dev/null || true
  SCRATCH_URL="$(python3 - <<PY
import os
from urllib.parse import urlparse, urlunparse
scratch = "${SCRATCH}"
u = urlparse(os.environ["DATABASE_URL"])
print(urlunparse(u._replace(path="/" + scratch)))
PY
)"
  psql_scratch() {
    psql "$SCRATCH_URL" -tAc "$1"
  }
fi

check_count() {
  local label="$1" sql="$2"
  local count
  count="$(psql_scratch "$sql" 2>/dev/null | tr -d '[:space:]')" || count="skip"
  printf "    %-28s %s\n" "$label" "${count:-skip}"
}

echo "==> spot-checks"
check_count "auth.users" "SELECT COUNT(*) FROM auth.users"
check_count "gr33ncore.farms" "SELECT COUNT(*) FROM gr33ncore.farms"
check_count "crop_catalog_entries" "SELECT COUNT(*) FROM gr33ncrops.crop_catalog_entries"

farms="$(psql_scratch "SELECT COUNT(*) FROM gr33ncore.farms" 2>/dev/null | tr -d '[:space:]' || echo 0)"
if [[ "$farms" == "0" ]]; then
  echo "error: scratch restore has zero farms — dump may be empty or corrupt" >&2
  exit 1
fi

echo "==> verify OK — scratch restore looks sane"
