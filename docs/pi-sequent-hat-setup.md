# Pi + Sequent Microsystems Stacking HAT Setup

> **In-app guide:** `localhost:5173/pi-setup` (same content, rendered visually).
>
> **Companion docs:**
> - [`pi-integration-guide.md`](pi-integration-guide.md) — API wiring, Phase 51 platform sync, config.yaml reference
> - [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md) — OS packages, network topology, multi-Pi farms

---

## Why Sequent Microsystems?

The Sequent HAT ecosystem is the cleanest path to **industrial-grade I/O on a Raspberry Pi**:

- **One I²C bus** — all cards share 2 GPIO pins (SDA/SCL); the remaining 24 GPIO stay free for future expansion.
- **Stack to 64 relays per Pi** — add cards as the farm grows; no rewiring, just set a DIP switch.
- **Standard pluggable connectors** — 26–16 AWG stranded wire with ferrule crimps, easy in a DIN enclosure.
- **NO and NC contacts** — wire fail-safe (NC = load energized when Pi is off) for ventilation.
- **Fits any Pi** — Zero through Pi 5, same pinout.

---

## Parts list (starter grow room)

| Qty | Part | Notes |
|-----|------|-------|
| 1× | Raspberry Pi 4B or 5 | Pi 5 recommended. Zero 2W works for small setups. |
| 1× | [Sequent Microsystems 8-Relay HAT](https://sequentmicrosystems.com/products/eight-relays-stackable-card-for-raspberry-pi) | 8 relays, 4A/120VAC NO/NC. Stack level 0 (DIP: all OFF). |
| 1× | 5V / 8A power supply | Powers both Pi and relay card. Each relay ~80mA at turn-on. |
| — | 14–16 AWG stranded wire | Load wiring (pump, lights). |
| — | 22 AWG stranded wire | Signal wiring (sensor inputs). |
| — | Ferrule crimping kit | Required for pluggable terminal blocks on HAT. |
| 1× | DIN rail enclosure (optional) | Neat install alongside breakers and contactors. |

---

## Stack diagram

```
  ┌──────────────────────────────────┐
  │  Stack 2 — 8-Relay HAT           │  ch 16–23   DIP: OFF ON  ON
  ├──────────────────────────────────┤
  │  Stack 1 — 8-Relay HAT           │  ch  8–15   DIP: ON  OFF OFF
  ├──────────────────────────────────┤
  │  Stack 0 — 8-Relay HAT  ← start  │  ch  0–7    DIP: OFF OFF OFF
  ├──────────────────────────────────┤
  │  Raspberry Pi 4B / 5             │
  │  I²C: GPIO2 (SDA) + GPIO3 (SCL) │  ← only 2 pins used
  └──────────────────────────────────┘
```

Cards can be stacked in any physical order — the DIP address determines which card responds.

---

## DIP switch address table

Set the 3-bit DIP switch (`ID0 ID1 ID2`) on each card before stacking.
Every card in the same stack must have a unique address.

| Stack level | I²C address | ID0 | ID1 | ID2 | gr33n channels |
|-------------|-------------|-----|-----|-----|----------------|
| **0** ← first | 0x27 | OFF | OFF | OFF | ch 0 – 7 |
| 1 | 0x26 | ON  | OFF | OFF | ch 8 – 15 |
| 2 | 0x25 | OFF | ON  | OFF | ch 16 – 23 |
| 3 | 0x24 | ON  | ON  | OFF | ch 24 – 31 |
| 4 | 0x23 | OFF | OFF | ON  | ch 32 – 39 |
| 5 | 0x22 | ON  | OFF | ON  | ch 40 – 47 |
| 6 | 0x21 | OFF | ON  | ON  | ch 48 – 55 |
| 7 | 0x20 | ON  | ON  | ON  | ch 56 – 63 |

---

## Channel numbering in gr33n

Each relay maps to a gr33n `channel_id`:

```
stack 0 relay 1 → channel_id: 0
stack 0 relay 2 → channel_id: 1
...
stack 0 relay 8 → channel_id: 7
stack 1 relay 1 → channel_id: 8
stack 1 relay 2 → channel_id: 9
...
stack N relay R → channel_id: (N × 8) + (R − 1)
```

In gr33n config (Phase 51 platform sync — recommended):

```yaml
# config.yaml — minimal bootstrap; wiring pulled from dashboard
api:
  base_url: "http://192.168.1.100:8080"
  api_key: "replace-with-PI_API_KEY"
device:
  uid: "flower-room-01"
farm:
  farm_id: 1
```

In the dashboard (Controls → New actuator) set:
- **Actuator type:** relay
- **Channel ID:** 0–63 matching the table above
- **Zone:** the grow room this relay serves

---

## Typical 8-channel farm wiring plan

Assign channels before physical wiring so the dashboard labels match the hardware.

| Channel | Relay | Typical use | Load | Notes |
|---------|-------|-------------|------|-------|
| ch0 | Relay 1 | Main irrigation pump | 120VAC pump | Via contactor if >4A |
| ch1 | Relay 2 | Nutrient dosing pump A | 24VDC peristaltic | |
| ch2 | Relay 3 | Nutrient dosing pump B | 24VDC peristaltic | pH up / pH down |
| ch3 | Relay 4 | Drain / return pump | 120VAC pump | Or CO₂ solenoid |
| ch4 | Relay 5 | Grow lights | LED / HID | Contactor for HID; SSR for dimmable LED |
| ch5 | Relay 6 | Exhaust fan | 120VAC fan | Use NC contact for fail-safe-on |
| ch6 | Relay 7 | Humidifier / dehumidifier | Appliance | |
| ch7 | Relay 8 | Spare / heater | — | Reserve for expansion |

> **⚡ Load exceeds 4A / 120VAC?** Wire the relay to a contactor coil or SSR — the relay switches the coil, not the full load.
> Typical: 25A HVAC contactor for HID lights, 40A DIN-rail SSR for high-wattage LEDs.

---

## Step-by-step setup

### 1. Flash Pi OS and enable I²C

```bash
sudo raspi-config  # Interface Options → I2C → Enable
sudo reboot
```

### 2. Set DIP switch, then stack the card

First card: all 3 bits **OFF** (stack level 0, I²C 0x27).
**Power off the Pi before stacking or unstacking.** Never connect HATs with power on.

### 3. Install Sequent relay drivers

```bash
cd ~
git clone https://github.com/SequentMicrosystems/8relind-rpi.git
cd 8relind-rpi
sudo make install

# Test: turn relay 1 on then off
8relind 0 write 1 on
8relind 0 write 1 off
```

### 4. Verify I²C addresses

```bash
sudo apt install -y i2c-tools
i2cdetect -y 1
# Stack 0 only  → shows 27
# Stack 0 + 1   → shows 27 and 26
```

### 5. Install gr33n Pi client

```bash
cd ~
git clone <your-gr33n-repo>
cd gr33n-platform/pi_client
cp config.bootstrap.example.yaml config.yaml
nano config.yaml   # set api.base_url, api.api_key, device.uid
```

### 6. Add actuators in the dashboard

Controls → New actuator → set channel_id to match the table above → assign to a zone.
The Pi pulls wiring from the platform at startup (Phase 51).

### 7. Wire physical loads

Use pluggable connectors (ferrule crimps on stranded wire):
- **NO terminal** — normally open, load off when idle (most loads)
- **NC terminal** — normally closed, load on when idle (exhaust fans, fail-safe)

COM → Common terminal. Load wire → NO or NC.

### 8. Run and verify

```bash
python3 gr33n_client.py --config config.yaml
```

Dashboard → Controls → your actuator → Manual pulse → 1 second.
Watch the relay LED on the card illuminate.

---

## Scaling

| Farm size | Cards | Relays | Notes |
|-----------|-------|--------|-------|
| Starter room | 1× | 8 | Pump, lights, fan, dosing, spare |
| Full farm | 2–4× | 16–32 | Multi-room: flower + veg + mothers on one Pi |
| Max per Pi | 8× | 64 | Warehouse scale. Add second Pi for redundancy |

---

## Adding sensor inputs

The relay HAT handles outputs. For inputs, add a matching Sequent card to the same stack:

| Card | Inputs | Use |
|------|--------|-----|
| **Eight HV Digital Inputs HAT** | 8 opto-isolated 3–240V AC/DC | Float switches, door sensors, flow pulse counters |
| **Building Automation HAT** | 8 universal (thermistor, 0–10V, dry contact) + 4 TRIAC/0–10V out | Temperature, humidity, EC probes (via signal conditioner), VFD control |
| **Industrial Automation HAT** | Mixed digital + analog + MOSFET | Modbus/RS485 devices, variable-speed pumps, CO₂ controllers |

Input cards share the same I²C bus and DIP addressing scheme, but each card *type* has its own address space — relay card at stack 0 and input card at stack 0 do **not** conflict.

---

## Troubleshooting

| Symptom | Check |
|---------|-------|
| `i2cdetect` shows nothing | I²C not enabled, or HAT not seated fully on header |
| `8relind` command not found | `sudo make install` not completed, or `PATH` missing `/usr/local/bin` |
| Relay clicks but load doesn't switch | Load wired to NC instead of NO (or vice versa), or load exceeds 4A — add contactor |
| Pi offline in dashboard | API key wrong, base_url unreachable, or Pi client not running |
| Config stale badge in UI | Pi can't reach API for config pull — check network and API key |

---

*Hardware: [sequentmicrosystems.com](https://sequentmicrosystems.com) · Relay driver: [8relind-rpi](https://github.com/SequentMicrosystems/8relind-rpi)*
