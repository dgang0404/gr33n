#!/usr/bin/env bash
# Idempotent local bootstrap: prerequisites, env template, database schema (+ optional seed), UI deps.
# Does not start long-running servers — use `make dev`, `make run`, or `docker compose` after this.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

usage() {
  cat <<'EOF'
Usage: scripts/bootstrap-local.sh [options]

  (default)     Native Postgres: apply schema, optional migrations, optional seed; npm ci in ui/
  --docker      Run `docker compose up -d` (DB + API + UI in containers); npm ci in ui/ for local tooling
  --seed        After schema, load db/seeds/master_seed.sql (demo farm_id=1; skip if using dashboard templates only)
  --skip-schema Skip db/schema/gr33n-schema-v2-FINAL.sql only (enums/tables already exist); migrations still run
  -h, --help    This message

Environment:
  DATABASE_URL  Postgres URL (default: postgres://$USER@/gr33n?host=/var/run/postgresql)

See docs/local-operator-bootstrap.md for full path and first-user steps.
EOF
}

USE_DOCKER=0
SEED=0
SKIP_SCHEMA=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --docker) USE_DOCKER=1 ;;
    --seed) SEED=1 ;;
    --skip-schema) SKIP_SCHEMA=1 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
  shift
done

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing command '$1' (install or add to PATH)" >&2
    exit 1
  }
}

echo "==> gr33n local bootstrap (repo: $ROOT)"

if [[ "$USE_DOCKER" -eq 1 ]]; then
  need docker
  echo "==> Starting Docker Compose (Postgres + API + UI)"
  docker compose up -d
  echo "    UI: http://localhost:5173  API: http://localhost:8080"
  echo "    For a host-run API/UI against this DB, set DATABASE_URL in .env, e.g.:"
  echo "    postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable"
else
  need psql
  if [[ -f "$ROOT/.env" ]]; then
    set -a
    # shellcheck disable=1091
    source "$ROOT/.env"
    set +a
  fi
  DATABASE_URL="${DATABASE_URL:-postgres://${USER}@/gr33n?host=/var/run/postgresql}"
  export DATABASE_URL

  echo "==> Checking database connection ($DATABASE_URL)"
  # Avoid racing `docker compose up -d db` — first init can take 30–60s.
  max_waits=60
  n=0
  until psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "SELECT 1" >/dev/null 2>&1; do
    n=$((n + 1))
    if [[ "$n" -ge "$max_waits" ]]; then
      cat <<EOF >&2
error: cannot connect to Postgres with DATABASE_URL after ${max_waits} attempts (~$((max_waits * 2))s).

If using Docker Compose db: wait until it is ready, then retry:
  sg docker -c 'docker compose exec db pg_isready -U gr33n -d gr33n'

Otherwise create the database and extensions first, for example:
  sudo -u postgres psql -c "CREATE DATABASE gr33n;"
  sudo -u postgres psql -d gr33n -c "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;"
  sudo -u postgres psql -d gr33n -c "CREATE EXTENSION IF NOT EXISTS vector;"

Peer auth (typical Linux): ensure a Postgres role exists for your OS user:
  sudo -u postgres psql -c "CREATE USER $USER WITH SUPERUSER;"

Then re-run this script or set DATABASE_URL (see INSTALL.md).
EOF
      exit 1
    fi
    echo "    waiting for Postgres (${n}/${max_waits})..."
    sleep 2
  done

  if [[ "$SKIP_SCHEMA" -eq 0 ]]; then
    echo "==> Applying db/schema/gr33n-schema-v2-FINAL.sql"
    if ! psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/db/schema/gr33n-schema-v2-FINAL.sql"; then
      seed_hint=""
      [[ "$SEED" -eq 1 ]] && seed_hint=" --seed"
      cat <<EOF >&2

error: schema SQL failed (see messages above). Typical causes:
  • Running the monolithic schema twice on a DB that already has enums/tables (stops on CREATE TYPE).
  • A previous run aborted mid-file — objects after the error line were never created.

Recovery options:
  • Fresh Compose volume (destructive):  sg docker -c 'cd $ROOT && docker compose down -v && docker compose up -d db'
    then:  $ROOT/scripts/bootstrap-local.sh${seed_hint}
  • DB already fully provisioned but you hit CREATE TYPE on re-run — skip only the big schema file (migrations still apply):
      $ROOT/scripts/bootstrap-local.sh --skip-schema${seed_hint}

Otherwise drop/recreate the database or restore from backup, then re-run this script.
EOF
      exit 1
    fi
  else
    echo "==> Skipping db/schema/gr33n-schema-v2-FINAL.sql (--skip-schema); migrations still run."
  fi

  echo "==> Applying db/migrations/*.sql (order: filename)"
  shopt -s nullglob
  for f in $(printf '%s\n' "$ROOT"/db/migrations/*.sql | LC_ALL=C sort); do
    echo "    -> $(basename "$f")"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
  done
  shopt -u nullglob

  if [[ "$SEED" -eq 1 ]]; then
    echo "==> Loading demo seed (db/seeds/master_seed.sql)"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$ROOT/db/seeds/master_seed.sql"
  else
    echo "==> Skipping demo seed (use --seed for master_seed.sql, or create farms from the dashboard)"
  fi
fi

if [[ ! -f "$ROOT/.env" ]]; then
  echo "==> Creating .env from .env.example (edit DATABASE_URL and secrets for production)"
  cp "$ROOT/.env.example" "$ROOT/.env"
else
  echo "==> .env already exists; not overwriting"
fi

if command -v npm >/dev/null 2>&1; then
  echo "==> Installing UI dependencies (npm ci --legacy-peer-deps)"
  (cd "$ROOT/ui" && npm ci --legacy-peer-deps)
elif [[ "$USE_DOCKER" -eq 1 ]]; then
  echo "==> npm not found; skipping ui/ install (Compose already runs the UI container)"
else
  echo "error: npm is required for native bootstrap (install Node.js) or use --docker" >&2
  exit 1
fi

echo ""
echo "Done."
echo "  • Full walkthrough: docs/local-operator-bootstrap.md"
echo "  • Run API + UI (native): make dev   (or make run + make ui)"
echo "  • First account: open the UI and register, or POST /auth/register (password ≥ 8 chars)"
echo "  • Insert Commons JSON rules: docs/insert-commons-pipeline-runbook.md"
