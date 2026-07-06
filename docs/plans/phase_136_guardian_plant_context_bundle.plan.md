---
name: Phase 136 — Guardian plant context bundle (Phase 82 WS7)
overview: >
  Fuse active grow, crop profile, genetics EC, recent sensors, fertigation program,
  light schedule, and grow_advisor into one read-tool block for grow-related questions.
  Closes Phase 82 deferred WS7; pairs with lookup_crop_targets and Phase 132 router.
todos:
  - id: ws1-bundle-readtool
    content: "WS1: plant_context_bundle read tool — renderPlantContextBundle(ctx, q, farmID, zoneID?, cycleID?) bounded prompt block"
    status: completed
  - id: ws2-router-hook
    content: "WS2: Tool router selects bundle on grow/stage/EC/VPD/DLI/flip/harvest intents + active cycle in context_ref"
    status: completed
  - id: ws3-symptom-bridge
    content: "WS3: When vision attachment or symptom keywords — cross-link lookup_crop_symptoms (Phase 106) in same bundle footer"
    status: completed
  - id: ws4-ui-starters
    content: "WS4: Zone/grow starters prefill context_ref cycle_id; chip 'How is this grow doing?'"
    status: completed
  - id: ws5-tests
    content: "WS5: readtools_plant_bundle_test.go; smoke grow question on demo farm; phase-82-closure update WS7 completed"
    status: completed
  - id: ws6-docs
    content: "WS6: architecture §7.0ag plant bundle; crop-knowledge runbook"
    status: completed
isProject: false
---

# Phase 136 — Guardian plant context bundle

**Status:** shipped · **Depends on:** [132](phase_132_guardian_read_tool_router.plan.md)

**Continues:** [Phase 82](phase_82_guardian_crop_grounding_hardening.plan.md) WS7

---

## Shipped

- **`plant_context_bundle`** read tool fuses cycle header, `lookup_crop_targets`, `grow_advisor`, `summarize_zone`, `summarize_zone_fertigation`, `summarize_zone_lighting`, and optional `lookup_crop_symptoms` footer
- **Router** adds bundle on grow/stage intents or `context_ref.crop_cycle_id`
- **Dedup** skips standalone crop/zone/grow tools when bundle ran
- **UI** zone grow strip chip: **How is this grow doing?** (prefills `crop_cycle_id`)
- **Cap** ~3200 runes; drops lighting/fertigation sections first when over budget

---

## Verify

```bash
go test ./internal/farmguardian/... -run PlantContext -count=1
go test ./cmd/api/... -run Phase136 -count=1   # needs migrated DB
cd ui && npm test -- --run src/__tests__/phase-136-plant-bundle.test.js
```

---

## Acceptance

- [x] "What stage is my veg grow?" on demo farm uses bundle without inventing cycle
- [x] EC answers still governed by Phase 97 structured-truth rule
- [x] Bundle skipped when `farm_id` absent (Quick chat path)
