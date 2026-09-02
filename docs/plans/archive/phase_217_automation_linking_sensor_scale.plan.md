---
name: Phase 217 — Automation resource linking + qualitative sensor scale
overview: >
  Surfaced 2026-09-02 reviewing the Comfort & Automation screens: schedule/rule labels like
  "Light OFF 16/8 Moderate Veg" or "AUTO Light ON 12/12 Flower" read fine in isolation but give
  the operator no way to jump to the zone/reservoir/actuator they actually touch, unlike Feed &
  Water and Natural Farming which already cross-link. Separately, raw sensor numbers (lux 28000,
  EC 1.2, fish tank DO) mean nothing without memorizing ranges — operators want a plain-language
  critical/bad/mid/good/awesome scale instead, with adjustable sensitivity, ideally sourced from
  the crop_library/field-guide data that already encodes ideal ranges per crop + stage.
todos:
  - id: ws1-rule-summary-bugfix
    content: "WS1: Fix ruleSummary.js rendering 'sensor is undefined undefined' for rules with malformed/empty predicates (missing op/value) or unresolvable sensor_id — guard instead of interpolating raw undefined"
    status: completed
  - id: ws2-target-resolver
    content: "WS2: Client-side resolver (schedules/rules -> zone/actuator/reservoir/lighting_program) since schedules and automation_rules don't carry target fields directly — must join through executable_actions.target_actuator_id -> actuators.zone_id, and lighting_programs.schedule_on_id/schedule_off_id -> zone_id, using data already loaded into the farm store (no new API calls)"
    status: completed
  - id: ws3-clickable-labels
    content: "WS3: Wire resolved targets into TargetsSchedulesPanel.vue and TargetsRulesPanel.vue as router-links with v-nav-hint (reusing the existing SensorTile.vue pattern) so hovering/clicking a schedule or rule label wiggles the matching sidebar section and navigates to the actual zone/reservoir"
    status: completed
  - id: ws4-five-tier-scale
    content: "WS4: Extend SensorTile's 3-tier ok/warn/danger classification to 5 tiers (critical/bad/mid/good/awesome) with plain-language labels and a color ramp, replacing the flat 15%-of-range warn band with graduated bands"
    status: completed
  - id: ws5-sensitivity-control
    content: "WS5: Adjustable sensitivity at two levels — a farm-wide default (how wide the 'mid' band is) and a per-sensor override, extending the existing alert_threshold_low/high override pattern rather than replacing it"
    status: completed
  - id: ws6-crop-library-autowiring
    content: "WS6: For a sensor whose zone has an active crop cycle, auto-derive ideal ranges from crop_library.yaml by crop_key + current_stage (ec_min/target/max, ph_min/max, temp_min/max_c, rh_min/max_pct) instead of the hardcoded SENSOR_META defaults; falls back to hardcoded defaults when no crop cycle, no stage match, or sensor_type has no library field mapping"
    status: completed
  - id: ws7-scope-note-aquaponics
    content: "WS7: Explicitly out of scope for this phase — aquaponics/animal domains (fish tank DO, etc.) have no equivalent ideal-range knowledge base yet; keep using hardcoded SENSOR_META defaults there and flag as a follow-on phase once such a source exists"
    status: completed
  - id: ws8-tests
    content: "WS8: Unit tests for the target resolver, the 5-tier classifier, and the sensitivity override precedence (per-sensor > farm default > hardcoded); UI test confirming a schedule/rule label renders as a link with v-nav-hint when a target resolves, and plain text when it doesn't"
    status: completed
isProject: false
---

# Phase 217 — Automation resource linking + qualitative sensor scale

**Status:** shipped 2026-09-02. All eight workstreams landed. See "Closure" at the bottom for
what shipped vs. explicitly deferred.

## Why this exists

Two related but separable UX gaps, raised together while reviewing the Comfort & Automation
screens against Feed & Water / Natural Farming (which already cross-link their cards to the
zones/reservoirs they reference):

