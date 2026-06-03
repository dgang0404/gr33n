---
domain: pi
safety_tier: safe
---

# Pi wiring basics (low-voltage control)

Use this guide for **3.3 V / 5 V DC control wiring** on the Raspberry Pi side only.

## Power

- Use the official 5 V USB-C supply (3 A recommended for Pi 4).
- **Green LED** on the board means the Pi has power.

## Common GPIO pins (BCM numbering)

| Role | BCM | Physical pin |
|------|-----|--------------|
| 3.3 V | — | 1 |
| 5 V | — | 2 |
| GND | — | 6 |
| GPIO 17 (example relay IN) | 17 | 11 |
| GPIO 4 (example 1-wire data) | 4 | 7 |

## Relay control module (typical 3-pin board)

- **IN** → a GPIO pin (e.g. GPIO 17).
- **VCC** → 5 V (pin 2) when the board expects 5 V logic.
- **GND** → GND (pin 6).

The **switched terminals** on the relay may carry **line voltage** — that side is **not** covered here; see `electrical-safety.md` and use a qualified electrician.

## Before you power on

1. Pi powered off or unplugged while moving wires.
2. Double-check GND is common between Pi, relay board, and sensor ground.
3. Confirm polarity on sensors that label VCC/GND/data.
