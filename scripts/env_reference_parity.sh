#!/usr/bin/env bash
# Phase 116 WS1 — every os.Getenv / os.LookupEnv in cmd/ + internal/ must appear
# in docs/environment-variables.md (or .env.example). Test-only vars are allow-listed.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

REF="$ROOT/docs/environment-variables.md"
EXAMPLE="$ROOT/.env.example"

if [[ ! -f "$REF" ]]; then
  echo "missing $REF" >&2
  exit 1
fi

ALLOW=(
  CI
  GITHUB_ACTIONS
  TEST_DATABASE_URL
  GR33N_HARDWARE_TEST
  GR33N_VISION_TEST
  HOME
  USER
  GR33N_REPO_ROOT
)

is_allowed() {
  local v="$1"
  for a in "${ALLOW[@]}"; do
    [[ "$v" == "$a" ]] && return 0
  done
  return 1
}

doc_vars() {
  # Backtick-wrapped names in environment-variables.md tables
  sed -n 's/.*`\([A-Z][A-Z0-9_]*\)`.*/\1/p' "$REF" | sort -u
  rg --no-filename -o '^#?\s*([A-Z][A-Z0-9_]+)=' "$EXAMPLE" -r '$1' | sort -u
}

code_vars() {
  rg --no-filename -o 'os\.(Getenv|LookupEnv)\("([^"]+)"\)' cmd internal -r '$2' | sort -u
  rg --no-filename -o 'getEnv\("([^"]+)"' cmd -r '$1' | sort -u
  rg --no-filename -o 'getenv\("([^"]+)"\)' internal/filestorage/config.go -r '$1' | sort -u
}

DOC_LIST="$(mktemp)"
code_vars | sort -u > "${DOC_LIST}.code"
doc_vars | sort -u > "$DOC_LIST"

missing=()
while IFS= read -r v; do
  [[ -z "$v" ]] && continue
  if is_allowed "$v"; then
    continue
  fi
  if ! grep -qx "$v" "$DOC_LIST"; then
    missing+=("$v")
  fi
done < "${DOC_LIST}.code"

rm -f "$DOC_LIST" "${DOC_LIST}.code"

if ((${#missing[@]})); then
  echo "Environment variables missing from docs/environment-variables.md or .env.example:" >&2
  printf '  - %s\n' "${missing[@]}" >&2
  exit 1
fi

echo "env reference parity OK"
