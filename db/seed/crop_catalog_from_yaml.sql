-- Generated from data/crop_library.yaml + docs/field-guides — do not edit by hand.
-- Regenerate: ./scripts/generate-crop-catalog-seed.sql.sh
-- crop_library version: 4

INSERT INTO gr33ncrops.crop_catalog_entries (
    crop_key, display_name, supported, category, source, substrate, watering_style,
    runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version
)
SELECT v.crop_key, v.display_name, v.supported, v.category, v.source, v.substrate, v.watering_style,
       v.runoff_pct_target, v.moisture_guidance, v.cousin_of, v.unsupported_reason, v.catalog_version
FROM (VALUES
    ('cannabis', 'Cannabis', true, 'flower', 'Curated indoor ranges; verify against your genetics', 'coco / rockwool', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('tomato', 'Tomato', true, 'fruiting', 'Hydroponic fruiting tomato references', 'coco / rockwool slab', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('pepper', 'Pepper (bell/chili)', true, 'fruiting', 'Similar to tomato, lower EC headroom', 'coco / rockwool', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('lettuce', 'Lettuce / leafy greens', true, 'leafy', 'Low-EC leafy production', 'DWC / NFT / rockwool', 'constant_feed', '5-10', NULL, NULL, NULL, 4),
    ('phalaenopsis', 'Orchid (Phalaenopsis)', true, 'epiphyte', 'Epiphyte — very low EC, high RH', 'orchid_bark', 'mist_epiphyte', '0', NULL, NULL, NULL, 4),
    ('basil', 'Basil / herbs', true, 'herb', 'Warm-weather herb baseline', 'coco / rockwool', 'constant_feed', '10-15', NULL, NULL, NULL, 4),
    ('strawberry', 'Strawberry', true, 'fruiting', 'Day-neutral strawberry baseline', 'coco / rockwool', 'pulse_dryback', '10-15', NULL, NULL, NULL, 4),
    ('eggplant', 'Eggplant', true, 'fruiting', 'Solanaceous fruiting; ~10% lower EC than tomato; hand-pollinate indoors', 'coco / rockwool slab', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('cucumber', 'Cucumber', true, 'fruiting', 'Vining fruiting cucumber; higher RH than tomato', 'coco / rockwool', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('kale', 'Kale', true, 'leafy', 'Leafy brassica; slightly higher EC than lettuce', 'DWC / NFT', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('spinach', 'Spinach', true, 'leafy', 'Cool-season leafy; bolts in heat', 'DWC / NFT', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('cilantro', 'Cilantro / coriander', true, 'herb', 'Cool-leaning herb; bolts in heat like basil but prefers lower temps', 'coco / rockwool', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('microgreens', 'Microgreens', true, 'leafy', 'Very low EC; 10–14 day turnover; shallow moisture', 'soilless_tray', 'top_water_drydown', NULL, NULL, NULL, NULL, 4),
    ('zucchini', 'Zucchini / summer squash', true, 'fruiting', 'Fruiting squash; modeled on cucumber', 'coco / rockwool', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('green_bean', 'Green bean', true, 'fruiting', 'Moderate EC warm legume; modeled on pepper', 'coco / rockwool', 'pulse_dryback', '10-15', NULL, NULL, NULL, 4),
    ('mint', 'Mint', true, 'herb', 'Aggressive roots — contain in pots; modeled on basil', 'coco / DWC', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('parsley', 'Parsley', true, 'herb', 'Biennial herb baseline; slightly cooler than basil', 'coco / rockwool', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('blueberry', 'Blueberry', true, 'fruiting', 'Acidic pH band 4.5–5.5; modeled on strawberry', 'peat / coco', 'pulse_dryback', '10-15', NULL, NULL, NULL, 4),
    ('hemp', 'Hemp (fiber/seed)', true, 'industrial', 'Separate from cannabis flower profile — vegetative fiber/seed baseline only', 'coco / rockwool', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('broccoli', 'Broccoli', true, 'leafy', 'Cool brassica; modeled on kale with lower temps', 'NFT / rockwool', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('melon', 'Melon / cantaloupe', true, 'fruiting', 'Warm vining melon; high transpiration; modeled on cucumber', 'coco / rockwool', 'pulse_dryback', '10-20', NULL, NULL, NULL, 4),
    ('arugula', 'Arugula / rocket', true, 'leafy', 'Fast turnover leafy; bolts in heat; modeled on lettuce', 'DWC / NFT', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('rose', 'Rose (cut flower)', true, 'flower', 'Cut-flower rose; moderate EC; long photoperiod', 'rockwool / coco', 'pulse_dryback', '10-15', NULL, NULL, NULL, 4),
    ('sunflower', 'Sunflower', true, 'flower', 'Short-cycle high-light flower; fast turnover', 'coco / rockwool', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('hops', 'Hops (bines)', true, 'industrial', 'Long vegetative bines — not cannabis flower profile', 'coco / rockwool', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('succulent', 'Succulents (general)', true, 'ornamental', 'Dry-down epiphyte/soilless — never constant wet', 'gritty_mix', 'dry_down_succulent', NULL, NULL, NULL, NULL, 4),
    ('san_pedro', 'San Pedro cactus', true, 'ornamental', 'Columnar cactus — minimal EC; dry winter rest', 'gritty_mix', 'dry_down_succulent', NULL, NULL, NULL, NULL, 4),
    ('houseplant', 'Houseplant (general)', true, 'ornamental', 'Conservative foliage baseline — many species; clone to customize', 'peat / soilless_mix', 'top_water_drydown', NULL, NULL, NULL, NULL, 4),
    ('chard', 'Swiss chard', true, 'leafy', 'Leafy beet relative; kale-class EC', 'DWC / NFT', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('bok_choy', 'Bok choy / pak choi', true, 'leafy', 'Cool Asian brassica; fast head crop', 'NFT / rockwool', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('radish', 'Radish', true, 'leafy', 'Fast root/leaf crop; low EC; 3–4 week turnover', 'DWC / NFT', 'constant_feed', NULL, NULL, NULL, NULL, 4),
    ('thyme', 'Thyme', true, 'herb', 'Woody herb; lower EC; lean dry-down between feeds', 'coco / rockwool', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('oregano', 'Oregano', true, 'herb', 'Mediterranean herb; moderate EC; drier than basil', 'coco / rockwool', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('rosemary', 'Rosemary', true, 'herb', 'Woody Mediterranean herb; lean feed; excellent drainage', 'coco / gritty_mix', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('lavender', 'Lavender', true, 'herb', 'Dry-lean aromatic; low EC; alkaline-leaning ok', 'gritty_mix / coco', 'dry_down_succulent', NULL, NULL, NULL, NULL, 4),
    ('chrysanthemum', 'Chrysanthemum (mum)', true, 'flower', 'Photoperiod-sensitive cut flower; moderate EC', 'rockwool / coco', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('apple', 'Apple (nursery / young tree)', true, 'fruit_tree', 'Container or greenhouse nursery — multi-year to bearing; not full orchard automation', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('citrus', 'Citrus (lemon / orange nursery)', true, 'fruit_tree', 'Warm greenhouse citrus nursery; container citrus production', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('fig', 'Fig (container)', true, 'fruit_tree', 'Warm container fig; breba/main crop cultivar-dependent', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('peach', 'Peach / nectarine (nursery)', true, 'fruit_tree', 'Deciduous stone fruit nursery; chill hours required', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('cherry', 'Cherry (nursery / sweet)', true, 'fruit_tree', 'Sweet cherry nursery; high light; chill dependent', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('grape', 'Grape (vine / nursery)', true, 'fruit_tree', 'Container or greenhouse vine nursery; trellis from year 1', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('avocado', 'Avocado (container nursery)', true, 'fruit_tree', 'Sensitive roots; long juvenile phase; warm greenhouse', 'coarse soilless_mix', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('pear', 'Pear (nursery)', true, 'fruit_tree', 'Deciduous nursery pear; similar to apple bench culture', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('plum', 'Plum (nursery / stone fruit)', true, 'fruit_tree', 'Stone fruit nursery; similar to peach', 'soilless_mix / large container', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('mango', 'Mango (container nursery)', true, 'fruit_tree', 'Tropical greenhouse nursery; warm only', 'coarse soilless_mix', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('ramps', 'Ramps (wild leek)', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Woodland spring ephemeral — not indoor fertigation; foraged crop, not bench automation', 4),
    ('mushroom', 'Mushroom / fungi', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Different production domain — substrate bags and humidity rooms; use husbandry module, not fertigation profiles', 4),
    ('in_ground_root', 'In-ground root crops (carrot / potato)', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Field and deep-soil crops — gr33n targets hydroponic and container; no bench EC curve for tubers or taproots', 4),
    ('ginseng', 'Ginseng (woodland medicinal)', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'Multi-year shaded woodland medicinal — not fertigation automation', 4)
) AS v(crop_key, display_name, supported, category, source, substrate, watering_style,
         runoff_pct_target, moisture_guidance, cousin_of, unsupported_reason, catalog_version)
ON CONFLICT (crop_key) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    supported = EXCLUDED.supported,
    category = EXCLUDED.category,
    source = EXCLUDED.source,
    substrate = EXCLUDED.substrate,
    watering_style = EXCLUDED.watering_style,
    runoff_pct_target = EXCLUDED.runoff_pct_target,
    moisture_guidance = EXCLUDED.moisture_guidance,
    unsupported_reason = EXCLUDED.unsupported_reason,
    catalog_version = EXCLUDED.catalog_version,
    updated_at = NOW();

-- cousin_of FK (second pass)
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'pepper';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'eggplant';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'cucumber';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'kale';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'spinach';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'basil' WHERE crop_key = 'cilantro';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'microgreens';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'cucumber' WHERE crop_key = 'zucchini';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'pepper' WHERE crop_key = 'green_bean';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'basil' WHERE crop_key = 'mint';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'basil' WHERE crop_key = 'parsley';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'strawberry' WHERE crop_key = 'blueberry';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'cannabis' WHERE crop_key = 'hemp';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'kale' WHERE crop_key = 'broccoli';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'cucumber' WHERE crop_key = 'melon';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'arugula';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'rose';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'sunflower';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'cannabis' WHERE crop_key = 'hops';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'succulent' WHERE crop_key = 'san_pedro';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'kale' WHERE crop_key = 'chard';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'kale' WHERE crop_key = 'bok_choy';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'arugula' WHERE crop_key = 'radish';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'oregano' WHERE crop_key = 'thyme';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'basil' WHERE crop_key = 'oregano';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'thyme' WHERE crop_key = 'rosemary';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'rosemary' WHERE crop_key = 'lavender';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'rose' WHERE crop_key = 'chrysanthemum';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'apple';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'citrus';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'fig';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'peach';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'cherry';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'tomato' WHERE crop_key = 'grape';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'citrus' WHERE crop_key = 'avocado';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'apple' WHERE crop_key = 'pear';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'peach' WHERE crop_key = 'plum';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'citrus' WHERE crop_key = 'mango';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'in_ground_root';

INSERT INTO gr33ncrops.crop_catalog_aliases (alias, crop_key)
SELECT v.alias, v.crop_key
FROM (VALUES
    ('allium_tricoccum', 'ramps'),
    ('apple_tree', 'apple'),
    ('aubergine', 'eggplant'),
    ('beefsteak_tomato', 'tomato'),
    ('carrot', 'in_ground_root'),
    ('cherry_tomato', 'tomato'),
    ('chrysanthemum_mum', 'chrysanthemum'),
    ('coriander', 'cilantro'),
    ('echinopsis', 'san_pedro'),
    ('fungi', 'mushroom'),
    ('grape_vine', 'grape'),
    ('grapevine', 'grape'),
    ('lemon', 'citrus'),
    ('lime', 'citrus'),
    ('mandarin', 'citrus'),
    ('marijuana', 'cannabis'),
    ('mum', 'chrysanthemum'),
    ('nectarine', 'peach'),
    ('orange', 'citrus'),
    ('orchid', 'phalaenopsis'),
    ('pak_choi', 'bok_choy'),
    ('pak_choy', 'bok_choy'),
    ('panax', 'ginseng'),
    ('potato', 'in_ground_root'),
    ('shiitake', 'mushroom'),
    ('sweet_potato', 'in_ground_root'),
    ('swiss_chard', 'chard'),
    ('trichocereus', 'san_pedro'),
    ('weed', 'cannabis'),
    ('wild_leek', 'ramps')
) AS v(alias, crop_key)
ON CONFLICT (alias) DO UPDATE SET crop_key = EXCLUDED.crop_key;

INSERT INTO gr33ncrops.agronomy_field_guides (
    slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order
)
SELECT v.slug, v.title, v.crop_key, v.guide_kind, v.domain, v.safety_tier, v.body_md, v.catalog_version, v.published, v.sort_order
FROM (VALUES
    ('pi-wiring-basics', 'Pi wiring basics', NULL, 'trades', 'pi', 'safe', $fg_pi_wiring_basics$# Pi wiring basics (low-voltage control)

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
3. Confirm polarity on sensors that label VCC/GND/data.$fg_pi_wiring_basics$, 4, TRUE, 0),
    ('relay-and-actuator-wiring', 'Relay and actuator wiring', NULL, 'trades', 'actuator', 'caution', $fg_relay_and_actuator_wiring$# Relay and actuator wiring

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
4. Load side installed by a qualified person and breaker on?$fg_relay_and_actuator_wiring$, 4, TRUE, 1),
    ('sensor-install-and-calibration', 'Sensor install and calibration', NULL, 'trades', 'sensor', 'safe', $fg_sensor_install_and_calibration$# Sensor install and calibration

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

See `field-troubleshooting.md` and run the **diagnose-sensor-no-reading** procedure when available.$fg_sensor_install_and_calibration$, 4, TRUE, 2),
    ('irrigation-and-plumbing-basics', 'Irrigation and plumbing basics', NULL, 'trades', 'plumbing', 'caution', $fg_irrigation_and_plumbing_basics$# Irrigation and plumbing basics

## Low-risk work (operators)

- **Drip tubing**, barbed fittings, and reservoir **gravity** lines at atmospheric pressure.
- Check for leaks at push-fit joints before leaving the site.
- Label **nutrient** vs **plain water** lines.

## Escalate to a qualified plumber

- **Pressurized** municipal or well lines, backflow preventers, and potable water tie-ins.
- Solenoid valves on **house pressure** — sizing and code are local.

Guardian will **not** give step-by-step pressurized or potable plumbing instructions.

## Pump + reservoir (typical gr33n site)

- Pump on relay or contactor — **control wire** from Pi is low voltage; **pump power** is often mains (electrician).
- Keep reservoir fill strainers clean; air locks show as “no flow” with a running pump.$fg_irrigation_and_plumbing_basics$, 4, TRUE, 3),
    ('electrical-safety', 'Electrical safety', NULL, 'trades', 'electrical', 'caution', $fg_electrical_safety$# Electrical safety for field installers

## What a non-electrician may do

- Wire **Pi GPIO**, **3.3 V / 5 V** sensor leads, and relay **IN/VCC/GND** control pins.
- Verify **power is off** before touching screw terminals on the **control** side.
- Take photos and notes for your electrician.

## What requires a qualified electrician

- **Mains AC** (120 V / 240 V) terminations, panel work, GFCI/AFCI, and conduit.
- Hard-wiring grow lights, contactors, or VFDs.
- Any work you are not licensed to perform in your jurisdiction.

## Lockout

- Unplug equipment or switch off the correct breaker; use a lock/tag when multiple people work the room.
- Never bypass breakers or fuse with wire.

Guardian **stops** if you ask for step-by-step mains wiring — it will tell you to stop and call a licensed electrician.$fg_electrical_safety$, 4, TRUE, 4),
    ('field-troubleshooting', 'Field troubleshooting', NULL, 'trades', 'general', 'safe', $fg_field_troubleshooting$# Field troubleshooting (symptom → checks)

| Symptom | First checks |
|---------|----------------|
| Sensor reads nothing | Pi power LED; 3-wire pinout; GPIO matches gr33n; sensor power |
| Actuator won't fire | Pi online; `pending_command`; relay IN pin; mains side by electrician |
| Pi offline in gr33n | Network/API key; `farm_id` in client env; offline queue backlog |
| Wrong zone data | Client `farm_id`; device registered to correct farm |
| Feed did not run | Program `schedule_id`; Pi/pump online; reservoir status; see `fertigation-troubleshooting.md` |
| EC/pH wrong after dose | `lookup_crop_targets`; last mix event; probe calibration; stage match |
| Grow light won't switch | `summarize_device_health`; relay channel vs `hardware_identifier`; demo map in `demo-farm-pi-layout.md` |

Use the live farm snapshot for device counts, program schedule posture, and unread alerts. Describe what you see — Guardian labels **operator-stated** facts separately from measurements.$fg_field_troubleshooting$, 4, TRUE, 5),
    ('fertigation-troubleshooting', 'Fertigation troubleshooting', NULL, 'trades', 'fertigation', 'safe', $fg_fertigation_troubleshooting$# Fertigation troubleshooting

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
- Writes (pause program, enqueue actuator test) require **propose → Confirm**.$fg_fertigation_troubleshooting$, 4, TRUE, 6),
    ('demo-farm-pi-layout', 'Demo farm pi layout', NULL, 'trades', 'pi', 'safe', $fg_demo_farm_pi_layout$# gr33n Demo Farm — edge device map (farm id 1 seed)

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
- Pi not in UI: `start procedure diagnose-pi-offline`$fg_demo_farm_pi_layout$, 4, TRUE, 7),
    ('crop-cannabis-nutrition', 'Cannabis nutrition (indoor hydro)', 'cannabis', 'crop_nutrition', 'general', 'safe', $fg_crop_cannabis_nutrition$# Cannabis nutrition (indoor hydro)

Cannabis is a **high-EC** crop relative to leafy greens. Targets ramp from ~1.0 mS/cm in early veg to ~1.6–2.0 mS/cm in mid-flower, then taper for flush.

**Why photoperiod matters:** 18/6 veg keeps plants vegetative; 12/12 triggers flower. Guardian cites structured targets from your assigned crop profile — never guess EC from memory.

**Common mistakes:** Lockout from pH drift (target 5.8–6.2 in hydro); running flower EC in veg (salt stress); skipping flush before harvest.$fg_crop_cannabis_nutrition$, 4, TRUE, 8),
    ('crop-tomato-nutrition', 'Tomato nutrition (fruiting hydro)', 'tomato', 'crop_nutrition', 'general', 'safe', $fg_crop_tomato_nutrition$# Tomato nutrition (fruiting hydro)

Tomatoes tolerate **much higher EC** than cannabis in fruiting — often 2.8–3.5 mS/cm when plants are loaded. Blossom-end rot is usually **calcium transport**, not only low EC — check VPD and irrigation consistency.

Guardian compares your readings to the tomato profile stage row assigned to the plant.$fg_crop_tomato_nutrition$, 4, TRUE, 9),
    ('crop-orchid-care', 'Phalaenopsis orchid care (epiphyte)', 'phalaenopsis', 'crop_care', 'general', 'safe', $fg_crop_orchid_care$# Phalaenopsis orchid care

Orchids are the **opposite** of high-EC fruiting crops: target EC often **0.4–0.8 mS/cm**, higher RH, lower DLI. Over-fertilizing burns roots quickly.

Use the phalaenopsis built-in profile for Guardian targets; narrative here explains *why* the numbers are low.$fg_crop_orchid_care$, 4, TRUE, 10),
    ('crop-pepper-nutrition', 'Pepper nutrition (fruiting hydro)', 'pepper', 'crop_nutrition', 'general', 'safe', $fg_crop_pepper_nutrition$# Pepper nutrition (fruiting hydro)

Peppers fruit at **moderate-high EC** — similar curve to tomato but **lower peak** (~2.2–3.0 mS/cm in heavy fruit). They want **warm** root zones; cold irrigation in winter rooms stalls set.

Guardian uses the pepper built-in profile stage row — not tomato peak EC by default.$fg_crop_pepper_nutrition$, 4, TRUE, 11),
    ('crop-lettuce-nutrition', 'Lettuce nutrition (leafy hydro)', 'lettuce', 'crop_nutrition', 'general', 'safe', $fg_crop_lettuce_nutrition$# Lettuce nutrition (leafy hydro)

Lettuce and leafy greens run **low EC** (~0.8–1.3 mS/cm) with **cool** temps and high turnover. Tip burn usually means EC or VPD too aggressive for cultivar — not always "more calcium."

Assign the lettuce profile in Start grow or Plants so Guardian cites structured mS/cm targets.$fg_crop_lettuce_nutrition$, 4, TRUE, 12),
    ('crop-basil-nutrition', 'Basil nutrition (warm herb)', 'basil', 'crop_nutrition', 'general', 'safe', $fg_crop_basil_nutrition$# Basil nutrition (warm herb)

Basil is a **warm-weather** continuous-harvest herb — EC ramps ~1.0–1.8 mS/cm through vegetative pulls. Cold irrigation or sub-18 °C nights stall growth and darken leaves.

Use the basil profile for Guardian feed targets; it is not interchangeable with cilantro or lettuce bands.$fg_crop_basil_nutrition$, 4, TRUE, 13),
    ('crop-strawberry-nutrition', 'Strawberry nutrition (day-neutral)', 'strawberry', 'crop_nutrition', 'general', 'safe', $fg_crop_strawberry_nutrition$# Strawberry nutrition (day-neutral)

Day-neutral strawberries run **moderate EC** (~1.0–2.0 mS/cm) with **shorter photoperiod** than tomato (often ~14 h). Crown health and consistent moisture matter as much as peak EC — oscillating dry/wet shrinks fruit size.

Guardian cites the strawberry profile; June-bearing vs day-neutral genetics may need a cloned farm override.$fg_crop_strawberry_nutrition$, 4, TRUE, 14),
    ('crop-eggplant-nutrition', 'Eggplant nutrition (fruiting hydro)', 'eggplant', 'crop_nutrition', 'general', 'safe', $fg_crop_eggplant_nutrition$# Eggplant nutrition (fruiting hydro)

Eggplant is **solanaceous** like tomato but runs roughly **10% lower EC** at fruiting — often 2.2–2.9 mS/cm when loaded. Indoors, flowers often need **hand pollination** (gentle vibration) — bees are rarely sufficient in sealed rooms.

Guardian cites mS/cm targets from the eggplant built-in profile; do not assume tomato numbers apply unchanged.$fg_crop_eggplant_nutrition$, 4, TRUE, 15),
    ('crop-cucumber-nutrition', 'Cucumber nutrition (vining fruiting)', 'cucumber', 'crop_nutrition', 'general', 'safe', $fg_crop_cucumber_nutrition$# Cucumber nutrition (vining fruiting)

Cucumbers are **high-transpiration vining** fruiting crops — EC and DLI sit in tomato territory but **humidity headroom is higher** (often 55–75% RH in veg). Long runs need trellis support and consistent irrigation; dry roots on hot days show as wilt fast.

Use the cucumber profile stage row for Guardian targets — never guess EC from cannabis or lettuce baselines.$fg_crop_cucumber_nutrition$, 4, TRUE, 16),
    ('crop-kale-nutrition', 'Kale nutrition (leafy hydro)', 'kale', 'crop_nutrition', 'general', 'safe', $fg_crop_kale_nutrition$# Kale nutrition (leafy hydro)

Kale sits **slightly above lettuce EC** (~1.0–1.5 mS/cm) and **tolerates cooler** root-zone temps than warm herbs. It is a brassica — watch for tip burn if EC climbs in summer rooms.

Guardian compares live readings to the kale profile assigned to the plant.$fg_crop_kale_nutrition$, 4, TRUE, 17),
    ('crop-spinach-nutrition', 'Spinach nutrition (cool leafy)', 'spinach', 'crop_nutrition', 'pi', 'safe', $fg_crop_spinach_nutrition$# Spinach nutrition (cool leafy)

Spinach is a **cool-season** leafy crop — low EC (~0.8–1.3 mS/cm) and prefers **15–20 °C** class temps. The main indoor failure mode is **bolting** (elongation, bitter leaves) when rooms run warm or daylength is long.

Harvest on schedule; do not treat spinach like basil heat tolerance.$fg_crop_spinach_nutrition$, 4, TRUE, 18),
    ('crop-cilantro-nutrition', 'Cilantro / coriander (herb)', 'cilantro', 'crop_nutrition', 'general', 'safe', $fg_crop_cilantro_nutrition$# Cilantro / coriander (herb)

Cilantro shares **moderate EC** with basil but prefers **cooler, shorter photoperiod** runs. Sustained heat triggers **bolting to seed** — succession planting beats fighting room temp.

Guardian cites the cilantro profile; "coriander" is the same crop for lookup purposes.$fg_crop_cilantro_nutrition$, 4, TRUE, 19),
    ('crop-microgreens-nutrition', 'Microgreens (fast tray crop)', 'microgreens', 'crop_nutrition', 'general', 'safe', $fg_crop_microgreens_nutrition$# Microgreens (fast tray crop)

Microgreens run **very low EC** (~0.4–0.85 mS/cm) on **10–14 day** cycles. Substrate stays **shallow** — top-water or mist, not deep flood-and-drain like fruiting hydro. Over-feeding causes salt crust and damping-off risk.

Targets are per tray turnover, not months-long veg stages.$fg_crop_microgreens_nutrition$, 4, TRUE, 20),
    ('crop-zucchini-nutrition', 'Zucchini nutrition (summer squash)', 'zucchini', 'crop_nutrition', 'general', 'safe', $fg_crop_zucchini_nutrition$# Zucchini nutrition (summer squash)

Summer squash follows a **cucumber-like EC curve** (~2.2–3.2 mS/cm in fruiting) with high humidity tolerance. Heavy fruit loads need trellis or slab support — wilt on hot afternoons is often transpiration, not always under-feeding.

Guardian cites mS/cm from the zucchini built-in profile.$fg_crop_zucchini_nutrition$, 4, TRUE, 21),
    ('crop-green-bean-nutrition', 'Green bean nutrition', 'green_bean', 'crop_nutrition', 'general', 'safe', $fg_crop_green_bean_nutrition$# Green bean nutrition

Green beans run **moderate EC** (~1.8–2.8 mS/cm) and want **warm** root zones — cold feed in winter rooms stalls pod set. Trellis pole or bush types both need consistent moisture during bloom.

Use the green bean profile for structured targets — not tomato peak EC.$fg_crop_green_bean_nutrition$, 4, TRUE, 22),
    ('crop-mint-nutrition', 'Mint (container herb)', 'mint', 'crop_nutrition', 'general', 'safe', $fg_crop_mint_nutrition$# Mint (container herb)

Mint matches **basil-class EC** (~1.0–1.8 mS/cm) but roots **spread aggressively** — isolate in pots or dedicated lines so it does not invade shared DWC or NFT channels.

Continuous harvest keeps plants vegetative; EC creep causes tip burn on tender leaves.$fg_crop_mint_nutrition$, 4, TRUE, 23),
    ('crop-parsley-nutrition', 'Parsley nutrition', 'parsley', 'crop_nutrition', 'general', 'safe', $fg_crop_parsley_nutrition$# Parsley nutrition

Parsley sits **slightly below basil EC** and prefers **cooler** temps (18–24 °C). Slow germination is normal — do not chase heat like warm-weather herbs.

Biennial plants may bolt in year two if rooms run hot; succession sow for steady supply.$fg_crop_parsley_nutrition$, 4, TRUE, 24),
    ('crop-blueberry-nutrition', 'Blueberry nutrition (acidic)', 'blueberry', 'crop_nutrition', 'general', 'safe', $fg_crop_blueberry_nutrition$# Blueberry nutrition (acidic)

Blueberries require an **acidic root zone — pH 4.5–5.5** — not the 5.5–6.0 band most fruiting hydro crops use. EC stays **moderate** (~1.0–1.9 mS/cm) similar to strawberry but lockout shows fast if pH drifts alkaline.

Monitor runoff pH weekly; peat/coco mixes buffer differently than pure rockwool.$fg_crop_blueberry_nutrition$, 4, TRUE, 25),
    ('crop-hemp-nutrition', 'Hemp (fiber/seed vegetative)', 'hemp', 'crop_nutrition', 'general', 'safe', $fg_crop_hemp_nutrition$# Hemp (fiber/seed vegetative)

This profile is **fiber and seed production** — long **18 h photoperiod** vegetative targets (~1.0–1.5 mS/cm). It is **not** the cannabis flower EC curve. Do not apply mid-flower cannabis targets to hemp fiber rooms.

For CBD/flower hemp genetics, clone and customize a farm profile from your agronomist.$fg_crop_hemp_nutrition$, 4, TRUE, 26),
    ('crop-broccoli-nutrition', 'Broccoli nutrition (cool brassica)', 'broccoli', 'crop_nutrition', 'general', 'safe', $fg_crop_broccoli_nutrition$# Broccoli nutrition (cool brassica)

Broccoli is a **cool brassica** — kale-class EC (~1.0–1.5 mS/cm) but **14–20 °C** is safer than warm-room lettuce runs. Heat pushes **premature bolting** and loose heads.

Harvest on schedule; extended warm veg wastes bench space.$fg_crop_broccoli_nutrition$, 4, TRUE, 27),
    ('crop-melon-nutrition', 'Melon nutrition (vining fruit)', 'melon', 'crop_nutrition', 'general', 'safe', $fg_crop_melon_nutrition$# Melon nutrition (vining fruit)

Melons combine **cucumber-like vining** with **very high transpiration** in fruiting — EC can reach ~2.9–3.3 mS/cm with warm temps (24–30 °C). Dry roots on hot days collapse vines quickly; consistency beats pulse starvation.

Guardian compares live readings to the melon profile stage row.$fg_crop_melon_nutrition$, 4, TRUE, 28),
    ('crop-arugula-nutrition', 'Arugula / rocket (fast leafy)', 'arugula', 'crop_nutrition', 'general', 'safe', $fg_crop_arugula_nutrition$# Arugula / rocket (fast leafy)

Arugula is a **fast turnover** crop (~3–4 weeks) at **lettuce-class EC** with **short photoperiod** (~12 h). Heat and long days trigger **bolt** and peppery bitterness — treat like a cool quick crop, not perpetual basil.

Harvest young leaves; succession plant for steady supply.$fg_crop_arugula_nutrition$, 4, TRUE, 29),
    ('crop-rose-care', 'Cut-flower rose care (moderate EC)', 'rose', 'crop_care', 'general', 'safe', $fg_crop_rose_care$# Cut-flower rose care (moderate EC)

Roses for stems need **steady moderate EC** — typically ~1.4–2.0 mS/cm through pull cycles — not fruiting-tomato peaks. Soft water and consistent irrigation matter more than chasing high salts.

Guardian compares your readings to the rose profile stage row assigned to the plant.$fg_crop_rose_care$, 4, TRUE, 30),
    ('crop-sunflower-nutrition', 'Sunflower nutrition (short cycle)', 'sunflower', 'crop_nutrition', 'general', 'safe', $fg_crop_sunflower_nutrition$# Sunflower nutrition (short cycle)

Sunflowers are a **short-cycle, high-light** crop — EC stays moderate (~1.2–1.8 mS/cm) because the run is fast. Under-lighting stretches stems; excess EC in a small root volume burns tips before harvest.

Guardian cites the sunflower profile for feed targets matched to your assigned growth stage.$fg_crop_sunflower_nutrition$, 4, TRUE, 31),
    ('crop-hops-nutrition', 'Hops nutrition (long veg bines)', 'hops', 'crop_nutrition', 'general', 'safe', $fg_crop_hops_nutrition$# Hops nutrition (long veg bines)

Hops spend **months in vegetative bine growth** before cone set — EC ramps gradually (~1.0–1.6 mS/cm), not cannabis-flower peaks. This is a tall trellis crop; root volume and nitrogen timing drive yield more than late salt push.

Use the hops profile in Guardian; do not copy indoor cannabis flower bands onto bines.$fg_crop_hops_nutrition$, 4, TRUE, 32),
    ('crop-succulent-care', 'Succulent care (dry-down low EC)', 'succulent', 'crop_care', 'general', 'safe', $fg_crop_succulent_care$# Succulent care (dry-down low EC)

Succulents want **dry-down cycles and low EC** — often ~0.3–0.8 mS/cm — with sharp drainage. Constant wet media at moderate EC invites rot faster than under-feeding.

Guardian compares live readings to the succulent profile assigned to the plant.$fg_crop_succulent_care$, 4, TRUE, 33),
    ('crop-san-pedro-care', 'San Pedro cactus care (minimal EC)', 'san_pedro', 'crop_care', 'general', 'safe', $fg_crop_san_pedro_care$# San Pedro cactus care (minimal EC)

San Pedro tolerates **minimal EC** — often ~0.2–0.6 mS/cm in active growth and near-zero feed through **winter rest**. Cold, dark dormancy plus salty irrigation is a common kill pattern.

Use the san pedro built-in profile for Guardian targets during each season row.$fg_crop_san_pedro_care$, 4, TRUE, 34),
    ('crop-houseplant-care', 'Houseplant care (conservative baseline)', 'houseplant', 'crop_care', 'general', 'safe', $fg_crop_houseplant_care$# Houseplant care (conservative baseline)

Most houseplants need a **conservative EC baseline** — often ~0.4–1.0 mS/cm — because light and root volume vary wildly indoors. Clone the profile per species; pothos bands are not ficus bands.

Guardian cites the species profile assigned to each plant, not a single generic houseplant target.$fg_crop_houseplant_care$, 4, TRUE, 35),
    ('crop-chard-nutrition', 'Swiss chard nutrition (kale-class)', 'chard', 'crop_nutrition', 'general', 'safe', $fg_crop_chard_nutrition$# Swiss chard nutrition (kale-class)

Chard sits in the **kale-class leafy band** — ~1.0–1.5 mS/cm — with continuous harvest pulls. It tolerates cooler roots than basil but still tip-burns if EC climbs in hot, low-airflow rooms.

Guardian compares your readings to the chard profile stage row assigned to the plant.$fg_crop_chard_nutrition$, 4, TRUE, 36),
    ('crop-bok-choy-nutrition', 'Bok choy nutrition (cool brassica)', 'bok_choy', 'crop_nutrition', 'general', 'safe', $fg_crop_bok_choy_nutrition$# Bok choy nutrition (cool brassica)

Bok choy is a **cool brassica** — target ~0.8–1.3 mS/cm and stable irrigation. Heat plus moderate EC pushes **premature bolting** and bitter leaves faster than salt stress alone.

Use the bok choy profile in Guardian; summer rooms may need lower EC and shorter cycles.$fg_crop_bok_choy_nutrition$, 4, TRUE, 37),
    ('crop-radish-nutrition', 'Radish nutrition (fast low EC)', 'radish', 'crop_nutrition', 'general', 'safe', $fg_crop_radish_nutrition$# Radish nutrition (fast low EC)

Radishes finish in weeks — keep EC **low to moderate** (~0.8–1.2 mS/cm) because roots are the crop. High EC in shallow media woody roots and splits bulbs before size targets.

Guardian cites the radish profile for feed targets matched to your short cycle stage.$fg_crop_radish_nutrition$, 4, TRUE, 38),
    ('crop-thyme-nutrition', 'Thyme nutrition (woody lean feed)', 'thyme', 'crop_nutrition', 'general', 'safe', $fg_crop_thyme_nutrition$# Thyme nutrition (woody lean feed)

Thyme is a **woody herb** that prefers **lean feed** — ~0.8–1.2 mS/cm — with excellent drainage. Over-fertilizing softens stems and dulls flavor oils compared with slightly stressed plants.

Guardian compares live readings to the thyme profile assigned to the plant.$fg_crop_thyme_nutrition$, 4, TRUE, 39),
    ('crop-oregano-nutrition', 'Oregano nutrition (Mediterranean lean)', 'oregano', 'crop_nutrition', 'general', 'safe', $fg_crop_oregano_nutrition$# Oregano nutrition (Mediterranean lean)

Oregano wants **drier, leaner feeding** than basil — ~0.8–1.4 mS/cm — with dry-down between irrigations. Wet, salty media produces lush leaves but weak aroma and more root issues.

Use the oregano profile in Guardian; do not swap in basil EC bands.$fg_crop_oregano_nutrition$, 4, TRUE, 40),
    ('crop-rosemary-nutrition', 'Rosemary nutrition (drainage first)', 'rosemary', 'crop_nutrition', 'general', 'safe', $fg_crop_rosemary_nutrition$# Rosemary nutrition (drainage first)

Rosemary fails from **root rot** before nutrient deficiency — lean EC ~0.8–1.2 mS/cm, fast drainage, and dry-down matter more than pushing salts. Constant wet feet at moderate EC kills plants quickly.

Guardian cites the rosemary profile for targets; fix drainage before raising EC.$fg_crop_rosemary_nutrition$, 4, TRUE, 41),
    ('crop-lavender-nutrition', 'Lavender nutrition (dry lean aromatic)', 'lavender', 'crop_nutrition', 'general', 'safe', $fg_crop_lavender_nutrition$# Lavender nutrition (dry lean aromatic)

Lavender is **dry, lean, and aromatic** — target ~0.6–1.0 mS/cm with infrequent irrigation. High EC plus humidity collapses oil quality and invites crown rot in dense media.

Guardian compares your readings to the lavender profile stage row assigned to the plant.$fg_crop_lavender_nutrition$, 4, TRUE, 42),
    ('crop-chrysanthemum-care', 'Chrysanthemum care (short-day bloom)', 'chrysanthemum', 'crop_care', 'general', 'safe', $fg_crop_chrysanthemum_care$# Chrysanthemum care (short-day bloom)

Mums are **photoperiod crops** — short days trigger bloom; long days keep vegetative growth. EC stays moderate (~1.2–1.8 mS/cm) through veg and bud, not fruiting-tomato peaks.

Guardian cites the chrysanthemum profile for stage targets tied to your photoperiod schedule.$fg_crop_chrysanthemum_care$, 4, TRUE, 43),
    ('crop-apple-nursery', 'Apple nursery (container / bench)', 'apple', 'crop_nutrition', 'general', 'safe', $fg_crop_apple_nursery$# Apple nursery (container / bench)

Young apple trees in **greenhouse nursery or large containers** — not full orchard automation. EC stays **moderate (≈1.0–1.6 mS/cm)** in early years; focus on **winter chill** for the cultivar and **dry-back** between pulses to avoid crown rot.

Fruit on bench-scale trees typically **years 3–5+**. Train central leader early; increase light (DLI) as wood hardens.$fg_crop_apple_nursery$, 4, TRUE, 44),
    ('crop-citrus-nursery', 'Citrus nursery (lemon / orange)', 'citrus', 'crop_nutrition', 'general', 'safe', $fg_crop_citrus_nursery$# Citrus nursery (lemon / orange)

Container citrus needs a **warm root zone** — never cold irrigation water. Keep pH **5.5–6.2** and watch **iron deficiency** if pH drifts high. EC **≈1.0–1.6 mS/cm** for young trees.

Fruit in pots often **year 2–4** depending on cultivar and graft. Reduce feed slightly if new flush is weak after repotting.$fg_crop_citrus_nursery$, 4, TRUE, 45),
    ('crop-fig-container', 'Fig (container)', 'fig', 'crop_nutrition', 'general', 'safe', $fg_crop_fig_container$# Fig (container)

Figs tolerate **dry-down** between feeds — soggy roots invite collapse. EC **≈0.9–1.5 mS/cm** for container culture; warm temps accelerate growth.

Breba vs main crop timing is **cultivar-dependent**. If trees go dormant, cut feed and let substrate dry more between irrigations.$fg_crop_fig_container$, 4, TRUE, 46),
    ('crop-peach-nursery', 'Peach / nectarine nursery', 'peach', 'crop_nutrition', 'general', 'safe', $fg_crop_peach_nursery$# Peach / nectarine nursery

Stone fruit nursery stock needs **chill hours** for the cultivar and **high light** during spring flush. EC **≈1.0–1.6 mS/cm**; watch **bacterial spot** when RH stays high on wet foliage.

Nectarine shares the same bench targets as peach. Fruiting in containers often **years 2–4**.$fg_crop_peach_nursery$, 4, TRUE, 47),
    ('crop-cherry-nursery', 'Cherry nursery (sweet)', 'cherry', 'crop_nutrition', 'general', 'safe', $fg_crop_cherry_nursery$# Cherry nursery (sweet)

Sweet cherry liners need **cool starts** and **high DLI** for quality wood. EC stays **moderate (≈1.0–1.5 mS/cm)** — avoid lush weak growth from over-feeding.

Rain at harvest causes **cracking** outdoors; in greenhouse, keep VPD stable during fruit swell if you push early crops.$fg_crop_cherry_nursery$, 4, TRUE, 48),
    ('crop-grape-vine', 'Grape (vine / nursery)', 'grape', 'crop_nutrition', 'general', 'safe', $fg_crop_grape_vine$# Grape (vine / nursery)

Greenhouse or container **vine nursery** — trellis from year one. EC **≈1.0–1.6 mS/cm** during cane training; increase light as bines lengthen.

First meaningful crop often **year 2–3** on greenhouse vines. Alias terms: grapevine, grape vine.$fg_crop_grape_vine$, 4, TRUE, 49),
    ('crop-avocado-nursery', 'Avocado (container nursery)', 'avocado', 'crop_nutrition', 'general', 'safe', $fg_crop_avocado_nursery$# Avocado (container nursery)

Avocado roots are **sensitive to waterlogging** — pulse to dry-back on coarse mix. EC **≈0.8–1.4 mS/cm**; **chloride-sensitive** — avoid salty source water.

Juvenile phase is long; fruit in pots may take **several years**. Graft union must stay above the substrate line.$fg_crop_avocado_nursery$, 4, TRUE, 50),
    ('crop-pear-nursery', 'Pear nursery', 'pear', 'crop_nutrition', 'general', 'safe', $fg_crop_pear_nursery$# Pear nursery

Bench pear culture mirrors **apple nursery** targets — moderate EC, good airflow. **Fire blight** risk rises with dense canopy and prolonged leaf wetness.

Train scaffolds early in containers; fruiting years similar to apple on bench scale.$fg_crop_pear_nursery$, 4, TRUE, 51),
    ('crop-plum-nursery', 'Plum nursery (stone fruit)', 'plum', 'crop_nutrition', 'general', 'safe', $fg_crop_plum_nursery$# Plum nursery (stone fruit)

Plum nursery stock follows **peach-class** feed and light — chill-dependent, spring flush management. EC **≈1.0–1.5 mS/cm** for young trees.

Fruit in containers typically **years 3–5**. Thin heavy sets on small trees to avoid branch break.$fg_crop_plum_nursery$, 4, TRUE, 52),
    ('crop-mango-nursery', 'Mango (container nursery)', 'mango', 'crop_nutrition', 'general', 'safe', $fg_crop_mango_nursery$# Mango (container nursery)

**Tropical warm-only** — cold roots stop growth fast. EC **≈1.0–1.6 mS/cm**; watch **anthracnose** when RH stays high on flush.

Juvenile mangoes in pots may fruit **years 3–5+**. Never ship cold-stressed liners into warm zones without acclimation.$fg_crop_mango_nursery$, 4, TRUE, 53),
    ('crop-unsupported-woodland', 'Unsupported woodland crops (ramps, ginseng)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_woodland$# Unsupported woodland crops (ramps, ginseng)

**Ramps** (wild leek) and **ginseng** are woodland ephemerals or multi-year shade medicinals — not indoor fertigation crops. gr33n does not ship EC, VPD, or photoperiod targets for them.

If an operator asks about bench automation, explain honestly: these are foraged or long-cycle outdoor/forest production. For general greenhouse questions, suggest a supported **leafy** or **herb** cousin only when they want a hydro starting point — never invent woodland feed schedules.$fg_crop_unsupported_woodland$, 4, TRUE, 54),
    ('crop-unsupported-mushroom', 'Mushroom production (unsupported fertigation profile)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_mushroom$# Mushroom production (unsupported fertigation profile)

Mushrooms and other **fungi** use bag/substrate colonization and humidity rooms — a different domain from plant EC/VPD profiles. gr33n crop targets do not apply.

Direct operators to husbandry / substrate workflows when available. Do not map shiitake or other fungi to cannabis or tomato nutrient curves.$fg_crop_unsupported_mushroom$, 4, TRUE, 55),
    ('crop-unsupported-field-roots', 'In-ground root crops (carrot, potato)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_field_roots$# In-ground root crops (carrot, potato)

**Carrots, potatoes, and sweet potatoes** are field or deep-container taproot/tuber crops. gr33n structured targets cover hydroponic and bench container production — not deep soil beds or field scale.

If an operator wants indoor hydro only, suggest cloning from **lettuce** (fast leafy baseline) or **tomato** (fruiting hydro) and adjusting manually — do not state fake EC targets for tubers.$fg_crop_unsupported_field_roots$, 4, TRUE, 56)
) AS v(slug, title, crop_key, guide_kind, domain, safety_tier, body_md, catalog_version, published, sort_order)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    crop_key = EXCLUDED.crop_key,
    guide_kind = EXCLUDED.guide_kind,
    domain = EXCLUDED.domain,
    safety_tier = EXCLUDED.safety_tier,
    body_md = EXCLUDED.body_md,
    catalog_version = EXCLUDED.catalog_version,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

