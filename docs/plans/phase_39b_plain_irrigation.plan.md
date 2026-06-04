---
name: Phase 39b — Plain irrigation (RO/well, no mix)
overview: >
  Farms that only pulse irrigation pumps (RO, well, municipal) without nutrient mixing
  need an honest program type and UI copy — not fake mix_batch or manual mixing events.
  Builds on Phase 39 command queue (WS1) and Phase 38 pulse; does not duplicate 39 mix math.
todos:
  - id: ws1-program-type
    content: "WS1: irrigation_only program flag or program_type — schedule fire enqueues pulse only (no mix_batch)"
    status: pending
  - id: ws2-worker-path
    content: "WS2: program_tick branch — irrigation_only skips mix calculator; single pulse enqueue with provenance"
    status: pending
  - id: ws3-zone-water-copy
    content: "WS3: Zone Water + Fertigation UI — show 'irrigation only' badge; hide mix preview when N/A"
    status: pending
  - id: ws4-docs-seed
    content: "WS4: operator-tour + workflow-guide + optional seed program; smoke irrigation_only tick"
    status: pending
isProject: false
---

# Phase 39b — Plain irrigation (RO / well)

## Status

**Planned.** Depends on **Phase 39 WS1** (device command queue). Referenced from [Phase 39](phase_39_edge_fertigation_execution.plan.md) out-of-scope table as **39b**.

**Indexed in:** [`pre_development_gaps_index.plan.md`](pre_development_gaps_index.plan.md) (gap **A4**).

---

## Problem

| Farm type | Today | Gap |
|-----------|--------|-----|
| Nutrient mixing | Manual mixing events; 39 adds automated `mix_batch` | Covered by 39 |
| RO / well / plain water | Operators use pulse (38) or programs with `run_duration_seconds` only | Programs UI still implies fertigation / EC / recipes; no first-class **irrigation-only** story |

Operators on plain-water systems should not see mix preview, base EC, or recipe requirements.

---

## Design principles

1. **Queue-based** — same `pulse` command type as Phase 38/39; no new executor.
2. **No mix calculator** — `irrigation_only` programs never call `internal/fertigation/mixplan`.
3. **Audit** — `fertigation_events` or a slim `irrigation_events` path TBD in WS1 (prefer reusing fertigation_events with `event_type=irrigation` if schema allows; document choice in WS1 spike).
4. **Defer** — peristaltic vendor protocols, interval pulse trains (39 v1 out of scope).

---

## WS1 — Program type / flag

**Goal:** Data model distinguishes mix+feed vs irrigate-only.

**Options (pick one in implementation spike):**

- `fertigation_programs.irrigation_only boolean default false`, or
- `program_type enum: fertigation | irrigation_only`

**Acceptance:** API rejects `application_recipe_id` on irrigation-only programs (400 + plain message).

---

## WS2 — Worker program tick

**Goal:** Schedule fire → enqueue **one** `pulse` on irrigation pump; no `mix_batch`.

**Tasks:**

1. Branch in [`dispatchProgramActuator`](../../internal/automation/program_tick.go) (or 39 WS5 pipeline) when `irrigation_only`.
2. Provenance: `source=program`, `program_id`, `schedule_id`.
3. Idempotency: same as 39 WS5.

**Acceptance:** Smoke: irrigation_only program tick creates exactly one queued pulse, zero mix_batch rows.

---

## WS3 — UI copy

**Goal:** Zone Water and Fertigation do not show mix UI for irrigation-only programs.

**Tasks:**

1. Program form: checkbox “Irrigation only (no nutrients)” with HelpTip.
2. Zone Water grow story (40 WS5): branch — last pulse / next run, no mix preview.
3. Guardian read tools: do not suggest mix for irrigation-only programs.

**Acceptance:** RO program on demo farm — Water tab has no “Preview mix” button.

---

## WS4 — Docs, seed, smoke

| Artifact | Content |
|----------|---------|
| **workflow-guide** | § automated vs manual vs irrigation-only |
| **operator-tour** | Subsection under plant needs / fertigation |
| **seed** | Optional one `irrigation_only` program on a well-water zone |
| **smoke** | `smoke_phase39b_irrigation_only_test.go` |

---

## Relationship to Phase 39 and 40

```mermaid
flowchart LR
  P39[Phase 39 queue + mix]
  P39b[Phase 39b irrigation_only]
  P40[Phase 40 Water grow story]
  P39 --> P39b
  P39b --> P40
```

- **39** must ship WS1 before **39b** worker path is safe.
- **40 WS5** should handle both mix and irrigation-only branches in copy.

---

## Out of scope (v1)

- Separate `irrigation_program` table (mentioned as future in 39 plan — defer unless WS1 spike requires it)
- Soil moisture / moisture-probe closed loop
- Multi-zone manifold routing

---

## Recommended order

WS1 → WS2 → WS3 → WS4 (after 39 WS1; can parallel 39 WS4 Pi work)

---

## Definition of done

- [ ] irrigation_only programs cannot attach recipes
- [ ] Program tick enqueues pulse only via queue
- [ ] UI hides mix affordances for irrigation-only
- [ ] Docs + smoke + optional seed

---

## Related

| Doc | Use |
|-----|-----|
| [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) | Queue prerequisite |
| [phase_40_unified_farmer_ux_zone_cockpit.plan.md](phase_40_unified_farmer_ux_zone_cockpit.plan.md) | Water grow story |
| [pre_development_gaps_index.plan.md](pre_development_gaps_index.plan.md) | Gap A4 |

---

## Using this plan in a new chat

> Implement Phase 39b after Phase 39 WS1 queue. Add irrigation_only program path — pulse enqueue only, no mix_batch. Update zone Water and Fertigation copy. See `docs/plans/phase_39b_plain_irrigation.plan.md`.
