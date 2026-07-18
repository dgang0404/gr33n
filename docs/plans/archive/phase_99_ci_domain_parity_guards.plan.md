---
name: Phase 99 — CI domain parity guards
overview: >
  make check-ui-domain-parity and CI gates — UI enum lists must match backend/OpenAPI;
  prevents SetpointRow-style drift from shipping silently.
todos:
  - id: ws1-script
    content: "WS1: scripts/check-ui-domain-parity.sh — growth stages, lighting presets, …"
    status: completed
  - id: ws2-make
    content: "WS2: make check-ui-domain-parity + CI workflow step"
    status: completed
  - id: ws3-fixtures
    content: "WS3: Golden files from OpenAPI / Go enum exports"
    status: completed
  - id: ws4-setpoint
    content: "WS4: Fail if SetpointRow default ≠ GROWTH_STAGES length"
    status: completed
  - id: ws5-phase88-link
    content: "WS5: After Phase 88 — UI fetches enums; parity checks API response shape"
    status: completed
isProject: false
---

# Phase 99 — CI domain parity guards

## Status

**Shipped.** Closes **blind spot #10** (enum drift ships for months undetected).

**Depends on:** [Phase 88](phase_88_domain_enums_api.plan.md) (ideal end state); parity guards fallback + backend sources.

**Closure:** [`phase-99-closure.md`](phase-99-closure.md) · **OC-99**

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

- [x] CI fails PR that drops a growth stage from UI only
- [x] Document in INSTALL.md / developer onboarding
- [x] Linked from Phase 88 closure

**Prompt loop:** **`phase 99`**.
