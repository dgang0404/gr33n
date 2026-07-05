-- Phase 127: fertigation + demo Pi field guides; refresh field-troubleshooting body.

INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order
)
VALUES
(
    'fertigation-troubleshooting',
    'Fertigation troubleshooting',
    NULL,
    'trades',
    'fertigation',
    'safe',
    $body$# Fertigation troubleshooting

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
- Writes (pause program, enqueue actuator test) require **propose → Confirm**.$body$,
    5,
    TRUE,
    15
),
(
    'demo-farm-pi-layout',
    'gr33n Demo Farm edge device map',
    NULL,
    'trades',
    'pi',
    'safe',
    $body2$# gr33n Demo Farm — edge device map (farm id 1 seed)

Reference layout for **development / demo** — physical wiring may differ; always verify with `summarize_device_health` and the Wiring UI.

## Devices (seed names)

| Device | Zone | `device_uid` | Role |
|--------|------|--------------|------|
| Veg Relay Controller | Veg Room | `demo-veg-relay-01` | Relay HAT — grow light |
| Flower Relay Controller | Flower Room | `demo-flower-relay-01` | Relay HAT — irrigation pump |

Both seed devices report **`simulation: true`** in config — suitable for laptop demo without real GPIO.

## Actuators (platform records)

| Actuator | Device | `hardware_identifier` | Type |
|----------|--------|----------------------|------|
| Veg Room Grow Light | Veg Relay Controller | `relay_1` (channel 1) | light |
| Flower Room Irrigation Pump | Flower Relay Controller | `relay_1` (channel 1) | pump |

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
- Pi not in UI: `start procedure diagnose-pi-offline` $body2$,
    5,
    TRUE,
    16
)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    guide_kind = EXCLUDED.guide_kind,
    domain = EXCLUDED.domain,
    safety_tier = EXCLUDED.safety_tier,
    body_md = EXCLUDED.body_md,
    catalog_version = EXCLUDED.catalog_version,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

UPDATE gr33ncrops.agronomy_field_guides
SET body_md = $ft$# Field troubleshooting (symptom → checks)

| Symptom | First checks |
|---------|----------------|
| Sensor reads nothing | Pi power LED; 3-wire pinout; GPIO matches gr33n; sensor power |
| Actuator won't fire | Pi online; `pending_command`; relay IN pin; mains side by electrician |
| Pi offline in gr33n | Network/API key; `farm_id` in client env; offline queue backlog |
| Wrong zone data | Client `farm_id`; device registered to correct farm |
| Feed did not run | Program `schedule_id`; Pi/pump online; reservoir status; see `fertigation-troubleshooting.md` |
| EC/pH wrong after dose | `lookup_crop_targets`; last mix event; probe calibration; stage match |
| Grow light won't switch | `summarize_device_health`; relay channel vs `hardware_identifier`; demo map in `demo-farm-pi-layout.md` |

Use the live farm snapshot for device counts, program schedule posture, and unread alerts. Describe what you see — Guardian labels **operator-stated** facts separately from measurements.$ft$,
    catalog_version = 5,
    updated_at = NOW()
WHERE slug = 'field-troubleshooting';
