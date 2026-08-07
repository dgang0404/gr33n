---
domain: pi
safety_tier: safe
---

# gr33n Demo Farm — edge device map (farm id 1 seed)

Reference layout for **development / demo** — physical wiring may differ; always verify with `summarize_device_health` and the Wiring UI.

## Devices (seed names)

| Device | Zone | `device_uid` | Role |
|--------|------|--------------|------|
| Veg Relay Controller | Veg Room | `demo-veg-relay-01` | Relay HAT — grow light |
| Flower Relay Controller | Flower Room | `demo-flower-relay-01` | Relay HAT — irrigation pump |

Both seed devices report **`simulation: true`** in config — suitable for laptop demo without real GPIO.

## Actuators (platform records)

| Actuator | Device | `hardware_identifier` | GPIO (BCM) | Type |
|----------|--------|----------------------|------------|------|
| Veg Room Grow Light | Veg Relay Controller | `relay_1` (channel 1) | **16** (phys 36) | light |
| Veg Room Irrigation Pump | Veg Relay Controller | `relay_2` (channel 2) | **20** (phys 38) | pump |
| Flower Room Irrigation Pump | Flower Relay Controller | `relay_2` (channel 2) | **27** (phys 13) | pump |

### Reserved on Veg LED sim / dual-use bench

| BCM | Physical | Role — do not assign platform relays here |
|-----|----------|-------------------------------------------|
| 17 | 11 | Heartbeat LED (`gpio_leds.heartbeat_pin`) |
| 27 | 13 | Fault LED (`gpio_leds.fault_pin`) — Flower pump uses 27 on a **different** Pi |

Low-voltage control wiring patterns: see `pi-wiring-basics.md` and `relay-and-actuator-wiring.md`.

## Fertigation programs (seed)

| Program | Zone | Notes |
|---------|------|-------|
| Veg Daily JLF Program | Veg Room | Pairs with veg light schedule |
| Flower Daily FFJ+WCA Program | Flower Room | Flower reservoir |
| Outdoor JLF Soil Drench | Outdoor | Often manual / unscheduled |

## Procedures

- New install: `start procedure wire-pi-relay-light`
- Actuator stuck: `start procedure diagnose-actuator-wont-fire`
- Pi not in UI: `start procedure diagnose-pi-offline`
