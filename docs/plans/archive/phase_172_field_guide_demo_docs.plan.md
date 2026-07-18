---
name: Phase 172 — Demo field guides + documentation sync
overview: >
  No further Today UI phases planned after 171. This phase expands Guardian
  field guides for demo-farm crops, adds marigold and geranium to the catalog,
  regenerates crop_catalog_from_yaml.sql, and syncs phase-14/current-state docs.
todos:
  - id: ws1-expand-demo-guides
    content: "WS1: Expand chrysanthemum, basil, pepper, strawberry field guides"
    status: completed
  - id: ws2-new-crops
    content: "WS2: Add marigold + geranium to crop_library.yaml, field guides, manifest, SQL seed"
    status: completed
  - id: ws3-docs
    content: "WS3: phase-14 164-171 table, current-state, field-guides README"
    status: completed
  - id: ws4-closure
    content: "WS4: phase-172-closure.test.js"
    status: completed
isProject: false
---

# Phase 172 — Demo field guides + documentation sync

**Status:** shipped · **Follows:** [171](phase_171_demo_zone_layouts_seed.plan.md)

## Verification

```bash
./scripts/generate-crop-catalog-seed.sql.sh --validate
cd ui && npm test -- --run src/__tests__/phase-172-closure.test.js
make rag-ingest-field-guides   # operator — after deploy
```
