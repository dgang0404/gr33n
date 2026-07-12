---
name: Phase 170 — Today Guardian one-tap farm counsel
overview: >
  Today starters (morning check, attention chips, zone quick actions) now open
  the Guardian drawer in Farm counsel mode and auto-send — matching the
  in-panel morning walkthrough UX. Setup starters still prefill only so
  operators can edit before sending.
todos:
  - id: ws1-starter-entry-lib
    content: "WS1: guardianStarterEntry.js — farm counsel + auto-send rules per starter"
    status: completed
  - id: ws2-panel-store
    content: "WS2: guardianPanel preferFarmCounsel + autoSendOnOpen on openDrawer"
    status: completed
  - id: ws3-drawer-autosend
    content: "WS3: GuardianChatPanel sendCounselStarter on drawer open; GuardianStarterChips + ZoneQuickActions wiring"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: Tests, phase-170-closure, current-state note"
    status: completed
isProject: false
---

# Phase 170 — Today Guardian one-tap farm counsel

**Status:** shipped · **Depends on:** [169](phase_169_today_attention_cockpit.plan.md)

## Problem

Dashboard Guardian chips (morning check, attention starters) only prefilled
the drawer — growers had to notice Farm counsel mode and tap Send. In-panel
morning starters already one-tapped; Today did not.

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `starterPrefersFarmCounsel` / `starterShouldAutoSend` in `guardianStarterEntry.js` |
| **WS2** | `guardianPanel.openDrawer({ farmCounsel, autoSend })` |
| **WS3** | `GuardianChatPanel.sendCounselStarter` on auto-send open; chips + quick sheet pass flags |
| **WS4** | `guardian-starter-entry.test.js`, `phase-170-closure.test.js` |

## Rules

- **Auto-send:** `morning_walkthrough`, `farm_counsel`, zone context refs
- **Farm counsel, no auto-send:** `setupMode` starters (edit first)
- **Quick chat:** weather / generic starters (unchanged)

## Verification

```bash
cd ui && npm test -- --run src/__tests__/guardian-starter-entry.test.js src/__tests__/phase-170-closure.test.js src/__tests__/guardian-panel.test.js
```
