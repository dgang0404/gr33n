#!/usr/bin/env bash
# Post-reboot / quick dev bring-up: start Compose Postgres only, wait (with progress), optional sanity report.
# Does NOT run bootstrap migrations or seed — use scripts/bootstrap-local.sh or scripts/dev-stack.sh when the schema changes.
# Does NOT start API/UI unless --serve (because first `go run` compile can take minutes — not an infinite loop).
#
# Usage (repo root):
#   ./scripts/restart-local.sh
#   ./scripts/restart-local.sh --serve              # then runs make dev-auth-test (API + UI)
#   ./scripts/restart-local.sh --skip-report        # faster if you trust the DB
#   ./scripts/restart-local.sh --quick              # docker compose up without --no-cache rebuild
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

SERVE=0
SKIP_REPORT=0
QUICK=1

usage() {
  cat <<'EOF'
Usage: scripts/restart-local.sh [options]

  (default)      docker compose up -d db, wait for Postgres, db sanity report
  --serve        Run make dev-auth-test after checks (skips if API+UI already up); starts local Ollama when loopback
  --skip-report  Skip scripts/db-sanity-report.sh
  --no-quick     docker compose build db --no-cache before up (slow; rare)
  -h, --help     This message

Tip: Slow startup after reboot is usually `go run` compiling the API. Run once:
  go build -tags dev -o ./bin/api ./cmd/api/
then start ./bin/api with the same env as make run-auth-test (advanced).

Requires .env with DATABASE_URL matching Compose (see .env.example).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --serve) SERVE=1 ;;
    --skip-report) SKIP_REPORT=1 ;;
    --no-quick) QUICK=0 ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

die() {
  echo "error: $*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing '$1' in PATH"
}

need docker
docker compose version >/dev/null 2>&1 || die "need Compose v2 (docker compose)"

[[ -f "$ROOT/.env" ]] || die "missing .env — copy .env.example"

