---
name: Guardian PR UX through farmer phases (40–45)
overview: >
  Product spec for how Guardian change requests (proposals) are triggered, suggested,
  and reviewed as the farmer UX arc ships. Documents current rule-assisted behavior,
  industry patterns, conversation starters, and per-phase deliverables. Prerequisite
  reading before Phase 40 implementation.
todos:
  - id: doc-operator-guide
    content: "Publish docs/guardian-change-requests-guide.md (operator + dev)"
    status: done
  - id: doc-this-plan
    content: "This plan — triggers, starters, inbox, LLM-tool gap, phase map"
    status: done
  - id: p40-zone-starters
    content: "Phase 40 — zone-context conversation starters + inline actions vs PR boundaries"
    status: pending
  - id: p41-morning-starters
    content: "Phase 41 — Dashboard morning starters; inbox badge tied to alerts/tasks"
    status: pending
  - id: p42-comfort-pr
    content: "Phase 42 — starters for comfort band + rule toggle; patch_rule summaries"
    status: pending
  - id: p44-setup-starters
    content: "Phase 44 — setup mode starters + grow_setup_pack from first-run checklist"
    status: pending
  - id: p45-sit-in-pr
    content: "Phase 45 — sit-in tests PR flows with non-technical farmers"
    status: pending
  - id: p40-contextual-ask-guardian
    content: "Phase 40 — replace generic Ask Guardian prefills (zone status) with snapshot-aware prompts + contextRef"
    status: pending
  - id: future-llm-tool-routing
    content: "Phase 46 — LLM structured tool proposals when matchers miss (NOT starter chips — see phase_46 plan)"
    status: pending
isProject: false
---

# Guardian PR UX through farmer phases (40–45)

## Status

**Documentation gate for Phase 40.** Code for starters may land in 40–44; this plan defines **behavior** so Guardian and zone cockpit stay one product.

**Operator guide:** [`../guardian-change-requests-guide.md`](../guardian-change-requests-guide.md)

---

## Mental model (frozen)

```text
User intent
    │
    ├─► Direct UI action (40 inline setpoint, run-now, form save) ──► API ──► DB
    │
    └─► Chat message or starter chip (sends same as typed message)
            │
            ├─► LLM answer (always, if AI on)
            └─► Rule-assisted matcher (optional) ──► proposal row ──► card + inbox
                        │
                        └─► User Confirm ──► tools.Execute ──► DB / pending_command
```

**v1 farmer arc does not require LLM-native tool calling for PRs.** Matchers + starters + inline UI cover most jobs. LLM-tool routing is **future** (see §8).

---

## How proposals work today (accurate as of Phase 34)

| Mechanism | Implementation |
|-----------|----------------|
| Insert proposal | `farmguardian.BuildRuleAssistedProposals` after chat turn |
| Matchers | Alert ack/read, setup pack, config tools (`proposals_config.go`), revise (`proposals_revise.go`) |
| LLM proposes PR | **No** — LLM text is separate; card only if matcher hits |
| Frozen args | JSON in `guardian_action_proposals.args` |
| Confirm | `POST /v1/chat/confirm` replays args |
| Inbox | `GET /v1/chat/proposals?farm_id=&status=pending` |
| TTL | 5 minutes; revise refreshes chain (max 8 revisions) |

**Gap to communicate honestly:** Operator asks in natural language for a tool with **no matcher** → Guardian may explain in prose **without** opening a PR. Roadmap §8 addresses that.

---

## Industry standard (what we adopt)

| Pattern | gr33n implementation |
|---------|---------------------|
| **Approval gate** | Confirm only |
| **Frozen payload** | Server-side args at propose time |
| **Audit trail** | `guardian_tool_executed` |
| **Risk labeling** | low / medium / high on card |
| **Contextual launch** | Ask Guardian + `context_ref` |
| **Suggested prompts** | **Planned** — phase map below |
| **Pending inbox** | Shipped — drawer Pending + `/guardian/requests` |
| **No silent agent writes** | Persona + platform context block |

**Not adopting for farmers v1:** Autonomous schedule changes from chat without Confirm; ChatGPT-style “Actions” that run immediately.

---

## Conversation starters — product rules

### What starters are

