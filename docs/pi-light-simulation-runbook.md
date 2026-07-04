# Pi LED simulation rig — demo runbook (Phase 125 WS4)

Hands-on walkthrough for rig v1: watch sensor comfort bands → automation → actuator
commands on the NeoPixel strip **and** in the gr33n UI. No live plants required.

**Prerequisites:** mapping spec [`pi-light-simulation-mapping.md`](pi-light-simulation-mapping.md),
hardware/setup [`../pi_client/README-simulation-rig.md`](../pi_client/README-simulation-rig.md),
headless Pi install [`pi-headless-first-boot.md`](pi-headless-first-boot.md).

---

## Before you start

- [ ] `make dev-stack-fresh` (or running API + UI against demo farm 1)
- [ ] Pi or laptop with `pi_client` deps: `pip install -r pi_client/requirements.txt`
- [ ] API key in env: `export PI_API_KEY=…` (from repo root `.env`)
- [ ] Sensor IDs match your DB: `./scripts/print-demo-sensor-ids.sh` — align
      `sensor_id` values in `config.yaml` if they differ from the example
- [ ] Copy config: `cp pi_client/config.simulation.example.yaml pi_client/config.yaml`
      and set `api.base_url` + `api.api_key`

**Off-Pi / no strip:** `gr33n_client` logs `LED[n] = (R,G,B)` in stub mode — demos
still work; skip “look at pixel N” steps and watch logs instead.

---

## Demo A — Automated moisture loop (primary path)

**Goal:** Pixel 0 drops out of band → alert in UI → pump LED reacts.

### 1. Start the rig

```bash
cd pi_client
python3 gr33n_client.py
```

**Expect:**
- Log line: `Light simulation started`
- Log line: `Synthetic sensor loop started` (if `synthetic_sensors` configured)
- GPIO 17 heartbeat toggling (or stub log on laptop)
- Pixels 0–4 mostly **green** within ~10s

### 2. Watch the moisture cycle (~3 minutes)

With `mode: demo_moisture` on sensor_id **7** (Media Moisture Indoor):

| Time (approx) | Pixel 0 | What’s happening |
|---------------|---------|------------------|
| 0:00–1:00 | Green solid | Reading ~55%, in band |
| 1:00–1:40 | Cyan blink | Drops below 25% alert threshold |
| 1:40–2:15 | Cyan blink | Holds low ~20% |
| 2:15–3:00 | Green solid | Recovers to ~55% |

### 3. Check the UI (while cyan is blinking)

- [ ] **Live Sensors** or zone cockpit — Media Moisture shows **ALERT** badge
- [ ] **Alerts** — new or updated low-moisture alert (if rules/schedules wired)
- [ ] **Pixel 6** (Veg Irrigation Pump) — amber pulse then blue blink if automation
      enqueued a pump command; pixel **7** white flash on activity

### 4. Pass criteria

- [ ] Pixel 0 color matches moisture band (green ↔ cyan)
- [ ] UI alert/threshold breach matches LED state within ~10s of POST
- [ ] If pump command fired: pixel 6 blinked and command visible under device queue
      in UI (Veg Relay Controller)

---

## Demo B — Manual nudge (operator control)

**Goal:** Prove one POST changes the rig instantly — good for live presentations.

With `gr33n_client` running:

```bash
# In band
python3 nudge_sensor.py --sensor-id 7 --value 55

# Trip low alert
python3 nudge_sensor.py --sensor-id 7 --value 18

# Recover
python3 nudge_sensor.py --sensor-id 7 --value 50
```

- [ ] Pixel 0: green → cyan blink → green within ~2s per nudge
- [ ] UI sensor tile matches each step

---

## Demo C — Scripted external loop (no synthetic_sensors in config)

Disable `simulation.synthetic_sensors` in `config.yaml` and run:

```bash
python3 run_demo_moisture_loop.py --sensor-id 7 --interval 5
```

Same pass criteria as Demo A. Useful when the daemon should not auto-generate data.

---

## Demo D — EC / pH band (pixels 1 and 4)

With synthetic sine on EC (sensor 8) or manual nudge:

```bash
python3 nudge_sensor.py --sensor-id 8 --value 0.3    # below 0.5 → cyan
python3 nudge_sensor.py --sensor-id 8 --value 4.0    # above 3.5 → red blink
python3 nudge_sensor.py --sensor-id 9 --value 5.0    # pH low
```

- [ ] Pixel 1 (EC) and pixel 4 (pH) change per mapping spec
- [ ] Sensor detail pages show matching WARN/ALERT badges

---

## Demo E — Grow light schedule (pixel 5)

If **Light ON 18/6 Veg** schedule is active on demo farm 1:

- [ ] At schedule fire time, pixel **5** yellow blink (grow light on)
- [ ] UI schedule last-run or actuator state shows **on**
- [ ] After OFF schedule, pixel 5 returns dim white (idle)

*Tip:* Temporarily shorten cron in UI for a faster demo, then restore.

---

## Demo F — Fault indicators

| Action | Expect |
|--------|--------|
| Stop API (`docker compose stop api`) | GPIO 27 fault LED on; sensor pixels may go gray (stale) |
| Restart API | Fault off; readings resume |
| Send invalid command (optional) | Pixel on failed actuator magenta fast-blink briefly |

---

## Troubleshooting

| Symptom | Check |
|---------|--------|
| All pixels dark | `simulation.enabled: true`? NeoPixel pin/power? |
| Pixel 0 stuck gray | No reading in cache — synthetic loop running? API reachable? |
| Wrong sensor LED | `./scripts/print-demo-sensor-ids.sh` vs `sensor_id` in YAML |
| Pump LED never blinks | Automation rule/schedule exists for moisture? Device `demo-veg-relay-01` online? |
| POST 401 | `PI_API_KEY` / device key matches farm |

---

## Quick reference — rig v1 pixel map

| Pixel | Entity |
|-------|--------|
| 0 | Media Moisture Indoor |
| 1 | EC Sensor |
| 2 | Air Temp Indoor |
| 3 | Air Humidity Indoor |
| 4 | pH Sensor |
| 5 | Veg Room Grow Light |
| 6 | Veg Room Irrigation Pump |
| 7 | Activity flash |

---

*Phase 125 WS4 — update this checklist when rig v2 or new demo farm layouts ship.*
