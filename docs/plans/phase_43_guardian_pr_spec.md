---
name: "Phase 43 — Guardian PR spec (low-stock read + ops starters)"
overview: >
  Implementation spec for Phase 43 Guardian slice only: optional farm-wide low-stock
  read enrichment and conversation starters on Supplies / Feeding / Money hubs.
  No new Confirm write tools; reuse create_task_from_alert for refill tasks.
  Not LLM-tool routing (Phase 46).
parent_plan: phase_43_operations_stock_feeding_finance.plan.md
status: planned
---

# Phase 43 — Guardian PR spec (low-stock read + ops starters)

**Parent:** [phase_43_operations_stock_feeding_finance.plan.md](phase_43_operations_stock_feeding_finance.plan.md)

**Not in this doc:** Comfort/automation matchers → [phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md). NL → PR when matchers miss → [phase_46_guardian_llm_tool_proposals.plan.md](phase_46_guardian_llm_tool_proposals.plan.md).

**Prerequisites:** [guardian-change-requests-guide.md](../guardian-change-requests-guide.md) · Phase 40 WS7b (contextual prompts) · Phase 41 WS4 (Dashboard chips, `?zone_id=`)

---

## 1. What Phase 43 adds to Guardian

| Deliverable | Type | Outcome |
|-------------|------|---------|
| **`summarize_farm_low_stock` read enrichment** | Go (`readtools.go`) | Chat answers “what’s running low?” with batch names, qty, thresholds — no new Confirm tool |
| **Conversation starters** | UI (`guardianStarters.js`) | Chips on Supplies, Feeding admin, Money hubs send **job-shaped** messages |
| **Low-stock → task PR (existing)** | Matcher (optional tighten) | Unread `inventory_low_stock` alert + “refill / restock / task from alert” → `create_task_from_alert` |
| **Persona bullets** | Go + docs | Say “supplies”, “feeding program”, “receipt” — not `input_batches`, `nf_recipes` |
| **No new Confirm tools** | — | Inventory PATCH / batch create stay UI-first; gaps → Phase 46 backlog |

Phase 43 is **read + starters**, not a new proposal surface for stock writes.

---

## 2. Read tool: `summarize_farm_low_stock`

### 2.1 Why

Operators ask Guardian before opening **Supplies**. Today:

- Low-stock **alerts** exist (`inventory_low_stock` worker — [lowstock.go](../../internal/automation/lowstock.go)).
- `list_unread_alerts` enrichment may mention them, but there is **no dedicated inventory summary** in [readtools.go](../../internal/farmguardian/readtools.go).
- `list_plants` regex can false-positive on “inventory” in plant-catalog questions — avoid overloading that matcher.

### 2.2 Behavior

| Field | Value |
|-------|--------|
| **Trigger** | `shouldRunSummarizeFarmLowStockReadIntent(question)` |
| **Query** | `q.ListLowStockBatchesByFarm(ctx, farmID)` |
| **Block header** | `summarize_farm_low_stock — {farmName}` |
| **Lines** | Per batch: `{inputName}` — `{remaining}` / threshold `{threshold}` `{unit}`; batch `#id` |
| **Empty** | `No batches below their low-stock threshold right now.` |
| **Dedup** | Do not run if question is clearly plant-catalog (`listPlantsIntent`) |

### 2.3 Intent examples (match)

| Operator phrase | Match? |
|-----------------|--------|
| What’s running low? | yes |
| Low stock / supplies low / out of OHN | yes |
| Do I need to restock anything? | yes |
| List my plants | no (plants intent) |
| Summarize zone Flower Room | no (zone summarize) |

### 2.4 Implementation sketch

```go
// readtools.go — add to EnrichPromptBlock after list_unread_alerts when farm-scoped
func matchSummarizeFarmLowStockIntent(question string) bool { ... }
func renderSummarizeFarmLowStock(ctx context.Context, q db.Querier, farmID int64) (string, error) { ... }
```

Log: `logReadToolUse(ctx, "summarize_farm_low_stock", farmID, "", 0)`.

**Not** registered in `tools/registry.go` (read-only enrichment only, same as `summarize_zone`).

### 2.5 Acceptance (read)

- [ ] Farm with one batch below threshold → block lists input name and quantities
- [ ] Farm with none → friendly empty line, no error
- [ ] “List plants” does not attach low-stock block

---

## 3. Conversation starters

### 3.1 Surfaces (Phase 43 UI)

| Surface key | Route / component | Max chips |
|-------------|-------------------|-----------|
| `supplies_hub` | `/operations/supplies` (WS2 — route TBD) | 4 |
| `feeding_admin` | `/operations/feeding` (WS3) | 3 |
| `money_hub` | `/operations/money` (WS4) | 3 |
| `supplies_hub_zone` | Supplies with `?zone_id=` | 3 |
| `dashboard_ops` | Dashboard (41) low-stock chip area | 2 |

Reuse `GuardianStarterChips.vue` + `buildStarters(surface, ctx)` from Phase 42 WS8.

### 3.2 Chip selection rules

Evaluate in order; show first N that pass `when(snapshot, routeQuery)`:

