# Raspberry Pi OS, edge units, and deployment topology

This guide complements [`pi-integration-guide.md`](pi-integration-guide.md) (Pi → API contract) with **where** each piece runs on real hardware and how deployments **grow** from a homestead to split servers and containers.

---

## 1. Topology ladder (mental model)

| Stage | Typical layout | Notes |
|--------|----------------|--------|
| **Dev laptop** | DB + API + UI loopback; Pi optional | [`INSTALL.md`](../INSTALL.md), [`local-operator-bootstrap.md`](local-operator-bootstrap.md) |
| **Single farm server** | One machine: Postgres + API + built or dev UI | Common “small farm” LAN server (NUC, old PC, NAS VM) |
| **Edge-only Pi** | Pi runs **only** `pi_client`; DB/API/UI elsewhere | Default for greenhouses — Pi is thin and survives power blips |
| **All-on-one Pi** | Postgres + API + UI **on the same Raspberry Pi** | Possible for demos or tiny sites; constrained RAM/SD wear |
| **Growing up** | DB on dedicated host/NAS/container; API and UI together or separate | Match load to iron; UI is often static files + CDN later |
| **Scaled** | DB HA, API replicas, UI static, Pis unchanged at edge | Pi still speaks HTTP to **`PI_API_KEY`** endpoints only |

Your likely path — **computer or server runs DB + API + UI**, Pis stay **edge-only** — matches the middle rows. If the farm grows substantially, moving **Postgres** to its own machine or container is usually the first split; carving out **API** or **UI** containers comes next under load or team process.

---

## 2. Headless Pi — edge daemon only (recommended field layout)

**Role:** GPIO/I2C sensors, actuators, offline queue; HTTP to your API with `X-API-Key`.

### 2.1 OS packages (Pi OS)

From a clone of this repository on the Pi:

```bash
./scripts/install-pi-edge-deps.sh
# make install-pi-edge-deps
```

This installs **git**, **Python 3 + venv + pip**, **libgpiod2**, **i2c-tools** (same baseline as [`pi_client/setup.sh`](../pi_client/setup.sh)). It does **not** install Postgres, Node, or Go.

Then install the client:

```bash
cd pi_client
./setup.sh
```

Edit **`pi_client/config.yaml`**: set `api.base_url` to the **LAN or VPN address** of your API (e.g. `http://192.168.1.50:8080`), and `api.api_key` to match the server’s **`PI_API_KEY`**. The machine running the API must accept TCP from the Pi on that port (firewall rules).

See [`pi-integration-guide.md`](pi-integration-guide.md) for routes, offline queue behavior, and MQTT alternatives.

### 2.2 MQTT bridge on the Pi

If you run [`pi_client/mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py) on the same Pi, install broker/client deps per [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md). Python requirements remain `pi_client/requirements.txt` after `setup.sh`.

---

## 3. Full stack on one Raspberry Pi (DB + API + UI + Pi code)

**When it fits:** demos, offline-first sheds, or a **single** Pi 4/5 with **8GB RAM**, **fast SD or USB SSD**, and acceptance of **heavy** Postgres/Timescale/pgvector on flash.

**Cautions:**

- **Memory:** Postgres + Timescale + API + `npm run dev` for the UI is tight on 4GB; prefer **8GB** and avoid RAG embeddings on-device unless you add swap and accept latency.
- **Storage:** PostgreSQL on SD cards wears NAND; use **SSD** or external drive for `pgdata` if this is more than a toy.
- **UI in dev mode:** `docker compose` here runs Vite dev server — fine for LAN trials. For always-on farms, prefer **built static UI** (`npm run build`) behind **nginx** or Caddy (document your own reverse proxy).

### 3.1 Docker Compose (same repo)

On the Pi (after **`install-pi-edge-deps.sh --with-docker`** or manual Docker install):

```bash
docker compose up -d --build
```

Defaults bind **5432**, **8080**, **5173** on `0.0.0.0`. From **another** laptop on the LAN, open the dashboard at `http://<pi-ip>:5173`.

**Important:** The browser talks to the API using **`VITE_API_URL`**. For Compose as shipped, `ui` uses `http://localhost:8080` inside the UI container — that works only when the browser runs **on the Pi**. From another PC, set (for example) in **`ui/.env`** before build, or adjust Compose env:

- `VITE_API_URL=http://<pi-ip>:8080`

then rebuild/restart the UI service so the SPA calls the API at a URL the **browser** can reach.

Secrets: change **`POSTGRES_PASSWORD`** and **`PI_API_KEY`** / **`JWT_SECRET`** for anything reachable beyond your LAN; use **`AUTH_MODE=production`** when not in a trusted lab.

### 3.2 Native stack on Pi (advanced)

You may run [`install-system-deps-debian.sh`](../scripts/install-system-deps-debian.sh) on Pi OS **Bookworm** (aarch64) **if** PGDG/Timescale publish packages for **arm64** for your codename — mirror the [`INSTALL.md`](../INSTALL.md) bare-metal path. Go **1.25+** may require the official tarball from [go.dev/dl](https://go.dev/dl/) for `linux-arm64`. This path is heavier to maintain than Compose; prefer Compose unless you know you need bare metal.

---

## 4. Growing beyond one box

Rough order operators often follow:

1. **Keep Pis edge-only** — no DB on the field unit; reduce brick risk.
2. **Move Postgres** to a NAS, VM, or managed host **on the LAN or VPC** — backup/restore story improves immediately.
3. **Containerize API** — same image as today’s `Dockerfile`; point **`DATABASE_URL`** at the DB service.
4. **Split UI** — ship static assets to nginx/Object Storage; JWT calls to API origin; same model as any SPA.

Network rules: Pis and dashboards only need routes to **API :443/:8080** (TLS in production); DB port should **not** be exposed to the internet.

---

## 5. Related scripts and docs

| Goal | Script / doc |
|------|----------------|
| Pi OS apt (edge only) | [`scripts/install-pi-edge-deps.sh`](../scripts/install-pi-edge-deps.sh) |
| Debian/Ubuntu dev server (Postgres + Node, not Go) | [`scripts/install-system-deps-debian.sh`](../scripts/install-system-deps-debian.sh), [`INSTALL.md`](../INSTALL.md) |
| First clone / bootstrap | [`scripts/setup-first-clone.sh`](../scripts/setup-first-clone.sh), [`local-operator-bootstrap.md`](local-operator-bootstrap.md) |
| Pi ↔ API protocol | [`pi-integration-guide.md`](pi-integration-guide.md) |
| MQTT | [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) |
| Workflow diagram | [`workflow-guide.md`](workflow-guide.md) §1 |

---

*Operational truth for tables and extensions remains [`db/schema/gr33n-schema-v2-FINAL.sql`](../db/schema/gr33n-schema-v2-FINAL.sql) — informal ERDs may lag; see [`database-schema-overview.md`](database-schema-overview.md).*
