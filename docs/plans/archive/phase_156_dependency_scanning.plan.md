---
name: Phase 156 — dependency & vulnerability scanning
overview: >
  Dependabot, govulncheck in CI, npm audit in UI CI, and make vuln-check for
  local pre-release scans. Dependency bumps applied so the lane is green on ship.
todos:
  - id: ws1-dependabot
    content: "WS1: .github/dependabot.yml for go.mod and ui/package.json"
    status: completed
  - id: ws2-govulncheck
    content: "WS2: CI + make vuln-check — govulncheck ./..."
    status: completed
  - id: ws3-npm-audit
    content: "WS3: npm audit --audit-level=high in ui CI step"
    status: completed
  - id: ws4-docs
    content: "WS4: SECURITY.md, CONTRIBUTING.md, docs/vuln-allowlist.md"
    status: completed
isProject: false
---

# Phase 156 — dependency & vulnerability scanning

**Status:** shipped · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | [`.github/dependabot.yml`](../../.github/dependabot.yml) — weekly grouped Go + npm |
| **WS2** | `govulncheck` in CI `go` job; [`scripts/vuln-check.sh`](../../scripts/vuln-check.sh); `make vuln-check` |
| **WS3** | `npm audit --audit-level=high` in CI `ui` job |
| **WS4** | [SECURITY.md](../../SECURITY.md), [CONTRIBUTING.md](../../CONTRIBUTING.md), [vuln-allowlist.md](../vuln-allowlist.md) |

## Dependency bumps (ship gate)

- `go` 1.25.7 → **1.26.5** (crypto/tls GO-2026-5856)
- `pgx/v5`, `golang.org/x/net`, `go-jose/v4`, `otel/sdk` — patched to fixed versions
- `ui/package-lock.json` — `npm audit fix --legacy-peer-deps` (0 high vulnerabilities)

## Operator commands

```bash
make vuln-check
```

## Close when

- [x] Dependabot config committed
- [x] CI runs govulncheck + npm audit
- [x] `make vuln-check` exits 0
- [x] `ui/src/__tests__/phase-156-closure.test.js`
