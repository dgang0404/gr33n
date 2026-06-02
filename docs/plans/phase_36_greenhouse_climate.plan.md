---
name: Phase 36 — Greenhouse climate (shade, panels, fans)
overview: >
  Promote greenhouse environmental control from bootstrap-only actuators into core
  domain concepts: shade/UV screens, glazing (glass/plastic panels), and ventilation
  fans — still scoped to zones with the same properties as today. Automation rules
  for heavy sun (deploy shade), heat (fans), and optional humidity interlocks. Complements
  Phase 35 lighting (supplemental vs blocking sun are different controls).
todos:
  - id: ws1-zone-climate-profile
    content: "WS1: Zone climate profile — greenhouse_climate JSON schema on zone meta_data; validate zone_type=greenhouse; cover_type glass|polycarbonate|film"
    status: pending
  - id: ws2-actuator-taxonomy
    content: "WS2: Actuator taxonomy — first-class actuator_type enum extension: shade_screen, ridge_vent, exhaust_fan, circulation_fan, glazing_panel; config schema per type"
    status: pending
  - id: ws3-automation-templates
    content: "WS3: Automation templates — UV/high-temp rules (lux/par + temp predicates); deploy shade / open vent / fan on; safe retract at night"
    status: pending
  - id: ws4-ui-greenhouse-tab
    content: "WS4: Greenhouse UI — ZoneDetail Greenhouse tab: climate profile, actuator cards, manual override + rule status"
    status: pending
  - id: ws5-bootstrap-to-core
    content: "WS5: Bootstrap → core — migrate greenhouse_climate_v1 template to use core types; seed demo greenhouse zone"
    status: pending
  - id: ws6-sensors-interlocks
    content: "WS6: Sensor interlocks — schedule/rule preconditions for lux, temp, humidity; document when sensors missing (operator override only)"
    status: pending
  - id: ws7-guardian-read
    content: "WS7: Guardian read — summarize_zone_greenhouse_climate; optional enqueue shade/fan commands (existing actuator tools, medium tier)"
    status: pending
  - id: ws8-docs-tests
    content: "WS8: Docs + tests — operator-tour greenhouse section; architecture grow stack; smokes for rule fire + manual shade command"
    status: pending
isProject: false
---

# Phase 36 — Greenhouse climate (shade, panels, fans)

## Status

**Not started.** Depends on **zones**, **actuators**, **automation_rules** + **schedules** worker. **Complements Phase 35** (supplemental lighting vs **blocking** sun). Can run in parallel with Phase 35 after shared automation TZ fix (Phase 35 WS4) is desirable for shade time rules.

**Preconditions:**

- `zone_type` includes `greenhouse` ([`Zones.vue`](../../ui/src/views/Zones.vue), schema)
- Bootstrap template `greenhouse_climate_v1` in [`20260423_farm_bootstrap_templates.sql`](../../db/migrations/20260423_farm_bootstrap_templates.sql) — `shade_cloth_motor`, `exhaust_fan`, etc. as **generic** actuators today
- Rules worker + predicates ([`internal/automation/rules.go`](../../internal/automation/rules.go), [`predicates.go`](../../internal/automation/predicates.go))
- Lux/PAR/temp/humidity sensor types in seed

**Today (gap):** Greenhouse hardware is **template-only** — no core `greenhouse_climate` profile, no typed actuator semantics, no operator UI for shade vs fans vs panels.

---

## Why this phase

Greenhouses need **UV/heavy sun management** (shade cloth, screens), **glazing** awareness (glass vs plastic thermal behavior), and **ventilation** (ridge vents, exhaust/circulation fans). All of this still lives in a **zone** with the same CRUD, sensors, programs, and crop cycles as indoor or outdoor zones.

| Today | After Phase 36 |
|-------|----------------|
| `shade_cloth_motor` as free-text actuator_type | **Typed** `shade_screen` with deploy/retract semantics |
| Bootstrap-only greenhouse pack | **Core** climate profile on `zone.meta_data` |
| No greenhouse-specific UI | **Greenhouse tab** on ZoneDetail |
| Rules are generic predicates | **Templates**: high lux → shade; high temp → fans |
| Operator blind to panel type | **cover_type** + notes in profile (glass / polycarbonate / film) |

**Not in v1:** motorized roof geometry CAD, weather API auto-forecast (defer), multi-zone windward/leeward climate zones.

---

## Design principles

