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
    ('marigold', 'Marigold (bedding)', true, 'flower', 'Short-cycle bedding flower; outdoor and bench', 'peat / coco / field', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
    ('geranium', 'Geranium (zonal bedding)', true, 'flower', 'Sun-loving pot and bench bedding; drainage critical', 'peat / coco', 'pulse_dryback', NULL, NULL, NULL, NULL, 4),
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
    ('rice', 'Rice (aquaponics / shallow water)', true, 'grain', 'Warm shallow-water grain; aquaponics raft or paddy tray', 'aquaponics raft / shallow tray', 'constant_feed', NULL, NULL, NULL, NULL, 4),
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
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'sunflower' WHERE crop_key = 'marigold';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'marigold' WHERE crop_key = 'geranium';
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
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'rice';
UPDATE gr33ncrops.crop_catalog_entries SET cousin_of = 'lettuce' WHERE crop_key = 'in_ground_root';

INSERT INTO gr33ncrops.crop_catalog_aliases (alias, crop_key)
SELECT v.alias, v.crop_key
FROM (VALUES
    ('allium_tricoccum', 'ramps'),
    ('apple_tree', 'apple'),
    ('aubergine', 'eggplant'),
    ('basmati', 'rice'),
    ('beefsteak_tomato', 'tomato'),
    ('carrot', 'in_ground_root'),
    ('cherry_tomato', 'tomato'),
    ('chrysanthemum_mum', 'chrysanthemum'),
    ('coriander', 'cilantro'),
    ('echinopsis', 'san_pedro'),
    ('ficus', 'houseplant'),
    ('fungi', 'mushroom'),
    ('grape_vine', 'grape'),
    ('grapevine', 'grape'),
    ('jasmine_rice', 'rice'),
    ('lemon', 'citrus'),
    ('lime', 'citrus'),
    ('mandarin', 'citrus'),
    ('marijuana', 'cannabis'),
    ('monstera', 'houseplant'),
    ('mum', 'chrysanthemum'),
    ('nectarine', 'peach'),
    ('orange', 'citrus'),
    ('orchid', 'phalaenopsis'),
    ('paddy', 'rice'),
    ('pak_choi', 'bok_choy'),
    ('pak_choy', 'bok_choy'),
    ('panax', 'ginseng'),
    ('pelargonium', 'geranium'),
    ('philodendron', 'houseplant'),
    ('potato', 'in_ground_root'),
    ('pothos', 'houseplant'),
    ('shiitake', 'mushroom'),
    ('sweet_potato', 'in_ground_root'),
    ('swiss_chard', 'chard'),
    ('tagetes', 'marigold'),
    ('trichocereus', 'san_pedro'),
    ('weed', 'cannabis'),
    ('wild_leek', 'ramps'),
    ('zonal_geranium', 'geranium')
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

Peppers fruit at **moderate-high EC** — similar curve to tomato but **lower peak** (~2.2–3.0 mS/cm in heavy fruit). They want **warm** root zones; cold irrigation in spring outdoor beds stalls set.

## Feed

| Stage | EC target (mS/cm) | Notes |
|-------|-------------------|--------|
| Early veg | ~1.6 | Establish canopy |
| Late veg | ~2.0 | First flowers |
| Early flower | ~2.4 | Fruit set |
| Late flower | ~2.8 | Peak fruit load |

pH 5.5–6.0. Outdoor raised beds with drip: match run length to soil moisture — sandy mixes need shorter, more frequent pulses than heavy loam.

## Climate

- **Temp:** 21–28 °C day for fruit set; below 15 °C nights drop fruit.
- **RH:** 50–60% in fruiting — very high RH reduces pollen viability.
- **Light:** Full sun outdoors; greenhouse peppers need high DLI.

## Outdoor bed notes

- Mulch reduces oscillating dry/wet that causes blossom drop.
- Calcium is mobile — inconsistent watering shows as blossom-end rot on fruit (like tomato).
- Watch for aphids and spider mites on stressed, dusty foliage.

Guardian uses the pepper built-in profile stage row — not tomato peak EC by default.$fg_crop_pepper_nutrition$, 4, TRUE, 11),
    ('crop-lettuce-nutrition', 'Lettuce nutrition (leafy hydro)', 'lettuce', 'crop_nutrition', 'general', 'safe', $fg_crop_lettuce_nutrition$# Lettuce nutrition (leafy hydro)

Lettuce and leafy greens run **low EC** (~0.8–1.3 mS/cm) with **cool** temps and high turnover. Tip burn usually means EC or VPD too aggressive for cultivar — not always "more calcium."

Assign the lettuce profile in Start grow or Plants so Guardian cites structured mS/cm targets.$fg_crop_lettuce_nutrition$, 4, TRUE, 12),
    ('crop-basil-nutrition', 'Basil nutrition (warm herb)', 'basil', 'crop_nutrition', 'general', 'safe', $fg_crop_basil_nutrition$# Basil nutrition (warm herb)

Basil is a **warm-weather** continuous-harvest herb — EC ramps ~1.0–1.8 mS/cm through vegetative pulls. Cold irrigation or sub-18 °C nights stall growth and darken leaves.

## Feed

| Stage | EC target (mS/cm) | Notes |
|-------|-------------------|--------|
| Seedling | ~1.0 | Warm, humid dome |
| Early veg | ~1.4 | First harvest cuts |
| Late veg | ~1.6 | Continuous harvest |

pH 5.5–6.0. Basil tolerates **constant feed** or short pulse cycles; aim 10–15% runoff if you measure it.

## Climate

- **Temp:** 22–28 °C day — below 18 °C nights cause chilling injury.
- **RH:** 55–65% in veg; avoid saturated canopy overnight.
- **Light:** 16 h photoperiod, DLI ~20–22 for bench herbs.

## Gravity drip / plain water

Perpetual herb beds often run **plain irrigation** (no reservoir mix) on a morning schedule — enough to rewet the root zone without flooding. If leaves cup or roots smell sour, shorten run time or add dry-back between feeds.

## Common issues

- **Blackening / downy mildew:** cool, wet nights — warm the zone and improve airflow.
- **Tip burn:** EC too high for small roots or inconsistent irrigation.
- **Flowering:** long days + stress — pinch flower spikes on culinary lines.

Guardian uses the basil profile for feed targets; it is not interchangeable with cilantro or lettuce bands.$fg_crop_basil_nutrition$, 4, TRUE, 13),
    ('crop-strawberry-nutrition', 'Strawberry nutrition (day-neutral)', 'strawberry', 'crop_nutrition', 'general', 'safe', $fg_crop_strawberry_nutrition$# Strawberry nutrition (day-neutral)

Day-neutral strawberries run **moderate EC** (~1.0–2.0 mS/cm) with **shorter photoperiod** than tomato (often ~14 h). Crown health and consistent moisture matter as much as peak EC — oscillating dry/wet shrinks fruit size.

## Feed

| Stage | EC target (mS/cm) | Notes |
|-------|-------------------|--------|
| Early veg | ~1.2 | Crown establishment |
| Late veg | ~1.6 | Runner control |
| Early flower | ~1.8 | First fruit |
| Late flower | ~2.0 | Peak harvest |

pH 5.5–6.0. Drip on perennial matted rows: keep crown **above** the wet line — buried crowns rot.

## Climate

- **Temp:** 18–24 °C — heat above 30 °C shrinks fruit and pauses flowering.
- **RH:** 50–60% — botrytis risk in dense, wet canopy.
- **Light:** ~14 h photoperiod for day-neutral lines in controlled environments.

## Perennial patch

- Renovate beds after peak season — thin old crowns, refresh drip emitters.
- Gray mold (botrytis) on ripe fruit: pick frequently, improve airflow, avoid overhead irrigation.

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

Mums are **photoperiod crops** — uninterrupted short nights trigger bloom; long days keep vegetative growth. The demo farm runs **18/6 veg** and **12/12 bloom**, matching standard cut-flower practice.

## Photoperiod

- **Veg (long day):** 16–18 h light — vegetative growth, cuttings root, branches fill out.
- **Bloom (short day):** 12 h light / 12 h dark — **no light leaks** during the dark period or buds revert or stall.
- **Propagation:** tip cuttings under dome, 24 h light until roots show — then move to veg photoperiod.

## Feed and EC

Moderate EC through the run — not tomato fruiting peaks:

| Stage | EC target (mS/cm) | Notes |
|-------|-------------------|--------|
| Early veg | ~1.6 | Long-day growth |
| Late veg | ~2.0 | Pinch for branchiness before flip |
| Early bloom | ~2.2 | Short-day — photoperiod is the main driver |

pH 5.5–6.0 on rockwool or coco. Pulse irrigation with dry-back between runs; constant wet feet invite root issues.

## Climate

- **Veg:** 18–24 °C, RH 55–65%, moderate VPD.
- **Bloom:** Slightly cooler (16–22 °C), **lower RH 45–55%** — high humidity in dense canopy invites powdery mildew (common on mums and roses in bloom).

If humidity reads above band (e.g. 72% RH in bloom), improve airflow, shorten irrigation, or run dehumidification before chasing feed changes.

## Harvest cues

Check **bloom openness and stem length** — not trichomes. Cut when outer petals show color and stems are firm enough to ship. Over-mature blooms shatter in the cooler.

## Guardian

Guardian cites the chrysanthemum profile for stage targets tied to your photoperiod schedule and compares live sensor readings to the bloom-stage humidity band.$fg_crop_chrysanthemum_care$, 4, TRUE, 43),
    ('crop-marigold-care', 'Marigold care (bedding flower)', 'marigold', 'crop_care', 'general', 'safe', $fg_crop_marigold_care$# Marigold care (bedding flower)

French and African marigolds are **short-cycle bedding flowers** — moderate EC, high light, fast turnover. They tolerate outdoor beds and greenhouse benches; primary failures are over-watering and low light stretch.

## Feed

EC ~1.2–1.8 mS/cm through vegetative growth; leaner than fruiting tomato. pH 5.8–6.2. Constant wet media causes root rot on young plugs.

## Climate

- **Temp:** 18–26 °C — marigolds handle warm days; frost kills outdoor plantings.
- **Light:** High DLI — under-lighting produces weak stems and delayed bloom.
- **RH:** 50–65% — avoid saturated overnight canopy.

## Outdoor use

Common as **companion rows** near vegetables — plant after last frost. Drip or hand water at the root line; wet foliage overnight encourages mildew on dense African types.

Guardian cites the marigold profile for stage targets when a plant row is assigned `crop_key: marigold`.$fg_crop_marigold_care$, 4, TRUE, 44),
    ('crop-geranium-care', 'Geranium care (bedding / pot)', 'geranium', 'crop_care', 'general', 'safe', $fg_crop_geranium_care$# Geranium care (bedding / pot)

Zonal geraniums (*Pelargonium*) are **sun-loving bedding plants** — moderate feed, excellent drainage, and dry-down between irrigation. They fail from root rot before salt deficiency.

## Feed

EC ~1.0–1.6 mS/cm — leaner than heavy fruiting crops. pH 5.8–6.2. Let the top of the root zone dry slightly between runs; geraniums hate constant saturation.

## Climate

- **Temp:** 18–24 °C day — cool nights (10–15 °C) improve flower color on many lines.
- **Light:** Full sun to very high DLI in greenhouse; shade stretches and reduces bloom count.
- **RH:** 45–55% — humid, stagnant benches invite botrytis on spent flowers.

## Bench and pot culture

- Remove spent bloom heads to redirect energy.
- Yellow lower leaves often mean **over-watering**, not nitrogen deficiency — check drainage first.

Guardian cites the geranium profile when plants use `crop_key: geranium`.$fg_crop_geranium_care$, 4, TRUE, 45),
    ('crop-apple-nursery', 'Apple nursery (container / bench)', 'apple', 'crop_nutrition', 'general', 'safe', $fg_crop_apple_nursery$# Apple nursery (container / bench)

Young apple trees in **greenhouse nursery or large containers** — not full orchard automation. EC stays **moderate (≈1.0–1.6 mS/cm)** in early years; focus on **winter chill** for the cultivar and **dry-back** between pulses to avoid crown rot.

Fruit on bench-scale trees typically **years 3–5+**. Train central leader early; increase light (DLI) as wood hardens.$fg_crop_apple_nursery$, 4, TRUE, 46),
    ('crop-citrus-nursery', 'Citrus nursery (lemon / orange)', 'citrus', 'crop_nutrition', 'general', 'safe', $fg_crop_citrus_nursery$# Citrus nursery (lemon / orange)

Container citrus needs a **warm root zone** — never cold irrigation water. Keep pH **5.5–6.2** and watch **iron deficiency** if pH drifts high. EC **≈1.0–1.6 mS/cm** for young trees.

Fruit in pots often **year 2–4** depending on cultivar and graft. Reduce feed slightly if new flush is weak after repotting.$fg_crop_citrus_nursery$, 4, TRUE, 47),
    ('crop-fig-container', 'Fig (container)', 'fig', 'crop_nutrition', 'general', 'safe', $fg_crop_fig_container$# Fig (container)

Figs tolerate **dry-down** between feeds — soggy roots invite collapse. EC **≈0.9–1.5 mS/cm** for container culture; warm temps accelerate growth.

Breba vs main crop timing is **cultivar-dependent**. If trees go dormant, cut feed and let substrate dry more between irrigations.$fg_crop_fig_container$, 4, TRUE, 48),
    ('crop-peach-nursery', 'Peach / nectarine nursery', 'peach', 'crop_nutrition', 'general', 'safe', $fg_crop_peach_nursery$# Peach / nectarine nursery

Stone fruit nursery stock needs **chill hours** for the cultivar and **high light** during spring flush. EC **≈1.0–1.6 mS/cm**; watch **bacterial spot** when RH stays high on wet foliage.

Nectarine shares the same bench targets as peach. Fruiting in containers often **years 2–4**.$fg_crop_peach_nursery$, 4, TRUE, 49),
    ('crop-cherry-nursery', 'Cherry nursery (sweet)', 'cherry', 'crop_nutrition', 'general', 'safe', $fg_crop_cherry_nursery$# Cherry nursery (sweet)

Sweet cherry liners need **cool starts** and **high DLI** for quality wood. EC stays **moderate (≈1.0–1.5 mS/cm)** — avoid lush weak growth from over-feeding.

Rain at harvest causes **cracking** outdoors; in greenhouse, keep VPD stable during fruit swell if you push early crops.$fg_crop_cherry_nursery$, 4, TRUE, 50),
    ('crop-grape-vine', 'Grape (vine / nursery)', 'grape', 'crop_nutrition', 'general', 'safe', $fg_crop_grape_vine$# Grape (vine / nursery)

Greenhouse or container **vine nursery** — trellis from year one. EC **≈1.0–1.6 mS/cm** during cane training; increase light as bines lengthen.

First meaningful crop often **year 2–3** on greenhouse vines. Alias terms: grapevine, grape vine.$fg_crop_grape_vine$, 4, TRUE, 51),
    ('crop-avocado-nursery', 'Avocado (container nursery)', 'avocado', 'crop_nutrition', 'general', 'safe', $fg_crop_avocado_nursery$# Avocado (container nursery)

Avocado roots are **sensitive to waterlogging** — pulse to dry-back on coarse mix. EC **≈0.8–1.4 mS/cm**; **chloride-sensitive** — avoid salty source water.

Juvenile phase is long; fruit in pots may take **several years**. Graft union must stay above the substrate line.$fg_crop_avocado_nursery$, 4, TRUE, 52),
    ('crop-pear-nursery', 'Pear nursery', 'pear', 'crop_nutrition', 'general', 'safe', $fg_crop_pear_nursery$# Pear nursery

Bench pear culture mirrors **apple nursery** targets — moderate EC, good airflow. **Fire blight** risk rises with dense canopy and prolonged leaf wetness.

Train scaffolds early in containers; fruiting years similar to apple on bench scale.$fg_crop_pear_nursery$, 4, TRUE, 53),
    ('crop-plum-nursery', 'Plum nursery (stone fruit)', 'plum', 'crop_nutrition', 'general', 'safe', $fg_crop_plum_nursery$# Plum nursery (stone fruit)

Plum nursery stock follows **peach-class** feed and light — chill-dependent, spring flush management. EC **≈1.0–1.5 mS/cm** for young trees.

Fruit in containers typically **years 3–5**. Thin heavy sets on small trees to avoid branch break.$fg_crop_plum_nursery$, 4, TRUE, 54),
    ('crop-mango-nursery', 'Mango (container nursery)', 'mango', 'crop_nutrition', 'general', 'safe', $fg_crop_mango_nursery$# Mango (container nursery)

**Tropical warm-only** — cold roots stop growth fast. EC **≈1.0–1.6 mS/cm**; watch **anthracnose** when RH stays high on flush.

Juvenile mangoes in pots may fruit **years 3–5+**. Never ship cold-stressed liners into warm zones without acclimation.$fg_crop_mango_nursery$, 4, TRUE, 55),
    ('crop-rice-nutrition', 'Rice nutrition (aquaponics / shallow water)', 'rice', 'crop_nutrition', 'general', 'safe', $fg_crop_rice_nutrition$# Rice nutrition (aquaponics / shallow water)

Rice in bench-scale aquaponics or shallow trays runs **low EC** (~0.5–1.0 mS/cm) with **warm roots** and **constant shallow water**. Fish waste often supplies nitrogen — watch for **iron chlorosis** if water is too alkaline.

Assign the rice profile in Start grow or Plants so Guardian cites structured mS/cm targets. Use **Variety / cultivar** for the strain (Basmati, Jasmine, etc.) after picking Rice from the catalog.$fg_crop_rice_nutrition$, 4, TRUE, 56),
    ('crop-unsupported-woodland', 'Unsupported woodland crops (ramps, ginseng)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_woodland$# Unsupported woodland crops (ramps, ginseng)

**Ramps** (wild leek) and **ginseng** are woodland ephemerals or multi-year shade medicinals — not indoor fertigation crops. gr33n does not ship EC, VPD, or photoperiod targets for them.

If an operator asks about bench automation, explain honestly: these are foraged or long-cycle outdoor/forest production. For general greenhouse questions, suggest a supported **leafy** or **herb** cousin only when they want a hydro starting point — never invent woodland feed schedules.$fg_crop_unsupported_woodland$, 4, TRUE, 57),
    ('crop-unsupported-mushroom', 'Mushroom production (unsupported fertigation profile)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_mushroom$# Mushroom production (unsupported fertigation profile)

Mushrooms and other **fungi** use bag/substrate colonization and humidity rooms — a different domain from plant EC/VPD profiles. gr33n crop targets do not apply.

Direct operators to husbandry / substrate workflows when available. Do not map shiitake or other fungi to cannabis or tomato nutrient curves.$fg_crop_unsupported_mushroom$, 4, TRUE, 58),
    ('crop-unsupported-field-roots', 'In-ground root crops (carrot, potato)', NULL, 'unsupported', 'general', 'safe', $fg_crop_unsupported_field_roots$# In-ground root crops (carrot, potato)

**Carrots, potatoes, and sweet potatoes** are field or deep-container taproot/tuber crops. gr33n structured targets cover hydroponic and bench container production — not deep soil beds or field scale.

If an operator wants indoor hydro only, suggest cloning from **lettuce** (fast leafy baseline) or **tomato** (fruiting hydro) and adjusting manually — do not state fake EC targets for tubers.$fg_crop_unsupported_field_roots$, 4, TRUE, 59),
    ('natural-farming-jms', 'JADAM Microbial Solution (JMS)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_jms$# JADAM Microbial Solution (JMS)

## What it is (1 paragraph)

JMS is the foundation JADAM microbial inoculant — diverse bacteria and fungi from forest leaf mold, activated with potato starch water. It builds soil and leaf-surface biology and suppresses pathogens when used at Cho dilutions.

## When to use

- Soil drenches every 2 weeks in active season; before transplant
- Foliar sprays in vegetative and early flower for leaf-surface microbes
- Combined with JLF in peak-season weekly drenches

## Ingredients (list with amounts)

- Leaf mold humus (local forest floor), ~1 cup per 10–20 L batch
- 1 potato boiled in non-chlorinated water; use cooled potato water as base
- Pinch of sea salt
- Non-chlorinated water to 10–20 L total volume

## Step-by-step preparation

1. Boil potato in non-chlorinated water; cool completely.
2. Place potato in a mesh bag; suspend in a bucket with leaf mold and pinch of salt.
3. Fill to 10–20 L with non-chlorinated water; cover loosely (not airtight).
4. Ferment 24–72 h at 20–30 °C until peak foam activity.
5. Use at peak — strain if needed for sprayers.

## Ferment / wait timeline

- **24–72 h** active fermentation to peak foam
- **Use within 6–12 h of peak** — not after sitting a full week idle

## Ready signs (smell, foam, color)

- Vigorous foam at surface at peak activity
- Earthy, not putrid, smell
- Cloudy water; potato breaks down in bag

## Storage

Use at peak foam; do not store active JMS long-term. Make fresh batches weekly during season.

## Safety & water (non-chlorinated, PPE)

Chlorinated tap water kills microbes — rain, RO, or de-chlorinated water only.

## How to apply (link to application recipe name)

- **JMS Soil Drench** — primary soil inoculant
- **JMS Foliar Spray** — leaf biology (+ JWA for coverage)
- **JLF and JMS Combined Drench** — weekly peak-season pass

## Dilution table (start conservative → stronger)

| Use | Dilution (JMS:water) | Notes |
|-----|----------------------|-------|
| Soil drench | **1:10** | 2–4 L per sqm root zone |
| Foliar | **1:20** + JWA | Early morning; both leaf sides |
| With JLF tank | **1:10** in same water as JLF 1:20 | Apply same day |

## Common mistakes

- Storing finished JMS a week+ after peak — weak or anaerobic
- Using 1:500 dilution (old drift) — far too weak per Cho
- Skipping JWA on foliar — poor leaf coverage$fg_natural_farming_jms$, 4, TRUE, 60),
    ('natural-farming-jlf-general', 'JLF from weeds and grass (general)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_jlf_general$# JLF from weeds and grass (general)

## What it is (1 paragraph)

JLF (JADAM Liquid Fertilizer) from local weeds and grasses returns native minerals to soil. It is the primary fertility input in JADAM programs — much stronger than JMS; dilute carefully.

## When to use

- Primary soil fertility weekly to biweekly in active growth
- Outdoor beds and indoor reservoirs on JADAM-style programs
- Base method for extension materials (e.g. goldenrod) — see [goldenrod JLF](natural-farming-goldenrod-jlf.md)

## Ingredients (list with amounts)

- Fresh untreated weeds/grass clippings — fill container **2/3**
- Handful leaf mold (microbial starter)
- Non-chlorinated water to top

## Step-by-step preparation

1. Chop weeds; fill ferment vessel 2/3 full.
2. Add leaf mold starter.
3. Top with non-chlorinated water; seal (burp if needed).
4. Ferment 7–14 days; stir every few days.
5. Strain through cloth before use.

## Ferment / wait timeline

- Minimum usable **7–14 days**; can mature weeks to months for richer brew
- Strain before applying

## Ready signs (smell, foam, color)

- Earthy fermented smell — not rotten-egg anaerobic
- Plant material softened; liquid amber to brown

## Storage

Strained: use within **30 days**. Sealed concentrate up to **3 months** cool/shaded.

## Safety & water (non-chlorinated, PPE)

No herbicide-treated clippings. Non-chlorinated water only.

## How to apply (link to application recipe name)

- **JLF General Soil Drench** — main fertility
- **JLF Seedling Drench** — gentler 1:30
- **JLF and JMS Combined Drench** — with JMS 1:10
- **JLF Foliar Feed** — stress only, finely strained

## Dilution table (start conservative → stronger)

| Situation | JLF:water | Notes |
|-----------|-----------|-------|
| First time / unsure | **1:100** | Start here per Cho/FigJam |
| Tested on your soil | **1:20** | Primary experienced default |
| Seedlings | **1:30** | See JLF Seedling Drench |

## Common mistakes

- Starting at 1:20 on unknown soil — leaf burn or salt shock
- Using diseased or sprayed weeds — pathogen carryover
- Foliar without fine strain — clogged sprayer$fg_natural_farming_jlf_general$, 4, TRUE, 61),
    ('natural-farming-jlf-crop-specific', 'JLF from same-crop residue', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_jlf_crop_specific$# JLF from same-crop residue

## What it is (1 paragraph)

Crop-specific JLF uses residue from the **same crop** you will feed — the most targeted JADAM fertilizer (tomato trimmings for tomatoes, corn stalks for corn).

## When to use

- Recurring fertility for a single crop through a season
- When you have clean, healthy residue volume

## Ingredients (list with amounts)

- Same-crop residue (stems, leaves — not fruit or seed) — 2/3 vessel
- Leaf mold handful
- Non-chlorinated water to top

## Step-by-step preparation

1. Chop residue small; fill container 2/3.
2. Add leaf mold; fill with water; seal.
3. Ferment **10–14 days**; stir occasionally.
4. Strain before use; label crop + date.

## Ferment / wait timeline

**10–14 days** minimum; use within same growing season.

## Ready signs (smell, foam, color)

Earthy ferment smell; residue broken down; brown liquid.

## Storage

Use within season; label crop type. Do not mix crop-specific batches blindly.

## Safety & water (non-chlorinated, PPE)

**Never** use diseased plant material. Non-chlorinated water.

## How to apply (link to application recipe name)

Apply via **JLF General Soil Drench** dilution bands (start **1:100**, up to **1:20** when tested).

## Dilution table (start conservative → stronger)

| Stage | Dilution | Notes |
|-------|----------|-------|
| Start | 1:100 | Conservative first pass |
| Established | 1:20–1:30 | Match crop vigor |

## Common mistakes

- Cross-crop residue — loses targeted benefit
- Diseased trimmings — spreads pathogens
- Over-applying on fruiting plants — too much N$fg_natural_farming_jlf_crop_specific$, 4, TRUE, 62),
    ('natural-farming-jlf-spring-nettle-comfrey', 'Spring JLF — nettle and comfrey', NULL, 'trades', 'natural_farming', 'caution', $fg_natural_farming_jlf_spring_nettle_comfrey$# Spring JLF — nettle and comfrey

## What it is (1 paragraph)

High-nitrogen spring JLF from **dynamic accumulator** biomass (nettle, comfrey) — deep-mining herbs, not nitrogen-fixing legumes. Strong vegetative push.

## When to use

- Early spring vegetative growth
- Before heavy fruiting — avoid over-N on fruiting plants later

## Ingredients (list with amounts)

- Fresh stinging nettle tops and/or comfrey leaves — 2/3 vessel
- Leaf mold handful
- Non-chlorinated water

## Step-by-step preparation

1. Harvest nettle wearing gloves; chop with comfrey.
2. Fill 2/3; add leaf mold; top with water; seal.
3. Ferment **7–10 days**; strain.

## Ferment / wait timeline

**7–10 days**; use strained liquid within **2 weeks**.

## Ready signs (smell, foam, color)

Rich earthy smell; dark green-brown liquid after strain.

## Storage

Use within 2 weeks of straining — high N degrades in storage.

## Safety & water (non-chlorinated, PPE)

Gloves for nettle harvest. High N — do not over-apply to fruiting crops.

## How to apply (link to application recipe name)

**JLF General Soil Drench** — start **1:100**, test before **1:20**.

## Dilution table (start conservative → stronger)

| Pass | Dilution | Notes |
|------|----------|-------|
| First | 1:100 | Spring push, watch leaf color |
| Follow-up | 1:30–1:20 | Only if plants respond well |

## Common mistakes

- Calling nettle "nitrogen-fixing" — it is a dynamic accumulator
- Strong dilution on fruit trees in bloom — vegetative push at wrong time$fg_natural_farming_jlf_spring_nettle_comfrey$, 4, TRUE, 63),
    ('natural-farming-ffj', 'Fermented Fruit Juice (FFJ) — KNF', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_ffj$# Fermented Fruit Juice (FFJ) — KNF

## What it is (1 paragraph)

FFJ is a **KNF** sugar ferment of ripe fruit — enzymes, sugars, and potassium for flowering and early fruit. Often paired with JADAM programs but it is sugar-based KNF, not JADAM core.

## When to use

- Transition to flowering through early fruit set
- With WCA in **FFJ and WCA Flowering Boost** foliar program

## Ingredients (list with amounts)

- Ripe or overripe fruit (banana peels work well)
- Brown sugar — **1:1 by weight** with fruit
- **No added water** (KNF standard)

## Step-by-step preparation

1. Chop fruit; mix 1:1 with brown sugar by weight.
2. Pack in jar; cover with breathable cloth (not airtight).
3. Ferment ~7 days at room temperature.
4. Strain liquid; bottle.

## Ferment / wait timeline

~**7 days** ferment; strain when juice separates.

## Ready signs (smell, foam, color)

Sweet-sour ferment smell; liquid syrup; fruit collapsed.

## Storage

Refrigerate after straining; use within **6 months**.

## Safety & water (non-chlorinated, PPE)

Do not add water. Avoid moldy fruit.

## How to apply (link to application recipe name)

**FFJ and WCA Flowering Boost** — FFJ 1:500 + WCA 1:1000 + JWA in same tank, weekly from first buds.

## Dilution table (start conservative → stronger)

| Use | Dilution | Notes |
|-----|----------|-------|
| Flowering foliar | 1:500 FFJ | With WCA 1:1000 |
| Hot weather | 1:800–1:1000 | Lighter pass |

## Common mistakes

- Adding water to ferment — not KNF FFJ
- Using in heavy veg — promotes wrong growth stage
- Labeling as pure JADAM — honesty: KNF input$fg_natural_farming_ffj$, 4, TRUE, 64),
    ('natural-farming-fpj', 'Fermented Plant Juice (FPJ) — KNF', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_fpj$# Fermented Plant Juice (FPJ) — KNF

## What it is (1 paragraph)

FPJ is **KNF** fermented growing tips (comfrey, nettle, mugwort, bamboo) with brown sugar — plant hormones and amino acids for vegetative growth. Sugar-based; label as KNF when used beside JADAM.

## When to use

- Vegetative stage only — **stop at flower transition**
- **FPJ Vegetative Foliar** every 7–14 days

## Ingredients (list with amounts)

- Fresh fast-growing plant tips
- Brown sugar **1:1 by weight**

## Step-by-step preparation

1. Chop tips; layer equal weight sugar.
2. Seal jar with breathable cover.
3. Ferment 3–7 days; strain and bottle.

## Ferment / wait timeline

**3–7 days** at room temp.

## Ready signs (smell, foam, color)

Sweet ferment; liquid separates; tips collapsed.

## Storage

Refrigerate; **6–12 months** sealed.

## Safety & water (non-chlorinated, PPE)

Accurate 1:1 sugar ratio; no moldy material.

## How to apply (link to application recipe name)

**FPJ Vegetative Foliar** — 1:500 (1:1000 hot weather) + JWA 1:1000.

## Dilution table (start conservative → stronger)

| Conditions | FPJ:water |
|------------|-----------|
| Normal | 1:500 |
| Hot / stress | 1:1000 |

## Common mistakes

- Continuing FPJ after flowers — wrong stage
- Calling it JADAM core — it is KNF (sugar ferment)$fg_natural_farming_fpj$, 4, TRUE, 65),
    ('natural-farming-lab', 'Lactic Acid Bacteria serum (LAB) — KNF', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_lab$# Lactic Acid Bacteria serum (LAB) — KNF

## What it is (1 paragraph)

LAB serum from soured rice wash cultured in milk — golden layer of lactic acid bacteria that suppresses harmful soil microbes and improves structure. KNF input, often paired with JADAM.

## When to use

- Soil conditioning every 2–4 weeks
- Before transplanting
- **LAB Soil Conditioner** drench

## Ingredients (list with amounts)

- Rice wash (first rinse water)
- Fresh whole milk (non-UHT) — 10 parts milk to 1 part soured rice wash

## Step-by-step preparation

1. Ferment rice wash 3–5 days until soured.
2. Mix 1 part soured rice wash into 10 parts milk.
3. Wait 5–7 days; collect **golden serum** from bottom.
4. Mix equal part raw sugar to preserve (optional).

## Ferment / wait timeline

Rice wash **3–5 d**; milk culture **5–7 d**.

## Ready signs (smell, foam, color)

Golden translucent serum layer below curds; sour-clean smell.

## Storage

Refrigerated with sugar preservative **6–12 months**. Use serum only — discard curds and white top.

## Safety & water (non-chlorinated, PPE)

Use golden layer only.

## How to apply (link to application recipe name)

**LAB Soil Conditioner** — **1:1000** LAB:water; water in lightly.

## Dilution table (start conservative → stronger)

| Use | Dilution |
|-----|----------|
| Soil drench | **1:1000** |

## Common mistakes

- Using curds or top milk layer — wrong fraction
- Stronger than 1:1000 — unnecessary; KNF standard is dilute$fg_natural_farming_lab$, 4, TRUE, 66),
    ('natural-farming-ohn', 'Oriental Herbal Nutrient (OHN) — KNF', NULL, 'trades', 'natural_farming', 'caution', $fg_natural_farming_ohn$# Oriental Herbal Nutrient (OHN) — KNF

## What it is (1 paragraph)

OHN is a potent **KNF** extract of garlic, ginger, angelica, cinnamon and other aromatics — immune support and pest deterrent in **very small** doses.

## When to use

- Preventative soil drench every 2–4 weeks
- Pest pressure — weekly at labeled dilution only

## Ingredients (list with amounts)

- Garlic, ginger, angelica root, cinnamon bark
- Brown sugar 1:1 with chopped herbs
- Alcohol ~25% ABV — equal volume to fermented herb mix after first ferment

## Step-by-step preparation

1. Chop herbs; layer 1:1 sugar; ferment 7 days.
2. Add equal volume alcohol; ferment 7 more days.
3. Strain; combine individual herb extracts into OHN stock.

## Ferment / wait timeline

**7 d** sugar ferment + **7 d** with alcohol per herb component.

## Ready signs (smell, foam, color)

Strong aromatic extract; clear to amber liquid after strain.

## Storage

Sealed **1–2 years**. Extremely concentrated.

## Safety & water (non-chlorinated, PPE)

**Never exceed 1:1000** application. Avoid inhaling concentrate.

## How to apply (link to application recipe name)

**OHN Pest and Immunity Drench** — strictly **1:1000** OHN:water.

## Dilution table (start conservative → stronger)

| Use | Dilution | Max |
|-----|----------|-----|
| All applications | **1:1000** | Never stronger |

## Common mistakes

- Full-strength or 1:500 — burn and phytotoxicity
- Treating as JADAM core — KNF sugar/alcohol extract$fg_natural_farming_ohn$, 4, TRUE, 67),
    ('natural-farming-wca-wcs', 'Water-soluble calcium inputs (WCA and WCS) — KNF', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_wca_wcs$# Water-soluble calcium inputs (WCA and WCS) — KNF

## What it is (1 paragraph)

**WCA** dissolves calcium from roasted eggshells in brown rice vinegar (1:10). **WCS** (WCAP) dissolves phosphorus and calcium from white-ashed bones in vinegar (1:10). Both are **KNF** mineral extracts used in flowering and cell-strength programs.

## When to use

- WCA with FFJ during flower (**FFJ and WCA Flowering Boost**)
- WCA with BRV before stress (**BRV and WCA Cell Strengthener**)
- WCS when root/flower phosphorus support is needed (foliar at 1:1000 band)

## Ingredients (list with amounts)

**WCA:** roasted eggshells + unpasteurized BRV 1:10 (shells covered by vinegar)

**WCS:** beef/pork bones charred to white ash + BRV 1:10

## Step-by-step preparation

**WCA**

1. Roast eggshells light brown; cool.
2. Cover with BRV 1:10 in breathable container.
3. Fizz 7 days; strain.

**WCS**

1. Char bones to white ash completely.
2. Dissolve in BRV 1:10 for 7 days; strain.

## Ferment / wait timeline

**7 days** extraction each; gases evolve — breathable lid.

## Ready signs (smell, foam, color)

Fizzing slows; clear to amber extract; ash/shells mostly dissolved.

## Storage

Breathable container; use within **30 days** after strain.

## Safety & water (non-chlorinated, PPE)

Vinegar acid — eye protection. Roast shells/bones fully.

## How to apply (link to application recipe name)

- **FFJ and WCA Flowering Boost** — WCA **1:1000** with FFJ 1:500
- **BRV and WCA Cell Strengthener** — WCA **1:1000** with BRV 1:800

## Dilution table (start conservative → stronger)

| Input | Typical foliar dilution |
|-------|------------------------|
| WCA | **1:1000** |
| WCS | **1:1000** (same band as WCA in programs) |

## Common mistakes

- Sealed tight jar on WCA — gas buildup
- Partially charred bones — inconsistent P and off flavors
- Undiluted on leaves — burn$fg_natural_farming_wca_wcs$, 4, TRUE, 68),
    ('natural-farming-jwa-js-jhs', 'JWA, JHS, and JS — wetting and pest inputs (JADAM)', NULL, 'trades', 'natural_farming', 'expert', $fg_natural_farming_jwa_js_jhs$# JWA, JHS, and JS — wetting and pest inputs (JADAM)

## What it is (1 paragraph)

**JWA** is JADAM wetting-agent soap (wood-ash lye + oil). **JHS** is boiled herbal concentrate for pest deterrent sprays. **JS** is exothermic **JADAM sulfur concentrate** (~25% sulfur) — not garden wettable sulfur. JWA is added to foliar mixes for coverage; JHS and JS are pest/disease programs at seed dilutions.

## When to use

- JWA: surfactant in JMS foliar, JHS/JS sprays, FFJ/FPJ tanks
- JHS + JWA: weekly preventative pest spray
- JS + JWA: powdery mildew, rust, mites at first sign (≤32 °C)

## Ingredients (list with amounts)

**JWA:** wood ash lye water + plant oil (soy/canola/coconut) → soap

**JHS:** 1 kg fresh herb (wormwood, artemisia, garlic chives, hot pepper, neem, or Jerusalem artichoke) + 4–5 L water

**JS concentrate:** elemental sulfur, caustic soda (NaOH), red clay, phyllite, sea salt — Cho exothermic batch (~25% sulfur concentrate)

## Step-by-step preparation

**JWA:** boil ash for lye water; filter; mix 1:1 with oil; boil to soap.

**JHS:** boil 1 kg plant in mesh bag in 4–5 L water **4–5 hours**; strain very fine.

**JS:** follow Cho exothermic batch method; label concentrate; dilute only at spray time.

## Ferment / wait timeline

JWA soap keeps dry indefinitely. JHS use within **2 weeks** refrigerated. JS concentrate stored sealed; mix spray **same day**.

## Ready signs (smell, foam, color)

JWA: firm soap paste. JHS: dark aromatic broth, fine strain required. JS: labeled concentrate strength.

## Storage

JWA dry soap; JHS refrigerated short term; JS concentrate sealed labeled.

## Safety & water (non-chlorinated, PPE)

**JWA:** lye caustic — gloves, no sun spray burns.

**JS:** caustic soda batch — full PPE, ventilation; **do not apply above 32 °C**.

**JHS:** avoid open blooms — deters pollinators.

## How to apply (link to application recipe name)

- **JMS Foliar Spray** — add JWA for coverage
- **JHS and JWA Natural Pesticide** — JHS 1:50 + JWA 1:500
- **JS Fungicide Spray** — **0.5–2 L JS concentrate per 500 L water** + JWA 1:500
- **JWA Insecticide Spray** — JWA 1:500 alone for soft-bodied insects

## Dilution table (start conservative → stronger)

| Product | Application dilution |
|---------|---------------------|
| JWA alone | 1:500 |
| JHS + JWA | 1:50 + 1:500 |
| JS concentrate | 0.5–2 L per 500 L water + JWA |

## Common mistakes

- Wettable sulfur labeled JS — wrong input (pre-audit drift)
- JHS cold steep 1–3 h — not Cho boil method
- JS spray in hot midday sun — sulfur burn$fg_natural_farming_jwa_js_jhs$, 4, TRUE, 69),
    ('natural-farming-brv', 'Brown rice vinegar (BRV) — purchased input', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_brv$# Brown rice vinegar (BRV) — purchased input

## What it is (1 paragraph)

Unpasteurized organic brown rice vinegar (4–8% acidity) — purchased input for WCA/WCS extraction and foliar cell-strengthening with WCA. Not a ferment you make on farm in v1 seed data.

## When to use

- WCA and WCS extraction solvent (1:10 with shells/bones)
- **BRV and WCA Cell Strengthener** foliar before stress

## Ingredients (list with amounts)

- Organic unpasteurized BRV — purchase ready-made

## Step-by-step preparation

1. Purchase unpasteurized organic BRV.
2. Use directly for extracts or dilute per application recipe.

## Ferment / wait timeline

N/A — purchased product.

## Ready signs (smell, foam, color)

Clear amber vinegar; live culture sediment normal in unpasteurized bottles.

## Storage

Sealed at room temperature indefinitely.

## Safety & water (non-chlorinated, PPE)

Undiluted on foliage burns — always follow recipe dilution.

## How to apply (link to application recipe name)

**BRV and WCA Cell Strengthener** — BRV **1:800** + WCA **1:1000**; do not exceed BRV rate.

## Dilution table (start conservative → stronger)

| Recipe | BRV dilution |
|--------|--------------|
| Cell strengthener foliar | **1:800** with WCA 1:1000 |

## Common mistakes

- Pasteurized clear vinegar for WCA — weak extraction
- Full-strength foliar — leaf burn$fg_natural_farming_brv$, 4, TRUE, 70),
    ('natural-farming-faa', 'Fish amino acid (FAA) — KNF', NULL, 'trades', 'natural_farming', 'caution', $fg_natural_farming_faa$# Fish amino acid (FAA) — KNF

## What it is (1 paragraph)

FAA is **KNF** fish scrap ferment with brown sugar — high nitrogen and trace minerals. Long ferment until bones dissolve; apply only at high dilution.

## When to use

- Supplemental nitrogen foliar or soil at **1:1000** minimum
- Paired with JADAM programs when fish waste is available

## Ingredients (list with amounts)

- Fresh fish scraps (no salt)
- Brown sugar **1:1 by weight**

## Step-by-step preparation

1. Layer fish and brown sugar 1:1 in ferment vessel.
2. Cover breathable; ferment **3–6 months** until bones soften/dissolve.
3. Strain; bottle.

## Ferment / wait timeline

**3–6 months**; longer in cool climates.

## Ready signs (smell, foam, color)

Fish breaks down; bones crumble; sauce-like liquid.

## Storage

Refrigerate after strain; **6–12 months**.

## Safety & water (non-chlorinated, PPE)

Strong odor — ventilate outdoor ferment. Salted fish scraps ruin batch.

## How to apply (link to application recipe name)

Soil or foliar at **≥1:1000** FAA:water (KNF standard — never stronger).

## Dilution table (start conservative → stronger)

| Use | Dilution |
|-----|----------|
| All passes | **1:1000 minimum** |

## Common mistakes

- Short ferment — incomplete breakdown
- Strong smell indoors without ventilation
- Undiluted application — salt/ammonia burn$fg_natural_farming_faa$, 4, TRUE, 71),
    ('natural-farming-compost-tea-aact', 'Actively aerated compost tea (AACT)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_compost_tea_aact$# Actively aerated compost tea (AACT)

## What it is (1 paragraph)

AACT is **not JADAM** — an Elaine Ingham–style aerobic compost extract brewed with air stone and molasses to multiply beneficial microbes. Complements JMS but follows a **4-hour use window** after brew finishes.

## When to use

- Soil drench or foliar when soil food web boost is needed
- Disease suppression support alongside good compost source

## Ingredients (list with amounts)

- Finished quality compost (mesh bag)
- Unsulfured molasses — ~1 tbsp per 4 L water
- Optional kelp meal
- De-chlorinated water

## Step-by-step preparation

1. Suspend compost bag in bucket with air stone running.
2. Add molasses (and kelp if used).
3. Brew **24–48 h** with continuous aeration.
4. Use entire batch within **4 hours** of stopping aeration.

## Ferment / wait timeline

Brew **24–48 h** aerated; apply within **4 h** of finish.

## Ready signs (smell, foam, color)

Earthy smell; slight foam; no anaerobic rotten odor.

## Storage

**Do not store** brewed tea — microbes crash without O₂.

## Safety & water (non-chlorinated, PPE)

Use finished mature compost; E. coli risk if compost is immature or aeration fails.

## How to apply (link to application recipe name)

Apply as soil drench or foliar using standard compost-tea dilution for your volume (undiluted to 1:10 depending on compost strength — start weak).

## Dilution table (start conservative → stronger)

| Pass | Guidance |
|------|----------|
| First use | Weak tea / longer water ratio — watch plant response |
| Follow-up | Stronger only if no burn and good compost source |

## Common mistakes

- Storing tea overnight — anaerobic crash
- Turning off air early — pathogen bloom risk
- Calling it JADAM JMS substitute — different tradition and timing$fg_natural_farming_compost_tea_aact$, 4, TRUE, 72),
    ('natural-farming-application-recipes', 'Natural farming application recipes (canon)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_application_recipes$# Natural farming application recipes (canon)

## What it is (1 paragraph)

Reference table for all **14 application recipes** in farm seed data — dilutions audited against Cho 2016 / KNF standards in Phase 208 WS0. Use with input batch inventory and fertigation programs.

## When to use

Look up how to apply a ready batch before mixing a tank or spraying. Cross-link input guides for making each concentrate.

## Ingredients (list with amounts)

See input guides per component — this table is **application** only.

## Step-by-step preparation

1. Identify recipe name below matching your program or task.
2. Strain concentrates as needed (JLF, JHS).
3. Mix at listed dilution in non-chlorinated water same day.
4. Add JWA when recipe notes coverage.

## Ferment / wait timeline

N/A — application step; batches must already be at ready status.

## Ready signs (smell, foam, color)

Input batches at **ready_for_use** per batch notes before applying these dilutions.

## Storage

Mixed tank: use same day. Do not store diluted spray overnight.

## Safety & water (non-chlorinated, PPE)

Follow each input's safety tier — OHN and JS never above labeled dilution.

## How to apply (link to application recipe name)

| # | Recipe | Type | Dilution (canon) | Frequency |
|---|--------|------|------------------|-----------|
| 1 | JMS Soil Drench | soil_drench | **1:10** JMS:water | Every 2 weeks |
| 2 | JLF General Soil Drench | soil_drench | 1:20 (start **1:100**) | Weekly–biweekly |
| 3 | JLF Seedling Drench | soil_drench | **1:30** | Weekly seedlings |
| 4 | JLF and JMS Combined Drench | soil_drench | JLF 1:20 + JMS **1:10** | Weekly peak season |
| 5 | LAB Soil Conditioner | soil_drench | **1:1000** | Every 2–4 weeks |
| 6 | OHN Pest and Immunity Drench | soil_drench | **1:1000** max | Preventative / pressure |
| 7 | JMS Foliar Spray | foliar_spray | **1:20** + JWA | Every 1–2 weeks |
| 8 | FPJ Vegetative Foliar | foliar_spray | 1:500–1:1000 + JWA | Veg only |
| 9 | FFJ and WCA Flowering Boost | foliar_spray | FFJ 1:500 + WCA 1:1000 + JWA | Flower → early fruit |
| 10 | BRV and WCA Cell Strengthener | foliar_spray | BRV 1:800 + WCA 1:1000 | Before stress |
| 11 | JHS and JWA Natural Pesticide | foliar_spray | JHS 1:50 + JWA 1:500 | Weekly preventative |
| 12 | JS Fungicide Spray | foliar_spray | **0.5–2 L JS conc. / 500 L** + JWA | ≤32 °C; repeat 5–7 d |
| 13 | JLF Foliar Feed | foliar_spray | 1:30–1:50 + JWA | Stress only |
| 14 | JWA Insecticide Spray | foliar_spray | **1:500** JWA | Active soft-bodied pests |

## Dilution table (start conservative → stronger)

JLF drenches: always **start 1:100** if unsure (see [JLF general](natural-farming-jlf-general.md)). JMS never weaker than **1:10** soil / **1:20** foliar per audit.

## Common mistakes

- Using pre-audit **1:500 JMS** — wrong (WS0 fixed)
- OHN or FAA stronger than 1:1000
- JS as 0.5% wettable sulfur — wrong input; use JADAM JS concentrate recipe$fg_natural_farming_application_recipes$, 4, TRUE, 73),
    ('natural-farming-indoor-photoperiod-program', 'Indoor photoperiod JADAM programs (bootstrap)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_indoor_photoperiod_program$# Indoor photoperiod JADAM programs (bootstrap)

## What it is (1 paragraph)

Maps the **`jadam_indoor_photoperiod_v1`** bootstrap template — veg (18/6), flower (12/12), and outdoor JLF drench programs wired to audited application recipes on demo and new farms.

## When to use

After applying bootstrap template or when mirroring demo farm fertigation layout. Commercial EC programs stay on **Feed & water** — these are parallel natural paths.

## Ingredients (list with amounts)

Program-linked batches (template IDs **TPL-JLF-GEN-001**, **TPL-JMS-001**, **TPL-FFJ-001**, **TPL-WCA-001**) — see input guides.

## Step-by-step preparation

1. Apply bootstrap `jadam_indoor_photoperiod_v1` or use demo farm seed.
2. Confirm batches exist for JLF, JMS, FFJ, WCA.
3. Refresh reservoir mix tasks before scheduled irrigations.

## Ferment / wait timeline

Maintain rolling JMS at peak foam; JLF/FFJ batches per input guide storage windows.

## Ready signs (smell, foam, color)

Batch status **ready_for_use** in inventory before linking to a mix event.

## Storage

Veg **Main Nutrient Reservoir** and flower **Flower Nutrient Reservoir** — mix same day as irrigation schedule.

## Safety & water (non-chlorinated, PPE)

Non-chlorinated make-up water in reservoirs. JMS at **1:10** with JLF **1:20** in combined veg tank (not legacy 1:500 JMS).

## How to apply (link to application recipe name)

| Program | Zone | Recipe | Schedule |
|---------|------|--------|----------|
| **Veg Daily JLF Program** | Veg Room 18/6 | **JLF and JMS Combined Drench** | Water Late Veg Daily |
| **Flower Daily FFJ+WCA Program** | Flower Room 12/12 | **FFJ and WCA Flowering Boost** | Water Early Flower Daily |
| **Outdoor JLF Soil Drench** | Outdoor Garden | **JLF General Soil Drench** | Water Outdoor Garden Daily |

See [application recipes](natural-farming-application-recipes.md) for dilutions.

## Dilution table (start conservative → stronger)

Combined veg tank: JLF **1:20** + JMS **1:10**. Flower tank: FFJ 1:500 + WCA 1:1000. Outdoor: JLF start **1:100** if unsure.

## Common mistakes

- Expecting EC 1.6–1.8 mS/cm Mericle targets without converting mindset — tune volume/cron, not bottle A/B
- Legacy JMS 1:500 in mix notes — pre-audit; refresh to 1:10$fg_natural_farming_indoor_photoperiod_program$, 4, TRUE, 74),
    ('natural-farming-goldenrod-jlf', 'Goldenrod JLF (extension method)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_goldenrod_jlf$# Goldenrod JLF (extension method)

## What it is (1 paragraph)

**Not a named Cho goldenrod recipe.** Valid **extension**: apply the standard [JLF general](natural-farming-jlf-general.md) method to **Solidago** (Canadian goldenrod) biomass — dynamic accumulator weed usable for orchard understory fertility. Compatible with dye harvest if biomass is collected responsibly.

## When to use

- Orchard understory (cherry, apple, plum) where goldenrod is present
- Operator keeps goldenrod for dyes **and** wants fertigation use — not prescriptive removal

## Ingredients (list with amounts)

- Fresh goldenrod biomass (stems/leaves — not necessarily flowers if reserved for dye)
- Leaf mold handful
- Non-chlorinated water — same ratios as JLF general (2/3 vessel biomass)

## Step-by-step preparation

Follow [JLF general](natural-farming-jlf-general.md) steps 1–5 using goldenrod as the weed feedstock.

## Ferment / wait timeline

**7–14 days** ferment; strain before use.

## Ready signs (smell, foam, color)

Earthy ferment — same ready signs as general JLF.

## Storage

Strained: **30 days** cool/shaded.

## Safety & water (non-chlorinated, PPE)

Do not use sprayed roadside plants. Label batch **goldenrod JLF extension**.

## How to apply (link to application recipe name)

**JLF General Soil Drench** dilution bands — start **1:100**, stronger only after plant response (up to **1:30** experienced).

Guardian must say: *"JLF from goldenrod using the standard JLF method"* — never *"Cho's goldenrod recipe."*

## Dilution table (start conservative → stronger)

| Pass | Dilution | Notes |
|------|----------|-------|
| First | **1:100** | Cherry understory conservative start |
| Tested OK | 1:30–1:20 | Only with observed plant response |

## Common mistakes

- Claiming cho_named source tier — must stay **extension_method**
- Starting at 1:20 on unknown understory — too strong
- Conflicting with dye harvest — plan biomass cuts so both uses fit operator intent

See also [forest garden understory](natural-farming-forest-garden-understory.md).$fg_natural_farming_goldenrod_jlf$, 4, TRUE, 75),
    ('natural-farming-forest-garden-understory', 'Forest garden understory (cherry, blackberry, goldenrod)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_forest_garden_understory$# Forest garden understory (cherry, blackberry, goldenrod)

## What it is (1 paragraph)

Counsel for **forest-garden / orchard understory** questions — cherry with goldenrod, blackberries, and mixed volunteers. gr33n does **not** ship EC/VPD programs for wild understory polycultures; this guide gives honest ecology and links to natural-farming extensions without inventing bottle-nutrient schedules.

## When to use

- Operator asks about cherry + goldenrod + blackberry coexistence
- Guardian smoke/regression forest-garden prompts with **farm context on** after RAG ingest

## Ingredients (list with amounts)

N/A — ecology guide, not a ferment recipe.

## Step-by-step preparation

1. Identify operator goals: fruit quality, dye harvest, blackberry keep/remove, goldenrod management.
2. For goldenrod biomass → [goldenrod JLF extension](natural-farming-goldenrod-jlf.md) at **1:100** start.
3. For sweet cherry production targets see [cherry nursery guide](crop-cherry-nursery.md) — not identical to backyard forest garden.
4. For unsupported woodland forage crops see [crop-unsupported-woodland.md](crop-unsupported-woodland.md).

## Ferment / wait timeline

If using goldenrod JLF — see extension guide ferment timeline.

## Ready signs (smell, foam, color)

N/A for ecology; for JLF extension see goldenrod guide.

## Storage

N/A.

## Safety & water (non-chlorinated, PPE)

Blackberry thorns — PPE for clearing. Do not recommend herbicide blanket on polyculture without operator consent.

## How to apply (link to application recipe name)

Optional understory fertility: **JLF General Soil Drench** via goldenrod extension — **1:100** conservative around cherry root zone; never claim EC match to indoor veg programs.

## Dilution table (start conservative → stronger)

| Material | Suggestion |
|----------|------------|
| Goldenrod → JLF | Start **1:100** drench |
| Blackberry | Management choice — not a fertigation recipe |

## Common mistakes

- Inventing EC 1.8 veg feed for forest garden cherry
- Telling operator they must eradicate goldenrod — operator may keep for dye + JLF biomass
- Confusing nursery cherry production guide with backyard understory$fg_natural_farming_forest_garden_understory$, 4, TRUE, 76),
    ('natural-farming-livestock-plant-feed', 'Livestock plant feed (simple inputs)', NULL, 'trades', 'natural_farming', 'safe', $fg_natural_farming_livestock_plant_feed$# Livestock plant feed (simple inputs)

## What it is (1 paragraph)

Simple on-farm **animal_feed** inputs — comfrey slurry, sprouted grain, chop-and-drop — tracked in `gr33nnaturalfarming` **animal_feed** category. **Not** total mixed ration (TMR) balancing or veterinary formulation.

## When to use

- Chickens, goats, or other livestock with on-farm plant feed supplements
- Linking comfrey or grain sprouts to inventory batches (see demo chicken bootstrap for flock context)

## Ingredients (list with amounts)

**Comfrey slurry:** fresh comfrey leaves + water — wilt/blend to slurry (operator volume by flock size)

**Sprouted grain:** grain soak 8–12 h, drain, sprout 2–5 days until short tails

## Step-by-step preparation

**Comfrey:** harvest comfrey; chop; soak or blend with water; feed fresh within 24 h.

**Sprouts:** rinse daily; feed when sprout tail appears; discard moldy trays.

## Ferment / wait timeline

Comfrey slurry: use **same day**. Sprouts: **2–5 days** from soak to feed-ready.

## Ready signs (smell, foam, color)

Sprouts: white root tails, no mold. Comfrey: fresh green smell.

## Storage

Do not store comfrey slurry long — anaerobic spoilage. Sprouts refrigerated max 1–2 days.

## Safety & water (non-chlorinated, PPE)

Comfrey contains pyrrolizidine alkaloids — **moderation** for poultry; not sole diet. Moldy sprouts — discard.

## How to apply (link to application recipe name)

Record as **animal_feed** input batch in Natural farming inventory — not fertigation application recipes.

## Dilution table (start conservative → stronger)

| Feed | Guidance |
|------|----------|
| Comfrey | Small supplement — not majority of ration |
| Sprouts | Treat as treat/supplement with balanced grain/forage |

## Common mistakes

- Using this guide as complete ration math — out of scope v1
- Feeding comfrey as unlimited primary forage
- Confusing with JLF comfrey ferment for plants — different use path

Cross-link [JLF spring nettle/comfrey](natural-farming-jlf-spring-nettle-comfrey.md) for **plant** fertility, not livestock ration.$fg_natural_farming_livestock_plant_feed$, 4, TRUE, 77)
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

