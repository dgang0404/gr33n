#!/usr/bin/env bash
# Port-aware dev stack bring-up: start only what's missing, or leave a healthy stack alone.
# Used by make dev-auth-test and restart-local.sh --serve.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

die() {
  echo "error: $*" >&2
  exit 1
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

  cd "$ROOT"
  exec make dev-auth-test-run
}

[[ -f "$ROOT/.env" ]] || die "missing .env — copy .env.example"
maybe_serve_api_ui