1. **Still a zone** — no parallel "greenhouse" entity; `zone_type='greenhouse'` + `meta_data.greenhouse_climate` profile.
2. **Actuators do the work** — shade/fan/vent commands use existing `control_actuator` + `pending_command`; new types are **taxonomy + config**, not a new Pi protocol.
3. **Block sun ≠ add light** — shade automation is separate from Phase 35 `lighting_programs` (supplemental). Document both on the same zone.
4. **Sensors optional, honesty required** — rules use lux/temp when present; UI + Guardian say when interlocks are skipped (aligns with Phase 34 operator-stated facts pattern).
5. **Safe defaults** — retract shade at night (cron or lux low); fan minimum cooldown; no auto-deploy without predicate unless operator enables "manual schedule only."

---

## Architecture

```
Zone (zone_type=greenhouse, meta_data.greenhouse_climate)
   ├─ cover_type: glass | polycarbonate | film
   ├─ shade_actuator_id, fan_actuator_ids[], vent_actuator_id (optional FK refs)
   └─ automation_policy: auto | manual | schedule_only

Actuators (typed)
   ├─ shade_screen  → commands: deploy | retract | stop
   ├─ exhaust_fan   → on | off
   ├─ circulation_fan → on | off
   ├─ ridge_vent    → open | close | percent (numeric if supported)
   └─ glazing_panel → metadata only in v1 (no motor) OR open/close if motorized

Rules (conditions_jsonb)
   ├─ high_lux → deploy shade (cooldown 30m)
   ├─ high_temp → exhaust_fan on (setpoint scope)
   └─ night cron → retract shade

Operator / Guardian
   └─► manual override or Confirm enqueue_actuator_command (medium tier)
```

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Zone climate profile | `zone.meta_data` schema; validation in zone handler |
| **WS2** | Actuator taxonomy | migration CHECK or enum doc; actuator create validation; OpenAPI |
| **WS3** | Automation templates | rule templates API or seed; predicate examples for lux/temp |
| **WS4** | Greenhouse UI | ZoneDetail tab; actuator cards with typed commands |
| **WS5** | Bootstrap → core | template + demo seed greenhouse zone |
| **WS6** | Sensor interlocks | preconditions; missing-sensor UX |
| **WS7** | Guardian | read tool; reuse actuator propose tools |
| **WS8** | Docs + tests | operator-tour, smokes |

---

## Work-stream detail

### WS1 — Zone greenhouse climate profile

**Goal:** Structured config on existing zones.

**Tasks:**

1. Document JSON schema `meta_data.greenhouse_climate`:
   ```json
   {
     "cover_type": "glass",
     "shade_actuator_id": 12,
     "vent_actuator_id": 13,
     "fan_actuator_ids": [14, 15],
     "automation_policy": "auto",
     "notes": "East-facing polycarbonate end wall"
   }
   ```
2. Validate on zone create/update when `zone_type=greenhouse` (warn if missing actuators when policy=auto).
3. API returns parsed profile in zone GET; PATCH merges meta.

**Acceptance:** Create greenhouse zone with profile; GET echoes typed fields; non-greenhouse zone ignores schema.

### WS2 — Actuator taxonomy (core)

**Goal:** First-class types instead of free-text-only.

**Tasks:**

1. Extend allowed `actuator_type` values: `shade_screen`, `ridge_vent`, `exhaust_fan`, `circulation_fan`, `glazing_panel` (keep legacy `shade_cloth_motor` mapped in UI).
2. Per-type `config` JSON schema (channel, normally_open, max_run_seconds for motors).
3. UI `ActuatorCard` shows type-appropriate command buttons (deploy/retract vs on/off).
4. Pi client: map deploy/retract to same GPIO on/off with config polarity.

**Acceptance:** Create `shade_screen` actuator; manual deploy sends `pending_command`; event logged with command text.

### WS3 — Automation templates

**Goal:** Heavy UV / heat responses out of the box.

**Tasks:**

1. Seed or `POST /farms/{id}/automation/rule-templates/greenhouse` clones:
   - **High PAR/lux** → `control_actuator` deploy shade (predicate on `par_umol` or `lux`, threshold configurable)
   - **High temp** → exhaust fan on (setpoint predicate)
   - **Night retract** → schedule + retract shade (pairs with Phase 35 TZ if available)
2. Templates set `trigger_configuration.zone_id` and populate `conditions_jsonb` (worker ignores trigger_source metadata today — predicates are real).
3. Cooldown defaults (30–60 min) to prevent shade flutter.

