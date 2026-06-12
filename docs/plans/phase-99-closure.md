# Phase 99 — closure (OC-99)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_99_ci_domain_parity_guards.plan.md`](phase_99_ci_domain_parity_guards.plan.md)

**Depends on:** [Phase 88](phase_88_domain_enums_api.plan.md) (UI loads enums from API; parity guards fallback + backend sources).

**Closes:** Blind spot **#10** — enum drift (e.g. SetpointRow missing `transition` / `flush`) shipping undetected.

---

## The one job (done)

> **`make check-ui-domain-parity`** and CI gates ensure UI enum lists match backend/OpenAPI — PRs that drop a growth stage from UI-only code fail before merge.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | `scripts/check-ui-domain-parity.sh` | growth stages, lighting denylist, Vitest |
| **WS2** | `make check-ui-domain-parity` + CI | `.github/workflows/ci.yml` go job |
| **WS3** | Go golden checks vs OpenAPI + croplibrary | `internal/platform/domainenums/parity_test.go` |
| **WS4** | SetpointRow / fallback length guard | `ui/src/__tests__/ui-domain-parity.test.js` |
| **WS5** | Post–Phase 88 API-shape alignment | fallback `domainEnums.fallback.js` ↔ `enums.go` |

---

## What CI checks

| Check | Source A | Source B |
|-------|----------|----------|
| Growth stages | `croplibrary.ValidGrowthStages` | `domainenums.All()` |
| Growth stages | OpenAPI `GrowthStageEnum` | `domainenums.All()` order |
| UI fallback | 11 canonical stages | `FALLBACK_GROWTH_STAGE_VALUES` = `GROWTH_STAGES` |
| SetpointRow | imports `FALLBACK_GROWTH_STAGE_VALUES` | full 11-stage spread default |
| Lighting presets | `lighting.PresetList()` keys | no hardcoded keys in `ui/src` (except loader) |
| Crop categories | API picker groups | no dead `CATEGORY_ORDER` in UI |

---

## Operator / developer usage

```bash
make check-ui-domain-parity   # local pre-push
```

Documented in [`INSTALL.md`](../INSTALL.md) § Go tests. UI Vitest parity also runs in the **`ui`** CI job (`npm test`).

---

## Automated tests

| Test | Path |
|------|------|
| Go enum ↔ croplibrary ↔ OpenAPI | `internal/platform/domainenums/parity_test.go` |
| Lighting preset keys | `internal/handler/lighting/presets_parity_test.go` |
| UI fallback + SetpointRow | `ui/src/__tests__/ui-domain-parity.test.js` |

---

## OC-99

Phase 99 is **closed** when CI fails PRs that drift UI enums from backend/OpenAPI. Completes **Arc B** (Phases 88–92 + **99**).
