# Pattern playbooks (farm bootstrap templates)

> **Audience:** operators setting up a new farm or wiring hardware after choosing a **starter template** at farm creation (Settings → Create farm) or **Apply starter pack** on an existing farm.
>
> **What this is:** each template is an idempotent bundle of zones, sensors, actuators, schedules, automation rules, and starter tasks. Keys match `gr33ncore.apply_farm_bootstrap_template(farm_id, template_key)` and the UI picker in `ui/src/constants/bootstrapTemplates.js`. Re-applying the same key returns `already_applied` and does not duplicate rows.

**Related:** [Operator workflow guide — §2 Zones & §3 Automation](workflow-guide.md), [Pi integration](pi-integration-guide.md), API [`openapi.yaml`](../openapi.yaml) (`POST /farms` with `bootstrap_template`, `POST /farms/{id}/bootstrap-template`).

---

## Conventions (all patterns)

- **Sensors** are logical channels on the farm; the **Pi client** posts readings on an interval. Actuators need a **device** (usually the Pi) so `pending_command` can reach hardware.
- **Automation rules** from templates are created **inactive** (`is_active = false`) so nothing fires until you have real readings and are happy with thresholds — then enable per rule in **Operate → Rules**.
- **Schedules** from templates are usually **inactive** until you confirm cron + timezone; turn them on in **Schedules** when ready.
- **Sensor types** (`sensor_type` text) and **actuator types** are intentionally free-form so you can match your wiring labels without schema migrations.

---

## Indoor photoperiod starter (`jadam_indoor_photoperiod_v1`)

**Best for:** indoor veg / flower rooms with photoperiod lighting, JADAM-style liquid inputs, fertigation reservoirs, and mixing history.

**What you get (summary):** four zones (seedling, veg, flower, outdoor), lighting + irrigation schedules, natural-farming input definitions and starter batches, application recipes, three reservoirs, EC targets, fertigation programs linked to schedules, mixing events, crop cycles, and tasks tied to irrigation.

**Hardware:** follow your room layout — one Pi (or one Pi per relay board) per zone cluster is typical. Map each relay to an **actuator** row; map each physical probe to a **sensor** row. The template does not create Pi `config.yaml` for you — align `sensor_id` / `actuator_id` in the Pi config with the IDs shown in the UI after bootstrap.

**Tuning:** enable schedules one room at a time; adjust cron and `preconditions` (Phase 19) before linking high-volume irrigation to reservoirs.

---

## Chicken coop (`chicken_coop_v1`)

**Best for:** small flock layer or broiler housing with hopper, drinker, ventilation, and optional heat.

**What you get:** one zone **Chicken Coop**; sensors **Coop water level**, **Coop feed level** (% full), **Coop air temperature**, **Coop air humidity**; actuators **feeder hopper**, **water valve**, **exhaust fan**, **heat lamp**; two feed reminder schedules (06:00 / 18:00, inactive); rules (inactive) for low water / low feed → **tasks**, hot → fan on, cold → heat lamp on; weekly egg-collection task.

**Hardware notes:**

- **Water / feed level** — usually a distance or weight sensor scaled to 0–100% in the Pi or in signal conditioning. Name stays generic so you can swap vendor.
- **Relays** — opto-isolated boards are common; match **active_high** / wiring to how `gr33n_client` drives GPIO (see Pi client comments).
- **Temperature / humidity** — one DHT22-style probe can back both channels on the Pi if you mirror two sensor rows to the same pin read (or use separate probes for redundancy).

**Tuning:** set alert thresholds on sensors once you know normal ranges; enable rules only after confirming predicate values against live readings.

---

## Greenhouse / tent climate (`greenhouse_climate_v1`)

**Best for:** controlled environment agriculture — shade screen, ridge vent, exhaust/circulation fans, humidifier, dehumidifier, optional CO₂. **Complements** supplemental lighting ([Phase 35](plans/archive/phase_35_lighting_domain.plan.md)) — blocking sun is not the same control as adding light.

**What you get (after migration `20260603_phase36_greenhouse_climate_v2.sql`):** zone **Greenhouse** with `zone_type=greenhouse` and `meta_data.greenhouse_climate` profile (cover type, linked actuator ids, `automation_policy`); sensors air temp, RH, CO₂, **dew point**, **VPD**, **lux**; typed actuators **shade_screen**, **ridge_vent**, **exhaust_fan**, **circulation_fan**, humidifier, dehumidifier, CO₂ injector; threshold rules including high-lux → deploy shade and high-temp → fan (all **inactive** by default); weekly CO₂ checklist task.

**Legacy farms:** older bootstrap applies may still have `shade_cloth_motor` and `zone_type=indoor`; re-apply `greenhouse_climate_v1` after the Phase 36 migration to pick up typed actuators and the climate profile (idempotent).

