# Vulnerability allowlist (Phase 156)

Maintainers document **accepted** `govulncheck` or `npm audit` findings here when:

- upstream has no fixed version yet, **or**
- the vulnerable symbol is not reachable in gr33n's code paths, **or**
- the fix requires a breaking major bump deferred to a dedicated phase

**Do not** allowlist without a GitHub issue link and review date.

## Format

```markdown
### GO-YYYY-NNNN or GHSA-xxxx — short title
- **Package:** module or npm package@version
- **Why accepted:** one sentence
- **Issue:** https://github.com/gr33n-platform/gr33n/issues/N
- **Review by:** YYYY-MM-DD
```

## Active entries

_None — `make vuln-check` is green on `main` as of Phase 156 ship (Go 1.26.5 + dependency bumps)._

## Triage workflow

1. **Dependabot PR** — CI (`go` + `ui` jobs) must pass; merge when green.
2. **`govulncheck` failure** — `go get` fixed module or bump `go` in `go.mod`; else add entry above.
3. **`npm audit --audit-level=high` failure** — `cd ui && npm audit fix --legacy-peer-deps`; if peer conflict, bump the direct dep in `package.json`.
4. **Local check** — `make vuln-check` before release.

See [SECURITY.md](../SECURITY.md) § Dependency vulnerabilities.
