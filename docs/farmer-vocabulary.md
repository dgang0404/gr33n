# Farmer vocabulary — gr33n UI language contract

**Status:** Published (Phase 47 WS5). **Vocabulary v2 (zones not rooms)** — **shipped** in [Phase 45 WS3](plans/phase_45_farmer_validation_whole_app_polish.plan.md#ws3--copy-pass-v2). Enforced on grow routes by `ui/src/__tests__/farmer-vocabulary-grow-path.test.js` and `phase-45-ws3-closure.test.js`.

**Audience:** Product, UX, and developers shipping Phases 40–47 and 45 sit-in validation.

**Purpose:** Farmers understand **EC, irrigation, fertigation, photoperiod, humidity** in the grow room. They should not need to learn **setpoint, cron, predicate, executable_action**, or six admin apps to run gr33n.

**Canonical feeding phase:** [phase_47_feeding_water_plain_language.plan.md](plans/phase_47_feeding_water_plain_language.plan.md)

---

## Rule

> On **grow routes**, use **job language**. On **Advanced / agronomist routes**, technical terms are allowed with HelpTips.

**Grow routes (v1):** `/zones`, `/zones/:id`, `/feeding` (47), Dashboard morning widgets, Operations Supplies (43) — not `/setpoints`, `/automation`, `/schedules` as primary nav labels.

---

## Replacements (required on grow paths)

| Do not show (label or CTA) | Use instead |
|----------------------------|-------------|
| Setpoint / Setpoints | **Comfort target** (climate) or **Target range** |
| "Add under Setpoints" | **Set comfort target here** or **Add target for humidity** |
| Fertigation (sidebar) | **Feed & water** |
| Manage → Fertigation | **Feeding plan** (stay on zone Water) |
| Schedule (alone) | **What runs when** or **Next feed** |
| Rule (alone) | **Automation** |
| Automation rules (nav) | **Automations** (Advanced) or sentence on card |
| executable_action | **Step in feeding plan** |
| predicate | *(omit — show rule sentence)* |
| cron / cron_expression | *(omit — show "Every day at 6:00 AM")* |
| application_recipe_id | **Recipe** (link to Supplies) |
| zone_setpoints | *(never — internal only)* |
| metadata.steps | **Feeding steps** (Advanced only) |
| input_batches | **Supplies on hand** |
| Triggering event source | *(never in UI)* |
| **Room** (generic grow-area label) | **Zone** / **My zones** / **this zone** — see [Vocabulary v2](#vocabulary-v2--zones-not-rooms-phase-45-ws3) |

---

## Concepts farmers already know (use freely)

| Term | UI usage |
|------|----------|
| EC | "EC 1.1–1.3" on feeding plan |
| pH | With EC on mix/feed lines |
| Irrigation | **Water only** / plain irrigation (39b) |
| Fertigation / feeding | **Feed** / **Feeding plan** (zone-scoped) |
| Zone | **Grow area** — default product word (greenhouse bay, bench, field block, or indoor room) |
| Reservoir | **Reservoir** — Ready / Needs top-up |
| Run now | Button label (product backlog ✅) |
| Pulse | **Run pump for N seconds** (plain English sublabel) |
| Photoperiod / lights on | Light tab (Phase 35) |
| Humidity / temperature | Climate tab; comfort band (42) |

---

## Vocabulary v2 — zones not rooms (Phase 45 WS3)

Phase 47 used **room** in grow-path copy and nav (**My rooms**) to match indoor demo farms. Sit-in feedback and broader ag use (greenhouse bays, propagation benches, field blocks, drying rooms — not only indoor **rooms**) favor **zone** as the default product word. The API and schema stay `zone`; only user-visible labels change.

### Rule

| Context | Use | Do not use |
|---------|-----|------------|
| Nav, page titles, empty states, hub cards | **Zone** / **My zones** / **this zone** | **Room** as the generic label for every grow area |
| A specific zone’s display name | The name as stored (e.g. **Flower Room**, **North Bench**) | Renaming zones to drop “Room” |
| Guardian starters and body copy | **Zone name** or **this zone** | **This room** when no name is loaded |
| Feeding hub | **One card per zone** | **One card per room** |

### Target label map (v2)

| Surface (today after 47) | Target (45 WS3) |
|--------------------------|-----------------|
| Sidebar **My rooms** | **My zones** (navTitle: *Grow areas — water, light, and climate per zone*) |
| Mobile bottom nav **Rooms** | **Zones** |
| `/zones` H1 **My rooms** | **My zones** |
| Dashboard section **Zones** vs nav **rooms** | **My zones** everywhere (one label) |
| Feeding hub empty state **create zones first… each room's** | **…each zone's** |
| `guardianStarters.js` fallback **this room** | **this zone** |
| `ZoneWaterGrowStory` default **This room** | **This zone** |

### Implementation notes (45 WS3)

- Centralize grow-path labels in `ui/src/lib/farmerVocabulary.js` (export map + optional Vitest for **room** as generic nav/body term on grow routes).
- Extend `farmer-vocabulary-grow-path.test.js` — fail on generic **My rooms** / **one card per room** / **this room** (allow **Room** inside zone **names** and seed data like `Flower Room`).
- Update [operator-tour.md](operator-tour.md), Guardian starters, and nav tests (`nav-groups.test.js`) in the same PR.

**Sit-in still valid:** “When is the next feed for **Flower Room**?” — tests a **named zone**, not the word “room” as the product term.

---

## Navigation labels (target — Phase 40 WS7 + 47 → v2 in 45)

| Today (Advanced-heavy) | Farmer nav (v2) |
|------------------------|-----------------|
| Zones | **My zones** |
| Dashboard | **Today** |
| Fertigation | **Feed & water** (47 hub) |
| Setpoints | **Targets** (42) — Advanced: Comfort bands (table) |
| Schedules | Inside feeding plan / Targets — Advanced: Schedules |
| Automation / Rules | **Automations** — Advanced only |
| Inventory | **Supplies** (43) — `/operations/supplies` |
| Fertigation (admin) | **Feeding (details)** (43) — `/operations/feeding`; daily grow → **Feed & water** (47) |
| Costs | **Money** (43) — `/operations/money` |

---

## Guardian chat

Guardian should mirror this doc ([platform_context](farm-guardian-persona-platform-context.md) + RAG ingest). Say **feeding plan**, **next feed**, **comfort target** — not `patch_fertigation_program` to operators.

---

## Validation (Phase 45 sit-in + CI)

**Automated (47 WS5):** Vitest scans grow-path Vue templates and `plantNeeds.js` for the ban list above. Run `npm test` in `ui/`.

Facilitator **fail** if tester must open a page titled **Setpoints** or the six-tab **Fertigation** console to complete:

- "When is the next feed for Flower Room?"
- "Change feed volume to 0.3 L"
- "Run plain irrigation only"

---

## Grow advisor addendum (Phase 62)

Guardian grow-science starters and answers on the zone grow strip and post-harvest flow:

| Prefer | Avoid (unless operator uses the term) |
|--------|----------------------------------------|
| **flip** | transition to 12/12 |
| **light hours** | photoperiod (OK in Advanced / crop profile detail) |
| **harvest window** | day of senescence |
| **VPD** | Long vapor-pressure lecture unless asked |

Targets always come from the assigned crop profile (`lookup_crop_targets` / `grow_advisor`) — never invented.

---

## Related

| Doc | Use |
|-----|-----|
| [farmer_ux_roadmap_40_plus.plan.md](plans/farmer_ux_roadmap_40_plus.plan.md) | Full arc |
| [phase_20_9b_terminology_and_copy_pass.plan.md](plans/phase_20_9b_terminology_and_copy_pass.plan.md) | Earlier pass |
| [operator-tour.md](operator-tour.md) | Operator narrative |
