---
name: Phase 73 — Guardian change-request discoverability + read-tool reliability
overview: >
  Make Guardian "pull requests" (change requests) easy to find and act on, and
  make Guardian's read tools fire reliably so it stops saying "I don't have that"
  when it actually does. Today proposals only appear when a server-side matcher
  fires, the Pending inbox is buried in a drawer, dismiss is client-only, and read
  tools (site_weather, lookup_crop_targets) are gated behind regex intents + need
  site coords / a crop profile. This phase adds a global pending badge, an
  empty-zone proposal nudge, a server-side dismiss, and grounding reliability:
  prompt farms for coordinates, prompt grows for a crop profile, and broaden /
  optionally LLM-route tool selection. Backend (Go) + UI.
todos:
  - id: ws1-global-pending
    content: "WS1: Global pending-change badge on the Guardian launcher (nav + TopBar) from guardianProposals store; not buried in the drawer Pending tab"
    status: pending
  - id: ws2-empty-zone-nudge
    content: "WS2: Empty-zone/empty-farm proactive proposal — when a zone has no active grow, Guardian offers an apply_grow_setup_pack PR the user can Confirm (uses existing starters + setup-pack matcher)"
    status: pending
  - id: ws3-server-dismiss
    content: "WS3 (backend): real POST /v1/chat/proposals/{id}/dismiss so dismiss persists (today it's UI-only; DB row lingers pending until TTL)"
    status: pending
  - id: ws4-readtool-reliability
    content: "WS4 (backend): widen read-tool intent matching + optional LLM tool-selection so site_weather/lookup_crop_targets fire on natural phrasing; clear 'why no data' messaging (missing coords / no crop profile)"
    status: pending
  - id: ws5-grounding-prereqs
    content: "WS5: Setup nudges for the data the tools need — prompt to set farm lat/long (site_weather) and assign a crop profile to a grow (lookup_crop_targets); surfaced in Settings + zone + empty states"
    status: pending
  - id: ws6-docs-tests
    content: "WS6: guardian-change-requests-guide update, Go matcher/dismiss tests, smoke_phase73_test.go, phase-73-closure.test.js, operator-tour; OC-73"
    status: pending
isProject: false
---

# Phase 73 — Guardian change-request discoverability + read-tool reliability

## Status

**Planned.** Last phase of the [SPA workspace arc](phase_68_73_spa_workspace_roadmap.plan.md). Backend (Go) + UI. Builds on the Guardian proposal system ([Phase 29](phase_29_guardian_agent_layer.md)/[30](phase_30_guardian_change_requests.plan.md)/[34](phase_34_guardian_pr_iteration.plan.md)/[45](phase_45_guardian_pr_spec.md)/[46](phase_46_guardian_llm_tool_proposals.plan.md)/[55](phase_55_guardian_pr_spec.md)) and the read tools from [Phase 55/64/66](phase_66_weather_site_context.plan.md).

**Closure:** **OC-73** — tracked in this plan's DoD + [arc hub OC table](phase_68_73_spa_workspace_roadmap.plan.md#operational-closure-oc-rows). Do not add to the archived Phase 35 closure doc.

---

## The one job

> **Guardian notices something (no plants in a zone, low light today), offers a change you can accept in one tap, and you can always find your pending changes — and it stops saying "I don't have that" when it does.**

---

## Problem

Two operator-reported gaps, both confirmed in the code.

### A. "There's no clear way to get pull requests going in the UI."

"Pull requests" = **Guardian change requests** = rows in `gr33ncore.guardian_action_proposals` ([migration](../db/migrations/20260521_phase29_guardian_proposals.sql)), confirmed via `POST /v1/chat/confirm`. The UI exists ([`GuardianActionProposal.vue`](../ui/src/components/GuardianActionProposal.vue) cards, a **Pending** tab in [`GuardianDrawer.vue`](../ui/src/components/GuardianDrawer.vue), `/chat?tab=pending`), but it's hard to find because:

1. **A card only appears when a server-side matcher fires** — generic questions get text only. Triggers live in [`internal/farmguardian/proposals*.go`](../internal/farmguardian/proposals.go), **not** in the persona. (Answering the operator's question: PRs are **not** made by phrases in the persona — they're created after each chat turn by Go intent-matchers, plus an optional LLM path gated off by `GUARDIAN_LLM_PROPOSALS`.)
2. **The Pending inbox is nested in the drawer** — there's a nudge dot but no clear global count of "you have N changes waiting."
3. **Dismiss is client-only** — there's no dismiss API; the DB row lingers `pending` until its ~5-min TTL.

