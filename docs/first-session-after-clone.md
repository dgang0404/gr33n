# First session after clone

**Audience:** New contributor, evaluator, or you on a fresh laptop — get the dashboard running and poke the demo farm **once**, without treating the repo like a black box.

**Not a timed challenge.** Wall-clock varies wildly: Docker image pulls, first `go run` compile (often **several minutes** on a cold machine), `npm ci`, and whether Postgres was already installed. Some people finish in one sitting; others spread steps across two. This doc names **order**, not promises.

**Deeper paths:** [local-operator-bootstrap.md](local-operator-bootstrap.md) · [INSTALL.md](../INSTALL.md) · [operator tour](operator-tour.md) · [Guardian real-grow readiness](guardian-real-grow-readiness.md) (before live plants)

---

## What “done” looks like

You can check all of these:

- [ ] `GET http://localhost:8080/health` → OK
- [ ] UI at **http://localhost:5173** loads
- [ ] Login works (**`admin`** / password you set — see below)
- [ ] **gr33n Demo Farm** (or your farm) shows **zones** and **Live Sensors** with values (not permanent NO DATA)
- [ ] Optional: open Guardian drawer — chat works or shows a clear Lite/config message (not a silent hang)

You do **not** need Ollama, RAG ingest, or a Pi to declare first session success.

---

## Pick a path

| Path | Best when | You need |
|------|-----------|----------|
| **A — Docker DB** (recommended) | Fresh Ubuntu/Debian laptop; want parity with CI | Docker + Compose, Go, Node 22 |
| **B — First-clone script** | Native Postgres already installed with extensions | Timescale + PostGIS + pgvector |
| **C — Manual** | Debugging install | [INSTALL.md](../INSTALL.md) step by step |

Most new contributors should use **Path A**.

---

## Path A — Docker DB + demo seed (happy path)

From repo root after `git clone`:

```bash
# 1) One-time: env + schema + UI deps (add --docker if DB not running yet)
./scripts/setup-first-clone.sh --docker
# or: make first-clone-docker

# 2) Ensure .env points at Compose Postgres (port 5433 in docker-compose.yml)
#    DATABASE_URL=postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable
#    AUTH_MODE=auth_test
#    JWT_SECRET=dev-secret-change-me
#    PI_API_KEY=dev-pi-key

# 3) Seed demo farm + migrations (if not already from setup-first-clone)
./scripts/bootstrap-local.sh --docker --seed

# 4) Dev admin password (login user: admin)
echo -n 'password' | go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash

# 5) API + UI
make dev-auth-test
```

**Open:** http://localhost:5173 → log in → select demo farm → Dashboard / Zones.

**If Docker permission denied:** `sudo usermod -aG docker "$USER"` then `newgrp docker` or log out/in.

**After reboot (same volume):** `make restart-local-serve` — no re-clone, no re-seed.

---

## Path B — Native Postgres

```bash
./scripts/setup-first-clone.sh
# Debian/Ubuntu without Postgres yet:
# ./scripts/setup-first-clone.sh --install-system-deps

./scripts/bootstrap-local.sh --seed
echo -n 'password' | go run scripts/gen-admin-hash.go > ~/.gr33n/admin.hash
make dev-auth-test
```

See [INSTALL.md](../INSTALL.md) if `CREATE EXTENSION` fails (Timescale / PostGIS / pgvector).

---

## First 15 minutes in the UI (suggested tour)

Order matches [operator tour](operator-tour.md) — skim, don’t memorize.

1. **Dashboard** — farm checklist, Today strip if seeded
2. **Zones** → open one zone → **Water / Light / Climate** tabs (plant-needs cockpit)
3. **Alerts** — demo farm may have unread rows after seed
4. **Guardian** (✨ drawer) — ask *“What unread alerts do I have?”* if Full mode is configured; otherwise note Lite banner
5. **Help** workspace — operator glossary; not the same as RAG Knowledge search

**Guardian writes:** anything that changes data shows a **Confirm** card — [change requests guide](guardian-change-requests-guide.md).

---

## Optional same session (only if you have time)

| Step | Command / action | Requires |
|------|------------------|----------|
| RAG demo ingest | `make rag-ingest-demo` | `EMBEDDING_API_KEY` in `.env` |
| Field guides | `make rag-ingest-field-guides` | Same |
| Fresh clean demo | `make dev-stack-fresh-rag` | Wipes Compose DB volume |
| Run tests | `make test` | Go + DB; long on first run |
| Pi edge | [pi-integration-guide.md](pi-integration-guide.md) | Separate hardware |

Skip these until the base dashboard works.

---

## Common blockers

| Symptom | Likely fix |
|---------|------------|
| API exits on start | `LLM_BASE_URL` set but Ollama down → fix URL or set `AI_ENABLED=false` for first session |
| 401 on farm routes | Run admin hash step; check `AUTH_MODE=auth_test` |
| Sensors NO DATA | Seed not loaded → `bootstrap-local.sh --seed` or `make dev-stack-fresh` |
| UI slow after weeks of dev | Junk in DB → [dev-reset-farm](local-operator-bootstrap.md#slow-ui-and-dev-db-hygiene) |
| `go run` “hangs” minutes | First compile — normal; use `go build` once for faster restarts |

More: [operator-troubleshooting.md](operator-troubleshooting.md)

---

## What to read next

| Goal | Doc |
|------|-----|
| Daily dev workflow | [local-operator-bootstrap.md](local-operator-bootstrap.md) |
| Guardian + Ollama | [farm-guardian-ollama-setup.md](farm-guardian-ollama-setup.md) |
| Real grow / live plants | [guardian-real-grow-readiness.md](guardian-real-grow-readiness.md) |
| Phase plans / roadmap | [phase-14-operator-documentation.md](phase-14-operator-documentation.md) |
| Hardware sizing | [recommended-hardware-and-sizing.md](recommended-hardware-and-sizing.md) |

---

## For maintainers evaluating the project

If you are deciding whether to star, fork, or run this on a bench:

- ** AGPL** — inspect everything; network use triggers copyleft obligations on *modified* app code ([LICENSE](../LICENSE)).
- **Confirm gate** — Guardian does not silently drive GPIO or change programs.
- **Phase 82/83** — crop-trust and enterprise bootstrap are **planned**; core ops + Confirm ship today.
- **Stars vs clones** — GitHub shows stargazers; anonymous `git clone` counts are aggregate traffic only, not usernames.

Honest status beats a feature laundry list — that is intentional.
