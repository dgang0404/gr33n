# Contributing to gr33n

Thank you for helping improve an open-source farm OS. This guide covers how we work in this repo — not generic GitHub etiquette.

---

## Before you open a PR

1. **Discuss large changes** — new schemas, breaking API changes, or multi-phase features should start as a plan in `docs/plans/` or an issue so operators know what's coming.
2. **Run local gates** (from repo root):

```bash
make test              # Go tests + cmd/api smokes (needs DATABASE_URL + migrated DB)
make test-unit         # fast unit lane (no cmd/api DB smokes)
make vuln-check        # govulncheck + npm audit --audit-level=high (Phase 156)
make lint              # go vet
make audit-openapi     # routes.go ↔ openapi.yaml
make audit-env         # os.Getenv ↔ docs/environment-variables.md
npm --prefix ui run build
npm --prefix ui test -- --run
make check-ui-test-baseline  # Phase 205 — fails only on failures NOT already tracked in ui/test-baseline-known-failures.json
```

**Refactors and "janitorial" phases especially:** run the *full* UI suite (`npm --prefix ui test -- --run`), not just the new phase's own closure test, before marking a phase shipped. `make check-ui-test-baseline` is what actually enforces this — it fails if your change adds a failure that isn't already in the known-debt list, so collateral breakage in an unrelated component gets caught immediately instead of silently joining the pile. See `docs/plans/phase_205_pre_existing_test_debt.plan.md` for the debt this is paying down and why the baseline exists instead of just fixing everything up front.

3. **Migrations** — SQL under `db/migrations/`; never edit applied migration files in place. Regenerate sqlc with `make sqlc` when queries change.
4. **OpenAPI** — every new `mux.Handle` in `cmd/api/routes.go` needs a matching path in `openapi.yaml` (or an entry in `routesIntentionallyUndocumented` with a comment). Sync embed copy: `cp openapi.yaml internal/openapiui/openapi.yaml` when the spec changes.

---

## Plan lifecycle

| Stage | Where |
|-------|--------|
| Planned | `docs/plans/phase_NNN_*.plan.md` with `status: pending` todos |
| In progress | Same file; todos updated in PRs |
| Shipped | Plan header **Shipped** + todos completed; index in `docs/phase-14-operator-documentation.md` |

Phases are numbered sequentially (113 security, 114 Pi, 115 schema, 116 docs, …). Keep each PR focused on one phase or workstream when possible.

---

## Code conventions

- **Go** — follow existing handler patterns (`internal/handler/*`), farm auth via `internal/farmauthz`, JSON via `internal/httputil`.
- **Vue** — Pinia stores, workspace routes in `ui/src/lib/navGroups.js`, Tailwind utility classes matching surrounding components.
- **Comments** — only for non-obvious business logic; phase numbers in commit messages are fine but not required in code.
- **Scope** — smallest correct diff; don't refactor unrelated files in the same PR.

---

## Auth & dev builds

- `make dev` / `make run` — `AUTH_MODE=dev`, `-tags dev`, auth bypass for local UI work.
- `make dev-auth-test` — real JWT + Pi key enforcement against local `.env` secrets.
- Never ship production images with `-tags dev` or `AUTH_MODE=dev`.

---

## Documentation

- Operator docs: `docs/` (markdown, no site generator).
- Env vars: [docs/environment-variables.md](docs/environment-variables.md) + [`.env.example`](.env.example).
- User-facing changes: add a bullet to [CHANGELOG.md](CHANGELOG.md) under **Unreleased**.

---

## Security

See [SECURITY.md](SECURITY.md). Do not commit secrets, real `.env` files, or production credentials.

**Dependency vulnerabilities (Phase 156):** run `make vuln-check` before release; triage Dependabot PRs when CI is green. Document accepted findings in [docs/vuln-allowlist.md](docs/vuln-allowlist.md).

---

## License

Contributions are accepted under the same [AGPL v3](LICENSE) as the project.
