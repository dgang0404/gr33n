#!/usr/bin/env bash
# Phase 156 — local dependency vulnerability scan (Go + UI).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "==> govulncheck (Go)"
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

echo "==> npm audit (ui, high+)"
(cd ui && npm audit --audit-level=high)

echo "==> vuln-check OK"
