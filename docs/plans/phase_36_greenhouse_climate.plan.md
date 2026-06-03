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
    content: "WS1: Zone climate profile — greenhouse_climate JSON schema on zone meta_data; validate zone_type=greenhouse; cover_type glass|polycarbonate|film (999bff1)"
    status: done
  - id: ws2-actuator-taxonomy
    content: "WS2: Actuator taxonomy — shade_screen, ridge_vent, exhaust_fan, circulation_fan, glazing_panel; config validation; POST/GET actuators (999bff1). UI ActuatorCard + Pi deploy/retract polarity → WS4"
    status: done
  - id: ws3-automation-templates
    content: "WS3: Automation templates — high-lux deploy shade, high-temp fan, night retract; POST /farms/{id}/automation/rule-templates/greenhouse + bootstrap rules (0916aba)"
    status: done
  - id: ws4-ui-greenhouse-tab
    content: "WS4: Greenhouse UI — ZoneDetail Greenhouse tab: climate profile, actuator cards, manual override + rule status"
    status: pending
  - id: ws5-bootstrap-to-core
    content: "WS5: Bootstrap → core — greenhouse_climate_v1 v2: zone_type=greenhouse, typed actuators, lux sensor, meta profile (20260603_phase36_greenhouse_climate_v2.sql, 0916aba)"
    status: done
  - id: ws6-sensors-interlocks
    content: "WS6: Sensor interlocks — schedule/rule preconditions for lux, temp, humidity; document when sensors missing (operator override only)"
    status: pending
  - id: ws7-guardian-read
    content: "WS7: Guardian — summarize_zone_greenhouse_climate read tool; enqueue_actuator_command deploy/retract/open/close/stop (f686d76)"
    status: done
  - id: ws8-docs-tests
    content: "WS8: Docs + tests — operator-tour greenhouse section; OpenAPI; architecture grow stack; smokes for rule fire + manual shade command (OC-36)"
    status: pending
isProject: false
---

# Phase 36 — Greenhouse climate (shade, panels, fans)

## Status

**In progress — backend shipped (WS1–WS3, WS5, WS7).** Commits `999bff1` (profile + actuator taxonomy), `0916aba` (bootstrap v2 + rule templates), `f686d76` (Guardian read + extended enqueue commands). **Open:** WS4 (ZoneDetail Greenhouse tab), WS6 (missing-sensor UX), WS8 (operator-tour, OpenAPI, smokes — **OC-36**). Track rollup in [`phase_35_37_operational_closure.plan.md`](phase_35_37_operational_closure.plan.md).

Depends on **zones**, **actuators**, **automation_rules** + **schedules** worker. **Complements Phase 35** (supplemental lighting vs **blocking** sun). Phase 35 WS4 timezone fix is available for future cron-based night retract (bootstrap still uses temp proxy rule today).

**Preconditions:** met.

- `zone_type` includes `greenhouse` ([`Zones.vue`](../../ui/src/views/Zones.vue), schema)
- Bootstrap `greenhouse_climate_v1`: original in [`20260504_phase205_husbandry_climate_bootstraps.sql`](../../db/migrations/20260504_phase205_husbandry_climate_bootstraps.sql); **Phase 36 upgrade** in [`20260603_phase36_greenhouse_climate_v2.sql`](../../db/migrations/20260603_phase36_greenhouse_climate_v2.sql) (`shade_screen`, `ridge_vent`, profile in `meta_data`, lux rules)
- Rules worker + predicates ([`internal/automation/rules.go`](../../internal/automation/rules.go), [`predicates.go`](../../internal/automation/predicates.go))
- Lux/temp/humidity sensor types in seed and bootstrap

**Remaining gap:** No **Greenhouse** tab on ZoneDetail, no missing-sensor banners, no operator-tour/OpenAPI/smokes (WS8). Apply migration `20260603_phase36_greenhouse_climate_v2.sql` on each environment before re-running bootstrap.

---

## Why this phase

Greenhouses need **UV/heavy sun management** (shade cloth, screens), **glazing** awareness (glass vs plastic thermal behavior), and **ventilation** (ridge vents, exhaust/circulation fans). All of this still lives in a **zone** with the same CRUD, sensors, programs, and crop cycles as indoor or outdoor zones.

