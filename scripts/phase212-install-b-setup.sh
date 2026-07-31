#!/usr/bin/env bash
# Phase 212 WS1 — bring up Install B as a sibling clone on remapped ports.
# Laptop path = hybrid (Compose DB :5434 + host API :8081), same style as Install A.
# Never push that clone. Re-runnable; does not touch Install A.
#
# Usage:
#   ./scripts/phase212-install-b-setup.sh
#   INSTALL_B_DIR=~/gr33n-platform-b ./scripts/phase212-install-b-setup.sh
#
# After setup, start API (separate terminal / background):
#   cd ~/gr33n-platform-b && ./scripts/phase212-install-b-serve.sh
set -euo pipefail

INSTALL_A_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_B_DIR="${INSTALL_B_DIR:-$HOME/gr33n-platform-b}"
CLONE_SRC="${CLONE_SRC:-$INSTALL_A_ROOT}"

echo "==> Phase 212 Install B → $INSTALL_B_DIR"
echo "    clone source: $CLONE_SRC"

if [[ ! -d "$INSTALL_B_DIR/.git" ]]; then
  git clone "$CLONE_SRC" "$INSTALL_B_DIR"
else
  echo "    clone already present; leaving git state as-is"
fi

# Sync seeds/scripts from A so B gets Phase 212 SQL even if clone was earlier.
mkdir -p "$INSTALL_B_DIR/db/seeds" "$INSTALL_B_DIR/scripts"
cp -f "$INSTALL_A_ROOT/db/seeds/farm_b_seed.sql" "$INSTALL_B_DIR/db/seeds/farm_b_seed.sql" 2>/dev/null || true
cp -f "$INSTALL_A_ROOT/scripts/phase212-install-b-serve.sh" "$INSTALL_B_DIR/scripts/phase212-install-b-serve.sh" 2>/dev/null || true
chmod +x "$INSTALL_B_DIR/scripts/phase212-install-b-serve.sh" 2>/dev/null || true

# Dedicated compose file (not merge/override) — Compose appends ports from the base
# file, which would steal Install A's :5433. See docker-compose.phase212-b.yml.
cp -f "$INSTALL_A_ROOT/docker-compose.phase212-b.yml" "$INSTALL_B_DIR/docker-compose.phase212-b.yml"
# Remove leftover override from earlier WS1 attempts (port-merge trap).
rm -f "$INSTALL_B_DIR/docker-compose.override.yml"

if [[ ! -f "$INSTALL_B_DIR/.env" ]]; then
  cp "$INSTALL_B_DIR/.env.example" "$INSTALL_B_DIR/.env"
fi

python3 - <<'PY' "$INSTALL_B_DIR/.env"
import re, sys
from pathlib import Path
p = Path(sys.argv[1])
text = p.read_text()
def set_kv(text, key, value):
    pat = re.compile(rf'(?m)^{re.escape(key)}=.*$')
    line = f'{key}={value}'
    if pat.search(text):
        return pat.sub(line, text)
    return text.rstrip() + '\n' + line + '\n'
text = set_kv(text, 'DATABASE_URL', 'postgres://gr33n:gr33n@127.0.0.1:5434/gr33n?sslmode=disable')
text = set_kv(text, 'PORT', '8081')
text = set_kv(text, 'CORS_ORIGIN', 'http://localhost:5174')
text = set_kv(text, 'AI_ENABLED', 'false')
text = set_kv(text, 'AUTH_MODE', 'auth_test')
p.write_text(text)
print(f'patched {p}')
PY

mkdir -p "$INSTALL_B_DIR/ui"
if [[ ! -f "$INSTALL_B_DIR/ui/.env" ]] && [[ -f "$INSTALL_B_DIR/ui/.env.example" ]]; then
  cp "$INSTALL_B_DIR/ui/.env.example" "$INSTALL_B_DIR/ui/.env"
fi
if [[ -f "$INSTALL_B_DIR/ui/.env" ]]; then
  python3 - <<'PY' "$INSTALL_B_DIR/ui/.env"
import re, sys
from pathlib import Path
p = Path(sys.argv[1])
text = p.read_text() if p.exists() else ''
line = 'VITE_API_URL=http://localhost:8081'
pat = re.compile(r'(?m)^VITE_API_URL=.*$')
text = pat.sub(line, text) if pat.search(text) else text.rstrip() + '\n' + line + '\n'
p.write_text(text)
print(f'patched {p}')
PY
fi

cd "$INSTALL_B_DIR"
echo "==> Compose DB only on :5434 (docker-compose.phase212-b.yml)"
docker compose -f docker-compose.phase212-b.yml down 2>/dev/null || true
docker compose -f docker-compose.phase212-b.yml up -d

echo "==> Wait for Postgres on :5434"
for i in $(seq 1 60); do
  if docker compose -f docker-compose.phase212-b.yml exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
    echo "    ready ($i)"
    break
  fi
  sleep 2
done

echo "==> Schema + master_seed against Install B DATABASE_URL"
# --docker only starts compose; schema/seed need the native path with B's DATABASE_URL.
set -a
# shellcheck disable=1091
source "$INSTALL_B_DIR/.env"
set +a
export DATABASE_URL
./scripts/bootstrap-local.sh --seed

if [[ -f "$INSTALL_B_DIR/db/seeds/farm_b_seed.sql" ]]; then
  echo "==> Applying farm_b_seed.sql (Organization B + Farm B)"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$INSTALL_B_DIR/db/seeds/farm_b_seed.sql"
fi

echo ""
echo "Install B DB ready on :5434 (Farm B / Organization B seeded)."
echo "Start API next:"
echo "  cd $INSTALL_B_DIR && ./scripts/phase212-install-b-serve.sh"
echo "Optional UI:"
echo "  cd $INSTALL_B_DIR/ui && npm run dev -- --host 0.0.0.0 --port 5174"
