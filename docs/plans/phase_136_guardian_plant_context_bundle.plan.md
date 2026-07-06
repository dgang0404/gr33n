---
name: Phase 136 — Guardian plant context bundle (Phase 82 WS7)
overview: >
  Fuse active grow, crop profile, genetics EC, recent sensors, fertigation program,
  light schedule, and grow_advisor into one read-tool block for grow-related questions.
  Closes Phase 82 deferred WS7; pairs with lookup_crop_targets and Phase 132 router.
todos:
  - id: ws1-bundle-readtool
    content: "WS1: plant_context_bundle read tool — renderPlantContextBundle(ctx, q, farmID, zoneID?, cycleID?) bounded prompt block"
    status: pending
  - id: ws2-router-hook
    content: "WS2: Tool router selects bundle on grow/stage/EC/VPD/DLI/flip/harvest intents + active cycle in context_ref"
    status: pending
  - id: ws3-symptom-bridge
    content: "WS3: When vision attachment or symptom keywords — cross-link lookup_crop_symptoms (Phase 106) in same bundle footer"
    status: pending
  - id: ws4-ui-starters
    content: "WS4: Zone/grow starters prefill context_ref cycle_id; chip 'How is this grow doing?'"
    status: pending
  - id: ws5-tests
    content: "WS5: readtools_plant_bundle_test.go; smoke grow question on demo farm; phase-82-closure update WS7 completed"
    status: pending
  - id: ws6-docs
    content: "WS6: architecture §7.0ag plant bundle; crop-knowledge runbook"
    status: pending
isProject: false
---

# Phase 136 — Guardian plant context bundle

**Status:** planned · **Depends on:** [132](phase_132_guardian_read_tool_router.plan.md)

**Continues:** [Phase 82](phase_82_guardian_crop_grounding_hardening.plan.md) WS7 (deferred)

---

## Problem

Grow questions scatter across `lookup_crop_targets`, zone snapshot lines, and RAG. Model must synthesize without a single "this grow right now" block.

---

## Bundle contents (capped)

For active `crop_cycle_id` (from context_ref or inferred zone):

| Section | Source |
|---------|--------|
| Plant + crop_key + variety | `plants`, catalog |
| Profile targets (EC, VPD, DLI, photoperiod) | `lookup_crop_targets` logic inlined |
| Genetics override | Phase 94 |
| Stage + days in stage | cycle row |
| Latest comfort readings | zone sensors (if any) |
| Active fertigation program name | snapshot programs line |
| Light schedule summary | schedules read slice |
| grow_advisor one-liner | flip/harvest readiness if applicable |

Max ~800 tokens — trim sections if over budget.

---

## Router triggers

- `shouldRunPlantContextBundle`: grow, stage, harvest, flip, EC, VPD, DLI, "how is my", crop name from active cycles
- `context_ref.crop_cycle_id` → always bundle

---

## Acceptance

- [ ] "What stage is my veg grow?" on demo farm uses bundle without inventing cycle
- [ ] EC answer uses mS/cm from profile when RAG conflicts (Phase 97 rule preserved)
- [ ] Bundle skipped on Quick chat (no farm_id)

---

## Non-goals

- New crop profiles in YAML (catalog work stays Phase 84+)
- Automatic cycle inference across whole farm in one turn (single cycle focus v1)