- **UI chips** that set the chat input (or send immediately) with a **known-good prompt**.
- They are **not** proposals themselves — they trigger the same pipeline as typing.
- Max **3–5 visible** per surface; rotate by **snapshot state** (unread alerts > empty comfort band > offline device).

### What starters are not

- Random agronomy trivia (“Tell me about calcium”).
- Hidden auto-PR creation without a user send event.
- Replacement for wizards (44) on linear setup.

### Starter categories

| Category | Example chip | Likely matcher / outcome |
|----------|--------------|---------------------------|
| **Explain** | “What should I do about this alert?” | Read + text; may propose ack if alert id in context |
| **Fix** | “Set up feeding for this room” | `apply_grow_setup_pack` or `create_fertigation_program` if intents match |
| **Hardware** | “Help me connect my Pi” | Text + link to 44 wizard; optional `create_task` |
| **Review** | “Summarize today in this room” | Read tools + snapshot text |

### Placement map (by farmer phase)

| Phase | Surfaces | Starter examples |
|-------|----------|------------------|
| **40** | Zone Overview, Water, Climate, alert strip | “Acknowledge latest alert”, “Explain today’s schedule”, “Queue a 30s pump pulse” (pulse = UI or enqueue PR) |
| **41** | Dashboard morning strip | “What should I do first today?”, “Show unread alerts” |
| **42** | Comfort targets hub | “Set humidity comfort band for this room”, “Turn off shade rule until tomorrow” |
| **43** | Supplies hub | “What’s running low?”, “Log a mix for Flower Room” (→ link wizard or PR) |
| **44** | First-run checklist, setup wizard | “Set up indoor veg starter pack”, “Add my first grow room” |
| **45** | Sit-in findings | Copy and chips adjusted from user tests |

**Implementation note:** Starters can live in `ui/src/lib/guardianStarters.js` keyed by `surface + snapshot flags` (no schema).

---

## User-triggered PR flows (canonical)

### Flow A — Chat ask

1. Operator opens drawer or `/chat`.
2. Types or picks starter → `POST /v1/chat`.
3. Reads streamed answer.
4. If `proposals[]` on `done`, reviews card → Confirm / Dismiss / Refine.

### Flow B — Context button (shipped)

1. Operator on zone or alert → **Ask Guardian**.
2. `guardianPanel.open({ prefilledMessage, contextRef })`.
3. Same as Flow A when they send.

### Flow C — Inbox first (shipped)

1. TopBar badge → Pending tab.
2. Open proposal → Confirm (may not be in originating session if TTL allows).

### Flow D — Inline vs PR (40+)

| Action | Prefer |
|--------|--------|
| Ack alert on zone | **Inline** (40 WS4) — direct API |
| Ack alert from chat | PR `ack_alert` |
| Edit comfort band | **Inline** (40 WS2 / 42) |
| Patch program EC | PR `patch_fertigation_program` or Fertigation form |
| Run program now | **Direct** `run-now` API — not PR |
| Deploy shade | PR `enqueue_actuator_command` (high) or zone Climate button |

**Rule:** If the zone cockpit already has a one-tap action, **do not require PR** for the same job.

---

## Per-phase Guardian deliverables

### Phase 40 — Zone cockpit

| Item | Type |
|------|------|
| Starters on `ZoneDetail` / `ZoneNeedSection` | UI |
| Document when inline ack replaces PR | Doc in 40 WS8 |
| Guardian `context_ref` includes zone tab (water/climate) | API optional |
| No new write tools required for 40 v1 | — |

### Phase 41 — Farm hub

| Item | Type |
|------|------|
| Dashboard starters | UI |
| Pending badge visible on morning path | UI polish |
| Deep link: alert chip → Guardian with alert prefill | UI |

### Phase 42 — Comfort & automation

**Canonical spec:** [phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md)

| Item | Type |
|------|------|
| Starters on comfort / schedules / rules surfaces | UI — §2 of spec |
| `matchComfortAutomationIntent` for patch_* tools | Backend — §3 of spec |
| Persona + impact copy | WS6 in parent plan |
| operator-tour §5c + §6e, architecture §7.0h | Docs — written (planned ship) |

### Phase 43 — Operations

**Canonical spec:** [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md)

