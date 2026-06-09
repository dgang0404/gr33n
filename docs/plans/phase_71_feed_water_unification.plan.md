---
name: Phase 71 — Feed & Water unification SPA
overview: >
  Merge the three-tier feeding domain — Feed & water (daily), Feeding admin
  (farm-wide summary), and Fertigation (power-user console) — plus the supplies
  *mixing* surface into one Feed & Water workspace with progressive-disclosure
  tabs. Same data, one place: daily per-zone status → programs & tanks → nutrients
  & mix → advanced console. Removes the "Feed & water / Feeding admin /
  Fertigation all explain the same thing" overlap the operator flagged. UI-only.
todos:
  - id: ws1-daily-tab
    content: "WS1: Daily tab = FeedingHub (per-zone next-run cards), the farmer entry; deep-links to zone Water tab kept"
    status: pending
  - id: ws2-programs-tanks-tab
    content: "WS2: Programs & tanks tab = FeedingAdminHub cards + Fertigation reservoirs/EC-targets, edit-in-place (promote read-only admin cards to editors)"
    status: pending
  - id: ws3-nutrients-mix-tab
    content: "WS3: Nutrients & mix tab = supplies mixing + recipes + mixing log (the supplies↔fertigation overlap), one mixing surface"
    status: pending
  - id: ws4-advanced-tab
    content: "WS4: Advanced tab = full Fertigation console (events, crop cycles, raw program editor) behind progressive disclosure"
    status: pending
  - id: ws5-redirects-vocab
    content: "WS5: Redirect /feeding,/operations/feeding,/fertigation → /feed-water?tab=; keep farmer vocab on Daily/Programs, allow 'fertigation' only on Advanced; update wiggle"
    status: pending
  - id: ws6-docs-tests
    content: "WS6: feed-water tabs Vitest, farmer-vocabulary-grow-path guard, phase-71-closure.test.js; operator-tour; OC-71"
    status: pending
isProject: false
---

# Phase 71 — Feed & Water unification SPA

## Status

**Planned.** Builds on the [workspace shell](phase_68_workspace_shell_spa_nav.plan.md) (Phase 68, which declared the `/feed-water` workspace + tabs). UI-only — no schema, no API, no Pi. Reuses the shipped feeding/fertigation/supplies components and endpoints.

**Closure:** **OC-71** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## The one job

> **One place for everything about feeding water and nutrients — from "did Flower Room get watered today" down to the raw fertigation program — without three near-identical sidebar entries.**

---

## Problem

The feeding domain is split across **three sidebar entries** that, to the operator, "explain the same thing," plus a fourth overlap with Supplies:

| Today | Route | View | Really is |
|-------|-------|------|-----------|
| Feed & water | `/feeding` | [`FeedingHub.vue`](../ui/src/views/FeedingHub.vue) | Daily per-zone watering status (farmer) |
| Feeding admin | `/operations/feeding` | [`FeedingAdminHub.vue`](../ui/src/views/FeedingAdminHub.vue) | Farm-wide programs / tanks / EC targets (read-only cards) |
| Fertigation | `/fertigation` | [`Fertigation.vue`](../ui/src/views/Fertigation.vue) | 6-tab power-user console (reservoirs, EC, programs, mixing log, crop cycles, events) |
| Supplies (mixing) | `/operations/supplies` | [`SuppliesHub.vue`](../ui/src/views/SuppliesHub.vue) | "Log a mix" + recipes — overlaps Fertigation mixing |

These are the **same backend domain** at three zoom levels (`hub → admin → editor`) — `FeedingAdminHub` even footer-links "Full feeding editor →" into the matching Fertigation tab. The tiering is intentional, but as **separate sidebar items** it reads as duplication. Phase 71 keeps the tiering as **tab order inside one workspace**.

---

## Design principles

1. **Progressive disclosure as tabs.** Daily (farmer) first → Programs & tanks → Nutrients & mix → Advanced (power-user) last. The depth is opt-in, not three doors.
2. **Reuse components.** Each tab hosts the existing view/sections; no rewrite of feeding logic.
3. **Vocabulary boundary preserved.** Plain language on Daily/Programs (farmer grow-path vocab ban still applies); "fertigation," "setpoint," "EC" allowed on the **Advanced** tab only (continues [Phase 47](phase_47_feeding_water_plain_language.plan.md)/[Phase 49](phase_49_sidebar_nav_polish.plan.md) rules).
4. **Daily stays linked to zones.** The per-zone cards keep deep-linking to the zone Water tab (Phase 69); zone Water tab keeps `ZoneWaterGrowStory`, with "advanced feeding" pointing into this workspace's Advanced tab.
5. **Contract-safe.** Old routes redirect; no schema/API change.

---

## WS1 — Daily tab (farmer entry)

- Tab body = [`FeedingHub.vue`](../ui/src/views/FeedingHub.vue): one card per zone, next-run status, attention flags.
- Keeps deep-links to zone Water tab (`/zones/:id?tab=water`) and the supplies "log a mix" CTA (now an in-workspace tab jump to Nutrients & mix).
- This is the default tab when landing on `/feed-water`.

---

## WS2 — Programs & tanks tab

