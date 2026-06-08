---
name: Phase 65 — Guardian Pi & hardware diagnostics
overview: >
  Guardian gains a live read tool that can see exactly how each Pi is wired —
  which sensor is on which GPIO pin, which actuator is on which relay channel,
  whether the Pi is online, and how fresh each reading is. Replaces the current
  "Guardian cannot see the wiring; ask the worker to confirm" wall with structured,
  queryable device health data. Enables directed wiring troubleshooting: wrong
  sensor on wrong pin, Pi offline, config stale, duplicate GPIO conflict.
todos:
  - id: ws1-read-tool
    content: "WS1: summarize_device_health read tool — device status, sensors (GPIO/source/last-reading-age), actuators (channel/status), config sync age"
    status: completed
  - id: ws2-intent
    content: "WS2: Intent matching for wiring/device diagnostic questions: offline, wrong reading, not responding, channel wrong"
    status: completed
  - id: ws3-grounding-update
    content: "WS3: Update fieldGuideGrounding + context_ref hints — Guardian can now see platform wiring; remove 'ask them to confirm'"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: architecture §7.0v, phase-65-closure.test.js, OC-65"
    status: completed
isProject: false
---

# Phase 65 — Guardian Pi & hardware diagnostics

## Status

**Shipped.** After [Phase 57](phase_57_pi_device_api_keys.plan.md) (wiring data is stored per-device) and before [Phase 66](phase_66_weather_site_context.plan.md) (weather). **Shipped before [Phase 67](phase_67_guardian_field_assistant.plan.md)** — voice + photo are most useful when Guardian can cross-reference wiring without asking the operator to read it back.

**Depends on:** Phase 50 (wiring model), Phase 51 (config sync / staleness tracking), Phase 57 (device keys, per-device identity).

---

## The one job

> **You say "why is my temperature stuck at 85°F?" — Guardian checks the platform wiring, sees DHT22 on GPIO 4 of the Veg Room Pi, sees the Pi is online but the last reading is 4 hours old, and tells you the signal wire may have pulled loose from GPIO 4.**

Today: *"Guardian cannot see the wiring; ask the worker to confirm what they observe."*
After: Guardian reads the structured wiring from the platform and cross-references it with device status and reading freshness.

---

## Problem today

Guardian's `fieldGuideGrounding` constant explicitly admits it is blind:

```
Guardian cannot see the wiring; ask the worker to confirm what they observe.
Operator-stated facts are labeled, not measured.
```

This is accurate — the wiring has been structured in the DB since Phase 50 and config sync freshness since Phase 51, but no read tool ever queries it. The operator has to narrate their own wiring to Guardian, which:

- Slows down every troubleshooting conversation
- Misses obvious issues Guardian could spot instantly (stale reading vs offline Pi vs wrong GPIO)
- Makes voice-mode troubleshooting (Phase 67) unworkably verbose

---

## WS1 — `summarize_device_health` read tool

New read tool — no Confirm, never writes.

**Input:** farm_id + optional device_id or device_uid hint from intent matching.

**Output (plain text for system prompt):**

```
summarize_device_health — Veg Room Pi (demo-veg-relay-01)
Status: online · last heartbeat 6m ago · config synced 6m ago (v12)

Sensors (4):
- Air Temp Indoor: DHT22 · BCM GPIO 4 · last reading 4h ago ⚠ STALE
- Air Humidity Indoor: DHT22 · BCM GPIO 4 · last reading 4h ago ⚠ STALE (shares GPIO with Air Temp — conflict?)
- CO₂ Sensor Indoor: MH-Z19 · /dev/ttyS0 · last reading 2m ago ✓
- EC Sensor: ADS1115 · I2C ch 0 · last reading 3m ago ✓

Actuators (3):
- Irrigation pump: relay HAT ch 0 (stack 0, relay 1) · zone: Veg Room
- Grow lights: relay HAT ch 4 (stack 0, relay 5) · zone: Veg Room
- Exhaust fan: relay HAT ch 5 (stack 0, relay 6) · zone: Veg Room
```

**Staleness thresholds** (configurable, not hardcoded):
- Sensor reading stale: > 3× polling interval, or > 15 min if interval unknown
- Config stale: `last_config_fetch_at` > 30 min
- Pi offline: `last_heartbeat` > 5 min

**Conflict detection:** flag when two sensors on the same device share a GPIO pin — this is stored in the platform and is often the cause of one sensor always reading wrong.

---

## WS2 — Intent matching

Fire `summarize_device_health` when the question matches wiring/device diagnostic patterns:

