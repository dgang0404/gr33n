#!/usr/bin/env bash
# Port-aware dev stack bring-up: start only what's missing, or leave a healthy stack alone.
# Used by make dev-auth-test, make laptop-up (restart-local.sh --serve).
#
# When :8080/health and :5173 already respond, compares .gr33n/dev-serve-stamp to
# `git describe --always --dirty` and restarts both if the repo changed (git pull,
# new commit, or dirty tree). Set GR33N_FORCE_DEV_RESTART=1 to restart anyway.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STAMP_FILE="$ROOT/.gr33n/dev-serve-stamp"

die() {
  echo "error: $*" >&2
  exit 1
}

dev_code_stamp() {
  if command -v git >/dev/null 2>&1 && git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$ROOT" describe --always --dirty --broken 2>/dev/null || echo "unknown"
  else
    echo "unknown"
  fi
}

read_dev_stamp() {
  if [[ -f "$STAMP_FILE" ]]; then
    tr -d '[:space:]' < "$STAMP_FILE"
  else
    echo ""
  fi
}

write_dev_stamp() {
  mkdir -p "$ROOT/.gr33n"
  dev_code_stamp > "$STAMP_FILE"
}

stamp_stale() {
  local current stored
  current=$(dev_code_stamp)
  stored=$(read_dev_stamp)
  [[ -z "$stored" || "$current" != "$stored" ]]
}

# Kill listeners on a TCP port (best-effort; ss shows the process holding the socket).
kill_listener_on_port() {
  local port=$1
  local pid
  if ! command -v ss >/dev/null 2>&1; then
    return 0
  fi
  while read -r pid; do
    [[ -z "$pid" || ! "$pid" =~ ^[0-9]+$ ]] && continue
    echo "    stopping pid $pid (port :$port)"
    kill "$pid" 2>/dev/null || true
  done < <(ss -tlnp 2>/dev/null | grep -E ":${port}\\b" | sed -n 's/.*pid=\([0-9]*\).*/\1/p' | sort -u)
  sleep 1
}

maybe_restart_for_new_code() {
  local port=$1
  local api_ok=$2
  local ui_ok=$3
  local force=${GR33N_FORCE_DEV_RESTART:-}

  if [[ "$api_ok" -eq 0 && "$ui_ok" -eq 0 ]]; then
    return 1
  fi
  if [[ "$force" != "1" ]] && ! stamp_stale; then
    return 1
  fi

  local stored current
  stored=$(read_dev_stamp)
  current=$(dev_code_stamp)
  if [[ "$force" == "1" ]]; then
    echo "==> GR33N_FORCE_DEV_RESTART=1 — restarting dev API + UI"
  else
    echo "==> Code changed (${stored:-none} -> ${current}) — restarting dev API + UI"
  fi
  kill_listener_on_port "$port"
  kill_listener_on_port 5173
  return 0
}

# Avoid a second API/UI dev stack when ports are already serving gr33n.
maybe_serve_api_ui() {
  set -a
  # shellcheck disable=SC1091
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

  if maybe_restart_for_new_code "$port" "$api_ok" "$ui_ok"; then
    api_ok=0
    ui_ok=0
  fi

  if [[ "$api_ok" -eq 1 && "$ui_ok" -eq 1 ]]; then
    echo "==> API (:${port}) and UI (:5173) already running (code stamp: $(read_dev_stamp))."
    echo "    Open http://localhost:5173/"
    echo "    Clean DB slate: make laptop-up-fresh   (or: make dev-stack-fresh && make laptop-up)"
    echo "    Force restart (same code): GR33N_FORCE_DEV_RESTART=1 make laptop-up"
    return 0
  fi

  if [[ "$api_ok" -eq 1 && "$ui_ok" -eq 0 ]]; then
    echo "==> API (:${port}) already running — starting UI only (:5173)."
    cd "$ROOT/ui"
    exec npm run dev
  fi

  if [[ "$api_ok" -eq 0 && "$ui_ok" -eq 1 ]]; then
    echo "==> UI (:5173) already running — starting API only (:${port}, AUTH_MODE=${GR33N_DEV_AUTH_MODE:-auth_test})."
    write_dev_stamp
    cd "$ROOT"
    export AUTH_MODE="${GR33N_DEV_AUTH_MODE:-auth_test}"
    exec make dev-auth-test-run
  fi

  if command -v ss >/dev/null 2>&1; then
    if ss -tln 2>/dev/null | grep -qE ":${port}\\b"; then
      die "port :${port} is in use but /health did not respond — free the port or fix the other process"
    fi
    if ss -tln 2>/dev/null | grep -qE ':5173\b'; then
      die "port :5173 is in use but UI did not respond — free the port or fix the other process"
    fi
  fi

  write_dev_stamp
  cd "$ROOT"
  export AUTH_MODE="${GR33N_DEV_AUTH_MODE:-auth_test}"
  exec make dev-auth-test-run
}

[[ -f "$ROOT/.env" ]] || die "missing .env — copy .env.example"
maybe_serve_api_ui