**Acceptance:** Smoke: inject sensor reading above threshold → rule fires → `actuator_events` row; cooldown blocks re-fire.

### WS4 — Greenhouse UI tab

**Tasks:**

1. `ZoneDetail.vue` — **Greenhouse** tab when `zone_type=greenhouse`:
   - Climate profile form (cover type, linked actuators)
   - Live sensor strip (lux, temp, RH) with "no sensor" placeholders
   - Active rules summary + last shade/fan events
2. Link to Automation view filtered to zone.

**Acceptance:** Demo greenhouse zone shows tab; deploy shade button works for operator role.

### WS5 — Bootstrap → core migration

**Tasks:**

1. Update `greenhouse_climate_v1` bootstrap to create typed actuators + zone profile + rule templates (not orphan generic names).
2. Demo seed: one greenhouse zone with shade + fan wired.

**Acceptance:** `make dev-stack` or seed apply leaves a testable greenhouse zone.

### WS6 — Sensor interlocks & missing-sensor UX

**Tasks:**

1. Rule templates document required sensors; if missing, show banner "Auto shade disabled — no PAR sensor" and `automation_policy=manual`.
2. Schedule `preconditions` optional on night retract.
3. Cross-doc with Phase 34: operator can state "no lux meter" — Guardian must not propose auto shade rules without sensor or explicit operator fact.

**Acceptance:** Zone without lux sensor cannot enable high-lux template without override flag.

### WS7 — Guardian

**Tasks:**

1. Read tool `summarize_zone_greenhouse_climate` — profile, actuator states, active rules, recent shade/fan events.
2. Reuse `enqueue_actuator_command` for manual deploy/retract (propose→Confirm); no new autonomous shade writes.

**Acceptance:** Chat "is shade deployed in GH-1?" uses read tool + snapshot.

### WS8 — Docs + tests

**Tasks:**

- `operator-tour.md` — greenhouse setup (profile, link actuators, enable high-sun rule)
- `farm-guardian-architecture.md` — grow environment: soil (inventory), fertigation, watering (note), lighting (35), greenhouse (36)
- Smokes: template apply, rule fire, manual deploy

**Acceptance:** Docs in phase-14 index; tests green.

---

## Plant environment stack (cross-phase note)

For the **generic grow program** mental model, document in WS8 how phases fit:

| Need | Phase / area |
|------|----------------|
| Soil / amendments | Inventory (`input_definitions` / batches) — largely done; worm casting = new definition row |
| Fertigation | `gr33nfertigation.programs` + JADAM seed |
| Plain watering (RO/well) | **Future slice** — `water_source` on program or separate irrigation_program (not 36) |
| Lighting | Phase 35 |
| Greenhouse shade/fans/panels | Phase 36 (this) |

---

## Out of scope (this phase)

- Cooling pads / heaters / CO₂ enrichment
- Weather API (rain, wind) automation
- Multi-bay independent climate zones inside one parent zone (use `parent_zone_id` later)
- Guardian revise-specific greenhouse bundles (use Phase 34 loop on actuator proposals)

---

## Recommended order

WS1 → WS2 → WS5 → WS3 → WS6 → WS4 → WS7 → WS8. WS5 early so demo farm is testable.

---

## Definition of done (phase ship)

- [ ] Greenhouse zones carry `greenhouse_climate` profile; typed actuators validated
- [ ] Rule templates for high sun / high temp / night retract
- [ ] ZoneDetail Greenhouse tab + manual typed commands
- [ ] Bootstrap/seed demo greenhouse wired
- [ ] Missing-sensor honesty in UI + Guardian read tool
- [ ] Docs + tests

---

## Using this plan in a new chat

> Implement Phase 36 from `docs/plans/phase_36_greenhouse_climate.plan.md`. Start WS1 zone meta profile + WS2 actuator types. Keep everything zone-scoped. Shade/fan execution uses existing pending_command path. Do not conflate with Phase 35 lighting_programs.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Supplemental photoperiod |
| [phase_34_guardian_pr_iteration.plan.md](phase_34_guardian_pr_iteration.plan.md) | Operator blind-spot facts |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Zone + cycle setup |
| [20260423_farm_bootstrap_templates.sql](../../db/migrations/20260423_farm_bootstrap_templates.sql) | Prior greenhouse_climate_v1 template |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | Wiring shade motor / fan + plumbing = guided Phase 37 procedures |