1. **Automation labels are dead-end text.** "Light OFF 16/8 Moderate Veg" tells you *what*
   fires but not *where* — no path back to the zone or actuator it controls. The app already
   has exactly the right tool for this (`v-nav-hint` + the sidebar wiggle, see WS3) sitting
   unused outside a handful of views.
2. **Raw sensor numbers require memorization.** "PAR 620.0 µmol/m²/s" or "EC 1.2 mS/cm" mean
   nothing to an operator who doesn't already know the ideal band for that reading. A plain
   critical/bad/mid/good/awesome scale removes that memorization burden — and since
   `data/crop_library.yaml` already encodes ideal EC/pH/temp/RH/VPD ranges per crop + stage,
   most of the data this needs already exists; it's just never been wired to sensor display.

## WS1 — Fix the "sensor is undefined undefined" rendering bug

`ui/src/lib/ruleSummary.js`'s `conditionPhrase()` interpolates `OP_LABEL[p.op] || p.op` and
`p.value` directly into the summary string. When a rule's `conditions_jsonb.predicates` entry
is missing `op`/`value` (or the whole predicates array is malformed), those come through as the
literal string `"undefined"` — combined with `sensorName()`'s `'sensor'` fallback for an
unresolvable `sensor_id`, the result is the "Close coop gate... When sensor is undefined
undefined" row seen in Comfort & Automation → Automations. Fix: skip a predicate entirely
(or say "a condition" generically) when it's missing the fields needed to describe it, rather
than interpolating raw `undefined`.

## WS2 — Target resolver (no new API calls)

Neither `gr33ncore.schedules` nor `gr33ncore.automation_rules` carry a direct
zone/actuator/reservoir reference — they reach their target through
`gr33ncore.executable_actions.target_actuator_id` (joined to `actuators.zone_id`), or for
lighting specifically through `gr33ncore.lighting_programs.schedule_on_id`/`schedule_off_id`
(which already has `zone_id` + `actuator_id` directly). Since `store.zones`, `store.actuators`,
and program/schedule data are already loaded farm-wide via `store.loadAll`, this resolves
entirely client-side — no new endpoints, no N+1 API calls. Output: a `{ path, label }` or
`null` per schedule/rule, `null` meaning "render as plain text, nothing to link to."

## WS3 — Wire it into the UI

`TargetsSchedulesPanel.vue` and `TargetsRulesPanel.vue` wrap the resolved-target case in a
`router-link` with `v-nav-hint`, mirroring `SensorTile.vue`'s existing pattern exactly (that
component already links a sensor tile to `/sensors/{id}` with a `v-nav-hint="'/sensors'"`).
Unresolvable targets (no linked actuator, e.g. a notification-only rule) stay plain text —
no dead links.

## WS4 — Five-tier sensor scale

Replace `SensorTile.vue`'s `ok/warn/danger/unknown` (3 real tiers) with
`critical/bad/mid/good/awesome` (5 tiers), computed as graduated bands within
`[thresholdLow, thresholdHigh]` rather than a single 15%-of-range warn margin on each edge.
Labels and colors change; the underlying threshold-resolution precedence (per-sensor DB
override → farm default → hardcoded `SENSOR_META`) is extended, not replaced, in WS5.

## WS5 — Sensitivity control (farm default + per-sensor override)

Two knobs, in this precedence order:
1. **Per-sensor override** — extends the existing `alert_threshold_low`/`alert_threshold_high`
   pattern; a sensor with explicit thresholds set keeps using them untouched.
2. **Farm-wide default sensitivity** — a new setting (e.g. narrow/normal/wide) that scales how
   wide the "mid" (acceptable-but-not-ideal) band is for every sensor that doesn't have a
   per-sensor override. Narrow = stricter bands, more readings flagged bad/critical sooner;
   wide = more forgiving.

## WS6 — Auto-pull ideal ranges from crop_library.yaml

