# Farmer vocabulary — gr33n UI language contract

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

---

## Concepts farmers already know (use freely)

| Term | UI usage |
|------|----------|
| EC | "EC 1.1–1.3" on feeding plan |
| pH | With EC on mix/feed lines |
| Irrigation | **Water only** / plain irrigation (39b) |
| Fertigation / feeding | **Feed** / **Feeding plan** (room-scoped) |
| Reservoir | **Reservoir** — Ready / Needs top-up |
| Run now | Button label (product backlog ✅) |
| Pulse | **Run pump for N seconds** (plain English sublabel) |
| Photoperiod / lights on | Light tab (Phase 35) |
| Humidity / temperature | Climate tab; comfort band (42) |

---

## Navigation labels (target — Phase 40 WS7 + 47)

| Today (Advanced-heavy) | Farmer nav |
|------------------------|------------|
| Zones | **My rooms** |
| Dashboard | **Today** |
| Fertigation | **Feed & water** (47 hub) |
| Setpoints | **Targets** (42) — Advanced: Comfort bands (table) |
| Schedules | Inside feeding plan / Targets — Advanced: Schedules |
| Automation / Rules | **Automations** — Advanced only |
| Inventory | **Supplies** (43) |

---

## Guardian chat

Guardian should mirror this doc ([platform_context](farm-guardian-persona-platform-context.md) + RAG ingest). Say **feeding plan**, **next feed**, **comfort target** — not `patch_fertigation_program` to operators.

---

## Validation (Phase 45 sit-in)

Facilitator **fail** if tester must open a page titled **Setpoints** or the six-tab **Fertigation** console to complete:

- "When is the next feed for Flower Room?"
- "Change feed volume to 0.3 L"
- "Run plain irrigation only"

---

## Related

| Doc | Use |
|-----|-----|
| [farmer_ux_roadmap_40_plus.plan.md](plans/farmer_ux_roadmap_40_plus.plan.md) | Full arc |
| [phase_20_9b_terminology_and_copy_pass.plan.md](plans/phase_20_9b_terminology_and_copy_pass.plan.md) | Earlier pass |
| [operator-tour.md](operator-tour.md) | Operator narrative |
