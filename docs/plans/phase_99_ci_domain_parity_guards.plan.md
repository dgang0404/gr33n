---
name: Phase 99 — CI domain parity guards
overview: >
  make check-ui-domain-parity and CI gates — UI enum lists must match backend/OpenAPI;
  prevents SetpointRow-style drift from shipping silently.
todos:
  - id: ws1-script
    content: "WS1: scripts/check-ui-domain-parity.sh — growth stages, lighting presets, …"
    status: pending
  - id: ws2-make
    content: "WS2: make check-ui-domain-parity + CI workflow step"
    status: pending
  - id: ws3-fixtures
    content: "WS3: Golden files from OpenAPI / Go enum exports"
    status: pending
  - id: ws4-setpoint
    content: "WS4: Fail if SetpointRow default ≠ GROWTH_STAGES length"
    status: pending
  - id: ws5-phase88-link
    content: "WS5: After Phase 88 — UI fetches enums; parity checks API response shape"
    status: pending
isProject: false
---

# Phase 99 — CI domain parity guards

## Status

**Planned.** Closes **blind spot #10** (enum drift ships for months undetected).

**Depends on:** [Phase 88](phase_88_domain_enums_api.plan.md) (ideal end state); can start earlier with static extraction.

**Closure:** **OC-99**

---

## Blind spot #10

`SetpointRow` missing 2 growth stages shipped for months — no CI guard.

---

## WS1 — Parity script checks

| Check | Source A | Source B |
|-------|----------|----------|
| Growth stages | `internal/croplibrary/catalog.go` ValidGrowthStages | `ui/lib/growHub.js` GROWTH_STAGES |
| Growth stages | OpenAPI GrowthStageEnum | SetpointRow default prop (grep/AST) |
| Lighting presets | `lighting.PresetList()` keys | No hardcoded keys in ui/src (grep denylist) |
| Crop categories | `picker.go categoryOrder` | No dead CATEGORY_ORDER in UI |

Exit non-zero on mismatch.

---

## Acceptance

- [ ] CI fails PR that drops a growth stage from UI only
- [ ] Document in INSTALL.md / developer onboarding
- [ ] Linked from Phase 88 closure

**Prompt loop:** **`phase 99`**.