| Priority | Condition | Chip label (example) | Message sent (example) |
|----------|-----------|----------------------|-------------------------|
| 1 | `ListLowStockBatchesByFarm` non-empty (API or snapshot flag) | What’s running low? | `What supplies are below their low-stock threshold on this farm?` |
| 2 | Unread alert `source_type === inventory_low_stock` | Turn alert into refill task | `Create a refill task from alert #{id} for {inputName}` |
| 3 | `?zone_id=` set + zone has active fertigation program | Feeding setup for this room | `Summarize feeding programs and reservoirs for {zoneName} — what should I check before the next run?` |
| 4 | On feeding admin, programs exist | When does feeding run next? | `When does the fertigation schedule for {zoneName} run next, in plain language?` |
| 5 | On money hub | Explain this month’s spend | `Summarize what I spent this month in plain language — no accounting jargon` |
| 6 | On supplies, recipes linked in ctx | Which recipe uses this input? | `Which mixing recipes use {inputName} and what should I reorder?` |
| 7 | (fallback supplies) | How do I log a mix? | `How do I log a nutrient mix and tie it to inventory on this farm?` |

**Chip click behavior:** Same as Phase 42 §2.2 — open drawer, `prefilledMessage`, `contextRef: { type: 'farm' | 'zone', surface: 'supplies_hub', ... }`, **no auto-send**.

### 3.3 Anti-patterns (reject in review)

| Bad chip | Why |
|----------|-----|
| “What’s the status of Inventory?” | Generic status — use job language |
| “Open `/inventory`” | Route jargon |
| “PATCH input_batch” | Schema language |

### 3.4 Acceptance (starters)

- [ ] Supplies hub shows “What’s running low?” when low-stock rows exist
- [ ] Dashboard chip (41) opens Guardian with same message when low-stock banner visible
- [ ] Chip does not Confirm without user send + Confirm tap

---

## 4. Proposals (PR cards) in Phase 43

### 4.1 In scope (existing tools only)

| Operator job | Primary path | Guardian PR |
|--------------|--------------|-------------|
| Restock reminder | Supplies hub list + alert banner | `create_task_from_alert` when starter/message matches unread low-stock alert |
| Generic follow-up | Money / receipt UI | `create_task` if message says “create task …” (existing matcher) |
| Ack low-stock alert | Alerts page / Dashboard | `ack_alert` (existing — not 43-specific) |

### 4.2 Optional matcher tighten (43, small)

Extend `pickAlertForIntent` preference when message contains `restock`, `refill`, `reorder`, `low stock`:

- Prefer alert where `TriggeringEventSourceType == inventory_low_stock`.

Document in `proposals_config_test.go` with fixture alert subjects like `Inventory low: OHN at …`.

### 4.3 Explicitly out of scope

| Ask | Phase |
|-----|-------|
| Adjust batch quantity / threshold via chat | UI on Supplies; else **46** |
| Create mixing event via Confirm | UI + 39 mix flow; no tool today |
| Post cost / receipt via Confirm | UI on Money hub; else **46** |

---

## 5. Persona & impact (WS6)

### 5.1 Persona (`platform_context.go`)

- Prefer **Supplies**, **Feeding details**, **Money** nav labels over Inventory / Fertigation / Costs.
- Low stock: cite **input name** and **remaining qty**, link mentally to Supplies hub (not SQL table).
- Do not promise Guardian can **change stock levels** — only read + task-from-alert.

### 5.2 Impact lines

No new tools — verify existing:

| Tool | Copy check |
|------|------------|
| `create_task_from_alert` | Impact mentions alert subject; refill tasks show input name when source is `inventory_low_stock` |

### 5.3 Deep links in chat footers (optional UI)

When `summarize_farm_low_stock` block is present, Guardian footer may show:

- **Open Supplies →** `/operations/supplies` (or `/inventory` until WS2 route ships — document redirect in parent plan).

---

## 6. Inline UI vs Guardian (Phase 43)

| Operator job | Primary path | Guardian |
|--------------|--------------|------------|
| See what’s low | Supplies hub banner + list | Read tool + starter |
| Log mix | Mixing log / zone Water | Starter “How do I log a mix?” (advice) |
| Attach receipt | Money hub camera/upload | Starter for spend summary only |
| Create refill task | Alert → Create task button | `create_task_from_alert` via chat |

**Wizards and forms win** over PR for writes.

---

## 7. Workstream mapping (parent plan)

| Parent WS | Guardian slice |
|-----------|----------------|
| WS1–WS5 | Hubs + cross-links; starters attach in WS2–WS4 |
| WS6 | §5 persona; this spec §2 read tool |
| WS7 | operator-tour §7 + §6f + architecture §7.0i |
| **WS8 (add to parent)** | This spec: read tool + starters + optional alert picker + Vitest |

---

## 8. Definition of done (Guardian slice only)

- [ ] `summarize_farm_low_stock` in readtools + tests
- [ ] Starters on supplies / feeding / money (+ dashboard when low stock)
- [ ] Optional: `pickAlertForIntent` prefers `inventory_low_stock` on refill phrases
- [ ] Persona bullets for operations vocabulary
- [ ] operator-tour §6f documents read vs PR
- [ ] No new registry Confirm tools
- [ ] No dependency on Phase 46

---

## Related

| Doc | Use |
|-----|-----|
| [guardian_pr_ux_through_farmer_phases.plan.md](guardian_pr_ux_through_farmer_phases.plan.md) | Cross-phase PR model |
| [workflow-guide.md](../workflow-guide.md) | Low-stock alert pipeline |
| [phase_42_guardian_pr_spec.md](phase_42_guardian_pr_spec.md) | Feeding schedule starters overlap — keep messages consistent |
