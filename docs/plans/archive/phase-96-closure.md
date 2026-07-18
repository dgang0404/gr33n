# Phase 96 — closure (OC-96)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_96_grow_feeding_program_validation.plan.md`](phase_96_grow_feeding_program_validation.plan.md)

**Depends on:** [Phase 86](phase_86_grow_ops_catalog_chain.plan.md) grow context; [Phase 102](phase_102_fertigation_program_catalog_metadata.plan.md) program `metadata` tags (replaces v1 name heuristics).

**Closes:** Blind spot **#7** — EC strip shows flower profile while pump runs veg recipe.

---

## The one job (done)

> **Attach-time guardrail:** if grow stage or `crop_key` doesn’t fit the linked fertigation program, show a **clear warning** before the operator confirms — Guardian says the same thing in chat.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Active cycle has `plant_id` + `crop_key` + stage | Phase 86 chain |
| **WS1** | Validation via program `metadata` tags | `programmeta.CheckFit` |
| **WS2** | `POST/PATCH crop-cycles` → `program_fit_warnings`; strict mode 422 | `cropcycle/handler.go`, `STRICT_PROGRAM_STAGE_MATCH` |
| **WS3** | Start grow + Water tab mismatch banner | `StartGrowWizard.vue`, `ZoneWaterGrowStory.vue` |
| **WS4** | Guardian zone / fertigation context | `ProgramFitHintLine` in `context_ref.go`, `readtools.go` |
| **WS5** | Smokes | `smoke_phase96_test.go`, `smoke_phase102_test.go` |
| **WS6** | Phase 102 handoff | `programfit` + `programFit.js` read `recommended_*` metadata |

---

## Operator behavior

| Surface | Mismatch UX |
|---------|-------------|
| **Start grow** | Amber banner + ⚠ on program dropdown before submit |
| **Water tab** | Amber banner with **Edit program →** link |
| **API attach** | `program_fit_warnings` array on create/update response |
| **Strict env** | `STRICT_PROGRAM_STAGE_MATCH=1` → HTTP 422 |
| **Guardian** | Zone focus + `summarize_zone_fertigation` cite mismatch |

Programs without Phase 102 metadata tags do not warn (unknown fit).

---

## Automated tests

| Test | Path |
|------|------|
| Crop cycle attach warnings | `cmd/api/smoke_phase96_test.go` |
| Metadata + filter contract | `cmd/api/smoke_phase102_test.go` |
| UI fit helpers | `ui/src/__tests__/program-fit.test.js` |

---

## OC-96

Phase 96 is **closed** when smokes pass and operators see feeding program mismatch warnings at **Start grow**, on the **Water tab**, and in **Guardian** zone context.
