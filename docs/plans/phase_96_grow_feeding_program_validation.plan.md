---
name: Phase 96 — Grow feeding program validation
overview: >
  Validate fertigation program matches crop_key and growth stage at start grow and
  on program attach — UI warnings + Guardian honest mismatch alerts.
todos:
  - id: ws1-rules
    content: "WS1: program validation rules — crop_key tag on programs or stage band metadata"
    status: pending
  - id: ws2-api
    content: "WS2: POST crop-cycles warns/blocks primary_program_id mismatch"
    status: pending
  - id: ws3-ui
    content: "WS3: Start grow + Water tab — mismatch banner (veg program + flower stage)"
    status: pending
  - id: ws4-guardian
    content: "WS4: Guardian read tool or prompt block when program stage ≠ cycle stage"
    status: pending
  - id: ws5-smokes
    content: "WS5: smoke — flower stage + veg JLF program → warning visible"
    status: pending
isProject: false
---

# Phase 96 — Grow feeding program validation

## Status

**Planned.** Closes **blind spot #7** (EC strip says flower; pump runs veg recipe).

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md).

**Closure:** **OC-96**

---

## Blind spot #7

| Layer | Can show |
|-------|----------|
| Zone strip | Flower EC from profile stage |
| Fertigation | Veg JLF program (unchanged) |

Operator trusts strip; reservoir runs wrong recipe.

---

## WS1 — Validation rules

Minimum v1 (no new tables):

- Program **name** or **meta** includes intended stages (`early_veg`, `late_veg`) or crop keys
- On attach: if `cycle.current_stage` not in program’s stage band → **`warning`** (soft) or **`422`** (strict mode env)

Better v2: `fertigation_programs.recommended_crop_keys` + `recommended_stages` JSONB seeded in migrations.

---

## WS4 — Guardian

When snapshot shows active cycle + program mismatch:

> "Your grow is in **early_flower** but the linked program **Veg JLF** targets vegetative stages. EC targets on the strip come from the crop profile; the pump recipe may differ. Check Water tab or switch program."

Inject via read-tool enrichment when `primary_program_id` set.

---

## Acceptance

- [ ] Start grow with mismatched program shows visible warning before confirm
- [ ] Water tab links to program edit
- [ ] Guardian mentions mismatch when asked about feeding

**Prompt loop:** **`phase 96`**.
