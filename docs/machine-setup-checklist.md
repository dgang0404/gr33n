# Machine setup checklist (multi-device hardening)

Use this when onboarding a **new laptop, VM, or Pi dev box** so each environment matches expectations without relying on one machine’s memory.

## Repository

- [ ] `git clone` … && `cd gr33n-platform`
- [ ] `git pull` on main (or your branch) before starting work

## Tooling

- [ ] **Go** 1.25+ (`go version`)
- [ ] **Node** 22+ (`node -v`) — UI / `npm ci`
- [ ] **psql** client — connects to Postgres from host
- [ ] **Docker** + Compose v2 — if using Compose DB (`docker compose version`)
- [ ] **`sg docker`** or user in **`docker`** group — no permission errors on `/var/run/docker.sock`

## Configuration files (never commit secrets)

- [ ] **`cp .env.example .env`** at repo root
- [ ] **`cp ui/.env.example ui/.env`** if missing
- [ ] **`.env`**: `DATABASE_URL` matches how you run Postgres:
  - Compose: `postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable`
  - Native peer: see [INSTALL.md](../INSTALL.md)
- [ ] **`AUTH_MODE`**, **`JWT_SECRET`**, **`PI_API_KEY`** set for `auth_test` / production paths
- [ ] **`CORS_ORIGIN`** matches browser origin (e.g. `http://localhost:5173` or `5174` if Vite chose another port)
- [ ] **`ADMIN_BIND_USER_ID` / `ADMIN_BIND_EMAIL`** if logging in as **`admin`** with `~/.gr33n/admin.hash`
- [ ] Optional: **`~/.gr33n/admin.hash`** for env-admin login

See also: **[example-env.md](example-env.md)** (readable copy of `.env.example`).

## Database

- [ ] Start DB: `sg docker -c 'docker compose up -d db'` (or native Postgres running)
- [ ] Wait for ready: `pg_isready` / bootstrap wait loop
- [ ] First-time or clean volume: `./scripts/bootstrap-local.sh --seed`
- [ ] Existing schema, migrations only: `./scripts/bootstrap-local.sh --skip-schema --seed`

## Run stack

- [ ] From repo root: **`make dev-auth-test`** (or `make dev` for AUTH_MODE=dev)
- [ ] **`curl -s http://127.0.0.1:8080/health`** → JSON `status":"ok"`
- [ ] Browser: UI URL from Vite (often `:5173`), **API online** in header
- [ ] Login: seeded **`dev@gr33n.local`** / **`devpassword`** or **`admin`** + hash

## Verification queries (optional)

```bash
psql "$DATABASE_URL" -c "SELECT count(*) FROM gr33ncore.farms;"
psql "$DATABASE_URL" -c "SELECT count(*) FROM gr33nnaturalfarming.application_recipes WHERE deleted_at IS NULL;"
```

## Second machine or browser profile

- [ ] **New device** — repeat the **Configuration files** and **Database** / **Run stack** sections; do not copy another machine’s **`.env`** wholesale if `DATABASE_URL` or ports differ.
- [ ] **Vite port** — if the UI is not on `5173`, set **`CORS_ORIGIN`** to the exact origin the browser uses (e.g. `http://localhost:5174`) and `ui/.env` **`VITE_API_URL`** to the API.
- [ ] **Offline queue** — [tasks-first-operator-guide.md](tasks-first-operator-guide.md) §3: each profile has its own **`gr33n_offline_write_queue_v2`**; queued writes on laptop A are invisible to laptop B until synced to the server.

## Troubleshooting pointers

- [INSTALL.md](../INSTALL.md) — native Postgres, extensions
- [local-operator-bootstrap.md](local-operator-bootstrap.md) — full operator path
- [ARCHITECTURE.md](../ARCHITECTURE.md) — dashboard mental model
