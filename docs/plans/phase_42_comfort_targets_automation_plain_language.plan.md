---
name: Phase 42 — Comfort targets & automation (plain language)
overview: >
  Replace the "Setpoints / Schedules / Rules are a database admin panel" feel with farmer
  jobs: comfort bands per room, "what runs when," and safe automation toggles. Reuses
  zone_setpoints, schedules, automation_rules APIs. Guardian slice: starters + matchers
  (not LLM) — see phase_42_guardian_pr_spec.md.
todos:
  - id: ws1-comfort-targets-hub
    content: "WS1: Grow → Targets hub — zone list, band status, why-empty; ?zone_id= from 41"
    status: pending
  - id: ws2-band-editor
    content: "WS2: ComfortBandEditor — too low / just right / too high; stage chip; reuse setpoint PATCH"
    status: pending
  - id: ws3-schedules-plain
    content: "WS3: Schedules farmer view — humanized next run; simple time picker → cron; active toggle"
    status: pending
  - id: ws4-rules-plain
    content: "WS4: Rules farmer view — plain summary; active toggle; GH template entry; Advanced link"
    status: pending
  - id: ws5-advanced-escape
    content: "WS5: Advanced → Power settings — legacy /setpoints, /automation, cron schedules"
    status: pending
  - id: ws6-guardian-persona-impact
    content: "WS6: Persona + impact.go — comfort band / pause schedule language on PR cards"
    status: pending
  - id: ws7-docs-tests-ui
    content: "WS7: operator-tour §6e + §6 comfort; architecture §7.0h; Vitest hub + band editor"
    status: pending
  - id: ws8-guardian-starters-matchers
    content: "WS8: Guardian starters + patch_rule/schedule/program matchers — phase_42_guardian_pr_spec.md"
    status: pending
isProject: false
---

# Phase 42 — Comfort targets & automation (plain language)

## Status

**Planned — doc complete for Guardian slice.** Implement after [Phase 40](phase_40_unified_farmer_ux_zone_cockpit.plan.md) + [Phase 41](phase_41_farm_hub_coherence.plan.md).

| Doc | Purpose |
|-----|---------|
| [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) | Arc position |
| **[phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md)** | **Starters + matchers (this phase’s Guardian work)** |

**Do not start Phase 40** until roadmap + Guardian guides are committed (done). Phase 42 code comes later in the arc.

---

## Problem

Operators report **setpoints are hard to understand**:

- Scope (`zone_id` vs `crop_cycle_id` vs `stage`) reads like a schema exercise.
- Rules and schedules expose **cron**, **predicate JSON**, and **executable_action** vocabulary.
- Phase 40 fixes **inline edit in the zone** but leaves `/setpoints`, `/automation`, `/schedules` as power-user tables.
- Guardian: `patch_*` tools exist but **matchers do not** open cards from typical “pause rule / change volume” chat (see guardian PR spec).

---

## Design principles

1. **Same data, different words** — `zone_setpoints` → **comfort band**.
2. **Cron is hidden** — farmers pick time + repeat pattern; API stores cron.
3. **Rules = one sentence** — human summary + toggle; JSON in Advanced only.
4. **Guardian complements UI** — starters + matchers; **Confirm unchanged**; not Phase 46 LLM tools.

---

## Site map (target)

```text
Grow
├── My rooms (40) — inline band wedge
└── Targets & schedules (42)  ← this phase
    ├── Comfort bands (hub)
    ├── What runs when (schedules)
    └── Automation (rules)

Advanced → Power settings
    ├── Setpoints (raw table)
    ├── Automation (RuleForm)
    └── Schedules (cron editor)
```

---

## WS1 — Comfort targets hub

**Route:** `/comfort-targets` (or **Grow → Targets**).

| UI | API / data |
|----|------------|
| Per-zone row: band status (ok / missing / out of range) | `GET /farms/{id}/setpoints`, recent sensor readings |
| Filter `?zone_id=` | Phase 41 pattern |
| Why-empty `no_setpoint` | Phase 41 `EmptyStateHint` |
| **Guardian starter chips** | [phase_42_guardian_pr_spec §2](phase_42_guardian_pr_spec.md#2-conversation-starters) |

**Acceptance:** Operator sets Flower Room humidity band without visiting raw Setpoints page.

---

## WS2 — Band editor

**Component:** `ComfortBandEditor.vue` — shared with Phase 40 zone tabs.

| API field | Farmer label |
|-----------|--------------|
| min | Too low |
| ideal | Just right |
| max | Too high |
| stage | For **{growth stage}** (from active cycle) |

`POST/PATCH` existing setpoint endpoints — no migration.

---

## WS3 — Schedules (plain)

| Feature | Notes |
|---------|--------|
| List | Name, **next run** (humanized), linked program/lighting |
| Toggle active | `PATCH` schedule |
| Simple create | Daily @ time + timezone → `buildCronExpressions` server-side |
| Starters | “When does feeding run next?” — [guardian PR spec](phase_42_guardian_pr_spec.md) |
| Advanced link | Full cron editor |

---

## WS4 — Rules (plain)

| Feature | Notes |
|---------|--------|
| List | `ruleSummary(conditions, actions)` one-liner |
| Toggle active | `patch_rule` API or dedicated PATCH |
| GH templates | Link to existing template apply (36) |
| Starters | “Pause shade rule…” — [guardian PR spec](phase_42_guardian_pr_spec.md) |

---

## WS5 — Advanced escape hatch

- Banner on legacy routes: “Power-user mode.”
- Redirects from `/setpoints`, `/schedules`, `/automation` optional — or dual nav entry.

---

## WS6 — Guardian persona & impact

See [phase_42_guardian_pr_spec §4](phase_42_guardian_pr_spec.md#4-persona--impact-ws6).

---

## WS7 — Docs, tests, closure (OC-42)

| Artifact | Content |
|----------|---------|
| **operator-tour §6** | Comfort bands + what runs when (farmer walkthrough) |
| **operator-tour §6e** | Guardian starters + patch matchers on Targets pages |
| **architecture §7.0h** | Comfort vs Advanced; Guardian patch tools |
| **Vitest** | `ComfortBandEditor`, schedule humanize helper |
| **Smokes** | Optional — setpoint PATCH unchanged |
| **OC-42** | [closure plan](phase_35_37_operational_closure.plan.md) row |

---

## WS8 — Guardian starters + matchers

**Canonical spec:** [phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md)

| Item | Owner |
|------|--------|
| Conversation starters on comfort / schedules / rules | UI WS8 |
| `matchComfortAutomationIntent` | Go WS8 |
| Tests | Go WS8 |

**Explicitly not WS8:** LLM tool proposals → Phase 46.

---

## Out of scope

- Phase 46 LLM structured proposals
- New automation engine
- Removing Confirm gate
- Replacing zone cockpit (40)

---

## Definition of done

- [ ] Targets hub + band editor (farmer labels)
- [ ] Schedules + rules farmer views (no cron/JSON first)
- [ ] Advanced group contains legacy CRUD
- [ ] Guardian WS8 per [phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md)
- [ ] operator-tour §6 + §6e + architecture §7.0h + OC-42

---

## Related

| Doc | Use |
|-----|-----|
| [phase_20_6_stage_scoped_setpoints.plan.md](phase_20_6_stage_scoped_setpoints.plan.md) | Underlying model |
| [phase_41_farm_hub_coherence.plan.md](phase_41_farm_hub_coherence.plan.md) | why-empty, zone_id |
| [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) | PR basics |
