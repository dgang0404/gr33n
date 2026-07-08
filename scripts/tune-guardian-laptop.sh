#!/usr/bin/env bash
# Phase 129 WS3 — idempotent Guardian laptop .env recommendations (CPU 16GB profile).
# Usage (repo root):
#   ./scripts/tune-guardian-laptop.sh              # print suggestions
#   ./scripts/tune-guardian-laptop.sh --apply      # merge into .env
#   ./scripts/tune-guardian-laptop.sh --profile gpu-server
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/.env"
APPLY=0
PROFILE="${GUARDIAN_TUNE_PROFILE:-cpu-16gb}"

usage() {
  cat <<'EOF'
Usage: scripts/tune-guardian-laptop.sh [--apply] [--profile cpu-16gb|gpu-server]

  cpu-16gb (default)  LLM_TIMEOUT_SECONDS>=1500, GUARDIAN_GROUNDED_TIMEOUT_SECONDS>=1800
  gpu-server          LLM_TIMEOUT_SECONDS>=666, LLM_RETRY_MAX_ATTEMPTS=1

Without --apply, prints recommended changes only.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --apply) APPLY=1 ;;
    --profile) PROFILE="${2:?}"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
  shift
done

[[ -f "$ENV_FILE" ]] || { echo "error: missing .env — copy .env.example" >&2; exit 1; }

read_env_val() {
  local key="$1"
  grep -E "^[[:space:]]*${key}=" "$ENV_FILE" 2>/dev/null | tail -1 | sed "s/^[[:space:]]*${key}=//" | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/" || true
}

MIN_LLM_TIMEOUT=1500
MIN_GROUNDED_TIMEOUT=1800
if [[ "$PROFILE" == "gpu-server" ]]; then
  MIN_LLM_TIMEOUT=666
  MIN_GROUNDED_TIMEOUT=900
fi

declare -a CHANGES=()

ensure_min() {
  local key="$1" min="$2" cur
  cur="$(read_env_val "$key")"
  if [[ -z "$cur" ]]; then
    CHANGES+=("${key}=${min}")
    return
  fi
  if [[ "$cur" =~ ^[0-9]+$ ]] && (( cur < min )); then
    CHANGES+=("${key}=${min}  # was ${cur}")
  fi
}

ensure_exact() {
  local key="$1" want="$2" cur
  cur="$(read_env_val "$key")"
  if [[ -z "$cur" ]]; then
    CHANGES+=("${key}=${want}")
  elif [[ "$cur" != "$want" ]]; then
    CHANGES+=("${key}=${want}  # was ${cur}")
  fi
}

ensure_min LLM_TIMEOUT_SECONDS "$MIN_LLM_TIMEOUT"
ensure_min GUARDIAN_GROUNDED_TIMEOUT_SECONDS "$MIN_GROUNDED_TIMEOUT"
ensure_exact LLM_RETRY_MAX_ATTEMPTS 1
if [[ "$PROFILE" == "cpu-16gb" ]]; then
  ensure_min GUARDIAN_EVAL_WARMUP_TIMEOUT 90
fi

if [[ ${#CHANGES[@]} -eq 0 ]]; then
  echo "ok  Guardian .env matches profile ${PROFILE} — no changes needed"
  exit 0
fi

echo "Guardian tune (profile=${PROFILE}):"
for line in "${CHANGES[@]}"; do
  echo "  $line"
done

if [[ "$APPLY" -ne 1 ]]; then
  echo ""
  echo "Run with --apply to write these keys to .env (existing secrets untouched)."
  exit 0
fi

for line in "${CHANGES[@]}"; do
  key="${line%%=*}"
  val="${line#*=}"
  val="${val%% #*}"
  val="$(echo "$val" | sed 's/^[[:space:]]*//')"
  if grep -qE "^[[:space:]]*${key}=" "$ENV_FILE"; then
    sed -i "s|^[[:space:]]*${key}=.*|${key}=${val}|" "$ENV_FILE"
  else
    echo "${key}=${val}" >> "$ENV_FILE"
  fi
  echo "applied ${key}=${val}"
done

echo "done — restart API (make restart-local-serve or make dev-auth-test) to pick up .env"
