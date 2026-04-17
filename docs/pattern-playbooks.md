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

**Best for:** controlled environment agriculture — fans, humidifier, dehumidifier, shade motor, optional CO₂.

**What you get:** zone **Greenhouse**; sensors air temp, RH, CO₂, **dew point**, **VPD** (stored in **pascals** in the template so a ~1.5 kPa “high VPD” threshold is `1500` — align your Pi or UI labels accordingly); actuators exhaust fan, humidifier, dehumidifier, shade motor, CO₂ injector; several threshold→actuator rules (inactive); weekly CO₂ checklist task.

**Derived channels on the Pi:** you do **not** need a physical “dew point probe” if the Pi already reads temperature and humidity. Configure a **`source: derived`** sensor in `pi_client/config.yaml` (see [pi-integration-guide.md](pi-integration-guide.md) and `pi_client/gr33n_client.py`) so **dew_point** / **vpd** / **heat_index** are computed at the edge and posted like any other sensor. Register matching sensor rows in the UI with the same `sensor_type` strings the rules use (`dew_point`, `vpd`).

**Tuning:** VPD and dew-point targets are crop- and stage-specific (flower vs. veg vs. dry/cure). Start with rules off; log readings for a week; then set thresholds and enable one rule at a time.

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
