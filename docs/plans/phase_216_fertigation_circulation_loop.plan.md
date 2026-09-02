---
name: Phase 216 — Fertigation circulation & dose-confirmation loop
overview: >
  Surfaced 2026-08-27 during LED-rig reconnection testing: gr33nfertigation.mixing_events
  and mixing_event_components can record discrete volumetric dose increments, but there is
  no modeled "circulate → wait → average reading → dose or stop" loop between increments.
  That loop was designed conversationally (peristaltic pump, 0.5ml increments, circulate to
  homogenize, wait, average N EC/PPM readings, decide) but never became a DB/worker feature.
  Deliberately being built out and validated against the LED simulation rig before any run
  touches a live reservoir — this plan tracks that work as its own phase.
todos:
  - id: ws1-schema-session-state
    content: "WS1: Add dosing session state tracking (status enum: idle/circulating/settling/sampling/evaluating/dosing/done/aborted, next_action_at) — new gr33nfertigation.dosing_sessions table linked 1:1 to a mixing_event, not new columns on mixing_events itself"
    status: not-started
  - id: ws2-worker-state-machine
    content: "WS2: Worker driver for the state machine — persists current state + next-transition timestamp in DB (not memory) so a restart mid-cycle resumes correctly instead of losing the session"
    status: not-started
  - id: ws3-circulation-actuator
    content: "WS3: New actuator_type 'circulation_pump', bound to a reservoir_id (not a zone_id like irrigation/growlight actuators) — wiring + simulation-mode support in pi_client mirrors the existing SimulationActuatorController pattern"
    status: not-started
  - id: ws4-sensor-averaging
    content: "WS4: Store every raw EC/PPM sensor_readings row during the sampling window (cheap, useful for debugging sensor noise); app-layer computes the rolling average that the state machine evaluates against — no new sensor table needed"
    status: not-started
  - id: ws5-safety-caps
    content: "WS5: Hard per-session max-dose-ml cap enforced before any pump fire (sum of mixing_event_components.volume_added_ml for the session); Guardian propose→confirm gate for any single dose above a configurable delta threshold"
    status: not-started
  - id: ws6-ui-session-panel
    content: "WS6: Fertigation view — live dosing session panel (current state, elapsed circulate/settle time, readings so far, cumulative ml vs cap, manual Abort button)"
    status: not-started
  - id: ws7-reservoir-idle-circulation
    content: "WS7: Reservoir idle-maintenance circulation (rain/bulk-capture tanks) — schedule-only reuse of the existing gr33nfertigation.programs + gr33ncore.schedules cron engine, NOT the WS1/WS2 dosing state machine; register circulation_pump in the actuator taxonomy (AllKnownActuatorTypes, ValidCommands, PulseDurationAllowed), allow irrigation_only programs with no ec_target to target a reservoir, and add a lightweight 'Add maintenance circulation' quick-create on the reservoir card"
    status: not-started
  - id: ws8-tests
    content: "WS8: Worker state machine unit tests with an injectable clock (no real sleeps), migration test, UI session panel test"
    status: not-started
isProject: false
---

# Phase 216 — Fertigation circulation & dose-confirmation loop

**Status:** planned — not started. Sequencing depends on the LED-rig validation pass (Phase 125
simulation rig) surfacing any other automation gaps first; do not start this against a live
reservoir until the state machine is unit-tested and dry-run-validated against simulated
actuators/sensors on the LED rig.

## Why this exists

The dosing design (peristaltic pump, ~0.5ml precision, calibrated ml/second per pump) already
has a home in the schema for **what was added** to a mix
(`gr33nfertigation.mixing_events` + `mixing_event_components`) and **the raw hardware pulse**
(`gr33ncore.actuator_events`). What's missing is the coordination loop *between* dose
increments:

```
dose small increment → circulate (homogenize solution) → wait →
sample EC/PPM N times, average → below target? dose again : done
```

Without this, "dosing" degrades to either (a) a single blind dose with no feedback, or (b)
manual operator judgement on when to stop — exactly the gap called out in the design
conversation this phase is based on (see conversation excerpt logged 2026-08-27).

## Why a new table instead of columns on `mixing_events`

`mixing_events` is an append-once record of a finished mix (`mixed_at`, `final_ec_mscm`, etc.).
The circulation loop is a **long-running, resumable, multi-step process** that exists *before*
a mixing_event is considered final — modeling it as session state on the same row conflates
"the finished record" with "the in-progress machine," and makes a mid-session worker restart
ambiguous (is a row with no `final_ec_mscm` yet unfinished, or just never got readings?).
A dedicated `dosing_sessions` row (one per mixing_event, 1:1) keeps that boundary clean:
session ends by writing back to its parent `mixing_event`'s `final_ec_mscm`/`final_ph`/
`ec_target_met` once `status = done`.

