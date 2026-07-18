# Phase 100 — closure (OC-100)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_100_offline_catalog_cache.plan.md`](phase_100_offline_catalog_cache.plan.md)

**Depends on:** Phase 85 picker API, Phase 88 domain enums API (offline fallback for both).

---

## The one job (done)

> **Plants picker and domain enums survive brief LAN/API outages** — IndexedDB cache keyed by farm + catalog version; 404 shows upgrade banner, network errors serve cache with honest offline/stale labels.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | IndexedDB cache for picker + domain enums | `ui/src/lib/catalogCache.js` |
| **WS2** | Stale banner when cached version < last known | `CropLibraryPicker.vue` + `isStaleCatalogVersion` |
| **WS3** | Loader uses cache on network failure; 404 ≠ offline | `cropLibraryLoader.js` |
| **WS4** | PWA / mobile path documented | [`crop-knowledge-operator-runbook.md`](../crop-knowledge-operator-runbook.md) § LAN-only |
| **WS5** | Vitest offline + stale + 404 vs cache | `ui/src/__tests__/catalog-cache.test.js` |

---

## Operator behavior

| Condition | UX |
|-----------|-----|
| **404** on picker | Amber **Knowledge base API outdated** — profile-only fallback (not full catalog) |
| **Network error** | Sky **Offline — showing cached knowledge base** if prior successful load |
| **Stale version** | Amber sub-line **New crops may be available — reconnect and reload** |
| **Online success** | Cache refreshed; `gr33n_last_catalog_version` updated in localStorage |

One online session populates the cache before going offline (warehouse Wi‑Fi / brief API restart).

---

## Automated tests

| Test | Path |
|------|------|
| Cache store/retrieve, stale, network detect | `ui/src/__tests__/catalog-cache.test.js` |
| Offline serve + 404 degraded + cache on success | same file — `crop library loader` describe block |

---

## OC-100

Phase 100 is **closed** when Vitest passes and operators follow the LAN section in the crop-knowledge runbook. No Go smoke required (UI-only cache layer).
