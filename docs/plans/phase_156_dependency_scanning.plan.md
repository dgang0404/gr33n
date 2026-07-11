---
name: Phase 156 — dependency & vulnerability scanning
overview: >
  Add Dependabot, govulncheck, and npm audit so operators and maintainers get
  early notice of CVEs in Go modules and UI packages. Proposed explicitly
  (unlike the reverted Guardian PR CI gate) with clear blocking vs advisory lanes.
todos:
  - id: ws1-dependabot
    content: "WS1: .github/dependabot.yml for go.mod and ui/package.json (weekly, grouped minor/patch)"
    status: pending
  - id: ws2-govulncheck
    content: "WS2: CI job or make target — govulncheck ./... with documented allowlist for false positives"
    status: pending
  - id: ws3-npm-audit
    content: "WS3: npm audit --audit-level=high in ui/ CI step; fail or warn per policy below"
    status: pending
  - id: ws4-docs
    content: "WS4: SECURITY.md + CONTRIBUTING.md — how to triage vuln alerts, when to bump vs suppress"
    status: pending
isProject: false
---

# Phase 156 — dependency & vulnerability scanning

**Status:** planned · **Hub:** [154–158 backlog](phase_154_158_infra_trust_gaps_backlog.plan.md)

---

## Why this phase

The repo today has:

- **No** `.github/dependabot.yml`
- **No** `govulncheck` in CI or Makefile
- **No** `npm audit` in the UI CI lane

gr33n handles auth (JWT), file uploads (receipt photos), device API keys, and Guardian change-request execution. A known CVE in a transitive dependency should surface in CI or weekly Dependabot PRs, not on a random `git pull` months later.

**Note on the reverted GitHub PR gate (Phase 153 lesson):** label-gated slow smoke CI is standard practice — the issue was consent/scope, not engineering quality. Phase 156 is **proposed here explicitly** so you can greenlight blocking vs advisory behavior before anything lands.

---

## Workstreams

### WS1 — Dependabot

**Target:** `.github/dependabot.yml`

```yaml
# Sketch — implementer fills versions/ecosystem blocks
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule: { interval: weekly }
    groups:
      go-minor-patch: { patterns: ["*"], update-types: [minor, patch] }
  - package-ecosystem: npm
    directory: /ui
    schedule: { interval: weekly }
```

- Group minor/patch Go bumps to avoid PR spam
- Major bumps stay separate PRs with changelog note in body template
- **Policy:** Dependabot PRs run existing CI (`go` + `ui` jobs) — no new Guardian/Ollama lane

### WS2 — govulncheck

**Options (pick one at implement time):**

| Lane | Behavior |
|------|----------|
| **A — blocking on `main`** | New CI step after `go test`; fails on known vulns in reachable code |
| **B — advisory** | `make vuln-check` locally; CI posts warning annotation only |

**Recommended default:** **A on `main` push + PR**, with a documented `docs/vuln-allowlist.md` (or inline `//govulncheck:ignore` only when upstream has no fix).

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### WS3 — npm audit

Add to existing UI CI job (after `npm ci`):

```bash
cd ui && npm audit --audit-level=high
```

- `moderate` → advisory only (log, don't fail) unless you prefer stricter
- Document override: `npm audit fix` vs manual bump when fix breaks Vue/Vite

### WS4 — Documentation

Update [`SECURITY.md`](../../SECURITY.md):

- How to report app-level security issues (already there — add dependency section)
- Maintainer triage: Dependabot PR → CI green → merge; govulncheck failure → bump or allowlist with issue link
- Link from [`CONTRIBUTING.md`](../../CONTRIBUTING.md)

---

## Acceptance

- [ ] Dependabot opens grouped weekly PRs for Go + npm (verify on fork or dry-run config)
- [ ] `govulncheck ./...` runs in CI or via `make vuln-check`
- [ ] `npm audit --audit-level=high` runs in UI CI lane
- [ ] SECURITY.md documents triage + allowlist process
- [ ] No Guardian/Ollama/model-dependent steps added to default PR CI (scope stays dependency-only)

## Non-goals

- Snyk/Trivy container scanning (Docker image not primary deploy path for most operators)
- License compliance scanning (AGPL is project license; deps are mostly standard OSS)
- Auto-merge Dependabot without human review

## Operator / maintainer commands (after ship)

```bash
make vuln-check          # local: govulncheck + npm audit summary
# Dependabot PRs — review in GitHub UI, merge when CI green
```
