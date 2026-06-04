---
name: "Phase 42 — Guardian PR spec (starters + matchers)"
overview: >
  Implementation spec for Phase 42 Guardian slice only: conversation starters on
  comfort/automation surfaces and rule-assisted matchers for patch_rule,
  patch_schedule, patch_fertigation_program. Not LLM-tool routing (Phase 46).
parent_plan: phase_42_comfort_targets_automation_plain_language.plan.md
status: planned
---

# Phase 42 — Guardian PR spec (starters + matchers)

**Parent:** [phase_42_comfort_targets_automation_plain_language.plan.md](phase_42_comfort_targets_automation_plain_language.plan.md)

**Not in this doc:** LLM structured tool proposals → [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md)

**Prerequisites:** [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) · Phase 40 WS7b (contextual prompts pattern) · Phase 41 WS4 (`EmptyStateHint`)

---

## 1. What Phase 42 adds to Guardian

| Deliverable | Type | Outcome |
|-------------|------|---------|
| **Conversation starters** | UI (`guardianStarters.js`) | Chips on comfort hub, schedules view, rules view send **specific** messages |
| **Rule-assisted matchers** | Go (`proposals_comfort.go` or extend `proposals_config.go`) | Natural-language asks for patch/disable schedule/rule/program open **Confirm cards** |
| **Persona + impact copy** | Go + docs | Cards say “comfort band” / “pause schedule” not `patch_rule` |
| **No new Confirm tools** | — | Reuse `patch_schedule`, `patch_rule`, `patch_fertigation_program` in registry |

Starters improve **questions**. Matchers improve **card appearance**. Neither replaces Confirm.

---

## 2. Conversation starters

### 2.1 Surfaces (Phase 42 UI)

| Surface key | Route / component | Max chips |
|-------------|-------------------|-----------|
| `comfort_hub` | `/comfort-targets` (WS1) | 4 |
| `schedules_farmer` | Farmer schedules view (WS3) | 3 |
| `rules_farmer` | Farmer rules view (WS4) | 3 |
| `comfort_hub_zone` | Hub filtered to one zone (`?zone_id=`) | 3 |

### 2.2 Chip selection rules

Evaluate in order; show first N that pass `when(snapshot, routeQuery)`:

| Priority | Condition | Chip label (example) | Message sent (example) |
|----------|-----------|----------------------|-------------------------|
| 1 | Zone has no humidity band + has RH sensor | Set humidity comfort band | `Help me set a humidity comfort band for {zoneName} at {stage} stage` |
| 2 | Zone has rule `is_active` + GH type | Pause shade automation | `Disable the greenhouse shade rule for {zoneName} until I turn it back on` |
| 3 | Unread alert mentions setpoint/threshold | Explain this alert | `Explain alert #{id} and whether I should change my comfort targets` |
| 4 | Schedule linked to zone program | When does feeding run next? | `When does the fertigation schedule for {zoneName} run next, in plain language?` |
| 5 | (fallback) | What should I fix in Targets? | `What comfort targets am I missing for {zoneName}?` |

**Chip click behavior:**

1. Open Guardian drawer (`guardianPanel.openDrawer({ tab: 'chat' })`).
2. Set `prefilledMessage` to message template (interpolate zone/stage/alert from farm store).
3. Set `contextRef`: `{ type: 'zone', zone_id, surface: 'comfort_hub' }` (extend `guardianRouteRef` if needed).
4. **Do not auto-send** on v1 — operator reviews and sends (same as Ask Guardian).

### 2.3 File layout (UI)

```
ui/src/lib/guardianStarters.js       # buildStarters(surface, ctx) → [{ id, label, message, contextRef? }]
ui/src/lib/guardianContextPrompts.js # shared interpolators (from Phase 40 WS7b)
ui/src/components/GuardianStarterChips.vue
```

Wire into comfort hub + schedules/rules pages in Phase 42 WS8 (UI), not Phase 40.

### 2.4 Acceptance (starters)

- [ ] No chip text equals generic “What’s the status of X?”
- [ ] At least one chip on comfort hub changes when zone lacks a band (demo seed)
- [ ] Sending chip message does **not** Confirm anything without user tap

---

## 3. Rule-assisted matchers (backend)

### 3.1 Gap today

`matchFreshProposal` → `matchConfigToolIntent` handles tasks and cycle stage only.

`patch_schedule`, `patch_rule`, `patch_fertigation_program` are **registered tools** (Confirm works) but **no fresh-message matcher** — operators only get cards via **revise** on an existing draft or manual UI.

Phase 42 adds **`matchComfortAutomationIntent`** called from `matchFreshProposal` after config intent, before return false.

### 3.2 Matcher: `patch_rule` (disable / enable)

