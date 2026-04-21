#!/usr/bin/env bash
# Back-compat wrapper — full logic lives in scripts/dev-stack.sh.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "$ROOT/scripts/dev-stack.sh"
