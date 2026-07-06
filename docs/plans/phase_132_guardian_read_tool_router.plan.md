---
name: Phase 132 — Guardian read-tool router v2 + walkthrough proposals
overview: >
  Replace fragile regex-only read-tool gating with a farm-counsel tool router: a core
  set always runs on relevant modes, expanded natural-language intents, and morning
  walkthrough findings that spawn ack/task proposals without matching write-intent regex.
todos:
  - id: ws1-tool-router-core
    content: "WS1: ReadToolRouter — farm_counsel plans tools from mode (morning_walkthrough, context_ref) + question; core set: snapshot always, walk_farm on walk mode"
    status: completed
  - id: ws2-expanded-intents
    content: "WS2: Widen intents — supplemental light, urgent issue, out-of-band readings, fertigation plain language; audit table in readtools_*_test.go"
    status: completed
  - id: ws3-always-on-minimal
    content: "WS3: Farm counsel default — always inject summarize_device_health + unread alert summary when farm has alerts (bounded caps)"
    status: completed
  - id: ws4-walkthrough-proposals
    content: "WS4: walk_farm warn findings → frozen ack_alert / create_task proposals (rule path, no LLM proposals flag)"
    status: completed
  - id: ws5-router-logging
    content: "WS5: Structured log guardian_tool_plan + eval/QA expect_tool field; smoke_phase132_test.go"
    status: completed
  - id: ws6-docs
    content: "WS6: architecture § read-tool router; operator-tour; Phase 128/131 smoke prompts cite router"
    status: completed
isProject: false
---

# Phase 132 — Guardian read-tool router v2

**Status:** **Shipped.** · **Depends on:** [130](phase_130_guardian_runtime_orchestration.plan.md), [131](phase_131_guardian_qa_harness.plan.md)

**Continues:** [Phase 73](phase_73_guardian_pr_discoverability.plan.md), [Phase 60](phase_60_guardian_morning_walkthrough.plan.md)

---

## Problem

Read tools enrich the system prompt only when regex matches. Natural questions miss triggers; Guardian honesty rules make it deny access to data that exists in DB/readtools.

---

## Design

### Tool router (new `internal/farmguardian/readtools_router.go`)

```go
type ToolPlan struct {
    ToolIDs []string
    Reason  map[string]string // tool_id → why selected
}

func PlanReadTools(question string, ref *ContextRef, snap Snapshot, mode string) ToolPlan
```

**Selection order:**

1. **Mode forced** — `context_ref.guardian_mode=morning_walkthrough` → `walk_farm` (+ optional device health)
2. **Core farm counsel** — if grounded: `summarize_unread_alerts` when `snap.UnreadAlerts > 0` (cap 3 subjects)
3. **Intent match** — existing `shouldRun*` functions (widened in WS2)
4. **Never** run all tools — stay within prompt budget (reuse `promptBudget`)

### Walkthrough → proposals (WS4)

After `walk_farm` render, if finding severity `warn` + category `alerts`:

- Emit rule-based `ack_alert` or `create_task` proposal (same frozen path as Phase 29 matchers)
- UI shows card: *"Humidity alert in Flower Room — acknowledge?"*

No `GUARDIAN_LLM_PROPOSALS` required.

---

## WS2 — Intent expansions (minimum)

| Tool | New phrases |
|------|-------------|
| `site_weather` | supplemental light, bright enough, frost tonight |
| `summarize_device_health` | offline, edge device, pi stale, not checking in |
| `summarize_zone_fertigation` | fertigation setup, tanks, programs plain language |
| `lookup_crop_targets` | active grows, what stage, behind schedule |
| `walk_farm` | urgent issue, most important, what needs attention |

---

## Acceptance

- [ ] "Do I need supplemental light today?" injects `site_weather` when coords set (regression from Phase 73)
- [ ] Morning walkthrough without exact phrase still runs `walk_farm` via `guardian_mode`
- [ ] "What is the most urgent issue?" runs `walk_farm` or alert summary without timeout-only failure
- [ ] Walkthrough with seed humidity alert offers ack proposal card
- [ ] `make guardian-qa-smoke` step 3 logs `tool_id` evidence for alerts

---

## Non-goals

- Full LLM tool-selection (defer; router is deterministic v1)
- New read tools (use existing registry)
