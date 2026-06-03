---
domain: sensor
safety_tier: safe
---

# Sensor install and calibration

## 3-wire analog / digital sensors (typical)

- **Red / VCC** → 3.3 V (pin 1) unless the datasheet says 5 V only.
- **Black / GND** → GND.
- **Yellow / data** → the GPIO configured in gr33n (e.g. GPIO 4).

## Placement

- Air temp/RH: shaded, away from direct lamp heat.
- EC/pH probes: submerged per manufacturer depth; rinse between solutions.

## Calibration (EC/pH)

Follow probe kit instructions. Record calibration date in your farm notes — Guardian can store operator-stated facts but does not replace probe maintenance logs.

## No reading in gr33n

See `field-troubleshooting.md` and run the **diagnose-sensor-no-reading** procedure when available.
