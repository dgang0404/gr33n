# Phase 71 — closure (OC-71)

**Status:** **Shipped** on `main` (v1 — workspace routes + tab shell; WS2–WS3 editor merge deferred).

**Canonical plan:** [`phase_71_feed_water_unification.plan.md`](phase_71_feed_water_unification.plan.md)

**Depends on:** [Phase 68](phase_68_workspace_shell_spa_nav.plan.md) workspace shell; [Phase 47](phase_47_feeding_water_plain_language.plan.md) vocabulary boundary.

**Closes:** `/feed-water` workspace with four tabs, legacy route redirects, and `/hardware` route restoration (pairs with [Phase 70](phase-70-closure.md)).

---

## The one job (v1)

> **One URL for feeding** — daily status, programs, nutrients, and advanced fertigation — instead of three sidebar entries that felt like duplicates.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Daily tab = `FeedingHub` | `FeedWaterWorkspace.vue` |
| **WS2** | Programs & tanks merged editors | **Deferred v2** — v1 hosts `FeedingAdminHub`; Fertigation reservoir/EC editors still on Advanced |
| **WS3** | Nutrients & mix unified | **Deferred v2** — v1 hosts `Inventory`; Supplies mixing consolidation follows |
| **WS4** | Advanced = full `Fertigation` console | `FeedWaterWorkspace.vue` |
| **WS5** | Legacy redirects + zone_id deep-links | `workspaces.js`, `router/index.js`, `navRelations.js` |
| **WS6** | Tests | `feed-water-tabs.test.js`, `phase-71-closure.test.js` |

---

## Routes

| Legacy | Target |
|--------|--------|
| `/feeding` | `/feed-water?tab=daily` |
| `/operations/feeding` | `/feed-water?tab=programs` |
| `/fertigation` | `/feed-water?tab=advanced` |
| `/feed-water?zone_id=N` | `/zones/N?tab=water` |

---

## Automated tests

| Test | Path |
|------|------|
| Tab model + deep-link | `ui/src/__tests__/feed-water-tabs.test.js` |
| Redirects + workspace routes | `ui/src/__tests__/phase-71-closure.test.js` |

---

## OC-71

Phase 71 is **closed** for v1 when `/feed-water` hosts four tabs, legacy feeding routes redirect correctly, and zone-scoped visits still land on the zone Water tab. Inline program editing and Supplies↔Fertigation mixing merge remain follow-ups.
