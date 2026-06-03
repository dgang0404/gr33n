#!/usr/bin/env bash
# Install pgvector for the OS PostgreSQL 14 cluster (PGDG apt).
# Run as your normal user — sudo will prompt for your password.
# Docker dev (port 5433) already includes pgvector; this is for native socket Postgres.
set -euo pipefail

if [[ "${EUID:-}" -eq 0 ]]; then
  echo "error: run as your normal user (script uses sudo)" >&2
  exit 1
fi

need_sudo() {
  if ! sudo -n true 2>/dev/null; then
    echo "You will be prompted for your sudo password."
  fi
}

need_sudo
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl gnupg lsb-release

CODENAME="$(lsb_release -cs)"
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -fsSL -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
  https://www.postgresql.org/media/keys/ACCC4CF8.asc
echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] \
  https://apt.postgresql.org/pub/repos/apt ${CODENAME}-pgdg main" \
  | sudo tee /etc/apt/sources.list.d/pgdg.list >/dev/null

sudo apt-get update -qq
sudo apt-get install -y postgresql-14-pgvector

echo ""
echo "==> Enabling extension on database gr33n (if it exists)"
if sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='gr33n'" | grep -q 1; then
  sudo -u postgres psql -d gr33n -v ON_ERROR_STOP=1 -c "CREATE EXTENSION IF NOT EXISTS vector;"
  sudo -u postgres psql -d gr33n -c "SELECT extname, extversion FROM pg_extension WHERE extname='vector';"
else
  echo "    (no native database 'gr33n' yet — skip; use Docker or createdb gr33n first)"
fi

echo ""
echo "Done. Native PG14 now has pgvector."
echo "Daily dev can still use Docker: DATABASE_URL=postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable"