| Before Phase 36 | Now (backend) | After ship (WS4 + WS8) |
|-----------------|---------------|-------------------------|
| `shade_cloth_motor` free-text | **Typed** `shade_screen`, `ridge_vent`, fans in API + bootstrap v2 | UI command buttons per type |
| Bootstrap-only pack | **Core** `meta_data.greenhouse_climate` + validation on zone PUT/POST | Operator edits profile in ZoneDetail |
| No greenhouse-specific UI | — | **Greenhouse tab** on ZoneDetail |
| Generic predicates only | **Templates** + bootstrap rules (lux → deploy, temp → fan, night retract) | Rule status + manual override in UI |
| Operator blind to panel type | `cover_type` + notes in profile (API) | Same in UI form |

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

| WS | Focus | Primary artifacts | Status |
|----|-------|-------------------|--------|
| **WS1** | Zone climate profile | `zone.meta_data` schema; validation in zone handler | ✅ |
| **WS2** | Actuator taxonomy | taxonomy + POST/GET actuators; OpenAPI → WS8 | ✅ backend |
| **WS3** | Automation templates | rule templates API + bootstrap rules | ✅ |
| **WS4** | Greenhouse UI | ZoneDetail tab; actuator cards with typed commands | pending |
| **WS5** | Bootstrap → core | `20260603_phase36_greenhouse_climate_v2.sql` | ✅ |
| **WS6** | Sensor interlocks | preconditions; missing-sensor UX | pending |
| **WS7** | Guardian | `summarize_zone_greenhouse_climate`; extended enqueue | ✅ |
| **WS8** | Docs + tests | operator-tour, OpenAPI, smokes (OC-36) | pending |

---

## Work-stream detail

### WS1 — Zone greenhouse climate profile ✅

**Goal:** Structured config on existing zones.

**Shipped:** [`internal/handler/zone/greenhouse.go`](../../internal/handler/zone/greenhouse.go) — `GreenhouseClimate`, `ValidateGreenhouseClimate`, validation on `POST /farms/{id}/zones` and `PUT /zones/{id}` when `zone_type=greenhouse`.

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

**Acceptance:** Create greenhouse zone with profile; GET echoes typed fields; non-greenhouse zone ignores schema. ✅ API validation; OpenAPI doc → WS8.

### WS2 — Actuator taxonomy (core) ✅ (backend)

**Goal:** First-class types instead of free-text-only.

**Shipped:** [`internal/handler/actuator/taxonomy.go`](../../internal/handler/actuator/taxonomy.go), `POST /farms/{id}/actuators`, `GET /actuators/{id}` with `valid_commands`. Legacy `shade_cloth_motor` still accepted.

**Tasks:**

1. Extend allowed `actuator_type` values: `shade_screen`, `ridge_vent`, `exhaust_fan`, `circulation_fan`, `glazing_panel` (keep legacy `shade_cloth_motor` mapped in UI). ✅
2. Per-type `config` JSON schema (channel, normally_open, max_run_seconds for motors). ✅
3. UI `ActuatorCard` shows type-appropriate command buttons (deploy/retract vs on/off). → **WS4**
4. Pi client: map deploy/retract to same GPIO on/off with config polarity. → **WS4** (Guardian already enqueues deploy/retract on `pending_command`)

**Acceptance:** Create `shade_screen` actuator; manual deploy sends `pending_command`; event logged with command text. ✅ via API + automation worker path; UI manual button → WS4.

### WS3 — Automation templates ✅

**Goal:** Heavy UV / heat responses out of the box.

**Shipped:** Bootstrap rules in [`20260603_phase36_greenhouse_climate_v2.sql`](../../db/migrations/20260603_phase36_greenhouse_climate_v2.sql); `gr33ncore.apply_greenhouse_rule_templates()`; `POST /farms/{id}/automation/rule-templates/greenhouse` ([`greenhouse_templates.go`](../../internal/handler/automation/greenhouse_templates.go)).

**Tasks:**

1. Seed or `POST /farms/{id}/automation/rule-templates/greenhouse` clones:
   - **High PAR/lux** → `control_actuator` deploy shade (predicate on `par_umol` or `lux`, threshold configurable)
   - **High temp** → exhaust fan on (setpoint predicate)
   - **Night retract** → schedule + retract shade (pairs with Phase 35 TZ if available)
2. Templates set `trigger_configuration.zone_id` and populate `conditions_jsonb` (worker ignores trigger_source metadata today — predicates are real).
3. Cooldown defaults (30–60 min) to prevent shade flutter.

**Acceptance:** Smoke: inject sensor reading above threshold → rule fires → `actuator_events` row; cooldown blocks re-fire. → **WS8** smokes.