### B. Guardian denies access to data it actually has.

The operator's convo: *"Do I need supplemental light today?"* → Guardian said it had no weather or crop-target access. But `site_weather` ([`readtools_weather.go`](../internal/farmguardian/readtools_weather.go)) and `lookup_crop_targets` ([`readtools_crop.go`](../internal/farmguardian/readtools_crop.go)) **are implemented**. They didn't fire because:

- Read tools are **regex-intent-gated** — they only inject data when the message matches; "do I need supplemental light today" may miss the trigger set.
- `site_weather` needs **farm site coordinates** (`GetFarmSiteCoords`); without lat/long it can't compute clear-sky DLI.
- `lookup_crop_targets` returns "no crop profile assigned" when the grow has no profile linked.
- Tools enrich the **system prompt** silently; if nothing is injected, the persona's honesty rule makes the model say "I don't have that."

So the fix is **discoverability + grounding reliability**, not new infrastructure.

---

## Design principles

1. **Proposals are first-class.** A global, always-visible pending count and a one-tap path to act — not a buried tab.
2. **Proactive where it's obviously helpful.** Empty zone / empty farm → offer the setup PR (the operator's exact scenario: "Guardian sees there are no plants and helps set up a nice pull request to accept").
3. **Grounding over guessing (unchanged rule).** Tools must fire when relevant; when data is genuinely missing, say *why* and offer to fix it (set coordinates / assign a crop profile) — never invent.
4. **Confirm-gated writes stay.** No silent changes; this phase only makes the propose→Confirm loop easier to reach.
5. **Reuse the registry.** Empty-zone nudge uses the existing `apply_grow_setup_pack` tool ([`tools/registry.go`](../internal/farmguardian/tools/registry.go)) and starters ([`guardianStarters.js`](../ui/src/lib/guardianStarters.js)).

---

## WS1 — Global pending-change badge

- Surface the pending-proposal count from [`ui/src/stores/guardianProposals.js`](../ui/src/stores/guardianProposals.js) (`GET /v1/chat/proposals?status=pending`) on the **Guardian launcher** ([`GuardianNavLaunch.vue`](../ui/src/components/GuardianNavLaunch.vue)) **and** the TopBar "Ask gr33n" button — a real count badge, not just a nudge dot.
- Clicking opens the drawer on the **Pending** tab ([`GuardianRequestsInbox.vue`](../ui/src/components/GuardianRequestsInbox.vue)).
- Optional: a compact "N changes waiting" strip on Today/Dashboard.

---

## WS2 — Empty-zone / empty-farm proposal nudge

- When a zone has **no active crop cycle** (the existing `zoneHasActiveCycle()` signal used in [`proposals_setup_pack.go`](../internal/farmguardian/proposals_setup_pack.go) and `guardianStarters.js`), surface a proactive Guardian card: *"Flower Room has no grow yet — want me to set one up?"* → produces an `apply_grow_setup_pack` proposal to **Confirm**.
- Wire it into the zone empty-state (Phase 69 Overview) and the farm empty-state, using the existing starter prefills so the matcher reliably fires.
- This is the operator's described flow: Guardian notices no plants → offers a PR → user accepts.

---

## WS3 — Server-side dismiss (backend)

- Add `POST /v1/chat/proposals/{id}/dismiss` ([handler](../internal/handler/chat/proposals.go)) setting `status = dismissed` (queries in [`db/queries/guardian_proposals.sql`](../db/queries/guardian_proposals.sql)).
- UI dismiss calls it so the row doesn't linger `pending` until TTL; the pending badge (WS1) stays accurate.
- Keep the existing TTL/expiry behavior for untouched proposals.

---

## WS4 — Read-tool reliability (backend)

- **Widen intent matching** for `site_weather` and `lookup_crop_targets` (and peers) so natural phrasings fire — e.g. "do I need supplemental light today," "is it bright enough," "what EC should I run." Centralize/expand the trigger sets in [`readtools_weather.go`](../internal/farmguardian/readtools_weather.go) / [`readtools_crop.go`](../internal/farmguardian/readtools_crop.go) / [`readtools.go`](../internal/farmguardian/readtools.go).
- **Optional LLM tool-selection:** when regex misses, let the model pick a read tool (gated like `GUARDIAN_LLM_PROPOSALS`, e.g. `GUARDIAN_LLM_READTOOLS`) so phrasing isn't a hard wall. Default-on can be decided after validation.
- **Honest "why no data":** when a tool *would* fire but its prerequisite is missing, inject a clear block — *"Weather needs this farm's location"* / *"This grow has no crop profile"* — so the model says the actionable thing, not a flat "no access," and can deep-link to the fix (WS5).

