---
name: Farmer sit-in protocol (Phase 45)
overview: >
  Scripted validation with 2–3 non-technical farmers after Phases 40–44 ship.
  Records confusion, time-on-task, and Guardian PR path success (ack, setup pack, dismiss).
status: planned
parent_plan: phase_45_farmer_validation_whole_app_polish.plan.md
---

# Farmer sit-in protocol (Phase 45 WS1)

**Parent:** [phase_45_farmer_validation_whole_app_polish.plan.md](../plans/phase_45_farmer_validation_whole_app_polish.plan.md)

**Guardian PR script detail:** [phase_45_guardian_pr_spec.md](../plans/phase_45_guardian_pr_spec.md)

**Related stream:** [sit-in-operator-experience.md](sit-in-operator-experience.md) (ongoing operator pain — this protocol is a **one-time farmer validation** gate for v1)

---

## 1. Goals

| Goal | Metric |
|------|--------|
| Daily loop without trainer | Completes morning + feed paths unaided |
| Guardian PR trust | Completes **ack**, **setup pack**, **dismiss** without fear of silent writes |
| Wizard over chat for setup | Fresh profile uses wizards first; chat is optional |
| Friction captured | P0/P1 backlog from verbatim notes |

---

## 2. Participants

| Role | Count | Profile |
|------|-------|---------|
| **Tester** | 2–3 | Non-technical grower; minimal SQL/API exposure |
| **Facilitator** | 1 | Notes only; does not drive mouse unless blocked >3 min |
| **Observer** | 0–1 | Optional dev/product |

**Environment:** Demo seed farm **or** fresh profile on staging; LAN Pi optional for device wizard session.

---

## 3. Sessions (90 min each)

### Session A — Returning operator (seeded farm)

