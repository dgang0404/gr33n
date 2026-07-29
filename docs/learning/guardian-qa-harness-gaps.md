# Guardian QA — harness vs counsel gaps

Short checklist for **why laptop smoke needed extra terminal work** and what the harness + fixtures now cover. Ops detail: [ci-guardian-qa.md](../ci-guardian-qa.md) · pipeline teaching: [guardian-pipeline-for-csharp-devs.md](guardian-pipeline-for-csharp-devs.md) · Phase [211.06](../plans/phase_211_06_guardian_smoke_reliability.plan.md).

## Harness gaps (fixed)

| Gap | Fix |
|-----|-----|
| Hard **4h** suite context killed NF prompts 8–16 | `GUARDIAN_EVAL_SUITE_TIMEOUT_HOURS` (default **12**) in `guardian-eval` |
| One **smoke-full** blob (17 Q&A) | `guardian-qa-smoke-all` batches: core → NF batch1 → NF batch2 → phase127 → change-requests |
| No re-warmup between suites | Preflight (`POST /guardian/warmup`) **before each batch** in `scripts/guardian-qa-smoke-all.sh` |
| Stale hung `guardian-eval` | Optional `GUARDIAN_QA_KILL_STALE=1` |
| End = log only | Script prints **pass/total rollup** from newest archives under `data/guardian_qa_runs/` |
| smoke-all said OK on 0/N fixtures | `GUARDIAN_QA_FAIL_ON_REGRESSION=1` default in smoke-all → `-fail-on-regression` |
| Rollup “(no archives)” despite files on disk | Do not pipe suite runners (bash subshell dropped rollup arrays); skip `partial_*.json` |
| Partial archive only on success | `RunSuite` rewrites partial after error/timeout rows too |
| Full smoke-all as only debug loop | `make guardian-qa-debug` = core + NF batch1; smoke-all = certification |

## Counsel / payload gaps (tests drive fixes)

| Gap | Test / fix |
|-----|------------|
| Recipe outcomes siloed from crop harvest analytics | `smoke-nf-history-compare` — expects recipe track record **and** crop analytics language |
| Model invents bridge math between tools | Score + topic-drift (`if we assume`); recipe tool footer |
| Invent / apology / instruction-soup mid-marathon | `AnswerLooksLikeInventStub` + one regenerate, then refuse (`invent_retry.go`) |
| NF dilution invent despite tool footer | `EnsureNFDilutionRatiosInAnswer` quotes catalog ratios (medium template) |
| ec-ph near-miss (EC + cites, no “pH” token) | `InjectPHFromChunks` + score soften via cite excerpts |
| Write proposal pass ≠ answer quality | `answer_relevant` on archives; notes `proposal_ok; answer_low_relevance` |

## Not a gap

Base LLM “knowing” the farm. Counsel quality = correct read-tool payloads + model obeying them. Server model + harness batching improve what we can **measure** and finish on a CPU laptop.

## Run one fixture / debug loop

```bash
make guardian-qa-debug MODEL=phi3:mini FARM_ID=1   # core + NF batch1
make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1   # core only
make guardian-qa-smoke-natural-farming MODEL=phi3:mini \
  # guardian-eval -suite smoke-natural-farming -prompt-ids smoke-nf-history-compare
```

Optional cool-down: `GUARDIAN_EVAL_PROMPT_COOLDOWN_SECONDS=30`.

Or add `-prompt-ids smoke-nf-history-compare` to the make target’s underlying `guardian-eval` invocation.