**API (operators / integrators):** `PUT /zones/{id}` with `meta_data.greenhouse_climate`; `POST /farms/{id}/actuators` with types like `shade_screen`; clone rules via `POST /farms/{id}/automation/rule-templates/greenhouse`. Commands (`deploy`, `retract`, `open`, `close`) reach the Pi through existing **`pending_command`** — map motor verbs to relay polarity in Pi config.

**Derived channels on the Pi:** you do **not** need a physical “dew point probe” if the Pi already reads temperature and humidity. Configure a **`source: derived`** sensor in `pi_client/config.yaml` (see [pi-integration-guide.md](pi-integration-guide.md) and `pi_client/gr33n_client.py`) so **dew_point** / **vpd** / **heat_index** are computed at the edge and posted like any other sensor. Register matching sensor rows in the UI with the same `sensor_type` strings the rules use (`dew_point`, `vpd`, `lux`).

**Tuning:** VPD, dew-point, and lux thresholds are crop- and glazing-specific. Start with rules off; log readings for a week; then set thresholds and enable one rule at a time. Without a lux sensor, do not enable the high-lux shade rule until a meter is installed (WS6 UI banner — planned).

---

## Drying / cure room (`drying_room_v1`)

**Best for:** post-harvest dry and cure with tight RH and dew-point control.

**What you get:** zone **Drying Room**; temp, RH, dew point sensors; **dehumidifier** and **circulation fan** actuators; rules (inactive) for dew-point high/low band and high-RH circulation; daily environment log task.

**Defaults disclaimer:** thresholds skew toward **cannabis-style** dry/cure (see template rule descriptions). Herbs (basil), ornamentals (orchids), or other crops often need different bands — change predicates before enabling rules.

**Hardware:** circulation prevents microclimates; dehumidification load depends on room volume and outdoor weather leakage — size equipment accordingly.

---

## Small aquaponics (`small_aquaponics_v1`)

**Best for:** single-loop hobby or pilot systems: fish tank + one grow bed.

**What you get:** zones **Fish Tank** and **Grow Bed**; tank sensors (water temperature, pH, ammonia, nitrate), bed sensors (pH, EC); **return pump** and **air pump** actuators; a row in **`gr33naquaponics.loops`** (**Main aquaponics loop**) with meta pointing at the two zone names; daily fish-feed schedule (inactive); rules (inactive) for ammonia spike → task and cold tank → task; daily feed reminder task.

**Hardware:** use off-the-shelf probes appropriate to your water chemistry; calibrate pH and EC on a schedule. Pumps should be on relays with **physical** overflow / leak safeguards — automation is not a substitute for plumbing failsafes.

**Tuning:** ammonia rule default is sensitive by design (early warning). Adjust after you know your biofilter maturity and stocking density.

---

## JADAM indoor starter vs. these patterns

You can apply **at most one** bootstrap template key per farm per key (tracked in `farm_bootstrap_applications`). The **JADAM** starter is the large fertigation + inventory pack; the four **20.5** patterns are smaller vertical slices (husbandry, climate, drying, aquaponics). Pick the template closest to your operation; you can always add zones and sensors manually later — templates are not exclusive “modes,” they are accelerators.

---

## Natural farming recipe packs (Phase 211 — shipped)

**Best for:** Mericle-style switchover or adding vetted JADAM/KNF definitions without retyping from books — after you already have a farm (with or without `jadam_indoor_photoperiod_v1` bootstrap).

**Two paths (same idempotent apply logic):**

| Path | When to use | API |
|------|-------------|-----|
| **Commons catalog import** | Browse vetted packs in **Natural farming → Start** or Help → Catalog | `POST /farms/{id}/commons/catalog-imports` with slug e.g. `jadam-indoor-starter-recipes-v1` |
| **Switchover pack apply** | Wizard maps your commercial EC pattern to a subset pack | `POST /farms/{id}/naturalfarming/apply-pack` with `pack_key` e.g. `mericle_veg_to_jlf_v1` |

**Pack keys (switchover YAML: `data/natural-farming-packs/switchover-packs.yaml`):**

| Pack key | Creates |
|----------|---------|
| `mericle_veg_to_jlf_v1` | JMS + veg JLF inputs, combined drench + JMS foliar recipes |
| `mericle_flower_to_ffj_v1` | FFJ + WCA inputs, flowering foliar recipe |
| `livestock_comfrey_feed_v1` | Comfrey slurry + sprouted grain `animal_feed` inputs (requires **Animals** module for library tab) |

**Idempotency:** imports skip existing input/recipe rows by **name** on that farm; recipe components are upserted. Re-import or re-apply is safe — you get `already_applied` / skipped counts, not duplicates.

**After import:** Natural farming → **Make batch** for JMS/JLF/FFJ; review **Recipes** and pause bottle EC programs in matching veg/flower zones when drenches are ready.

**Related:** [Natural farming studio](operator-tour.md) (§7u), closure test `ui/src/__tests__/phase-211-closure.test.js`, smoke `cmd/api/smoke_phase211_test.go`.