| Intent regex (examples) | Args | Summary |
|-------------------------|------|---------|
| `(disable\|turn off\|pause\|stop).*(rule\|automation\|shade\|vent)` | `is_active: false`, `rule_id` if parsed else first zone-matched rule | `Disable rule "{name}"` |
| `(enable\|turn on\|resume).*(rule)` | `is_active: true`, … | `Enable rule "{name}"` |

**Resolution:**

1. Parse `rule_id` from `rule #12` if present.
2. Else match rule name substring against `snap` (extend snapshot with active rules per zone in Phase 42 WS3 — **read path only**, optional lightweight query in matcher via `q.ListRulesByFarm` if snapshot insufficient).

**Risk:** `patch_rule` with `is_active: false` → **high** tier (existing `risk.go`).

### 3.3 Matcher: `patch_schedule`

| Intent | Args | Summary |
|--------|------|---------|
| `(pause\|disable\|stop).*(schedule\|feeding\|lights?)` | `is_active: false`, `schedule_id` | `Pause schedule "{name}"` |
| `run (at\|every)? \\d{1,2}(:\d{2})? ?(am\|pm)?` + schedule context | `cron_expression` derived | `Change schedule "{name}" to run at …` — **v1 optional**; may defer cron patch to Phase 42 WS3 UI only |

**v1 minimum:** disable/enable schedule by name or id only.

### 3.4 Matcher: `patch_fertigation_program`

| Intent | Args | Summary |
|--------|------|---------|
| `(set\|change\|update).*(volume\|feed).*(\\d+\\.?\\d*)\\s*l` | `program_id`, `total_volume_liters` | `Set program "{name}" volume to {n}L` |
| `(set\|change).*(ec\|conductivity).*(\\d+\\.?\\d*)` | `ec_trigger_low` | `Set EC target to {n}` |
| `irrigation only\|plain water` + program context | `irrigation_only: true` (if PATCH supports) | `Switch program to irrigation only` — verify handler fields |

**Resolution:** Match zone name in message → program linked to zone (`snap` programs list — extend snapshot in Phase 42 if needed).

### 3.5 Tests

| File | Cases |
|------|-------|
| `proposals_comfort_test.go` | disable rule, pause schedule, set volume 0.3L, no match on pure Q&A |
| Extend smoke (optional) | Chat POST with phrase → `proposals[0].tool` assert |

### 3.6 Acceptance (matchers)

- [ ] “Turn off the shade rule for Flower Room” → `patch_rule` proposal when rule exists
- [ ] “Set feed volume to 0.3 L for Flower Room” → `patch_fertigation_program` when program exists
- [ ] “What is EC?” → no proposal (text only)
- [ ] Confirm still replays frozen args (no regression)

---

## 4. Persona & impact (WS6)

### 4.1 `platform_context.go` / persona

Add bullet: prefer **comfort band**, **feeding schedule**, **automation rule** in summaries; never say `zone_setpoints` to operators.

### 4.2 `impact.go`

Extend `patch_rule` / `patch_schedule` / `patch_fertigation_program` lines:

| Tool | Impact line pattern |
|------|---------------------|
| `patch_rule` | `Pause automation rule "{name}" — it will stop firing until re-enabled` |
| `patch_schedule` | `Pause schedule "{name}" — no automatic runs until re-enabled` |
| `patch_fertigation_program` | `Update feeding program "{name}": {fields} — does not run the program now` |

### 4.3 Guardian starters in RAG

After ship: `make rag-ingest-platform-docs` (operator-tour §6e + this spec is planning-only unless added to manifest exclude).

---

## 5. Inline UI vs PR (Phase 42)

| Operator job | Primary path | Guardian PR |
|--------------|--------------|-------------|
| Edit comfort band | `ComfortBandEditor` (WS2) | Optional: “set band” only if no inline on page |
| Pause rule | Rules view toggle (WS4) | Matcher + starter for chat-first users |
| Pause schedule | Schedules view toggle | Same |
| Patch program volume | Fertigation or zone Water | Matcher for chat |

**Do not** remove toggles in favor of chat-only.

---

## 6. Workstream mapping (parent plan)

| Parent WS | Guardian slice |
|-----------|----------------|
| WS1–WS5 | UI; starters attach in WS1/WS3/WS4 |
| WS6 | §4 persona + impact |
| WS7 | operator-tour §6e + architecture §7.0h |
| **WS8 (add to parent)** | This spec: starters + matchers + Vitest |

---

## 7. Definition of done (Guardian slice only)

- [ ] `guardianStarters.js` + chips on comfort/schedules/rules surfaces
- [ ] `matchComfortAutomationIntent` + tests
- [ ] impact + persona bullets
- [ ] Documented in operator-tour §6e
- [ ] No dependency on Phase 46

---

## Related

| Doc | Use |
|-----|-----|
| [guardian_pr_ux_through_farmer_phases.plan.md](guardian_pr_ux_through_farmer_phases.plan.md) | Cross-phase PR model |
| [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md) | NL → card when matchers still miss |
