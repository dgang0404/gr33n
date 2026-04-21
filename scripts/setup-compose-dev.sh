#!/usr/bin/env bash
# After Docker is installed: bring up Compose Postgres only, wait for readiness,
# run bootstrap with --seed (sources .env DATABASE_URL — must be the Compose URL).
#
# Usage (repo root): ./scripts/setup-compose-dev.sh
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing '$1'. Install Docker first, e.g. (Ubuntu 22.04+ jammy):" >&2
    echo "  sudo apt-get update && sudo apt-get install -y docker.io docker-compose-v2" >&2
    echo "  (Stock Ubuntu uses package docker-compose-v2 for \`docker compose\`; docker-compose-plugin is from Docker's apt repo.)" >&2
    echo "  sudo systemctl enable --now docker && sudo usermod -aG docker \"\$USER\" && newgrp docker" >&2
    exit 1
  }
}

need docker
docker compose version >/dev/null 2>&1 || {
  echo "error: install Compose v2 — Ubuntu: sudo apt-get install -y docker-compose-v2" >&2
  exit 1
}

echo "==> Starting db service (first build can take several minutes)"
docker compose up -d db --build

echo "==> Waiting for Postgres (user gr33n)"
for _ in $(seq 1 90); do
  if docker compose exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
    echo "    Postgres is accepting connections."
    break
  fi
  sleep 2
done

if ! docker compose exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
  echo "error: db did not become ready. Try: docker compose logs db" >&2
  exit 1
fi

echo "==> Bootstrap (schema/migrations/seed — reads .env; use DATABASE_URL=postgres://gr33n:gr33n@127.0.0.1:5432/gr33n?sslmode=disable)"
./scripts/bootstrap-local.sh --seed

echo "==> check-stack"
./scripts/check-local-stack.sh

echo ""
echo "Next:  make dev-auth-test"
echo "       curl -s http://localhost:8080/health"
