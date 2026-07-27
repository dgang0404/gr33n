# Guardian QA — harness vs counsel gaps

Short checklist for **why laptop smoke needed extra terminal work** and what the harness + fixtures now cover. Ops detail: [ci-guardian-qa.md](../ci-guardian-qa.md) · pipeline teaching: [guardian-pipeline-for-csharp-devs.md](guardian-pipeline-for-csharp-devs.md).

## Harness gaps (fixed)

| Gap | Fix |
|-----|-----|
| Hard **4h** suite context killed NF prompts 8–16 | `GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS` (default **12**) in `guardian-eval` |
| One **smoke-full** blob (17 Q&A) | `guardian-qa-smoke-all` batches: core → NF batch1 → NF batch2 → phase127 → change-requests |
| No re-warmup between suites | Preflight (`POST /guardian/warmup`) **before each batch** in `scripts/guardian-qa-smoke-all.sh` |
| Stale hung `guardian-eval` | Optional `GUARDIAN_QA_KILL_STALE=1` |
| End = log only | Script prints **pass/total rollup** from newest archives under `data/guardian_qa_runs/` |

## Counsel / payload gaps (tests drive fixes)

| Gap | Test |
|-----|------|
| Recipe outcomes siloed from crop harvest analytics | `smoke-nf-history-compare` — expects recipe track record **and** crop analytics language; intents co-fire `summarize_recipe_outcomes` + `summarize_farm_crops_by_key` |
| Model invents bridge math between tools | Score + topic-drift (`if we assume`); recipe tool footer points at crop block when both present |
| Single-tool recipe smoke only | `smoke-nf-recipe-outcomes` (still useful alone) |

## Not a gap

Base LLM “knowing” the farm. Counsel quality = correct read-tool payloads + model obeying them. Server model + harness batching improve what we can **measure** and finish on a CPU laptop.

## Run one fixture

```bash
make guardian-qa-smoke-natural-farming MODEL=phi3:mini \
  # guardian-eval -suite smoke-natural-farming -prompt-ids smoke-nf-history-compare
```

Or add `-prompt-ids smoke-nf-history-compare` to the make target’s underlying `guardian-eval` invocation.
