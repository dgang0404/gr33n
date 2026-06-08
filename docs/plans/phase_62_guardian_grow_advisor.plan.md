---
name: Phase 62 — Guardian grow advisor
overview: >
  Make Guardian knowledgeable about actual grow science — VPD, DLI, strain-specific
  targets, stage transitions — grounded in the farm's active cycles and comfort bands.
  Read-only intelligence; no new write tools. Pairs with Phase 56 grow schema.
todos:
  - id: ws1-grow-knowledge
    content: "WS1: Go grow_advisor read tool — active cycle + comfort bands + VPD calc"
    status: pending
  - id: ws2-starters
    content: "WS2: Zone grow strip starters — VPD, DLI, stage advice, transition readiness"
    status: pending
  - id: ws3-persona-grow
    content: "WS3: Grow advisor persona block; vocabulary: VPD, DLI, PPFD, stage plain language"
    status: pending
  - id: ws4-compare-advisor
    content: "WS4: Post-harvest Guardian analysis — what went well vs last run, one recommendation"
    status: pending
  - id: ws5-docs-tests
    content: "WS5: farm-guardian-architecture §9 grow advisor; phase-62-closure; OC-62"
    status: pending
isProject: false
---

# Phase 62 — Guardian grow advisor

## Status

**Planned.** After [Phase 56](phase_56_grow_schema_harvest_analytics.plan.md) `plant_id` FK ships (needs strain data).

---

## The one job

> **Guardian knows this strain, this stage, this room — and gives targeted grow advice, not generic tips.**

---

## WS1 — `grow_advisor` read tool

```go
func GrowAdvisor(ctx, zoneID, cycleID) GrowAdvisorResult {
    cycle  := activeCycle(cycleID)        // stage, days_since_start, plant
    bands  := comfortBands(zoneID)        // temp, humidity
    latest := latestSensorReading(zoneID) // temp_c, rh_pct, co2_ppm
    vpd    := calcVPD(latest.temp, latest.rh)
    dli    := estimateDLI(cycle.lightHours, cycle.ppfd)
    return GrowAdvisorResult{cycle, bands, vpd, dli}
}
```

VPD target range looked up by stage (seedling/veg/flower/late-flower).

---

## WS2 — Starters on zone grow strip

| Chip | Message | When |
|------|---------|------|
| "Is my VPD on target?" | Current VPD vs stage target | Active cycle with sensor data |
| "How many days to flip?" | Days in veg + plant + readiness signals | Veg stage |
| "Ready to harvest?" | Days in flower + trichome guide | Late flower |
| "Optimize light hours" | DLI estimate vs strain target | Any active cycle |
| "Summarize this grow so far" | Cost + yield pace + anomalies | Day 14+ |

---

## WS3 — Grow advisor persona

`context_ref.go` grow tab hint adds:

```
Active cycle: {plant_name}, day {n} of {stage}.
Current VPD: {vpd} kPa (target {target_range} for {stage}).
DLI estimate: {dli} mol/m²/day.
Do not explain what VPD is unless asked. Use plain language for stage recommendations.
```

**Vocabulary rules** (farmer-vocabulary.md addendum):
- "flip" not "transition to 12/12"
- "harvest window" not "day of senescence"
- "light hours" not "photoperiod"
- "VPD" OK — most growers know it; define only on first mention

---

## WS4 — Post-harvest Guardian analysis

After harvest weigh-in (Phase 53 WS1.3):

**Starter:** "What should I do differently next run?"

Guardian compares:
- Yield vs prior cycle in same zone (if available)
- VPD deviations during run (comfort band breaches logged)
- Feeding consistency (autolog lines)
- One concrete recommendation (e.g., "VPD was high during week 3 — tighten humidity band")

---

## WS5 — Docs, tests, OC-62

- `farm-guardian-architecture.md` §9 grow advisor tool
- Vitest: VPD starter present when active cycle + sensor data
- `phase-62-closure.test.js`

---

## Definition of done

- [ ] "Is my VPD on target?" starter surfaces on zone grow strip
- [ ] Guardian answer cites actual current VPD vs stage target
- [ ] Post-harvest starter appears after weigh-in
- [ ] OC-62 closed

---

## Future (not Phase 62)

- Strain database (library of targets per cultivar)
- ML anomaly detection on sensor history
- PPFD sensor integration (requires hardware beyond Sequent HAT)