### WS4 — Greenhouse UI tab (pending)

**Tasks:**

1. `ZoneDetail.vue` — **Greenhouse** tab when `zone_type=greenhouse`:
   - Climate profile form (cover type, linked actuators)
   - Live sensor strip (lux, temp, RH) with "no sensor" placeholders
   - Active rules summary + last shade/fan events
2. Link to Automation view filtered to zone.

**Acceptance:** Demo greenhouse zone shows tab; deploy shade button works for operator role.

### WS5 — Bootstrap → core migration ✅

**Shipped:** [`20260603_phase36_greenhouse_climate_v2.sql`](../../db/migrations/20260603_phase36_greenhouse_climate_v2.sql) replaces `_bootstrap_greenhouse_climate_v1` body: `zone_type=greenhouse`, typed actuators, GH lux sensor, `meta_data.greenhouse_climate` profile, inactive lux/temp/vent rules.

**Tasks:**

1. Update `greenhouse_climate_v1` bootstrap to create typed actuators + zone profile + rule templates (not orphan generic names). ✅
2. Demo seed: one greenhouse zone with shade + fan wired. ✅ on new bootstrap apply; re-apply after migration on existing dev DBs.

**Acceptance:** `make dev-stack` or seed apply leaves a testable greenhouse zone. ✅ after migration + `greenhouse_climate_v1` bootstrap.

### WS6 — Sensor interlocks & missing-sensor UX

**Tasks:**

1. Rule templates document required sensors; if missing, show banner "Auto shade disabled — no PAR sensor" and `automation_policy=manual`.
2. Schedule `preconditions` optional on night retract.
3. Cross-doc with Phase 34: operator can state "no lux meter" — Guardian must not propose auto shade rules without sensor or explicit operator fact.

**Acceptance:** Zone without lux sensor cannot enable high-lux template without override flag.

### WS7 — Guardian ✅

**Shipped:** [`internal/farmguardian/tools/greenhouse.go`](../../internal/farmguardian/tools/greenhouse.go) — `summarize_zone_greenhouse_climate`; [`actuators.go`](../../internal/farmguardian/tools/actuators.go) — `enqueue_actuator_command` accepts `deploy`, `retract`, `open`, `close`, `stop`.

**Tasks:**

1. Read tool `summarize_zone_greenhouse_climate` — profile, actuator states, active rules, recent shade/fan events. ✅
2. Reuse `enqueue_actuator_command` for manual deploy/retract (propose→Confirm); no new autonomous shade writes. ✅

**Acceptance:** Chat "is shade deployed in GH-1?" uses read tool + snapshot. ✅

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

- [x] Greenhouse zones carry `greenhouse_climate` profile; typed actuators validated (API + bootstrap v2)
- [x] Rule templates for high sun / high temp / night retract (SQL + HTTP clone endpoint; rules inactive by default)
- [ ] ZoneDetail Greenhouse tab + manual typed commands (WS4)
- [x] Bootstrap `greenhouse_climate_v1` uses core types + profile (apply migration first)
- [ ] Missing-sensor honesty in UI (WS6); Guardian read tool ✅
- [ ] Docs + tests (WS8 / OC-36)

---

## Using this plan in a new chat

> Phase 36 backend (WS1–WS3, WS5, WS7) is shipped. Continue with **WS4** (ZoneDetail Greenhouse tab), **WS6** (missing-sensor UX), then **WS8** (operator-tour, OpenAPI, smokes — OC-36). Do not conflate with Phase 35 `lighting_programs`. Shade/fan execution stays on `pending_command`.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Supplemental photoperiod |
| [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | OC-36A ✅ bootstrap; OC-36B/C = WS8 docs/smokes |
| [phase_34_guardian_pr_iteration.plan.md](phase_34_guardian_pr_iteration.plan.md) | Operator blind-spot facts |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Zone + cycle setup |
| [20260504_phase205_husbandry_climate_bootstraps.sql](../../db/migrations/20260504_phase205_husbandry_climate_bootstraps.sql) | Original `greenhouse_climate_v1` bootstrap |
| [20260603_phase36_greenhouse_climate_v2.sql](../../db/migrations/20260603_phase36_greenhouse_climate_v2.sql) | Phase 36 typed actuators, profile, lux rules, `apply_greenhouse_rule_templates` |
| [phase_37_guardian_offline_field_assistant.plan.md](phase_37_guardian_offline_field_assistant.plan.md) | Wiring shade motor / fan + plumbing = guided Phase 37 procedures |