## Hard safety limits (non-negotiable, from the original design review)

1. **Max dose per session** — enforced in the Go handler before any pump-fire command is
   queued, not just documented. Sum `mixing_event_components.volume_added_ml` for the active
   `dosing_session_id`; refuse to queue another increment once the cap is hit, transition the
   session to `done` (capped) instead, and log a warning via `systemlog`.
2. **Guardian gate on outsized deltas** — a dose increment larger than a configured delta
   (e.g. more than 2x the calibrated normal increment) requires Guardian propose→confirm
   rather than firing automatically. Small maintenance corrections stay fully automatic.
3. **Fouled-sensor detection** — if N consecutive dose increments don't move the averaged
   EC/PPM reading in the expected direction, abort the session (`status = aborted`) and raise
   an alert instead of continuing to dose. This is the direct mitigation for the "probe
   fouled/drifted and dosing logic just keeps calling for more" failure mode.

## Sequencing

1. Land WS1–WS2 (schema + worker skeleton) against the **LED simulation rig** first —
   `simulation.actuators` already has a `SimulationActuatorController` pattern that can stand
   in for the circulation pump with no real hardware, and `synthetic_sensors.py` can generate
   a fake EC curve that responds to simulated "dose" events for the worker to react to.
2. Only after the state machine survives a restart mid-cycle in that dry-run, wire WS3
   (real circulation_pump actuator type) for physical hardware.
3. WS5 (safety caps) ships *with* WS1/WS2, not after — this is not being deferred given the
   real-world consequence (nutrient burn) of shipping the happy path first.

## Out of scope for this phase

- Spectrum-shifting / addressable LED lighting control (separate hardware track, discussed
  in the same design conversation but functionally unrelated)
- Per-probe calibration curve UI for moisture/EC sensors (worth its own phase once enough
  grow-cycle data exists to validate drift assumptions)
- Automated JLF/FPJ feeding *schedules* by crop stage (this phase is the dosing mechanism;
  stage-driven triggering is a follow-on once the mechanism itself is proven)

## WS7 — Reservoir idle-maintenance circulation (large outdoor rain tanks)

Surfaced 2026-09-02: a 500-gallon outdoor rain-capture reservoir can sit unused for weeks —
even over a month in a wet stretch — because the farm just isn't drawing from it. Standing
water in a tank that size wants periodic circulation to discourage algae/stagnation even
though the tank is kept covered/dark.

**This is intentionally *not* the WS1/WS2 dosing state machine.** There's no dose to confirm
and nothing to sample against — it's "run the circulation pump for a few minutes every few
days, unconditionally." That's a plain schedule, not a feedback loop, so it reuses
infrastructure that already exists rather than adding new schema/worker code:

- `gr33nfertigation.programs` already supports `reservoir_id` + `schedule_id` (cron) +
  `run_duration_seconds` + `irrigation_only` (no EC target / no nutrient dosing), and
  `program_tick.go` already fires timed pulse commands off a cron schedule for lighting and
  feeding programs today. A reservoir-scoped, `irrigation_only` program with no `ec_target_id`
  is exactly this feature with zero new tables.
- The only gap is the actuator taxonomy: `circulation_pump` (added in WS3) needs to be
  registered in `AllKnownActuatorTypes` and `ValidCommands` (on/off) in
  `internal/handler/actuator/taxonomy.go`, and added to `PulseDurationAllowed` in
  `internal/handler/actuator/command.go` so a timed burst (e.g. 10 minutes) is accepted
  instead of an indefinite "on."
- UI: rather than sending operators through the full feeding-program form (which assumes
  nutrients/EC targets), add a small "+ Add maintenance circulation" quick-create on the
  reservoir card in `Fertigation.vue` that pre-fills `irrigation_only=true`, a sensible cron
  default (e.g. every 2–3 days), and a short duration.

Explicitly out of scope here (no sensor to react to yet): detecting an actual algae bloom or
water-quality fault and adjusting the schedule automatically. Today's version is schedule-only,
same as a lighting program — a turbidity/ORP sensor feeding into `reservoirs` would be a
separate, later phase if this ever needs to become condition-based. Physical mitigations raised
alongside this (opaque/dark tank cover, mosquito fish) are operator/facility concerns, not
schema — a one-line note in the reservoir's existing `description` field is enough to record
them.

Sequencing: depends only on WS3 landing the `circulation_pump` actuator type; does not block
on or wait for WS1/WS2/WS4/WS5 (the dosing loop).