| Item | Type |
|------|------|
| Read enrichment `summarize_farm_low_stock` | Backend — spec §2 |
| Starters on Supplies / Feeding / Money (+ dashboard) | UI — spec §3 |
| `create_task_from_alert` for refill (existing; optional alert picker) | Backend — spec §4 |
| operator-tour §7 + §6f, architecture §7.0i | Docs — written (planned ship) |

### Phase 44 — Getting started

| Item | Type |
|------|------|
| Setup mode persona flag in chat when `?setup=1` or first-run | Handler optional |
| Starters: bootstrap template, grow setup pack | UI |
| Wizard primary; Guardian secondary | UX principle |

### Phase 45 — Validation

| Item | Type |
|------|------|
| Sit-in script includes 3 PR paths: ack, setup pack, dismiss | Protocol |
| Fix matcher gaps found in sit-in | Backlog → §8 |

---

## Starters vs LLM-tool routing (do not conflate)

| | **Starter chips / Ask Guardian (40–44)** | **LLM tool proposals (Phase 46)** |
|---|------------------------------------------|-----------------------------------|
| **Fixes** | Weak or generic **questions**; empty chat | Matcher **missed** a write intent |
| **Mechanism** | UI sends a **better message**; matchers unchanged | LLM emits **tool + args** → proposal row |
| **User still** | Sends / confirms send | Sends; may get card without exact phrase |
| **Example gap today** | Zone button: *"What's the current status of X?"* → obvious answer | *"Set feed volume to 0.3 L"* → advice text, **no card** |

**Your sit-in feedback** (generic Ask Guardian, obvious answers) is addressed in **Phase 40 WS7b + contextual prefills** — not by Phase 46 alone.

### Phase 40 — contextual Ask Guardian (shipped pattern, better copy)

Replace hard-coded prefills like `What's the current status of ${zone.name}?` with **snapshot-driven** prompts, e.g.:

- Unread alert present → *"Explain alert #N and what I should do in the next 10 minutes"*
- No comfort band → *"What humidity target should I set for {zone} at {stage}?"*
- Queue depth > 0 → *"What's queued for devices in {zone} and is it safe to run another pulse?"*

Implement via `ui/src/lib/guardianContextPrompts.js` (build message from props + farm store slice). Optional: send on chip click without opening drawer first.

---

## §8 — Future: LLM structured tool proposals (Phase 46)

**Plan:** [`phase_46_guardian_llm_tool_proposals.plan.md`](phase_46_guardian_llm_tool_proposals.plan.md)

**Problem:** Rule-assisted matchers do not cover all natural-language asks.

**Options (pick one in Phase 46):**

| Option | Pros | Cons |
|--------|------|------|
| **A. More matchers** | Deterministic, testable | Maintenance |
| **B. LLM JSON tool proposal** | Flexible language | Validation, safety, cost |
| **C. Hybrid** | Matcher first; LLM only if no match + Operate role | Complexity |

**Invariant unchanged:** Confirm gate + frozen args + audit.

**Not in 40–45 v1** — document so Phase 40 does not promise “any ask creates a PR.”

---

## OpenAPI / RAG

- OpenAPI: `GuardianActionProposal`, `POST /v1/chat/confirm`, `GET /v1/chat/proposals` (0.4.x).
- Re-ingest: add `guardian-change-requests-guide.md` to [`platform-doc-manifest.yaml`](../rag/platform-doc-manifest.yaml) when embedding next run.

---

## Definition of done (documentation gate)

- [x] Operator guide published
- [x] This plan linked from farmer roadmap + gaps index
- [ ] Phase 40 plan references PR boundaries (starters ≠ PR)
- [ ] operator-tour §6 cross-links new guide
- [ ] platform-doc-manifest row (optional before RAG run)

---

## Related

| Doc | Use |
|-----|-----|
| [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) | UI arc 40–45 |
| [phase_30_guardian_change_requests.plan.md](phase_30_guardian_change_requests.plan.md) | Shipped PR queue |
| [phase_34_guardian_pr_iteration.plan.md](phase_34_guardian_pr_iteration.plan.md) | Revise loop |
| [phase_44_getting_started_edge_wizard.plan.md](phase_44_getting_started_edge_wizard.plan.md) | Wizards vs Guardian |
