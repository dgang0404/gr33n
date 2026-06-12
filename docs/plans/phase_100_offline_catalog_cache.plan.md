---
name: Phase 100 — Offline catalog cache (LAN / mobile)
overview: >
  Picker and domain enums cache locally for LAN-only sites — Plants dropdown works
  when API briefly down; field assistant degrades gracefully without picker 404.
todos:
  - id: ws1-cache
    content: "WS1: IndexedDB cache for picker + domain-enums keyed by farm_id + catalog_version"
    status: pending
  - id: ws2-stale
    content: "WS2: Stale banner when cache catalog_version < API"
    status: pending
  - id: ws3-offline
    content: "WS3: Field assistant / Pi setup — use cache when GET picker fails (not silent fallback)"
    status: pending
  - id: ws4-pwa
    content: "WS4: Service worker optional — document mobile/LAN operator path"
    status: pending
  - id: ws5-smokes
    content: "WS5: Vitest — offline mode shows cached crops + stale warning"
    status: pending
isProject: false
---

# Phase 100 — Offline catalog cache

## Status

**Planned.** Closes **blind spot #11** (mobile / offline operators need picker without live API).

**Depends on:** [Phase 85](phase_85_catalog_bound_plants.plan.md), [Phase 88](phase_88_domain_enums_api.plan.md).

**Closure:** **OC-100**

---

## Blind spot #11

Plants dropdown requires live `GET /farms/{id}/crop-library/picker`. Phase 37 field assistant degrades without LLM — but not without catalog API.

LAN-only warehouse: brief API restart = empty Plants modal.

---

## WS1 — Cache strategy

On successful picker load:

- Store `{ farm_id, catalog_version, groups, fetched_at }` in IndexedDB
- On network failure: serve cache + **“Offline — showing cached knowledge base (date)”**
- Distinct from Phase 85 **404 fallback** (old API) — show **upgrade banner** when 404, cache when offline

---

## Relationship to picker 404 banner (Phase 85 WS6)

| Condition | UX |
|-----------|-----|
| 404 | “Knowledge base API outdated — run migrate & restart API” |
| Network error | Cached picker if available |
| Stale version | “New crops available — reconnect to refresh” |

---

## Acceptance

- [ ] Airplane mode after one successful load → dropdown still populated
- [ ] 404 never silently looks like full catalog
- [ ] Document in crop-knowledge-operator-runbook LAN section

**Prompt loop:** **`phase 100`**.
