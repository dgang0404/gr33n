#!/usr/bin/env bash
# Install OS-level dependencies for local gr33n development on Debian / Ubuntu (amd64/arm64).
# Uses sudo apt — you will be prompted for your password.
#
# Installs:
#   - PostgreSQL 16 (PGDG apt), PostGIS 3, pgvector, TimescaleDB 2
#   - Node.js 22 LTS (NodeSource) + global npm
# Does NOT install Go: the project requires Go 1.25+; see INSTALL.md or https://go.dev/dl/
#
# Usage (from repo root):
#   ./scripts/install-system-deps-debian.sh
#   ./scripts/install-system-deps-debian.sh --skip-node   # if you use nvm/fnm already
#
# After this, continue with INSTALL.md §2 (create DB, extensions, peer user) or
#   ./scripts/setup-first-clone.sh
set -euo pipefail

SKIP_NODE=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-node) SKIP_NODE=1 ;;
    -h|--help)
      grep -E '^#(|  | Usage)' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
  shift || true
done

if [[ "${EUID:-}" -eq 0 ]]; then
  echo "error: run this script as a normal user (it will invoke sudo); do not run as root." >&2
  exit 1
fi

if [[ ! -f /etc/os-release ]]; then
  echo "error: /etc/os-release not found — this script supports Debian/Ubuntu only." >&2
  exit 1
fi

# shellcheck source=/dev/null
. /etc/os-release

if [[ ! -f /etc/debian_version ]]; then
  echo "error: not a Debian-family system (/etc/debian_version missing). Use INSTALL.md or Docker." >&2
  exit 1
fi

# Ubuntu derivatives (Mint, Pop, …) often set UBUNTU_CODENAME for apt repos (PGDG / Timescale).
APT_CODENAME="${UBUNTU_CODENAME:-$VERSION_CODENAME}"
if [[ -z "${APT_CODENAME:-}" ]]; then
  echo "error: could not determine apt codename (VERSION_CODENAME / UBUNTU_CODENAME)." >&2
  exit 1
fi

need_sudo() {
  if ! sudo -n true 2>/dev/null; then
    echo "This step needs sudo (apt). You may be prompted for your password."
  fi
}

need_sudo

PG_MAJOR=16

echo "==> Installing base apt packages"
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl gnupg lsb-release wget

echo "==> Adding PostgreSQL PGDG apt repository (${APT_CODENAME}-pgdg)"
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -fsSL -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
  https://www.postgresql.org/media/keys/ACCC4CF8.asc
echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt ${APT_CODENAME}-pgdg main" \
  | sudo tee /etc/apt/sources.list.d/pgdg.list >/dev/null

echo "==> Adding TimescaleDB apt repository"
# Timescale hosts separate packagecloud paths; only stock Debian uses the debian URL.
if [[ "${ID:-}" == "debian" ]]; then
  TIMESCALE_BASE="https://packagecloud.io/timescale/timescaledb/debian"
else
  TIMESCALE_BASE="https://packagecloud.io/timescale/timescaledb/ubuntu"
fi
curl -fsSL "${TIMESCALE_BASE}/gpgkey" \
  | sudo gpg --dearmor -o /usr/share/keyrings/timescaledb-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/timescaledb-archive-keyring.gpg] ${TIMESCALE_BASE}/ ${APT_CODENAME} main" \
  | sudo tee /etc/apt/sources.list.d/timescaledb.list >/dev/null

sudo apt-get update -qq

echo "==> Installing PostgreSQL ${PG_MAJOR}, PostGIS, pgvector, TimescaleDB"
sudo apt-get install -y \
  "postgresql-${PG_MAJOR}" \
  "postgresql-contrib-${PG_MAJOR}" \
  "postgresql-${PG_MAJOR}-postgis-3" \
  "postgresql-${PG_MAJOR}-pgvector" \
  "timescaledb-2-postgresql-${PG_MAJOR}"

if [[ "$SKIP_NODE" -eq 0 ]]; then
  echo "==> Installing Node.js 22 LTS (NodeSource)"
  tmp="$(mktemp)"
  curl -fsSL https://deb.nodesource.com/setup_22.x -o "$tmp"
  sudo -E bash "$tmp"
  rm -f "$tmp"
  sudo apt-get install -y nodejs
else
  echo "==> Skipping Node.js (--skip-node)"
fi

echo "==> Starting PostgreSQL cluster (if needed)"
sudo pg_ctlcluster "${PG_MAJOR}" main restart 2>/dev/null \
  || sudo systemctl restart postgresql || true

echo ""
echo "Done. Installed:"
echo "  • PostgreSQL ${PG_MAJOR} (PGDG) + PostGIS + pgvector + TimescaleDB"
if [[ "$SKIP_NODE" -eq 0 ]]; then
  node -v 2>/dev/null && echo "  • Node $(node -v)"
fi
echo ""
echo "Go $(go version 2>/dev/null || echo 'not found') — project needs Go 1.25+."
if ! command -v go >/dev/null 2>&1 || ! go version | grep -qE 'go1\.(2[5-9]|[3-9][0-9])'; then
  echo ""
  echo "Install Go from https://go.dev/dl/ (extract to /usr/local/go and add PATH), or use your distro backports."
fi
echo ""
echo "sqlc (after Go works):  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
echo "Next:"
echo "  • INSTALL.md §2 — create database gr33n, CREATE EXTENSION timescaledb / vector, create peer OS user role"
echo "  • ./scripts/setup-first-clone.sh"