For a sensor on a zone with an active crop cycle: look up `crop_library.yaml` by
`crop_key` + `current_stage`, map the sensor's `sensor_type` to the matching library field
(`ec` → `ec_min/ec_target/ec_max`, `ph` → `ph_min/ph_max`, `temperature` → `temp_min_c`/
`temp_max_c`, `humidity` → `rh_min_pct`/`rh_max_pct`), and use that as the effective
low/high/target for the 5-tier scale — instead of the hardcoded `SENSOR_META` defaults.
No crop cycle, no stage match, or no field mapping for that `sensor_type` → fall back to
hardcoded defaults exactly as today.

## Out of scope for this phase

- **Aquaponics/animal domains** (fish tank dissolved oxygen, etc.) — no equivalent
  ideal-range knowledge base exists yet comparable to `crop_library.yaml` for crops. These
  sensors keep using hardcoded `SENSOR_META` defaults; a species/system profile source would
  be its own follow-on phase.
- **VPD/DLI-driven sensor types** — `crop_library.yaml` has `vpd_min_kpa`/`vpd_max_kpa` and
  `dli_target`, but there's no first-class `vpd`/`dli` sensor_type in wide use yet; deferred
  until that's confirmed as a real reading type, not computed client-side.
- **Server-side threshold persistence** — WS6's crop-library lookup happens at render time in
  the UI, not by writing computed values back onto `sensors.alert_threshold_low/high`. Keeping
  it read-only avoids the library silently overwriting an operator's explicit override, and
  keeps the crop cycle's stage transitions live without a backend sync step.

## Closure (2026-09-02)

All eight workstreams shipped:

- **WS1** — `ruleSummary.js` filters predicates to ones with both `op` and `value` present
  before rendering them, eliminating the literal "undefined" text.
- **WS2/WS3** — `ui/src/lib/automationTargetLink.js` (`resolveScheduleTarget`,
  `resolveRuleTarget`) resolves entirely from data already in the farm store; wired into
  `TargetsSchedulesPanel.vue`/`TargetsRulesPanel.vue` as `v-nav-hint` links, falling back to
  plain text when nothing resolves.
- **WS4** — `ui/src/lib/sensorQuality.js`'s `classifySensorQuality()` replaced the 3-tier
  check with 5 graduated bands around a target (defaults to range midpoint); new
  `badge-critical/bad/mid/good/awesome` CSS classes.
- **WS5** — `gr33ncore.sensors.quality_sensitivity` (per-sensor override) +
  `farms.meta_data.default_sensor_sensitivity` (farm-wide default, via the existing generic
  `MergeFarmMetaData` query — no new farms column). Precedence: per-sensor > farm default >
  `'normal'`. A sensitivity picker lives on the Live Sensors panel header in `Dashboard.vue`.
- **WS6** — `Dashboard.vue`'s `resolveCropStageByZone()` resolves each active crop cycle's
  plant → crop_key → `GET /farms/{id}/crop-profiles/effective` → stage row, passed to
  `SensorTile` as `cropStageRow`; `idealRangeFromCropStage()` maps `ec`/`ph`/`temperature`/
  `humidity` sensor_types to the matching stage fields. Best-effort throughout — no active
  cycle, unresolvable plant, or no stage match just falls back to hardcoded defaults.
- **WS7** — confirmed out of scope; no code changes.
- **WS8** — 19 classifier/resolver/sensitivity unit tests (`sensor-quality.test.js`), 9
  target-resolver tests (`automation-target-link.test.js`), 1 `ruleSummary` regression test,
  and a static-analysis panel-wiring test (`automation-label-linking.test.js`, matching the
  existing `phase-80-closure.test.js` convention for these same two panel files).

Full UI suite: 285 files / 1459 tests passing. Backend: `go build`/`vet` clean, full `cmd/api`
suite passing against a live Postgres, OpenAPI parity fixed for the new
`PATCH /farms/{id}/sensor-sensitivity` route.

