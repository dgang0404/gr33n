# gr33n-api — Local Development Setup

## Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.23+ | https://go.dev/dl/ or `snap install go --classic` |
| PostgreSQL | 14+ | `sudo apt install postgresql` |
| TimescaleDB | 2.x | https://docs.timescale.com/self-hosted/latest/install/ |
| sqlc | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| golang-migrate | latest | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |

---

## 1. Clone the repo

```bash
git clone https://github.com/YOUR_ORG/gr33n-api.git
cd gr33n-api
```

---

## 2. PostgreSQL setup

### 2a. Create the database

```bash
sudo -u postgres psql -c "CREATE DATABASE gr33n;"
```

### 2b. Enable TimescaleDB on the database

```bash
sudo -u postgres psql -d gr33n -c "CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;"
```

### 2c. Create a local dev user matching your Linux username

PostgreSQL on Linux uses **peer authentication** by default — the connecting
OS user must match a PostgreSQL role of the same name.

```bash
sudo -u postgres psql -c "CREATE USER $USER WITH SUPERUSER;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gr33n TO $USER;"
```

Verify it works (no password, no sudo needed):

```bash
psql -d gr33n -c "SELECT current_user, current_database();"
# Expected:  current_user | current_database
#            davidg       | gr33n
```

---

## 3. Run migrations

```bash
migrate -path ./migrations -database "postgres://$USER@/gr33n?host=/var/run/postgresql" up
```

---

## 4. Environment variables

The API reads one required env var at startup:

| Variable | Description | Default (dev) |
|----------|-------------|---------------|
| `DATABASE_URL` | PostgreSQL connection string | see below |
| `PORT` | HTTP listen port | `8080` |

For local development with peer auth (no password):

```bash
export DATABASE_URL="postgres://$USER@/gr33n?host=/var/run/postgresql"
```

Add this to your `~/.bashrc` or `~/.zshrc` to avoid typing it every time.

---

## 5. Build and run

```bash
go mod tidy
go run ./cmd/api/
```

Expected output:

```
2026/02/26 16:41:55 ✅ Connected to gr33n database
2026/02/26 16:41:55 🌱 gr33n API running on http://localhost:8080
```

---

## 6. Smoke test

```bash
# Health check
curl http://localhost:8080/health
# → {"service":"gr33n-api","status":"ok"}

# All units of measure
curl http://localhost:8080/units

# Units filtered by type
curl "http://localhost:8080/units?type=temperature"

# Devices
curl http://localhost:8080/devices
```

---

## 7. Code generation (sqlc)

If you modify any `.sql` query files under `internal/db/`, regenerate the
Go query layer:

```bash
sqlc generate
```

Generated files live in `internal/db/` — do **not** edit them by hand.

---

## Common issues

### `could not connect to database after 5 attempts`

The error message will now print the real cause on each attempt.
Most common root causes:

- **Peer auth mismatch** — your Linux username has no matching PostgreSQL role.
  Fix: run step 2c above.
- **Socket path wrong** — make sure `?host=/var/run/postgresql` is in the URL
  (not `localhost:5432`, which forces TCP and fails peer auth).
- **PostgreSQL not running** — `sudo systemctl start postgresql`

### `package gr33n-api/internal/platform/commontypes is not in std`

The `enums.go` file is missing. Copy it into place:

```bash
mkdir -p ~/gr33n-api/internal/platform/commontypes
cp ~/Downloads/enums.go ~/gr33n-api/internal/platform/commontypes/
go mod tidy
```

### `could not change directory … Permission denied` (sudo -u postgres)

Harmless warning — postgres can't `cd` into your home dir when you run `sudo`
from inside it. The command itself still executes correctly.

---

## Repository layout

```
gr33n-api/
├── cmd/
│   └── api/
│       ├── main.go          # Entry point, DB connection, server startup
│       └── routes.go        # HTTP route registration
├── internal/
│   ├── db/                  # sqlc-generated query layer (do not edit)
│   ├── handlers/            # HTTP handler functions
│   └── platform/
│       └── commontypes/
│           └── enums.go     # Shared enum types used by sqlc
├── migrations/              # SQL migration files (golang-migrate)
├── schema/                  # sqlc schema + query source files
├── sqlc.yaml
├── go.mod
└── go.sum
```
