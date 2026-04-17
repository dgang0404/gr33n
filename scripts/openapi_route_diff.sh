#!/usr/bin/env bash
# ============================================================
# Phase 20.95 WS6 — openapi ↔ routes.go diff guard.
#
# Compares the set of (METHOD PATH) pairs declared in
# cmd/api/routes.go against the operations listed in openapi.yaml.
# Prints a unified diff and exits non-zero if they disagree.
#
# Exit codes:
#   0 — routes.go and openapi.yaml are 1:1.
#   1 — disagreement (diff printed to stdout).
#   2 — extractor returned 0 routes (paranoia guard against a
#       future regex break silently reporting green against an
#       equally-empty openapi).
#
# Router note: cmd/api uses Go 1.22+ http.ServeMux method+path
# patterns (`mux.Handle("GET /x", …)` and
# `mux.HandleFunc("GET /x", …)`), NOT chi. The regex below is
# pinned to that shape; if we ever switch routers, update it.
# ============================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROUTES_FILE="$REPO_ROOT/cmd/api/routes.go"
OPENAPI_FILE="$REPO_ROOT/openapi.yaml"

ROUTES_TMP=$(mktemp)
OPENAPI_TMP=$(mktemp)
trap 'rm -f "$ROUTES_TMP" "$OPENAPI_TMP"' EXIT

# 1. Extract METHOD path pairs from routes.go.
#    Matches:  mux.Handle("GET /foo/{id}", …)
#              mux.HandleFunc("POST /bar", …)
python3 - "$ROUTES_FILE" <<'PY' | sort -u > "$ROUTES_TMP"
import re
import sys

pattern = re.compile(
    r'mux\.(?:Handle|HandleFunc)\("(GET|POST|PUT|PATCH|DELETE) ([^"]+)"'
)
with open(sys.argv[1]) as f:
    for m in pattern.finditer(f.read()):
        print(f"{m.group(1)} {m.group(2)}")
PY

ROUTE_COUNT=$(wc -l < "$ROUTES_TMP" | tr -d ' ')
if [ "$ROUTE_COUNT" -eq 0 ]; then
  echo "ERROR: extractor found zero routes in $ROUTES_FILE." >&2
  echo "       Regex probably drifted from the mux.Handle/HandleFunc pattern." >&2
  exit 2
fi

# 2. Extract METHOD path pairs from openapi.yaml.
python3 - "$OPENAPI_FILE" <<'PY' | sort -u > "$OPENAPI_TMP"
import sys
import yaml

with open(sys.argv[1]) as f:
    spec = yaml.safe_load(f)

for path, ops in (spec.get("paths") or {}).items():
    if not isinstance(ops, dict):
        continue
    for method in ("get", "post", "put", "patch", "delete"):
        if method in ops:
            print(f"{method.upper()} {path}")
PY

# 3. Diff.
if ! diff -u "$OPENAPI_TMP" "$ROUTES_TMP" > /tmp/openapi_route_diff.out; then
    echo "openapi.yaml (-)  vs  cmd/api/routes.go (+)"
    cat /tmp/openapi_route_diff.out
    exit 1
fi

echo "audit-openapi: $ROUTE_COUNT routes, in sync."
