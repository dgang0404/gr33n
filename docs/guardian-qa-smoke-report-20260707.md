# Guardian QA smoke report — 2026-07-07 (in progress)

**Suite:** `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`  
**Machine:** CPU laptop (`cpu_laptop` profile)  
**Status:** **IN PROGRESS** — 2/4 prompts completed with HTTP 200 (as of 2026-07-06 ~22:11 local)

---

## Run comparison

| Run | Started | Timeouts | API binary | Result |
|-----|---------|----------|------------|--------|
| **#1** | 20:12 | `LLM_TIMEOUT_SECONDS=777` | Stale `/tmp/gr33n-api-test` (no `/guardian/warmup`) | **0/4** — all `llm_timeout` @ 777s |
| **#2** | 21:15 | `1500` / `GUARDIAN_GROUNDED_TIMEOUT_SECONDS=1800` | Fresh `./bin/api` (current `main`) | **2/4 done** — prompts 1–2 HTTP 200; 3–4 pending |

**Pre-run fixes (run #2):**

- `make guardian-laptop-tune ARGS="--apply"`
- Rebuilt API (`smoke_phase135_test.go` `package main` fix unblocked `go build`)
- phi3:mini Ollama pre-warm (~7.6s load)

**Logs:** `/tmp/guardian-qa-smoke-run2.log` · `/tmp/gr33n-api.log`

---

## Completed prompts (run #2)

### 1. `smoke-cherry-forest` (ungrounded) — HTTP 200

| Field | Value |
|-------|-------|
| Started | 21:15:52 |
| Completed | 21:27:33 |
| Wall time | ~12 min (`elapsed_ms=700956`) |
| Tokens | prompt 2486 · completion 470 |
| Citations | 0 |
| Heuristic | **TBD** — need full answer text for cherry/goldenrod/blackberry check |

### 2. `smoke-morning-walk` (grounded) — HTTP 200

| Field | Value |
|-------|-------|
| Started | 21:28:26 |
| Completed | 21:51:31 |
| Wall time | ~23 min (`elapsed_ms=1385088`) |
| Timeout budget | `llm_timeout_seconds=1800` |
| Tokens | prompt 4096 · completion 491 |
| Context chunks | 5 |
| Citations | 0 |
| Tools | `walk_farm`, `list_unread_alerts` (planned + used) |
| Inline warmup | yes (`guardian: inline warmup on send`) |
| Heuristic | **Likely pass** — answer length > 40 expected; full text TBD |

**Warmup note:** `POST /guardian/warmup` returned **503** (`state: unavailable`) before grounded block; eval continued. Inline warmup on send compensated for prompt 2.

---

## In progress

### 3. `smoke-unread-alerts` (grounded) — running

| Field | Value |
|-------|-------|
| Started | 21:51:56 |
| Request ID | `09404786-5186-4b5e-bcd4-e75a6cbcdd47` |
| Timeout budget | 1800s |
| Tools at start | `list_unread_alerts` |

### 4. `smoke-ec-ph` (grounded) — pending

Field-guide prompt; heuristic expects RAG citation or EC/pH in answer.

---

## Heuristic summary (partial)

```
phi3:mini: 2/4 HTTP 200 so far · cite TBD · decline TBD · proposal TBD · avg latency TBD
```

Final archive (when complete): `data/guardian_qa_runs/<timestamp>_smoke_phi3-mini.json`  
Eval index: `data/guardian_model_eval.json`

---

## Side observations

- API logs repeat `column "guardian_counsel_model" does not exist` on background ticks — schema migration may be behind code (Phase 138); does not block chat smoke.
- Run #1 archive: `data/guardian_qa_runs/20260707T010518_smoke_phi3-mini.json` (gitignored).

---

## Next update

When run #2 finishes, append:

- Pass/fail per heuristic (`internal/farmguardian/eval/score.go`)
- Full answer excerpts from QA archive JSON
- Final `phi3:mini: grounded cite % · decline % · proposal % · latency` line from eval runner
