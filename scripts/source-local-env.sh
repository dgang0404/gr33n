#!/usr/bin/env bash
# Load tracked .env then gitignored .env.local into the current shell.
#
# Usage (must source — do not execute in a subshell):
#   cd ~/gr33n-platform
#   source scripts/source-local-env.sh
#   # or:  . scripts/source-local-env.sh
#
# Refresh smoke/eval JWT (login token expires):
#   source scripts/source-local-env.sh --refresh-eval-token
#
# Optional API base for --refresh-eval-token:
#   GR33N_API_URL=http://127.0.0.1:8080 source scripts/source-local-env.sh --refresh-eval-token

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  echo "source scripts/source-local-env.sh   # loads JWT_SECRET, PI_API_KEY, GUARDIAN_EVAL_TOKEN, …" >&2
  echo "Do not run with ./scripts/source-local-env.sh — use source or ." >&2
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

_load_env_file() {
  local f="$1"
  if [[ -f "$f" ]]; then
    set -a
    # shellcheck disable=SC1090
    . "$f"
    set +a
    return 0
  fi
  return 1
}

_refresh_eval_token() {
  local api="${GR33N_API_URL:-http://127.0.0.1:8080}"
  local local_file="$ROOT/.env.local"
  if ! command -v curl >/dev/null 2>&1 || ! command -v python3 >/dev/null 2>&1; then
    echo "source-local-env: curl and python3 required for --refresh-eval-token" >&2
    return 1
  fi
  local token
  token="$(curl -sf -X POST "${api%/}/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{"username":"dev@gr33n.local","password":"devpassword"}' \
    | python3 -c "import sys,json; print(json.load(sys.stdin).get('token',''))")" || true
  if [[ -z "$token" ]]; then
    echo "source-local-env: login failed — is API up at $api?" >&2
    return 1
  fi
  export GUARDIAN_EVAL_TOKEN="$token"
  if [[ -f "$local_file" ]]; then
    python3 - "$local_file" "$token" <<'PY'
import re, sys
from pathlib import Path
path, token = Path(sys.argv[1]), sys.argv[2]
text = path.read_text()
if re.search(r'^GUARDIAN_EVAL_TOKEN=', text, re.M):
    text = re.sub(r'^GUARDIAN_EVAL_TOKEN=.*$', 'GUARDIAN_EVAL_TOKEN=' + token, text, flags=re.M)
else:
    text = text.rstrip() + '\n\n# Refreshed by scripts/source-local-env.sh --refresh-eval-token\nGUARDIAN_EVAL_TOKEN=' + token + '\n'
path.write_text(text if text.endswith('\n') else text + '\n')
PY
    echo "source-local-env: GUARDIAN_EVAL_TOKEN refreshed in .env.local"
  else
    echo "source-local-env: GUARDIAN_EVAL_TOKEN set in shell (.env.local missing — not persisted)"
  fi
}

REFRESH_EVAL=0
for arg in "$@"; do
  case "$arg" in
    --refresh-eval-token) REFRESH_EVAL=1 ;;
    -h|--help)
      sed -n '2,14p' "${BASH_SOURCE[0]}"
      return 0
      ;;
    *)
      echo "source-local-env: unknown option: $arg" >&2
      return 1
      ;;
  esac
done

if _load_env_file "$ROOT/.env"; then
  :
else
  echo "source-local-env: warning — $ROOT/.env not found" >&2
fi

if _load_env_file "$ROOT/.env.local"; then
  :
else
  echo "source-local-env: warning — $ROOT/.env.local not found (copy .env.local.example)" >&2
fi

if [[ "$REFRESH_EVAL" -eq 1 ]]; then
  _refresh_eval_token || return 1
fi

_mask() {
  local name="$1" val="${!1:-}"
  if [[ -n "$val" ]]; then
    echo "  $name=set (${#val} chars)"
  else
    echo "  $name=unset"
  fi
}

echo "source-local-env: loaded from $ROOT"
_mask JWT_SECRET
_mask PI_API_KEY
_mask GUARDIAN_EVAL_TOKEN
