---
name: Phase 55 — Guardian ops / grow / money intelligence
overview: >
  Deepen Farm Guardian for operations jobs shipped in Phase 43 and 53 — read tools,
  starters, persona copy, and prompt blocks so operators get plain-language answers
  about stock, spend, and grow performance without opening Advanced editors.
  Complements Phase 46 LLM proposals (no new Confirm write tools in v1).
todos:
  - id: ws1-read-tools
    content: "WS1: Read tools — summarize_cycle_cost, summarize_farm_spending, restock_priority (wrap existing APIs)"
    status: completed
  - id: ws2-starters
    content: "WS2: Starters on Supplies, Money, grow strip, post-harvest, low-stock banner"
    status: completed
  - id: ws3-persona-prompt
    content: "WS3: Ops persona blocks for /operations/* routes + grow/money context_ref hints"
    status: completed
  - id: ws4-guardian-pr-spec
    content: "WS4: phase_55_guardian_pr_spec.md — matchers, banned phrases, impact previews"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: architecture §7.0s ops; smokes for read tools; phase-55-closure; OC-55"
    status: completed
isProject: false
---

# Phase 55 — Guardian ops / grow / money intelligence

## Status

**Shipped.** WS1–WS5 complete on `main`. Best after [Phase 53](phase_53_grow_stock_money_closure.plan.md) surfaces. **OC-55** closed.

**Spec pattern:** [phase_43_guardian_pr_spec.md](phase_43_guardian_pr_spec.md) · [phase_44_guardian_pr_spec.md](phase_44_guardian_pr_spec.md)

---

## The one job

> **Ask Guardian about stock, spend, or grow results — get grounded answers with numbers, not schema tours.**

---

## WS1 — Read tools (Go)

| Tool | Wraps | Example question |
|------|-------|------------------|
| `summarize_cycle_cost` | `GET /crop-cycles/{id}/cost-summary` | "What did Flower Room cost?" |
| `summarize_farm_spending` | costs summary + category rollup | "Spending this month by category" |
| `restock_priority` | low-stock batches + alerts | "What should I restock first?" |
| `summarize_active_grows` | active cycles per zone | "What's growing where?" |

Register in tool catalog; farm-scoped; JWT auth same as existing read tools.

---

## WS2 — UI starters

| Surface | Chips (max 4 each) |
|---------|-------------------|
| Supplies hub | Restock priority, What's running low, Log a mix help |
| Money hub | Month summary, Biggest spend category, Tag receipt help |
| Zone grow strip | Cost so far, Compare last run, Stage advice |
| Post-harvest screen | How did we do vs last time, Cost per gram |
| Dashboard low-stock chip | Open Supplies, Create refill task |

Reuse [guardianStarters.js](../../ui/src/lib/guardianStarters.js) builders.

---

## WS3 — Persona & prompt blocks

| Route | Prompt framing |
|-------|----------------|
| `/operations/supplies` | On-hand, batches, low-stock; never promise Guardian changes stock without Confirm |
| `/operations/money` | Receipts, month net, autolog lines; hide GL unless user on `/costs` |
| `/operations/feeding` | Admin vs daily Feed & water hub distinction |
| `/plants` | Plant catalog vs active grow run |
| Zone + active cycle context | Prefer cycle summary + cost over generic snapshot |

Update [platform_context.go](../../internal/farmguardian/platform_context.go) and [context_ref.go](../../internal/farmguardian/context_ref.go).

---

## WS4 — Guardian PR spec doc

Create `phase_55_guardian_pr_spec.md`:

- **No new Confirm tools** in v1 (restock/receipt stay UI)
- Matchers for "restock", "log receipt", "harvest" → point to UI wizards
- Banned: "Inventory module", "cost_transactions table"
- Impact preview templates for any future Phase 46 proposals

---

## WS5 — Docs, tests, OC-55

- `farm-guardian-architecture.md` §8 ops read tools table
- Go tests: read tool JSON shape, farm scope denial
- `ui/src/__tests__/phase-55-closure.test.js`
- Vitest: starter ids on hub surfaces

---

## Definition of done

- [x] Four read tools callable from grounded chat
- [x] Starters on Supplies, Money, grow strip, post-harvest, dashboard
- [x] phase_55_guardian_pr_spec.md published
- [x] OC-55 closed