| Block | Minutes | Script |
|-------|---------|--------|
| Morning | 25 | Open **Dashboard** → read morning summary → open one **alert** → complete **Guardian ack PR** ([§4.1](#41-path-1--ack_alert)) |
| Grow room | 25 | **Zones** → zone cockpit → read Water story → **Run now** or pulse (if device online) |
| Comfort | 15 | Adjust one **comfort band** (42 UI) or ask Guardian starter |
| Stock | 15 | **Supplies** hub → find low stock → optional refill task PR |
| Debrief | 10 | Verbatim quotes + P0/P1 tags |

### Session B — Fresh setup (blank or new farm profile)

| Block | Minutes | Script |
|-------|---------|--------|
| Wizards first | 35 | **Farm setup wizard** → **Add zone** → **Device wizard** (copy API key; skip live Pi if offline) |
| First-run checklist | 15 | Dashboard checklist — mark steps; prefer buttons over chat |
| Grow via Guardian | 25 | Empty zone → use starter → **setup pack PR** → Confirm → verify Plants + Fertigation ([§4.2](#42-path-2--apply_grow_setup_pack)) |
| Dismiss drill | 10 | Open any pending PR → **Dismiss** without Confirm ([§4.3](#43-path-3--dismiss-no-db-write)) |
| Debrief | 5 | |

### Session C — Mobile WebView (optional)

| Block | Minutes | Script |
|-------|---------|--------|
| Morning on phone | 20 | Dashboard + alert ack |
| Guardian drawer | 20 | Send message; Confirm/Dismiss tap targets (a11y WS6) |
| Debrief | 10 | |

---

## 4. Guardian PR paths (required)

Facilitator marks **pass / fail / skip** per path.

### 4.1 Path 1 — `ack_alert`

| Step | Operator does | Pass criteria |
|------|---------------|---------------|
| 1 | Open Guardian from alert row or **Ask Guardian** with contextual prefill (40) | Drawer opens |
| 2 | Send: *“Acknowledge alert #N”* or use starter | **Proposal card** appears (`ack_alert`) |
| 3 | Read **impact** line | Understands alert will be acknowledged |
| 4 | Tap **Confirm** | Alert acknowledged in Alerts UI |
| 5 | Check **Pending** inbox | Proposal status executed |

**Fail examples:** No card (matcher miss → log phrase for 46 backlog); Confirm without reading impact; operator thinks DB changed before Confirm.

### 4.2 Path 2 — `apply_grow_setup_pack`

| Step | Operator does | Pass criteria |
|------|---------------|---------------|
| 1 | Pick **empty zone** (no active cycle) | Zone name visible |
| 2 | Send grow-setup phrase or chip: *“Add my philodendron to {zone} with a light fertigation program”* | **Setup pack card** (`SetupPackProposalCard`) |
| 3 | Review numbered bundle (plant, zone, stage, program EC/volume) | Can explain one line in own words |
| 4 | Tap **Confirm** | Plant + cycle + program appear (high-tier warning seen) |
| 5 | Optional refine | *“use 0.3 L not 0.5”* → revised draft (Phase 34) before Confirm |

**Fail examples:** No card when zone already has cycle; operator expects instant plant without Confirm; facilitator had to type message.

### 4.3 Path 3 — Dismiss (no DB write)

| Step | Operator does | Pass criteria |
|------|---------------|---------------|
| 1 | Trigger any pending PR (ack or setup pack) **or** use demo pending row | Card visible |
| 2 | Tap **Dismiss** | Card shows **Dismissed**; **no** API confirm call |
| 3 | Verify farm data | No change from dismissed card (plants/alerts unchanged) |
| 4 | Explain to facilitator | “Nothing happened to my farm — I cancelled the suggestion” |

**Implementation note:** Dismiss is **client-side** ([`GuardianActionProposal.vue`](../../ui/src/components/GuardianActionProposal.vue)) — proposal may remain `pending` server-side until TTL; sit-in teaches **operator truth** (no write), not inbox hygiene.

**Fail examples:** Operator believes Dismiss = Confirm; afraid to dismiss high-tier card; cannot find Dismiss button on mobile.

---

## 5. What we record

| Field | Example |
|-------|---------|
| `session_id` | A1, B2, C1 |
| `task` | ack_alert / setup_pack / dismiss |
| `result` | pass \| fail \| skip |
| `blocker` | P0 \| P1 \| P2 |
| `quote` | Verbatim |
| `time_sec` | Optional |
| `route` | `/zones/3`, Dashboard, … |
| `matcher_gap` | Phrase that should have proposed but did not → Phase 46 backlog |

Log in spreadsheet or GitHub issues with label `sit-in-45`.

---

## 6. Triage rules (WS2)

| Priority | Definition | Example fix |
|----------|------------|-------------|
| **P0** | Cannot finish daily loop | Run now hidden; auth broken |
| **P1** | Finishes wrong page | Lands on `/setpoints` not Targets |
| **P2** | Copy/layout | Button label, contrast |

Prefer UI composition fixes; schema only if sit-in proves data model gap.

**Matcher gaps:** File under `sit-in-46-backlog` or [phase_46](phase_46_guardian_llm_tool_proposals.plan.md) — do not block 45 ship on 46.

---

## 7. Success criteria (Phase 45 closure)

- [ ] ≥2 sessions A + ≥1 session B completed
- [ ] All three PR paths **pass** for ≥2 testers (or documented skip with fix)
- [ ] P0 backlog empty
- [ ] P1 backlog triaged (fix or defer with reason)
- [ ] Findings linked from [phase_45_guardian_pr_spec.md](../plans/phase_45_guardian_pr_spec.md)

---

## Related

| Doc | Use |
|-----|-----|
| [operator-tour.md](../operator-tour.md) §6, §9 | Operator + validation tour |
| [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) | PR basics |
| [farmer_ux_roadmap_40_plus.plan.md](../plans/farmer_ux_roadmap_40_plus.plan.md) | Arc context |
