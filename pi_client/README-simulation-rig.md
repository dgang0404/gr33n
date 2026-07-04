# Pi LED simulation rig — hardware & driver swap (Phase 125 WS5)

Build a **lights-only dry-run bench** before wiring real relays and sensors.
Same `pi_client` + platform sync as production — only the bottom-layer driver
changes when you go live.

**See also:**
- Headless Pi first boot: [`../docs/pi-headless-first-boot.md`](../docs/pi-headless-first-boot.md)
- LED mapping: [`../docs/pi-light-simulation-mapping.md`](../docs/pi-light-simulation-mapping.md)
- Demo walkthrough: [`../docs/pi-light-simulation-runbook.md`](../docs/pi-light-simulation-runbook.md)
- Platform sync: [`../docs/pi-integration-guide.md`](../docs/pi-integration-guide.md)

---

## Parts list (rig v1)

| Part | Qty | Notes |
|------|-----|-------|
| Raspberry Pi 4 (2GB+) or Pi 5 | 1 | Pi Zero 2 W works for LED-only; more headroom with 4/5 |
| 5 V 3 A USB-C power supply | 1 | Official Pi supply recommended |
| microSD 16 GB+ | 1 | Raspberry Pi OS Lite 64-bit |
| WS2812B NeoPixel strip or 8-LED ring | 1 | 8 pixels for rig v1 |
| 330 Ω resistor | 1 | NeoPixel data line (DIN) |
| 1000 µF electrolytic capacitor | 1 | Across strip 5V/GND near first pixel (recommended) |
| 5 V level shifter (74AHCT125 or similar) | 0–1 | Optional; use for long strips or noisy 5V rail |
| Jumper wires | several | Pi GPIO 18 → DIN, GND common |
| 3 mm LEDs + 220 Ω resistors | 2 | Heartbeat (GPIO 17) + fault (GPIO 27) — optional if strip-only |
| Breadboard or Perma-Proto | 1 | Bench layout |

**Not required for simulation:** relay HAT, soil probes, pumps. Add those when
swapping to production (below).

### Software on the Pi

```bash
sudo apt update
sudo apt install -y python3-pip python3-venv i2c-tools
cd /opt/gr33n/pi_client   # or your clone path
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
# Optional NeoPixel on Pi:
pip install adafruit-circuitpython-neopixel
```

---

## Wiring diagram (rig v1)

```
                    Raspberry Pi
                 ┌─────────────────┐
    3.3V logic   │  GPIO 18 (D18)├───[330Ω]───► DIN  WS2812 strip
                 │  GND          ├──────────────► GND  strip + PSU
                 │  GPIO 17      ├───[220Ω]───► heartbeat LED → GND
                 │  GPIO 27      ├───[220Ω]───► fault LED → GND
                 │  5V (pin 2)   │              (optional: power strip
                 └─────────────────┘               from Pi — prefer
                                                   separate 5V supply)

NeoPixel strip power:
  External 5V ──► strip 5V
  External GND ──► strip GND + Pi GND (common ground required)
  Do NOT back-feed high-current 5V into the Pi 5V pin on long strips.
```

**Data pin:** BCM **GPIO 18** (default in `config.simulation.example.yaml`).

**Pixel order:** WS2812B is usually **GRB** — set `neopixel.pixel_order: GRB`.

---

## Configuration

1. Export platform wiring from Virtual Pi or use platform sync:

   ```yaml
   device:
     uid: "demo-veg-relay-01"   # matches gr33ncore.devices.device_uid
   farm:
     farm_id: 1
   api:
     base_url: "http://YOUR_API:8080"
     api_key: "gdev_…"          # per-device key (Phase 57) or PI_API_KEY
   ```

2. Enable simulation block — copy from `config.simulation.example.yaml`:

   ```yaml
   simulation:
     enabled: true
     neopixel: { pin: 18, count: 8, brightness: 0.4 }
     sensors: [ … ]              # pixel map + thresholds
     actuators: [ … ]
     synthetic_sensors: [ … ]    # optional WS3 loopback
   ```

3. Install systemd unit (production Pi):

   ```bash
   sudo cp gr33n.service /etc/systemd/system/
   sudo systemctl enable --now gr33n
   ```

4. Verify: `journalctl -u gr33n -f` — look for `Light simulation started`.

---

## Swap simulation → real relays (production)

No platform/API changes. Only Pi-side config and hardware.

### Step 1 — Install relay hardware

- Wire Sequent 8-relay HAT or GPIO relays per Virtual Pi export / hookup steps
- Assign actuators in dashboard wiring (`driver: relay_hat` or `gpio`)

### Step 2 — Disable simulation actuators

In `config.yaml`:

```yaml
simulation:
  enabled: false          # disables LED driver + SimulationActuatorController
```

Or remove the entire `simulation:` block.

### Step 3 — Keep platform sync

Leave `device.uid` and drop local `actuators:` override so wiring comes from API:

```yaml
api: { … }
device:
  uid: "demo-veg-relay-01"
farm:
  farm_id: 1
# No simulation: block
# No local actuators: — platform sync supplies relay channels
```

### Step 4 — Real sensors

- Remove `simulation.synthetic_sensors` (WS3 demo data)
- Wire physical sensors in Virtual Pi; export updates `GET /devices/by-uid/…/config`
- Notify Pi: **Notify Pi to reload** button or `POST /devices/{id}/push-config`

### Step 5 — Restart & validate

```bash
sudo systemctl restart gr33n
```

- [ ] `journalctl -u gr33n` shows real GPIO/HAT lines, not `[sim]`
- [ ] Manual ON/OFF in UI toggles physical relay
- [ ] Sensor readings appear in Live Sensors (not synthetic)

### Optional — LEDs + relays together

For bench debug only:

```yaml
simulation:
  enabled: true
  # … LED map …
```

And in platform wiring set `mirror_relay_gpio: true` on an actuator row (if
supported in export) **or** use real `ActuatorController` by setting
`simulation.enabled: false` while keeping a separate read-only LED process — the
supported path is **either** simulation **or** real relays, not both on the same
channel.

---

## Laptop dev (no hardware)

```bash
cd pi_client
cp config.simulation.example.yaml config.yaml
# set api.base_url + api_key from .env
python3 gr33n_client.py
```

NeoPixel and GPIO run in **stub mode** — LED states appear in debug logs.
Synthetic sensors still POST to the API and drive automation.

---

## File index

| File | Purpose |
|------|---------|
| `gr33n_client.py` | Main daemon; starts LED + synthetic loops when configured |
| `light_simulation.py` | NeoPixel / GPIO indicator driver |
| `synthetic_sensors.py` | WS3 loopback POST |
| `nudge_sensor.py` | One-shot manual reading |
| `run_demo_moisture_loop.py` | External scripted demo |
| `config.simulation.example.yaml` | Rig v1 reference config |

---

*Phase 125 WS5 — rig v1 for demo farm 1 / `demo-veg-relay-01`.*
