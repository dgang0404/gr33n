---
name: Phase 44 — Getting started & edge install wizard
overview: >
  In-app paths for "I have a new farm" and "connect my Pi" — surfacing Phase 15 bootstrap
  templates, device pairing, and Guardian-guided setup without requiring cron literacy or
  shell docs. Confirm-gated writes remain; wizards collect intent, Guardian/API execute.
todos:
  - id: ws1-farm-setup-wizard
    content: "WS1: Farm setup wizard — blank vs template cards; preview what gets created; POST apply-template"
    status: completed
  - id: ws2-zone-starter
    content: "WS2: Add zone wizard — name, type (greenhouse/indoor), optional bootstrap slice"
    status: completed
  - id: ws3-device-wizard
    content: "WS3: Edge device wizard — API key, test connection, assign zone; link pi-integration-guide steps in UI"
    status: completed
  - id: ws4-guardian-setup-mode
    content: "WS4: Guardian 'setup mode' prompts — grow_setup_pack, create_lighting_program, create_fertigation_program with checklists"
    status: completed
  - id: ws5-first-run-dashboard
    content: "WS5: First-run empty Dashboard — checklist (zones, device, comfort band, one schedule)"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: operator-tour §8 setup; architecture §7.0j; Vitest wizard flows; OC-44"
    status: completed
  - id: ws8-guardian-pr-slice
    content: "WS8: phase_44_guardian_pr_spec — wizards first; setup starters + setup-mode second"
    status: pending
isProject: false
---

# Phase 44 — Getting started & edge install wizard

## Status

**WS1–WS6 shipped** on `main`. WS8 (Guardian PR slice closure) pending. After [Phase 41](phase_41_farm_hub_coherence.plan.md) (empty states) and [Phase 42](phase_42_comfort_targets_automation_plain_language.plan.md) (comfort bands — wizard can set first band).

**Prerequisite API:** [Phase 15](phase_15_farm_onboarding.plan.md) bootstrap templates ✅.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

**Guardian slice (doc complete):** [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md) — wizards primary; starters + setup-mode secondary.

---

## Problem

| Job | Today |
|-----|--------|
| New farm | Settings/template exists but easy to miss; docs-heavy |
| New zone | Form fields, no guided path |
| Pi online | [`pi-integration-guide.md`](../pi-integration-guide.md) — not in-app |
| First grow | Guardian + bootstrap — assumes chat comfort |

---

## Design principles

1. **Wizards call existing APIs** — `apply_bootstrap_template`, device POST, zone POST.
2. **Guardian complements, does not replace** — wizards for linear steps; Guardian for questions.
3. **No silent writes** — template apply may still need admin RBAC; device keys shown once.
4. **Offline-aware copy** — point to Phase 37 field guides for Pi procedures.

---

## WS1 — Farm setup wizard

**Entry:** New farm created → wizard modal or `/farms/{id}/setup`.

| Step | Action |
|------|--------|
| Choose | Blank · Indoor veg template · Greenhouse climate template (cards with bullet preview) |
| Confirm | List zones/schedules/programs to be created |
| Apply | `POST /farms/{id}/bootstrap-template` |

Reuse Phase 15 UI patterns; improve copy and illustration.

---

## WS2 — Add zone wizard

| Step | Fields |
|------|--------|
| Basics | Name, zone type |
| Needs | Greenhouse profile if type=greenhouse (36) |
| Optional | Link device, pick starter lighting preset |

`POST /farms/{id}/zones` + optional lighting from-preset.

---

## WS3 — Edge device wizard

| Step | Content |
|------|---------|
| Register device | Name, UID, farm/zone |
| API key | Generate/show once; copy button |
| Test | Poll device status or `GET /devices` online |
| Actuators | Auto-discover or assign pump/light |

Embed checklist from pi-integration-guide (not PDF-only).

---

## WS4 — Guardian setup mode ✅

**Spec:** [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md) §5.

- Chat system hint when farm has zero zones or `?setup=1`.
- Starters send grow-setup phrases; **`apply_grow_setup_pack`** matcher unchanged (Phase 32).
- **Bootstrap** via wizard POST — not chat-first (`apply_bootstrap_template` stays admin PR).

**Shipped:** `internal/farmguardian/setup_mode.go` + `POST /v1/chat` `setup_mode` / `?setup=1`; `buildSetupStarters` on drawer, wizards, and `/chat?setup=1`.

---

## WS5 — First-run dashboard ✅

When farm has no zones / no devices / no setpoints:

```
Getting started
☐ Add a grow room
☐ Connect edge device
☐ Set comfort targets
☐ Turn on one schedule
```

Links to wizards; dismiss when complete.

**Shipped:** `GettingStartedChecklist` on Dashboard; `firstRunChecklist.js`; Guardian `first_run_dashboard` starters; auto-hides when all four steps complete or operator hides for now.

---

## WS6 — Docs, tests, closure (OC-44) ✅

operator-tour §8 + §6g, architecture §7.0j, Vitest wizard navigation, smoke bootstrap apply from UI path.

**Shipped:** `phase-44-closure.test.js`, `phase-44-wizard-navigation.test.js`, `TestPhase44WizardBootstrapApply`, operator-tour §8/§6g + architecture §7.0j updated from stubs.

---

## WS8 — Guardian PR slice

| Item | Owner |
|------|--------|
| Starters on checklist, wizards, empty zone | UI — spec §4 |
| Setup-mode persona | Handler — spec §5 |
| Grow-setup PR via existing matcher | Backend — no new tool |
| Bootstrap **not** via starter chips | UX rule — spec §3 |

---

## Out of scope

- Guardian as **only** UI (no one-tap without Confirm)
- Replacing `bootstrap-local.sh` for developers
- OTA firmware management

---

## Definition of done

- [x] Farm setup wizard (WS1) — `/farms/:id/setup`, template cards, preview, apply
- [x] Add zone wizard (WS2) — `/farms/:id/zones/new`, greenhouse profile, optional lighting preset
- [x] Edge device wizard (WS3) — `/farms/:id/devices/new`, Pi checklist, poll online, actuators
- [x] First-run checklist on Dashboard (WS5)
- [x] Docs, tests, OC-44 closure (WS6)
- [x] Pi steps reachable without leaving app (device wizard checklist)
- [ ] Guardian WS8 per [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md)
- [x] operator-tour §8 + §6g + architecture §7.0j

## Related

| Doc | Use |
|-----|-----|
| [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md) | Wizards vs starters |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Setup pack |
