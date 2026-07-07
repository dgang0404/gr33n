# Guardian QA smoke report — 2026-07-07 (complete)

**Suite:** `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`  
**Machine:** CPU laptop (`cpu_laptop` profile)  
**Run #2 finished:** 2026-07-06 21:15 → 22:37 local (~**81 min**)  
**Archive:** `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json`  
**Eval index:** `data/guardian_model_eval.json` (local, do not commit)

---

## Executive summary

| Metric | Run #1 (777s) | Run #2 (1500/1800s) |
|--------|---------------|---------------------|
| Heuristic pass | **0/4** | **4/4** |
| HTTP completions | 0 | 4 |
| Mean latency | — | **~20.4 min/prompt** |
| Grounded cite rate | 0% | **100%** (3/3 grounded) |
| Tools | partial | `walk_farm`, `list_unread_alerts` confirmed |

**Verdict:** Guardian smoke **passes heuristics** on phi3:mini CPU after laptop tune + fresh API. Answer **quality** has known gaps (prompt leak, fake URLs, weak pH coverage) — heuristics are lenient; operator review recommended.

---

## Per-prompt results (run #2)

| # | ID | Grounded | Pass | Latency | Citations | Tools |
|---|-----|----------|------|---------|-----------|-------|
| 1 | `smoke-cherry-forest` | No | ✅ | 11.7 min | 0 | — |
| 2 | `smoke-morning-walk` | Yes | ✅ | 24.0 min | 0 | `walk_farm` |
| 3 | `smoke-unread-alerts` | Yes | ✅ | 22.7 min | 3 | `list_unread_alerts` |
| 4 | `smoke-ec-ph` | Yes | ✅ | 23.2 min | 5 | RAG |

**Eval summary line:**
```
phi3:mini: grounded cite 100% · decline 0% · proposal 0% · latency 1224102ms
```

---

## Answer quality notes (human review)

### 1. `smoke-cherry-forest` — Good ✅
On-topic: cherry tree, goldenrod removal, blackberry/thorns, forest-garden framing. Sensible ungrounded counsel.

### 2. `smoke-morning-walk` — Pass but quality issues ⚠️
- **Prompt template leak:** Answer ends with `## Your task:Given the sources...` — instruction block echoed into user-visible reply.
- **Hallucinated links:** `https://gr33n.com/sources/field_guide`, `gr33n.com/tasks` — not real platform URLs.
- **Farm signal present:** Flower Room humidity, powdery mildew, veg EC 1.2–2.0 mS/cm — grounded in snapshot.
- **Tools:** `walk_farm` log evidence; heuristic pass also backed by tool scrape override.

### 3. `smoke-unread-alerts` — Strong ✅
Three concrete alerts: Flower Room humidity 72.4%, OHN-001 low stock, photoperiod transition. Actionable next steps. Citations `[2]`–`[4]`.

### 4. `smoke-ec-ph` — Pass, incomplete ⚠️
- EC ranges for lettuce/kale/spinach cited from docs.
- **pH barely addressed** — prompt asked EC **and** pH; answer is EC-heavy.
- Minor garbling: `1.0–1 endorsed`, `00.8–1.3` — small-model token noise.

---

## Infrastructure observations

| Issue | Severity | Notes |
|-------|----------|-------|
| `POST /guardian/warmup` → **503** before grounded block | Medium | Eval continued; inline warmup on send worked. Warmup likely `unavailable` because default env model is `tinyllama` (2048 ctx &lt; 8192 grounded floor). |
| `guardian_counsel_model` column missing | Medium | Background ticks error every 30s — **Phase 138 migration not applied** on local DB. |
| phi3 latency ~20 min/prompt | Operational | Expected on CPU; laptop tune required. Not a functional bug. |
| `data/guardian_model_eval.json` untracked | Low | Runtime artifact; should gitignore alongside `guardian_qa_runs/`. |

---

## Run comparison

| Run | Started | Timeouts | API | Result |
|-----|---------|----------|-----|--------|
| **#1** | 20:12 | 777s | Stale binary (warmup 404) | 0/4 `llm_timeout` |
| **#2** | 21:15 | 1500 / 1800 | `./bin/api` + phi3 pre-warm | **4/4 pass** |

**Pre-run fixes (run #2):** `make guardian-laptop-tune ARGS="--apply"`, API rebuild, `smoke_phase135_test.go` package fix, secrets → `.env.local`.

---

## Recommended fixes → Phase 143 planning

### Ship now (hygiene, no new features)

1. **Apply Phase 138 migration** — `guardian_counsel_model` / `guardian_quick_model` on local DB (`make migrate` or bootstrap).
2. **Gitignore** `data/guardian_model_eval.json`.
3. **Document** smoke pass ≠ quality pass in `docs/ci-guardian-qa.md` — link this report.

### Phase 143 — Guardian answer quality (proposed)

| WS | Title | Scope |
|----|-------|-------|
| WS1 | **Instruction leak guard** | Strip/detect `## Your task` and similar template echoes before persisting turn; optional eval heuristic `no_prompt_leak`. |
| WS2 | **Citation URL hygiene** | Reject or rewrite fake `gr33n.com` markdown links; prefer `[source#N]` only. |
| WS3 | **Warmup fix for eval** | `POST /guardian/warmup` with explicit `phi3:mini` or farm counsel model when eval model ≠ env default; fix 503 path. |
| WS4 | **Smoke heuristic hardening** | Morning-walk: fail on prompt leak / fake URL patterns; ec-ph: require `ph` in answer. |
| WS5 | **Manual review pass** | Settings → Guardian feedback on 4 smoke turns per `docs/guardian-feedback-review-runbook.md`. |

### Defer

- Full `make guardian-qa-regression` (~24 prompts, hours on CPU).
- LLM-as-judge (Phase 131 non-goal).
- Mandatory PR CI gate.

---

## Artifacts

| Path | Purpose |
|------|---------|
| `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json` | Full answers (gitignored) |
| `data/guardian_model_eval.json` | Eval index (local) |
| `/tmp/guardian-qa-smoke-run2.log` | Runner stdout |
| `/tmp/gr33n-api.log` | API timing + tools |

**Settings:** `GET /v1/guardian/qa/latest` should show this run (4 passed).
