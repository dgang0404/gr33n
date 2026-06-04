---
name: Phase 46 — Guardian LLM tool proposals (structured writes)
overview: >
  Close the gap where free-form chat gets good LLM answers but no proposal card because
  rule-assisted matchers missed the intent. Optional structured tool proposal from the
  LLM (validated, Confirm-gated) — not the same as conversation starter chips (40–44).
todos:
  - id: ws1-design
    content: "WS1: Design hybrid — matchers first; LLM tool JSON only on miss + Operate role"
    status: pending
  - id: ws2-schema
    content: "WS2: Tool-call schema aligned with registry.go; server validation + risk tier"
    status: pending
  - id: ws3-handler
    content: "WS3: Chat handler path — parse proposal from LLM; insert row; same Confirm flow"
    status: pending
  - id: ws4-safety
    content: "WS4: Tests — never execute without Confirm; reject hallucinated tool/args"
    status: pending
  - id: ws5-docs
    content: "WS5: Update guardian-change-requests-guide; operator-tour; OC-46"
    status: pending
isProject: false
---

# Phase 46 — Guardian LLM tool proposals

## Status

**Planned — after [Phase 45](phase_45_farmer_validation_whole_app_polish.plan.md).** Not part of Phases 40–44.

## This is NOT conversation starter chips

| Workstream | What it fixes | Phase |
|------------|---------------|-------|
| **Contextual Ask Guardian + starter chips** | Generic/obvious **prompts** and empty chat — chips send **better questions** | **40–44** ([guardian_pr_ux plan](guardian_pr_ux_through_farmer_phases.plan.md)) |
| **Better patch matchers** | More phrases recognized **without** LLM | **42** (and incremental) |
| **LLM tool proposals (this phase)** | Matcher miss but operator clearly asked for a **write** — LLM returns structured `tool` + `args` → same Confirm gate | **46** |

Starters improve **what you ask**. LLM-tool routing improves **whether a card appears** when you already asked for a change in natural language.

## Problem

Today: `BuildRuleAssistedProposals` only. Example: *"change my veg room feed to 0.3 L"* may get prose advice and **no** `patch_fertigation_program` card unless regex matches.

## Invariants (unchanged)

- Confirm only; frozen args; audit; risk tiers.
- Matchers run **first**; LLM proposal only when policy allows (hybrid recommended in [guardian_pr_ux plan §8](guardian_pr_ux_through_farmer_phases.plan.md)).

## Out of scope

- Removing Confirm.
- Autonomous writes.

## Related

[guardian-change-requests-guide.md](../guardian-change-requests-guide.md) · [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md)
