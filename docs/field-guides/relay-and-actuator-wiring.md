---
domain: actuator
safety_tier: caution
---

# Relay and actuator wiring

## Control side (Pi → relay) — safe for operators

Wire only the **low-voltage IN, VCC, GND** pins to the Pi as described in `pi-wiring-basics.md`.

When the gr33n UI sends `pending_command` to the Pi, the configured **GPIO pin** must match the wire on **IN**.

## Switched load side — qualified person required

The relay **COM / NO / NC** terminals often switch **mains AC** (120 V / 240 V) for grow lights, pumps, or contactors.

- **Do not** follow step-by-step mains wiring from chat.
- Hire a **licensed electrician** for line-voltage terminations, breakers, and enclosures.
- Always **unplug or lock out** upstream power before anyone opens a mains box.

## Actuator won't fire checklist

1. Pi online in gr33n (recent heartbeat)?
2. Command visible in `pending_command` on the device row?
3. Relay IN on the GPIO configured in gr33n?
4. Load side installed by a qualified person and breaker on?
