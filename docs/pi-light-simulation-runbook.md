# Pi LED simulation rig — demo runbook (Phase 125 WS4)

Hands-on walkthrough for rig v1: watch sensor comfort bands → automation → actuator
commands on LEDs **and** in the gr33n UI. No live plants required.

**Two hardware paths:**
- **Starter kit (breadboard LEDs)** — default. Use the [complete plug map](#starter-kit-breadboard-plug-map) below.
- **NeoPixel strip** — optional. See [`pi-light-simulation-mapping.md`](pi-light-simulation-mapping.md).

**Prerequisites:** mapping spec [`pi-light-simulation-mapping.md`](pi-light-simulation-mapping.md),
hardware/setup [`../pi_client/README-simulation-rig.md`](../pi_client/README-simulation-rig.md),
headless Pi install [`pi-headless-first-boot.md`](pi-headless-first-boot.md).

---

## Why this test

Prove the full edge loop **without** real soil probes or pumps:

`synthetic reading → API → Live Sensors / alerts → Pi client → physical LED`

| LED color (kit) | Meaning | Pass when… |
|-----------------|---------|------------|
| **Red** | Pi client heartbeat | Blinks ~1 Hz while `gr33n_client` is running |
| **Yellow** | Media Moisture Indoor | On when moisture in band; blinks when ALERT (demo drops below 25%) |
| **Blue** | Veg Room Irrigation Pump | Blinks when a pump command is on / queued |

Minimum bench: those three LEDs. Add the rest from the full map when you want EC/temp/etc.

---

## Starter-kit breadboard plug map

**Power OFF and unplug the Pi before any wiring.** Never hot-plug GPIO.

This map is the **known-working heartbeat circuit** from the bench build
(Pi pin 11 → `a13` → resistor `b13`–`b11` → LED `a11`/`a10` → blue → pin 6).
Yellow and blue use that same shape on new columns.

### Breadboard orientation (read this once)

Use the **`a`–`e` side** only (do **not** use `f`–`j` for this map).
Same row number on `a`–`e` is one metal strip. The center trench does not cross.

```
  a b c d e   |   f g h i j
  ↑ use this side only
```

Hole names are **letter + row number** (example: `a13`).

### Parts for the minimum (red + yellow + blue)

| Qty | Part |
|-----|------|
| 1 | Raspberry Pi (40-pin header) |
| 1 | Breadboard |
| 3 | LEDs: red, yellow, blue |
| 3 | 220 Ω resistors |
| 7 | Jumpers (3 signal + 3 ground returns + 1 Pi GND → blue rail) |

LED legs: **long = (+)** · **short = (−)** toward ground.

### Pi header pins

| Physical pin | BCM | Wire role |
|-------------:|-----|-----------|
| **6** | GND | → blue (−) rail (outer row, 3rd pin from pin-1 end) |
| **11** | GPIO17 | → red heartbeat (`a13`) |
| **12** | GPIO18 | → yellow moisture (`a20`) |
| **38** | GPIO20 | → blue pump (`a30`) |

### Shared ground (do once)

| Step | Part | From | To |
|-----:|------|------|-----|
| 1 | Jumper | Pi **pin 6** (GND) | Breadboard **blue (−) rail** |

### Complete hole list — red heartbeat (proven working)

Path: `pin 11 → a13 → resistor → a11 → LED → a10 → blue → pin 6`

| Step | Part | Hole / pin |
|-----:|------|------------|
| 1 | Jumper | Pi **pin 11** → **`a13`** |
| 2 | Resistor leg A | **`b13`** |
| 3 | Resistor leg B | **`b11`** |
| 4 | Red LED long (+) | **`a11`** |
| 5 | Red LED short (−) | **`a10`** |
| 6 | Jumper | **`b10`** (or `c10` / `e10` — same row as short leg, not the same hole) → **blue (−) rail** |

If red already matches this and blinks, leave it alone.

### Complete hole list — yellow moisture (pixel 0)

Same shape as red, on rows **20 / 18 / 17**.

| Step | Part | Hole / pin |
|-----:|------|------------|
| 1 | Jumper | Pi **pin 12** → **`a20`** |
| 2 | Resistor leg A | **`b20`** |
| 3 | Resistor leg B | **`b18`** |
| 4 | Yellow LED long (+) | **`a18`** |
| 5 | Yellow LED short (−) | **`a17`** |
| 6 | Jumper | **`b17`** (or `c17` / `e17`) → **blue (−) rail** |

### Complete hole list — blue pump (pixel 6)

Same shape as red, on rows **30 / 28 / 27**.

| Step | Part | Hole / pin |
|-----:|------|------------|
| 1 | Jumper | Pi **pin 38** → **`a30`** |
| 2 | Resistor leg A | **`b30`** |
| 3 | Resistor leg B | **`b28`** |
| 4 | Blue LED long (+) | **`a28`** |
| 5 | Blue LED short (−) | **`a27`** |
| 6 | Jumper | **`b27`** (or `c27` / `e27`) → **blue (−) rail** |

### One-page checklist (minimum three LEDs)

| Color | Role | Pi pin | Signal | Resistor | LED long (+) | LED short (−) | Ground jumper |
|-------|------|-------:|--------|----------|--------------|---------------|---------------|
| Red | Heartbeat | 11 | `a13` | `b13`–`b11` | `a11` | `a10` | `b10` → blue (−) |
| Yellow | Moisture | 12 | `a20` | `b20`–`b18` | `a18` | `a17` | `b17` → blue (−) |
| Blue | Pump | 38 | `a30` | `b30`–`b28` | `a28` | `a27` | `b27` → blue (−) |

Plus once: Pi **pin 6 (GND)** → **blue (−) rail**.

### Platform wiring vs indicator LEDs (do not collide)

Wiring UI / `pi-config` export for `demo-veg-relay-01` must **not** put relays on the
indicator pins:

| BCM | Physical | Platform (Veg) | LED sim indicator |
|----:|---------:|----------------|-------------------|
| 16 | 36 | Grow light relay | Pixel 5 (grow light LED) |
| 20 | 38 | Irrigation pump relay | Pixel 6 (pump LED) |
| **17** | **11** | *(reserved)* | **Heartbeat** |
| **27** | **13** | *(reserved on this Pi)* | **Fault** |

If the Wiring download shows grow light on GPIO **17**, that is a seed bug — fix before
going live with real relays + heartbeat LED on the same header.

### Full discrete LED bank (optional)

Same `a`/`b` pattern: signal **`aN`**, resistor **`bN`–`b(N−2)`**,
LED long **`a(N−2)`**, LED short **`a(N−3)`**, jumper **`b(N−3)`** → blue (−).

| Color (suggested) | Role | Pixel | BCM | Phys pin | Signal | Resistor | LED+ | LED− | Gnd jumper |
|-------------------|------|------:|----:|---------:|--------|----------|------|------|------------|
| Red | Heartbeat | — | 17 | 11 | `a13` | `b13`–`b11` | `a11` | `a10` | `b10`→blue |
| Yellow | Media Moisture | 0 | 18 | 12 | `a20` | `b20`–`b18` | `a18` | `a17` | `b17`→blue |
| Green | EC Sensor | 1 | 23 | 16 | `a25` | `b25`–`b23` | `a23` | `a22` | `b22`→blue |
| White | Air Temp | 2 | 24 | 18 | `a35` | `b35`–`b33` | `a33` | `a32` | `b32`→blue |
| Orange | Air Humidity | 3 | 25 | 22 | `a40` | `b40`–`b38` | `a38` | `a37` | `b37`→blue |
| Purple | pH Sensor | 4 | 12 | 32 | `a45` | `b45`–`b43` | `a43` | `a42` | `b42`→blue |
| Warm white | Grow light | 5 | 16 | 36 | `a50` | `b50`–`b48` | `a48` | `a47` | `b47`→blue |
| Blue | Irrigation pump | 6 | 20 | 38 | `a30` | `b30`–`b28` | `a28` | `a27` | `b27`→blue |
| White flash | Activity | 7 | 21 | 40 | `a55` | `b55`–`b53` | `a53` | `a52` | `b52`→blue |
| Amber | Fault (API down) | — | 27 | 13 | `a60` | `b60`–`b58` | `a58` | `a57` | `b57`→blue |

Config: `simulation.driver: gpio` + `discrete_leds` in
`scripts/run-led-simulation.sh` / Pi `config.yaml`.

### After wiring

1. Plug Pi power back in.
2. Start API on laptop + reverse SSH tunnel if Pi cannot reach the laptop LAN directly.
3. Start `gr33n_client` on the Pi (`simulation.enabled: true`, `driver: gpio`).
4. Watch: red blinks → yellow follows moisture demo → blue blinks if pump command fires.
5. Confirm **Live Sensors** updates for Media Moisture Indoor.

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

**Goal:** Moisture LED drops out of band → alert in UI → pump LED reacts.

**Breadboard:** yellow = moisture, blue = pump, red = heartbeat
([plug map](#starter-kit-breadboard-plug-map)).  
**NeoPixel:** pixel 0 = moisture, pixel 6 = pump (colors below).

### 1. Start the rig

```bash
# Prefer ID-resolving runner (gpio breadboard default):
./scripts/run-led-simulation.sh

# Or on the Pi with existing config.yaml:
cd /opt/gr33n/pi_client && ./venv/bin/python gr33n_client.py
```

**Expect:**
- Log line: `Light simulation started` (and `DiscreteLedBank` if `driver: gpio`)
- Log line: `Synthetic sensor loop started` (if `synthetic_sensors` configured)
- **Red** heartbeat blinking ~1 Hz (GPIO 17)
- Moisture LED on (in band) within ~10s

### 2. Watch the moisture cycle (~3 minutes)

With `mode: demo_moisture` on Media Moisture Indoor:

| Time (approx) | Breadboard yellow | NeoPixel pixel 0 | What’s happening |
|---------------|-------------------|------------------|------------------|
| 0:00–1:00 | Solid ON | Green solid | Reading ~55%, in band |
| 1:00–1:40 | Blink | Cyan blink | Drops below 25% alert threshold |
| 1:40–2:15 | Blink | Cyan blink | Holds low ~20% |
| 2:15–3:00 | Solid ON | Green solid | Recovers to ~55% |

### 3. Check the UI (while moisture is blinking / ALERT)

- [ ] **Live Sensors** or zone cockpit — Media Moisture shows **ALERT** badge
- [ ] **Alerts** — new or updated low-moisture alert (if rules/schedules wired)
- [ ] **Blue pump LED** (or NeoPixel pixel 6) — solid ON for the pulse window on
      breadboard LEDs (use ≥2s Run pulse); activity LED / pixel **7** flashes if wired

### 4. Pass criteria

- [ ] Moisture LED matches band (ON ↔ blink on breadboard; green ↔ cyan on NeoPixel)
- [ ] UI alert/threshold breach matches LED state within ~10s of POST
- [ ] If pump command fired: blue / pixel 6 blinked and command visible under device
      queue in UI (Veg Relay Controller)

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
