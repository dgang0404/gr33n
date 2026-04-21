#!/usr/bin/env bash
# First-time clone helper: optional Debian apt deps, Go deps, env templates, then bootstrap-local.sh.
# Intended for anyone who can follow README-style steps — run from the repository root after `git clone`.
#
# Usage:
#   ./scripts/setup-first-clone.sh
#   ./scripts/setup-first-clone.sh --install-system-deps   # Debian/Ubuntu: sudo apt (Postgres+Node; see script)
#   ./scripts/setup-first-clone.sh --docker
#
# Pass-through: arguments are forwarded to bootstrap-local.sh (e.g. --seed), except
#   --install-system-deps (consumed here).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: '$1' not found in PATH — install it and try again." >&2
    echo "  See INSTALL.md for versions (Go, Node, PostgreSQL or Docker)." >&2
    exit 1
  }
}

if [[ ! -f "$ROOT/go.mod" ]] || [[ ! -f "$ROOT/db/schema/gr33n-schema-v2-FINAL.sql" ]]; then
  echo "error: run this script from the gr33n repository root (expected go.mod and db/schema/)." >&2
  exit 1
fi

INSTALL_SYS=0
BOOTSTRAP_ARGS=()
for a in "$@"; do
  case "$a" in
    --install-system-deps) INSTALL_SYS=1 ;;
    *) BOOTSTRAP_ARGS+=("$a") ;;
  esac
done

if [[ "$INSTALL_SYS" -eq 1 ]]; then
  if [[ "$(uname -s)" != "Linux" ]] || [[ ! -f /etc/debian_version ]]; then
    echo "error: --install-system-deps only supports Linux with apt (Debian/Ubuntu family)." >&2
    echo "  Run ./scripts/install-system-deps-debian.sh manually after checking INSTALL.md." >&2
    exit 1
  fi
  echo ""
  "$ROOT/scripts/install-system-deps-debian.sh"
  echo ""
fi

echo "gr33n — first-time setup"
echo "------------------------"
echo "This will: prefetch Go modules, ensure .env templates, install UI deps, and run the DB bootstrap."
echo ""

need go
if [[ "${BOOTSTRAP_ARGS[0]:-}" != "--docker" ]]; then
  need psql
fi
need npm

echo "==> go mod download"
go mod download

if [[ ! -f "$ROOT/.env" ]]; then
  echo "==> Creating .env from .env.example (edit DATABASE_URL and secrets once)"
  cp "$ROOT/.env.example" "$ROOT/.env"
else
  echo "==> .env already present; leaving as-is"
fi

if [[ ! -f "$ROOT/ui/.env" ]] && [[ -f "$ROOT/ui/.env.example" ]]; then
  echo "==> Creating ui/.env from ui/.env.example (default API http://localhost:8080)"
  cp "$ROOT/ui/.env.example" "$ROOT/ui/.env"
else
  echo "==> ui/.env already present or no example; skipping ui/.env copy"
fi

echo "==> Running scripts/bootstrap-local.sh${BOOTSTRAP_ARGS[*]:+ ${BOOTSTRAP_ARGS[*]}}"
exec "$ROOT/scripts/bootstrap-local.sh" "${BOOTSTRAP_ARGS[@]}"
