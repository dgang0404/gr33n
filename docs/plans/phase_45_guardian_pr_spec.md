---
name: "Phase 45 — Guardian PR spec (sit-in paths + protocol)"
overview: >
  Phase 45 Guardian slice: validate ack_alert, apply_grow_setup_pack, and dismiss
  with non-technical farmers; triage matcher gaps to Phase 46. Protocol artifact
  in docs/workstreams/farmer-sit-in-protocol.md.
parent_plan: phase_45_farmer_validation_whole_app_polish.plan.md
status: completed
---

# Phase 45 — Guardian PR spec (sit-in paths + protocol)

**Parent:** [phase_45_farmer_validation_whole_app_polish.plan.md](phase_45_farmer_validation_whole_app_polish.plan.md)

**Protocol (WS1):** [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md) · **Dry-run log:** [sit-in-45-dry-run-log.md](../workstreams/sit-in-45-dry-run-log.md)

**Not in this doc:** Implementing Phase 46 — only **backlog** matcher misses found in sit-in.

---

## 1. What Phase 45 adds to Guardian

| Deliverable | Type | Outcome |
|-------------|------|---------|
| **Sit-in protocol** | Doc | Scripted ack / setup pack / dismiss paths |
| **Pass/fail criteria** | QA | Operators understand Confirm vs Dismiss |
| **Matcher gap backlog** | Process | Phrases → Phase 46 or incremental matcher PR |
| **Copy/a11y fixes** | UI | Impact lines, Dismiss visibility, mobile taps (WS3/WS6) |
| **No new tools** | — | Validation only unless P0 requires hotfix |

Phase 45 does **not** ship new proposal types — it **proves** 40–44 + Phase 32 paths work for farmers.

---

## 2. Three required PR paths

| Path | Tool | Risk | Primary UI alternative |
|------|------|------|------------------------|
| **Ack alert** | `ack_alert` | low | Alerts → Acknowledge button |
| **Grow setup pack** | `apply_grow_setup_pack` | high | Plants + Fertigation manual |
| **Dismiss** | *(none — UI only)* | — | Ignore suggestion |

### 2.1 Ack — acceptance

- [x] Card shows `Acknowledge: {subject}` impact
- [x] Confirm updates alert state in UI
- [x] Starter from Dashboard/alert row (40) produces same card as typed phrase

### 2.2 Setup pack — acceptance

- [x] `SetupPackProposalCard` readable without training
- [x] High-tier warning seen before Confirm
- [x] Revise loop tested once per session (*“0.3 L not 0.5”*) — `guardian-proposal.test.js` revision diff

### 2.3 Dismiss — acceptance

- [x] Operator states Dismiss does **not** change farm data (aria-label + dry-run)
- [x] Dismiss visible on mobile (min tap target WS6)
- [x] Facilitator documents if operator confused Dismiss with Confirm — none in dry-run

---

## 3. Sit-in integration (WS1–WS2)

| WS | Action |
|----|--------|
| WS1 | Run [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md) sessions A + B |
| WS2 | Triage issues; tag `matcher_gap` → [phase_46](phase_46_guardian_llm_tool_proposals.plan.md) §9 backlog |

**Dry-run evidence:** [sit-in-45-dry-run-log.md](../workstreams/sit-in-45-dry-run-log.md) · `./scripts/sit-in-dry-run.sh`

---

## 4. Copy pass v2 — Guardian strings (WS3)

Ban in farmer routes + proposal cards:

| Term | Replace with |
|------|----------------|
| `patch_rule` | Pause automation / resume rule |
| `apply_grow_setup_pack` | Set up grow in {zone} (card title) |
| `guardian_action_proposals` | *(never show)* |

Audit:

- [x] `GuardianActionProposal.vue` tool labels map
- [x] `guardianImpact.js` lines for ack + setup pack
- [x] HelpTips on Confirm / Dismiss / Refine (WS6 aria-labels)

---

## 5. Matcher gap → Phase 46 backlog

When sit-in records **fail** with “expected PR card”:

| Log field | Example |
|-----------|---------|
| `phrase` | “set feed to 300 ml” |
| `expected_tool` | `patch_fertigation_program` |
| `got` | text only |
| `owner` | 46 hybrid or 42/43 incremental matcher |

Dry-run: **no gaps filed.** Re-open if external farmer sit-in finds misses.

---

## 6. Workstream mapping

| Parent WS | Guardian slice |
|-----------|----------------|
| WS1 | Protocol + this spec §2 |
| WS2 | Triage + 46 backlog · [phase-45-ws2-friction-backlog.md](../workstreams/phase-45-ws2-friction-backlog.md) |
| WS3 | §4 copy |
| WS6 | Dismiss/Confirm a11y |
| WS7 | operator-tour §9 + README farmer-ready |
| **WS8** | Sit-in PR checklist (this doc §7 DoD) |

---

## 7. Definition of done (Guardian slice)

- [x] [farmer-sit-in-protocol.md](../workstreams/farmer-sit-in-protocol.md) executed (DR-A + DR-B dry-run)
- [x] ack + setup pack + dismiss **pass** documented
- [x] Matcher gaps filed for 46 (none in dry-run)
- [x] P0 empty; Guardian copy pass merged
- [x] operator-tour §9 links protocol

**Vitest:** `phase-45-ws8-guardian-closure.test.js`

---

## Related

| Doc | Use |
|-----|-----|
| [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md) | Wizards before setup-pack starter |
| [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md) | Post–sit-in NL → card |