- Combine [`FeedingAdminHub.vue`](../ui/src/views/FeedingAdminHub.vue) summary cards (Programs, Nutrient tanks, Strength/EC targets) with the matching **Fertigation** sub-editors (reservoirs, EC targets, programs).
- **Promote read-only admin cards to edit-in-place** so the operator doesn't bounce from "summary" to "editor" — the two tiers merge here. (The deep raw editor still lives on Advanced.)
- Surface program ↔ schedule ↔ actuator links (ties to the zone/hardware workspaces via wiggle).

---

## WS3 — Nutrients & mix tab

Resolve the **Supplies ↔ Fertigation mixing overlap** in one place:

- Mixing log + recipes (from Fertigation's mixing tab) and "log a mix" / recipe management (from [`SuppliesHub.vue`](../ui/src/views/SuppliesHub.vue)).
- Shows on-hand nutrient batches relevant to mixing (read), with a link to the Money workspace Supplies tab (Phase 72) for stock/cost management.
- One mental model: "what I mix, from what, by which recipe."

---

## WS4 — Advanced tab (power-user console)

- Full [`Fertigation.vue`](../ui/src/views/Fertigation.vue) console for the remaining power-user surfaces: events, crop cycles, raw program editor.
- This is where technical vocabulary lives. Behind progressive disclosure — most farmers never open it.

---

## WS5 — Redirects, vocabulary, wiggle

- Redirects (declared in Phase 68 WS4, confirm here): `/feeding → /feed-water?tab=daily`, `/operations/feeding → /feed-water?tab=programs`, `/fertigation → /feed-water?tab=advanced`. Supplies mixing entry points jump to `?tab=nutrients`.
- Keep `farmer-vocabulary-grow-path` guard green: Daily/Programs/Nutrients use plain language; "fertigation/EC/setpoint" allowed only on Advanced.
- Update `v-nav-hint`/[`navRelations.js`](../ui/src/lib/navRelations.js): zone Water tab "advanced feeding →" wiggles the Feed & Water workspace; Supplies "log a mix" wiggles the Nutrients tab.

---

## WS6 — Docs, tests, closure (OC-71)

| Artifact | Content |
|----------|---------|
| `ui/src/__tests__/feed-water-tabs.test.js` (new) | Four tabs render; `?tab=` deep-link; default = daily |
| [farmer-vocabulary-grow-path.test.js](../ui/src/__tests__/farmer-vocabulary-grow-path.test.js) | "fertigation/EC" absent on Daily/Programs/Nutrients; allowed on Advanced |
| `ui/src/__tests__/phase-71-closure.test.js` (new) | Programs cards editable inline; mixing consolidated; old routes redirect |
| [operator-tour.md](../operator-tour.md) | Feed & Water: one workspace, depth via tabs |

**OC-71** added and closed when WS1–WS6 ship.

---

## Out of scope

- Zone Water tab internals — stays as [Phase 69](phase_69_zone_workspace_hub.plan.md) (`ZoneWaterGrowStory`).
- Stock quantities / restock / receipts — those stay in the Money/Supplies workspace ([Phase 72](phase_72_money_unification.plan.md)); this workspace only *reads* on-hand for mixing.
- Any schema/API/Pi change.

---

## Definition of done

- [ ] `/feed-water` workspace has Daily / Programs & tanks / Nutrients & mix / Advanced tabs
- [ ] Programs admin cards are editable inline (admin + editor tiers merged)
- [ ] Mixing/recipes live in one Nutrients & mix tab (Supplies↔Fertigation overlap resolved)
- [ ] `/feeding`, `/operations/feeding`, `/fertigation` redirect into the workspace
- [ ] Farmer vocab guard green; technical terms confined to Advanced
- [ ] Vitest green; OC-71 closed

---

## Suggested implementation order

1. WS1 Daily (host FeedingHub) — default tab, lowest risk
2. WS4 Advanced (host Fertigation) — power-user path intact
3. WS2 Programs & tanks (merge admin + fertigation editors)
4. WS3 Nutrients & mix (resolve supplies overlap)
5. WS5 redirects + vocab + wiggle
6. WS6 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_68_workspace_shell_spa_nav.plan.md](phase_68_workspace_shell_spa_nav.plan.md) | Workspace shell + declared tabs |
| [phase_47_feeding_water_plain_language.plan.md](phase_47_feeding_water_plain_language.plan.md) | Vocabulary boundary |
| [phase_49_sidebar_nav_polish.plan.md](phase_49_sidebar_nav_polish.plan.md) | Feeding label disambiguation history |
| [phase_43_operations_stock_feeding_finance.plan.md](phase_43_operations_stock_feeding_finance.plan.md) | Hubs origin |
| [ui/src/views/Fertigation.vue](../ui/src/views/Fertigation.vue) | Advanced console hosted here |

---

## Using this in a new chat

> Read `docs/plans/phase_71_feed_water_unification.plan.md`. UI-only. Merge FeedingHub + FeedingAdminHub + Fertigation + supplies-mixing into the `/feed-water` workspace as progressive-disclosure tabs (Daily → Programs & tanks → Nutrients & mix → Advanced). Redirect old routes. Keep farmer vocab on early tabs; technical terms only on Advanced.
