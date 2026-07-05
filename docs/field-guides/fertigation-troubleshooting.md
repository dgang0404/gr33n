---
domain: fertigation
safety_tier: safe
---

# Fertigation troubleshooting

Use the **live farm snapshot** for program names and schedule posture. Use **`summarize_zone_fertigation`** and **`lookup_crop_targets`** for EC targets and setpoints — do not invent numbers.

## Program active but no dose

| Check | What it means |
|-------|----------------|
| `schedule_id` bound? | Unscheduled programs do not auto-run — operator or automation must trigger |
| Pi / pump actuator online? | `summarize_device_health` — heartbeat, relay channel, pending command |
| Reservoir status | Empty or `maintenance` reservoir blocks mix/dose |
| Zone EC trigger | Program may skip when substrate EC already above trigger |
| Worker running? | `automation/worker` must tick scheduled programs on the server |

## Wrong EC or pH after feed

| Check | Action |
|-------|--------|
| Compare to `lookup_crop_targets` | Structured targets beat narrative guesses |
| Reservoir mix vs inline dose | Confirm last `mixing_event` matches the program recipe |
| Sensor calibration | Stale or uncalibrated EC/pH probe — see sensor-install guide |
| Stage mismatch | Program EC target row must match crop **current_stage** |

## Pump runs but plants stay dry

- Relay **IN** GPIO must match gr33n actuator record — platform wiring can differ from physical wires.
- Listen for relay click on test command; no click → control side (Pi/GPIO).
- Load side (mains pump power) → qualified electrician only.

## Guardian boundaries

- Guardian **reads** programs, sensors, and device health — it does not silently start feeds.
- Writes (pause program, enqueue actuator test) require **propose → Confirm**.