| Pattern examples | Intent |
|-----------------|--------|
| "sensor is stuck / wrong / not updating" | stale reading investigation |
| "Pi is offline / not connecting" | device health check |
| "fan / pump not responding" | actuator channel check |
| "wrong reading / impossible value" | GPIO conflict or wrong pin |
| "wiring / gpio / channel / relay" combined with a sensor/actuator name | direct wiring lookup |
| "why is [sensor name] showing [value]" | cross-reference reading + wiring |

If a specific device_uid or device name is in the question, scope the tool to that device.
If no device is specified but the zone context (from `context_ref`) points at a zone, return devices in that zone.
Fall back to all farm devices if ambiguous.

---

## WS3 — Grounding update

**`fieldGuideGrounding`** (guardian.go): remove the `"Guardian cannot see their wiring"` sentence. Replace with:

> Guardian can look up how sensors and actuators are wired on each Pi (GPIO pin, relay channel, device assignment, last reading freshness) via the `summarize_device_health` read tool. When wiring looks correct in the platform but the operator reports wrong behaviour, ask them to verify the physical connection matches the platform record.

**`context_ref.go`** — when context is `/pi-setup`, `/sensors`, `/actuators`, or a sensor detail page:

- Add: "Guardian can now call `summarize_device_health` to see actual GPIO / relay channel assignments from the platform — no need to ask the operator to read back config."

**`field_assistant.go`** — update the field persona note:

- Before: "You cannot see their wiring; ask them to confirm what they observe."
- After: "You can see the platform wiring record. Cross-reference it with what the operator observes physically — the platform may be correct but the physical wire may differ."

---

## What Guardian can say after this phase

| Scenario | Before Phase 65 | After Phase 65 |
|----------|----------------|----------------|
| "Temp sensor stuck at 85°F" | "What GPIO pin is it on?" | "Your Air Temp (DHT22, GPIO 4) last reported 4h ago. The Pi is online — the signal wire may have pulled loose from GPIO 4, or the sensor has failed." |
| "Pi is offline" | "Check your config.yaml for the right API URL" | "Your Veg Room Pi last heartbeat was 2h ago. Config was synced 2h ago (v12 — current). Check power and network; the last known state was online." |
| "Fan not coming on" | "What channel is it on?" | "Your Exhaust fan is on relay channel 5 (stack 0, relay 6) of the Veg Room Pi. Confirm the Pi is online and the relay IN lead is connected." |
| "Two sensors reading same wrong value" | "Check your wiring manually" | "Air Temp and Air Humidity are both set to GPIO 4 — that's a pin conflict. One of them needs to move to a different pin." |

---

## WS4 — Docs, tests, OC-65

- `farm-guardian-architecture.md` §7.0v — Guardian Pi diagnostics
- `operator-tour.md` §6m
- `phase-65-closure.test.js` — intent fires on wiring questions; tool output contains GPIO + staleness
- OC-65 closed in `phase_35_37_operational_closure.plan.md`

---

## Definition of done

- [x] `summarize_device_health` read tool fires on wiring/device diagnostic intent
- [x] Output includes: device status, sensors with GPIO/source/last-reading-age, actuators with channel, GPIO sharing flagged
- [x] `fieldGuideGrounding` updated — Guardian no longer says it cannot see the wiring
- [x] OC-65 closed (`phase-65-closure.test.js`, `smoke_phase65_test.go`)

---

## Boundary

- **Read only** — Guardian reports what the platform says the wiring is. It does not auto-fix wiring or propose changes to GPIO pins (those go through the edit panel on Sensors / Controls).
- **Platform wiring ≠ physical wiring** — the platform stores operator-entered data. Guardian should always caveat: "this is what your platform wiring says — verify the physical wire matches."
- **No mains AC diagnosis** — the safety stop rule stays. Guardian will not walk through line-voltage troubleshooting regardless of what the wiring tool shows.
- **Phase 37 relationship** — [Phase 37 WS5](phase_37_guardian_offline_field_assistant.plan.md) field diagnostics used snapshot + procedure refs while asking the operator to confirm GPIO/channel. Phase 65 **supersedes that blind spot** (structured platform wiring + staleness). Guided procedures (`wire-pi-relay-light`, `diagnose-actuator-wont-fire`) and safety gating stay; this phase adds live read depth on top.

---

## Ship order note

**Ship 64 first** (crop knowledge base — more important, unblocks grow advisor).
Then **65** (wiring diagnostics — useful standalone, makes Phase 67 voice much better).
Then **66** (weather — independent).
Then **67** (voice + photo — uses 64 for crop grounding + 65 for wiring context when troubleshooting by voice).
