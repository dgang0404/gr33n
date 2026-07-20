---
name: Phase 210 — Guardian natural farming integration
overview: >
  Guardian read tools over process catalog + farm inputs/batches/plants; write
  tools that draft input definitions, batches, and application recipes as
  Confirm-gated proposals. Additive regression fixture for goldenrod→JLF —
  do NOT alter the four-step smoke suite while it is in use.
todos:
  - id: ws1-read-tools
    content: "WS1: Read tools — lookup_process_catalog, suggest_process_from_material, summarize_natural_farming_inventory"
    status: completed
  - id: ws2-read-router
    content: "WS2: PlanReadTools intent routing for ferment/recipe/JLF/JMS/goldenrod/material questions"
    status: completed
  - id: ws3-write-tools
    content: "WS3: Write tools — draft_input_definition, draft_application_recipe, draft_input_batch (proposal only)"
    status: completed
  - id: ws4-llm-allowlist
    content: "WS4: Phase 46 LLM proposal allowlist + JSON schema for new tools"
    status: completed
  - id: ws5-regression-fixture
    content: "WS5: regression-cherry-goldenrod-jlf in RegressionFixtures ONLY — new score branch, smoke untouched"
    status: completed
  - id: ws6-tests-docs
    content: "WS6: readtools tests, tool execute tests, phase-210-closure; farm-guardian-architecture § natural farming tools"
    status: pending
isProject: false
---

# Phase 210 — Guardian natural farming integration

**Status:** WS5 shipped · **Depends on:** [208 knowledge](phase_208_natural_farming_process_knowledge.plan.md) · **Soft depend:** [209 studio UI](phase_209_natural_farming_studio_ui.plan.md) (proposal deep-links)

## The one job

> When an operator asks about goldenrod, ferments, or switching off EC bottles,
> Guardian **reads** the process catalog and farm inventory, then **proposes**
> recipes/batches — never silent DB writes.

## Smoke test safety (repeat — non-negotiable)

| Suite | File | Rule |
|-------|------|------|
| **Smoke (4-step)** | [`fixtures_smoke.go`](../../internal/farmguardian/eval/fixtures_smoke.go) | **No edits** to prompts, order, or `smoke-cherry-forest` score in [`score.go`](../../internal/farmguardian/eval/score.go) while smoke runs are active |
| **Regression** | `Fixtures()` / `RegressionFixtures()` | Add `regression-cherry-goldenrod-jlf` here |
| **Promotion** | Future phase | Only move cherry+recipe bar to smoke after regression stable |

Current smoke pass (unchanged):

```113:117:internal/farmguardian/eval/score.go
	case in.Question.ID == "smoke-cherry-forest":
		res.Passed = len(a) > 80 && (strings.Contains(a, "goldenrod") || strings.Contains(a, "blackberry") || strings.Contains(a, "cherry"))
		if !res.Passed {
			res.Notes = "expected forest-garden answer mentioning cherry/goldenrod/blackberry"
```

## WS1 — Read tools

| Tool ID | Trigger | Returns |
|---------|---------|---------|
| `lookup_process_catalog` | "what is JLF", "how do I make JMS" | Steps, dilution bands from 208 catalog + field guides |
| `suggest_process_from_material` | "goldenrod", "comfrey", material name | Matching catalog entries + suggested process type + linked guide |
| `summarize_natural_farming_inventory` | "what ferments do I have", "ready batches" | Farm `input_batches` by status + low stock |

**Farm plant tie-in (best-effort v1):**

- Cross-reference `list_plants` crop names / free-text labels against catalog `common_names`
- ponytail: string match on plant name + crop catalog common names; no computer vision

Register in [`readtools.go`](../../internal/farmguardian/readtools.go) + render in `readtools_plan.go`.

## WS2 — Read router

Extend [`readtools_router.go`](../../internal/farmguardian/readtools_router.go) intents:

```go
// illustrative patterns
jlfIntent  = `(?i)\b(jlf|fermented plant juice|plant juice)\b`
materialProcessIntent = `(?i)\b(goldenrod|comfrey|make|ferment|recipe|drench|foliar)\b`
nfInventoryIntent = `(?i)\b(jms|ffj|wca|ferment|batch|natural farming)\b.*\b(have|ready|stock)\b`
```

**Ungrounded cherry prompt:** stays ungrounded in smoke — no forced farm context. Regression fixture may use `Grounded: true` with farm plants including goldenrod label.

## WS3 — Write tools (Confirm-gated)

Follow [Phase 46](../plans/archive/phase_46_guardian_llm_tool_proposals.plan.md) pattern — proposals land in Pending.

| Tool ID | Action | Risk |
|---------|--------|------|
| `draft_input_definition` | POST `/naturalfarming/inputs` shape | medium |
| `draft_input_batch` | POST `/naturalfarming/batches` | medium |
| `draft_application_recipe` | POST `/naturalfarming/recipes` + components | medium |

**`suggest_process_from_material` is read-only.** Writes are separate explicit proposals:

> "Want me to draft a JLF application recipe for goldenrod biomass (extension method — start 1:100)?"

Proposal must include `source_tier` in summary when `extension_method`.

Proposal summary links to `/natural-farming?tab=recipes` when 209 shipped.

Implement executors in [`internal/farmguardian/tools/`](../../internal/farmguardian/tools/) calling existing handlers/queries.

## WS4 — LLM allowlist

Add to [`proposals_llm.go`](../../internal/farmguardian/proposals_llm.go) allowlist + arg validation:

- Reject hallucinated `input_definition_id` / `farm_id`
- Require catalog `material_id` OR explicit user-provided name for drafts
- `draft_application_recipe` must include `dilution_ratio` and `target_application_type` enum

## WS5 — Regression fixture (additive)

New entry in **`RegressionFixtures()`** only:

```go
{
    ID:       "regression-cherry-goldenrod-jlf",
    Category: "natural_farming",
    Prompt:   "...", // same cherry/goldenrod prompt OR variant with farm context on
    Grounded: true,  // farm has labeled goldenrod / outdoor zone
    Model:    "phi3:mini",
    ExpectTool: "suggest_process_from_material", // optional log scrape
}
```

**Score heuristic (new branch in score.go — NOT smoke-cherry-forest):**

Pass if answer:

- mentions `JLF` or `fermented plant juice` or `JADAM Liquid Fertilizer`
- mentions dilution band (e.g. `1:100`, `1:30`) OR cites process catalog / field guide
- for goldenrod: uses **extension** framing (local weed JLF method), not "Cho's goldenrod recipe"
- does **not** invent EC mS/cm targets for woodland/forage context
- optional: proposal card with `draft_application_recipe` OR explicit "I can draft a recipe"

## WS6 — Tests & docs

- `readtools_naturalfarming_test.go`
- `tools/naturalfarming_draft_test.go`
- `phase-210-closure.test.js` — tool IDs listed in platform context
- Update [`farm-guardian-architecture.md`](../farm-guardian-architecture.md)

## Acceptance criteria

- [ ] `suggest_process_from_material("goldenrod")` returns JLF entry from 208 catalog
- [ ] Grounded question produces recipe proposal with Confirm gate
- [ ] `make guardian-qa-smoke` — **4/4 unchanged**
- [ ] `make guardian-qa-regression` — new fixture passes on dev stack
- [ ] No silent writes — all creates go through Pending
- [ ] `phase-210-closure` tests green

## Out of scope

- Changing smoke-cherry-forest prompt or pass bar
- Auto-executing bootstrap without Confirm
- Livestock ration optimization
- Mixing event execution (operator still logs mix in Feed & water)
