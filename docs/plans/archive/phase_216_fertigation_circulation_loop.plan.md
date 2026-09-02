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
    status: completed
  - id: ws2-worker-state-machine
    content: "WS2: Worker driver for the state machine — persists current state + next-transition timestamp in DB (not memory) so a restart mid-cycle resumes correctly instead of losing the session"
    status: completed
  - id: ws3-circulation-actuator
    content: "WS3: New actuator_type 'circulation_pump', bound to a reservoir_id (not a zone_id like irrigation/growlight actuators) — wiring + simulation-mode support in pi_client mirrors the existing SimulationActuatorController pattern"
    status: completed
  - id: ws4-sensor-averaging
    content: "WS4: Store every raw EC/PPM sensor_readings row during the sampling window (cheap, useful for debugging sensor noise); app-layer computes the rolling average that the state machine evaluates against — no new sensor table needed"
    status: completed
  - id: ws5-safety-caps
    content: "WS5: Hard per-session max-dose-ml cap enforced before any pump fire (sum of mixing_event_components.volume_added_ml for the session); Guardian propose→confirm gate for any single dose above a configurable delta threshold"
    status: completed
  - id: ws6-ui-session-panel
    content: "WS6: Fertigation view — live dosing session panel (current state, elapsed circulate/settle time, readings so far, cumulative ml vs cap, manual Abort button)"
    status: completed
  - id: ws7-reservoir-idle-circulation
    content: "WS7: Reservoir idle-maintenance circulation (rain/bulk-capture tanks) — schedule-only reuse of the existing gr33nfertigation.programs + gr33ncore.schedules cron engine, NOT the WS1/WS2 dosing state machine; register circulation_pump in the actuator taxonomy (AllKnownActuatorTypes, ValidCommands, PulseDurationAllowed), allow irrigation_only programs with no ec_target to target a reservoir, and add a lightweight 'Add maintenance circulation' quick-create on the reservoir card"
    status: completed
  - id: ws8-tests
    content: "WS8: Worker state machine unit tests with an injectable clock (no real sleeps), migration test, UI session panel test"
    status: completed
isProject: false
---

# Phase 216 — Fertigation circulation & dose-confirmation loop

**Status:** shipped 2026-09-02. All eight workstreams landed and are covered by tests (Go
smoke tests against a live DB with the worker in simulation mode, plus UI unit tests for the
dosing session panel helpers). See "Closure" below for what was actually validated vs. what
still needs a physical LED-rig dry run before touching a live reservoir.

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

## Closure (2026-09-02)

All eight workstreams shipped:

- **WS1** — `gr33nfertigation.dosing_sessions` table + status enum + sqlc queries.
- **WS2/WS4/WS5** — landed together as planned. `internal/automation/dosing_session_tick.go`
  drives idle→dosing→circulating→settling→sampling→evaluating, persisting state +
  `next_action_at` after every transition. Sampling accumulates a running EC sum
  (`samples_sum_ec_mscm`) for a fully-typed average. Safety caps: `max_total_dose_ml`
  checked before every fire, an outsized-delta guard (no override path exists yet to trigger
  it, but it's in place), and a fouled/no-effect-sensor abort after `max_no_improvement_cycles`.
- **WS3** — `circulation_pump` registered in the actuator taxonomy (type, valid commands,
  pulse-duration allowlist) and the device type registry. Confirmed `pi_client` needs zero
  code changes — `ActuatorController`/`SimulationActuatorController` dispatch by `driver`
  (gpio/simulation/relay_hat), not `actuator_type`.
- **WS6** — HTTP surface (`internal/handler/fertigation/dosing_sessions.go`: start/list/get/
  abort) + a live dosing sessions panel on the Fertigation mixing-log tab (status, elapsed
  time, mL dosed vs cap, latest avg EC, manual Abort, 15s poll).
- **WS7** — `reservoirs.circulation_actuator_id` column + a quick-create UI shortcut that
  chains the existing schedule/program/executable-action engine — no new worker code, since
  this is schedule-only maintenance, not the dosing loop.
- **WS8** — `cmd/api/smoke_phase216_dosing_session_test.go` drives the real worker
  (`simulation=true`) through a full cycle against a live Postgres, forcing `next_action_at`
  into the past between ticks instead of sleeping in real time. Covers the happy path (reaches
  target, writes back to `mixing_events`) and the already-above-target abort guard. UI helpers
  (`dosingStatusLabel`, `dosingElapsed`, `isDosingSessionTerminal`) extracted to
  `ui/src/lib/dosingSession.js` with 9 unit tests. OpenAPI parity fixed (5 new routes +
  `DosingSession`/`DosingSessionCreate` schemas).

**What's validated:** the state machine's logic (transitions, safety caps, target/abort
decisions) against a real database, worker in simulation mode.

**What's *not* yet validated, and still gates a live reservoir per the Sequencing section
above:** an actual physical dry run on the LED simulation rig with `synthetic_sensors.py`
feeding a fake EC curve and a real (simulated) circulation actuator pulse observed on the LED
bank. That operational dry run is a follow-up step, not blocked on any further code changes.

