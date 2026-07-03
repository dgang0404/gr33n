# Upgrade guide

How to update an existing gr33n install after `git pull`. For first-time setup see [INSTALL.md](../INSTALL.md) and [local-operator-bootstrap.md](local-operator-bootstrap.md).

---

## Standard upgrade (most releases)

From the repo root:

```bash
git pull
make migrate          # apply new db/migrations/*.sql
# restart API + UI
make dev-auth-test    # or your production process manager
```

**Order matters:** migrate **before** starting the new API binary so handlers match the schema.

After pull, **always restart the API** — new routes and worker behavior only register on process start.

---

## Checklist

| Step | Command / action |
|------|------------------|
| 1. Backup (production) | [backup-restore-runbook.md](backup-restore-runbook.md) — DB + file blobs together |
| 2. Pull code | `git pull` |
| 3. UI deps (if package.json changed) | `npm --prefix ui ci --legacy-peer-deps` |
| 4. Go deps (if go.mod changed) | `go mod download` |
| 5. Database | `make migrate` |
| 6. Optional seed deltas | only when release notes say so — never re-run full seed on production blindly |
| 7. RAG re-index (when docs/catalog changed) | `make rag-ingest-platform-docs` or farm-scoped ingest |
| 8. Restart | API, UI, automation worker (same process as API), Pi clients if edge protocol changed |
| 9. Smoke | `curl -s localhost:8080/health` · login · one farm dashboard load |

---

## Version-specific runbooks

| Topic | Doc |
|-------|-----|
| Receipt / S3 storage cutover | [receipt-storage-cutover-runbook.md](receipt-storage-cutover-runbook.md) (full detail) |
| Backup & restore | [backup-restore-runbook.md](backup-restore-runbook.md) |
| Pi client update | [pi-integration-guide.md](pi-integration-guide.md) |
| Ollama / Guardian models | [INSTALL.md](../INSTALL.md) § Ollama · Settings → Guardian usage |
| Env var changes | [environment-variables.md](environment-variables.md) |

---

## Docker Compose installs

```bash
git pull
docker compose build    # when Dockerfile or dependencies changed
make migrate            # against compose DATABASE_URL from .env
docker compose up -d
```

Ensure `.env` `DATABASE_URL` matches the compose `db` service (see [.env.example](../.env.example)).

---

## Lite vs Guardian installs

- **`AI_ENABLED=false`** — no `/v1/chat`; upgrades skip LLM/Ollama steps.
- **`AI_ENABLED=true`** — after upgrade, confirm `LLM_BASE_URL` / `LLM_MODEL` and run **Settings → Guardian** or `GET /v1/chat/health`.

---

## Rollback

1. Stop API (prevent writes).
2. Restore DB from backup taken before upgrade.
3. Restore file blobs if receipts/storage changed.
4. `git checkout <previous-tag-or-commit>` and rebuild.
5. Start API on the restored DB.

Do not run older migrations against a newer schema — restore the whole DB snapshot instead.

---

## CI parity (contributors)

Before merge: `make test`, `make audit-openapi`, `make audit-env` — see [CONTRIBUTING.md](../CONTRIBUTING.md).