# Laptop dev only — start local Ollama when LLM_BASE_URL points at this machine.
# Enterprise / split inference: LLM_BASE_URL aims at another host — skip auto-start.
maybe_start_local_ollama() {
  set -a
  # shellcheck disable=1091
  source "$ROOT/.env"
  set +a

  [[ "${AI_ENABLED:-true}" != "false" ]] || return 0

  local llm_base="${LLM_BASE_URL:-http://127.0.0.1:11434/v1}"
  if [[ ! "$llm_base" =~ ^https?://(127\.0\.0\.1|localhost)(:|/|$) ]]; then
    echo "==> Ollama: remote LLM_BASE_URL — not auto-starting a local service"
    return 0
  fi

  local ollama_base="${llm_base%/v1}"
  ollama_base="${ollama_base%/}"
  if curl -sf "${ollama_base}/api/tags" >/dev/null 2>&1; then
    echo "==> Ollama: already running (${ollama_base})"
    return 0
  fi

  if ! command -v systemctl >/dev/null 2>&1; then
    echo "==> Ollama: not running — start the Ollama app or service manually"
    return 0
  fi

  if ! systemctl list-unit-files ollama.service >/dev/null 2>&1; then
    echo "==> Ollama: not running — no ollama.service unit (open the Ollama app?)"
    return 0
  fi

  echo "==> Ollama: starting ollama.service (laptop dev)"
  if systemctl start ollama 2>/dev/null || sudo systemctl start ollama 2>/dev/null; then
    for _ in $(seq 1 15); do
      if curl -sf "${ollama_base}/api/tags" >/dev/null 2>&1; then
        echo "    Ollama ready."
        return 0
      fi
      sleep 1
    done
    echo "    Ollama service started — waiting for HTTP (Guardian may show unavailable briefly)"
    return 0
  fi

  echo "==> Ollama: could not start automatically — from any terminal run: systemctl start ollama"
  return 0
}

# Avoid a second API/UI dev stack when ports are already serving gr33n.
maybe_serve_api_ui() {
  set -a
  # shellcheck disable=1091
  source "$ROOT/.env"
  set +a

  local port="${PORT:-8080}"
  local api_ok=0 ui_ok=0

  if curl -sf "http://127.0.0.1:${port}/health" >/dev/null 2>&1; then
    api_ok=1
  fi
  if curl -sf "http://127.0.0.1:5173/" >/dev/null 2>&1 || curl -sf "http://localhost:5173/" >/dev/null 2>&1; then
    ui_ok=1
  fi

  if [[ "$api_ok" -eq 1 && "$ui_ok" -eq 1 ]]; then
    echo "==> API (:${port}) and UI (:5173) already running — leaving them up."
    echo "    Open http://localhost:5173/  ·  stop with Ctrl+C in the terminal that started make dev-auth-test"
    return 0
  fi

  if [[ "$api_ok" -eq 1 && "$ui_ok" -eq 0 ]]; then
    echo "==> API (:${port}) already running — starting UI only (:5173)."
    cd "$ROOT/ui"
    exec npm run dev
  fi

  if [[ "$api_ok" -eq 0 && "$ui_ok" -eq 1 ]]; then
    echo "==> UI (:5173) already running — starting API only (:${port}, AUTH_MODE=auth_test)."
    cd "$ROOT"
    exec make run-auth-test
  fi

  if command -v ss >/dev/null 2>&1; then
    if ss -tln 2>/dev/null | grep -qE ":${port}\\b"; then
      die "port :${port} is in use but /health did not respond — free the port or fix the other process"
    fi
    if ss -tln 2>/dev/null | grep -qE ':5173\b'; then
      die "port :5173 is in use but UI did not respond — free the port or fix the other process"
    fi
  fi

  exec make dev-auth-test
}

compose() {
  if docker info >/dev/null 2>&1; then
    docker compose "$@"
    return
  fi
  if sg docker -c 'docker info' >/dev/null 2>&1; then
    # shellcheck disable=SC2145
    sg docker -c "cd $(printf '%q' "$ROOT") && docker compose $(printf '%q ' "$@")"
    return
  fi
  die "cannot talk to Docker — try: newgrp docker or sg docker -c 'cd $(printf '%q' "$ROOT") && ./scripts/restart-local.sh'"
}

echo "==> restart-local: Compose db + wait + checks (repo: $ROOT)"

if [[ "$QUICK" -eq 1 ]]; then
  compose up -d db
else
  compose build db --no-cache
  compose up -d db
fi

echo -n "==> Waiting for Postgres (gr33n)"
ready=0
for i in $(seq 1 90); do
  if compose exec -T db pg_isready -U gr33n -d gr33n >/dev/null 2>&1; then
    ready=1
    echo ""
    echo "    Ready after ~$((i * 2))s."
    break
  fi
  echo -n "."
  sleep 2
done

[[ "$ready" -eq 1 ]] || die "Postgres not ready — docker compose logs db"

set -a
# shellcheck disable=1091
source "$ROOT/.env"
set +a

if [[ "$SKIP_REPORT" -eq 0 ]]; then
  echo ""
  ./scripts/db-sanity-report.sh
else
  echo "==> Skipping db sanity report (--skip-report)"
  ./scripts/check-local-stack.sh
fi

echo ""
echo "==> Done."
if [[ "$SERVE" -eq 1 ]]; then
  if [[ "${GUARDIAN_AUTO_TUNE:-}" == "1" ]] && [[ -x "$ROOT/scripts/tune-guardian-laptop.sh" ]]; then
    echo "==> GUARDIAN_AUTO_TUNE=1 — applying laptop Guardian .env recommendations"
    "$ROOT/scripts/tune-guardian-laptop.sh" --apply || true
  fi
  echo "Starting API + UI (make dev-auth-test) — first compile may take several minutes."
  echo ""
  maybe_start_local_ollama
  echo ""
  echo "Guardian tip: after login, open Farm Guardian — awakening preloads models."
  echo "  One-time laptop tune: make guardian-laptop-tune ARGS=\"--apply\""
  maybe_serve_api_ui
fi

echo "Next:"
echo "  ./scripts/check-local-stack.sh   # optional API /health ping"
echo "  make dev-auth-test               # API + UI (AUTH_MODE from Makefile)"
echo "  make dev                         # dev auth bypass"
