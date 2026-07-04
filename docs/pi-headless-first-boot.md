# Pi headless first boot (simulation rig)

One-time setup on a **Raspberry Pi OS Lite** headless unit so `gr33n_client.py`
posts sensor readings to your **laptop or server API** and drives the Phase 125
LED simulation rig.

**What runs where**

| Machine | Runs |
|---------|------|
| **Laptop / server** | Postgres, API (`:8080`), UI (`:5173`), Ollama, `make guardian-bootstrap-farm` |
| **Pi (headless)** | Only `pi_client/gr33n_client.py` — no `make`, no Ollama, no Postgres |

**See also:** [`pi-light-simulation-runbook.md`](pi-light-simulation-runbook.md) (demos after setup),
[`pi-integration-guide.md`](pi-integration-guide.md) (full edge contract),
[`../pi_client/README-simulation-rig.md`](../pi_client/README-simulation-rig.md) (wiring + parts).

---

## 1. Laptop first (every dev session)

From the repo root on the machine running the API:

```bash
cd ~/gr33n-platform

make compose-db-up          # if Docker Postgres is not up
make dev-auth-test          # API + UI — leave this terminal running
```

Optional — full Guardian RAG corpus (once per seed / doc change; **laptop only**):

```bash
cd ~/gr33n-platform
make guardian-bootstrap-farm FARM_ID=1 ARGS="--smoke"
```

Requires Ollama already running (`EMBEDDING_*` in `.env`). If `ollama serve` says
*address already in use*, Ollama is fine — skip that command.

Get sensor IDs for the Pi config (**laptop only** — needs DB access):

```bash
cd ~/gr33n-platform
./scripts/print-demo-sensor-ids.sh
```

Note your laptop's LAN IP (Pi will use this instead of `localhost`):

```bash
hostname -I | awk '{print $1}'
```

---

## 2. Pi one-time setup (SSH into the headless unit)

### 2.1 System packages

```bash
sudo apt update
sudo apt install -y python3-pip python3-venv git
```

Or, after cloning the repo on the Pi:

```bash
cd /opt/gr33n/gr33n-platform
./scripts/install-pi-edge-deps.sh
```

(`install-pi-edge-deps.sh` also installs GPIO helpers — recommended on real hardware.)

### 2.2 Get `pi_client` (once)

```bash
sudo mkdir -p /opt/gr33n
sudo chown "$USER":"$USER" /opt/gr33n
cd /opt/gr33n
git clone <your-repo-url> gr33n-platform
```

**Alternative:** copy only `pi_client/` from the laptop:

```bash
# on laptop
scp -r ~/gr33n-platform/pi_client pi@<PI_IP>:/opt/gr33n/gr33n-platform/
```

### 2.3 Python venv + dependencies (once)

```bash
cd /opt/gr33n/gr33n-platform/pi_client
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# NeoPixel on real Pi only (skip on laptop stub tests):
pip install adafruit-circuitpython-neopixel
```

**Shortcut:** from `pi_client/` after clone, `./setup.sh` creates the venv and
installs the systemd unit (paths assume user `pi` and `/home/pi/gr33n_pi` — edit
`gr33n.service` if you use `/opt/gr33n` instead).

### 2.4 Config (once, then tweak)

```bash
cd /opt/gr33n/gr33n-platform/pi_client
cp config.simulation.example.yaml config.yaml
nano config.yaml
```

Minimum bootstrap — set **`api.base_url`** to the **laptop LAN IP**, not
`127.0.0.1` (on the Pi, localhost is the Pi itself):

```yaml
api:
  base_url: "http://<LAPTOP_LAN_IP>:8080"
  api_key: "<PI_API_KEY from laptop .env>"

device:
  uid: "demo-veg-relay-01"

farm:
  farm_id: 1
```

Align `simulation.sensors[].sensor_id` (and `synthetic_sensors` if used) with
`./scripts/print-demo-sensor-ids.sh` output from the laptop. Fresh
`make dev-stack-fresh` seed order is documented in
`config.simulation.example.yaml`.

Set the API key on the Pi (either in YAML or env):

```bash
export PI_API_KEY='<same value as laptop .env>'
```

---

## 3. First run (manual test)

Laptop API must be up (`make dev-auth-test`). Pi and laptop on the same LAN.

```bash
cd /opt/gr33n/gr33n-platform/pi_client
source venv/bin/activate
python3 gr33n_client.py
```

**Expect:** log lines `Light simulation started`, `Synthetic sensor loop started`
(if configured), and Live Sensors updating in the UI within one interval.

Press Ctrl+C to stop. For laptop-only stub tests (no GPIO), the same command
logs `LED[n] = (R,G,B)` without a physical strip.

---

## 4. Auto-start on boot (once)

Edit [`pi_client/gr33n.service`](../pi_client/gr33n.service) if paths differ from
the defaults (`User=pi`, `WorkingDirectory=/home/pi/gr33n_pi`):

```ini
WorkingDirectory=/opt/gr33n/gr33n-platform/pi_client
ExecStart=/opt/gr33n/gr33n-platform/pi_client/venv/bin/python gr33n_client.py --config config.yaml
```

Then:

```bash
sudo cp gr33n.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now gr33n
journalctl -u gr33n -f
```

After this, **power-on = client starts** — you do not re-run pip or config on
every plug-in unless you change wiring or API URL.

---

## 5. Do not run on the Pi

These are **laptop/server only**:

```bash
ollama serve
make dev-auth-test
make guardian-bootstrap-farm ...
make dev-stack-fresh
./scripts/print-demo-sensor-ids.sh
```

---

## 6. Troubleshooting

| Symptom | Check |
|---------|--------|
| Pi cannot reach API | `curl http://<LAPTOP_LAN_IP>:8080/health` from the Pi; firewall on laptop |
| `401` on readings | `api.api_key` / `PI_API_KEY` must match laptop `.env` |
| Wrong sensor in UI | Re-run `print-demo-sensor-ids.sh` on laptop; fix `sensor_id` in `config.yaml` |
| No NeoPixel light | `pip install adafruit-circuitpython-neopixel`; wiring GPIO 18; run on real Pi OS |
| Dashboard empty | Laptop `make dev-auth-test` running; `farm_id: 1` matches seeded demo farm |

**Next:** walk through [`pi-light-simulation-runbook.md`](pi-light-simulation-runbook.md) Demo A–C.
