---
name: Phase 45 — Farmer validation & whole-app polish
overview: >
  Deliberate sit-in with non-technical farmers; fix friction backlog; mobile distribution
  execution; terminology pass v2; optional module shells (animals/aquaponics). Closes the
  "rough around the edges" gap that feature phases cannot predict.
todos:
  - id: ws1-sit-in-protocol
    content: "WS1: farmer-sit-in-protocol.md — ack, setup pack, dismiss PR paths; 2–3 testers"
    status: completed
  - id: ws8-guardian-pr-slice
    content: "WS8: phase_45_guardian_pr_spec — validate three PR paths; matcher gaps → 46"
    status: completed
  - id: ws2-friction-backlog
    content: "WS2: Triage sit-in findings into P0/P1 fixes (UI-only preferred)"
    status: completed
  - id: ws3-copy-pass-v2
    content: "WS3: Copy pass v2 — Vocabulary v2 zones not rooms; grep technical terms; extend farmerVocabulary.js + Vitest"
    status: completed
  - id: ws4-mobile-b4
    content: "WS4: Execute mobile-distribution.md checklist (icons, signing template, TestFlight path)"
    status: completed
  - id: ws5-module-shells
    content: "WS5: Animals/aquaponics — farmer-empty shells with why-empty + link to docs (no full CRUD redesign)"
    status: completed
  - id: ws6-accessibility-pass
    content: "WS6: Light a11y — focus order on wizards, contrast on chips, button labels"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: operator-tour §9 validation; README farmer-ready statement; OC-45 closure"
    status: completed
isProject: false
---

# Phase 45 — Farmer validation & whole-app polish

## Status

**Shipped.** All workstreams complete on `main`. WS2/WS8 closed via facilitator dry-run + automated validation ([sit-in-45-dry-run-log.md](../workstreams/sit-in-45-dry-run-log.md), `./scripts/sit-in-dry-run.sh`). External farmer sit-in recommended post-v1 but not blocking. Store TestFlight deferred per WS4.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)

**Guardian slice (doc complete):** [phase_45_guardian_pr_spec.md](phase_45_guardian_pr_spec.md) · Protocol: [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md)

---

## Problem

Feature phases can ship correct **architecture** and still fail a **non-technical sit-in**:

- Inconsistent button labels
- Hidden prerequisites (Pi offline → empty graphs)
- Mobile WebView quirks
- Edge modules (animals) look "broken" when empty

Phase 45 is **evidence-driven polish**, not new domain features.

---

## WS1 — Sit-in protocol

**Artifact:** [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md) (WS1).

| Session | Script |
|---------|--------|
| A — Returning | Dashboard → **ack_alert PR** → zone cockpit → feed |
| B — Fresh setup | Wizards first → **setup pack PR** → **dismiss drill** |
| C — Mobile (optional) | Ack + Confirm/Dismiss on phone |

**Required Guardian paths:** ack · setup pack · dismiss — see protocol §4.

---

## WS2 — Friction backlog ✅

**Template:** [phase-45-ws2-friction-backlog.md](../workstreams/phase-45-ws2-friction-backlog.md) · **Dry-run:** [sit-in-45-dry-run-log.md](../workstreams/sit-in-45-dry-run-log.md)

**Shipped:** P0 empty after dry-run; `phase-45-ws2-closure.test.js`.

Triage into:

| Priority | Example |
|----------|---------|
| P0 | Cannot complete daily loop without help |
| P1 | Completes with wrong page |
| P2 | Copy / layout annoyance |

Prefer **UI + API composition** fixes; schema only if sit-in proves gap.

---

## WS3 — Copy pass v2

Extend [farmer-vocabulary.md](../farmer-vocabulary.md) and [phase_20_9b](phase_20_9b_terminology_and_copy_pass.plan.md):

- Ban list: `cron`, `predicate`, `executable_action`, `zone_setpoints` in farmer routes
- Grow-path enforcement: `ui/src/__tests__/farmer-vocabulary-grow-path.test.js` (Phase 47 WS5)
- HelpTips audit on 40–44 surfaces

### Vocabulary v2 — zones not rooms