---

## WS5 — Grounding prerequisites (setup nudges)

Make the data the tools need easy to provide:

- **Farm coordinates** for `site_weather`: a lat/long field nudge in [`Settings.vue`](../ui/src/views/Settings.vue) (farm section) and a one-time prompt when weather is asked without coords. (Schema + `weather_data` already exist per [Phase 66](phase_66_weather_site_context.plan.md).)
- **Crop profile** for `lookup_crop_targets`: prompt to assign a crop profile to a grow from the zone/grow empty state and Plants ([`CropProfileDetail.vue`](../ui/src/views/CropProfileDetail.vue), [Phase 64 knowledge base](phase_64_crop_knowledge_base.plan.md)).
- These nudges can themselves be Guardian proposals where a write tool exists, or plain inline CTAs otherwise.

---

## WS6 — Docs, tests, closure (OC-73)

| Artifact | Content |
|----------|---------|
| [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) | Update: global badge, empty-zone nudge, server dismiss, "why no data" |
| Go tests | Widened matchers fire on new phrasings; dismiss endpoint sets status; "missing prereq" block emitted |
| `cmd/api/smoke_phase73_test.go` (new) | Propose → confirm → list; dismiss persists; weather tool fires + reports missing coords |
| `ui/src/__tests__/phase-73-closure.test.js` (new) | Pending badge count; empty-zone proposal card; dismiss calls API |
| [operator-tour.md](../operator-tour.md) | "Find your pending changes; Guardian sets up empty zones" |

**OC-73** added and closed when WS1–WS6 ship.

---

## Out of scope

- New write tools beyond what's registered (restock/receipt stay inline per [Phase 55](phase_55_guardian_pr_spec.md)).
- Replacing the proposal architecture with full OpenAI function-calling (WS4 is additive LLM routing, not a rewrite).
- Voice/vision (shipped in [Phase 67](phase_67_guardian_field_assistant.plan.md)).
- New weather providers (Phase 66 owns weather sourcing).

---

## Definition of done

- [ ] A global pending-change count is visible on the Guardian launcher + TopBar; one tap reaches the inbox
- [ ] Empty zones/farms surface a Confirm-able grow-setup proposal (Guardian "sees no plants → offers a PR")
- [ ] `POST /v1/chat/proposals/{id}/dismiss` persists; badge stays accurate
- [ ] `site_weather` / `lookup_crop_targets` fire on natural phrasing; the supplemental-light question now gets a grounded answer
- [ ] Missing coords / crop profile produce a clear "set this up" message + CTA, never a flat "no access"
- [ ] Go + Vitest green; smoke_phase73 green; OC-73 closed

---

## Suggested implementation order

1. WS1 global badge (immediate discoverability)
2. WS3 server dismiss (badge accuracy)
3. WS4 read-tool widening + "why no data" (fixes the reported convo)
4. WS5 grounding prereq nudges (coords + crop profile)
5. WS2 empty-zone proposal nudge (proactive PR)
6. WS6 closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_55_guardian_pr_spec.md](phase_55_guardian_pr_spec.md) | Change-request spec + write boundary |
| [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md) | LLM proposal path (flag) WS4 extends |
| [phase_66_weather_site_context.plan.md](phase_66_weather_site_context.plan.md) | site_weather + coordinates |
| [phase_64_crop_knowledge_base.plan.md](phase_64_crop_knowledge_base.plan.md) | crop targets grounding |
| [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) | Operator-facing PR guide |
| [internal/farmguardian/proposals.go](../internal/farmguardian/proposals.go) | Matcher chain |

---

## Using this in a new chat

> Read `docs/plans/phase_73_guardian_pr_discoverability.plan.md`. Make Guardian change requests discoverable (global pending badge, empty-zone proposal nudge, server-side dismiss) and make read tools fire reliably (widen intent matching + optional LLM routing; clear "missing coords / no crop profile" messaging with setup CTAs). Backend in Go + UI. Add smoke_phase73_test.go. PRs are created by server-side matchers, not persona phrases — keep that model.