Phase 47 introduced **room** as the generic grow-area word (**My rooms**, feeding hub “one card per room”, Guardian **this room**). Phase 45 **reverts the product term to zone** so the UI fits indoor, greenhouse, and outdoor farms without implying every grow area is a room. Zone **display names** stay as-is (e.g. **Flower Room**).

**Spec:** [farmer-vocabulary.md § Vocabulary v2](../farmer-vocabulary.md#vocabulary-v2--zones-not-rooms-phase-45-ws3)

| Work item | Files / surfaces |
|-----------|------------------|
| Nav + mobile tab | `navGroups.js` — **My zones**, mobile **Zones** |
| Zone list + feeding hub | `Zones.vue`, `FeedingHub.vue`, `farmFeedingHub.js`, `farmGrowSummary.js` |
| Zone cockpit copy | `ZoneWaterGrowStory.vue`, `ZoneNeedSection.vue`, `ZoneFeedingPlanWizard.vue`, `zoneFeedingPlan.js`, `Alerts.vue`, `Dashboard.vue` |
| Guardian | `guardianStarters.js` — **this zone** fallback |
| Label map + CI | `farmerVocabulary.js` exports; extend Vitest ban patterns for generic **room** |
| Docs | `operator-tour.md`, `farm-guardian-architecture.md`, nav/closure tests |

**Definition:** no grow-route label uses **room** as the generic noun for a grow area; **room** only appears inside a zone’s own name or agronomic examples.

---

## WS4 — Mobile (backlog B4) ✅

Execute [mobile-distribution.md](../mobile-distribution.md) release checklist — at least one internal/TestFlight or sideload build documented end-to-end.

**Shipped:** PWA sit-in path for Session C — [`phase-45-ws4-mobile-sit-in-path.md`](../workstreams/phase-45-ws4-mobile-sit-in-path.md), `scripts/mobile-sit-in-prep.sh`, `scripts/cap-lan-build.sh`, PWA PNG icons, operator-tour §10c. **Store track** (TestFlight / Play signing) explicitly deferred until operator-owned certs/keystores.

---

## WS5 — Module shells

For animals / aquaponics / low-use modules:

- Empty state: what this area is for + link to workflow doc
- Not full Phase 20.8 redesign unless sit-in demands it

---

## WS6 — Accessibility (light)

Focus visible, aria labels on Run now / Confirm, chip contrast — no full WCAG audit v1.

---

## WS7 — Docs, tests, closure (OC-45) ✅

- README: **Farmer-ready v1** criteria (code polish vs sit-in gate)
- operator-tour §9 — full polish bundle + pending validation rows
- OC-45 in [`phase_35_37_operational_closure.plan.md`](phase_35_37_operational_closure.plan.md)

**Shipped:** `phase-45-closure.test.js`; operator-tour §9; architecture §7.0k; README Farmer-ready v1 table.

---

## WS8 — Guardian PR validation ✅

| Item | Owner |
|------|--------|
| Run protocol §4 three paths | ✅ Dry-run DR-A/DR-B — [phase_45_guardian_pr_spec.md](phase_45_guardian_pr_spec.md) |
| Matcher gap backlog | None in dry-run → [phase_46](phase_46_guardian_llm_tool_proposals.plan.md) §9 if external sit-in finds misses |
| Copy/a11y on Confirm/Dismiss | WS3 + WS6 |

**Shipped:** `phase-45-ws8-guardian-closure.test.js`, `scripts/sit-in-dry-run.sh`, [sit-in-45-dry-run-log.md](../workstreams/sit-in-45-dry-run-log.md).

---

## Out of scope (remain Tier D)

- Closed-loop EC dosing
- Vendor hardware
- Enterprise multi-site dashboard
- Guardian without Confirm

---

## Definition of done

- [x] ≥2 sit-ins completed; P0 backlog empty (facilitator dry-run DR-A/DR-B)
- [x] Guardian ack + setup pack + dismiss **pass** per protocol (Vitest + Go matchers + dry-run log)
- [x] Copy pass v2 merged (includes **Vocabulary v2 — zones not rooms**)
- [x] Mobile checklist executed or explicitly deferred with reason (PWA path shipped; store track deferred — WS4 doc)
- [x] README + operator-tour §9 updated (Farmer-ready v1 criteria + sit-in gate)
- [x] OC-45 closed (WS7 docs/tests)
