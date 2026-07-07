# Guardian feedback review — operator runbook (Phase 141 · quality checklist Phase 143)

**Audience:** Farm admins and agronomy reviewers closing the Guardian quality loop after smoke tests or field use.

**Related:** [Phase 134 plan](plans/phase_134_guardian_answer_feedback.plan.md) · [Phase 131 QA harness](plans/phase_131_guardian_qa_harness.plan.md) · [Phase 143 answer quality](plans/phase_143_guardian_answer_quality.plan.md) · [ci-guardian-qa.md](ci-guardian-qa.md)

---

## When to review

| Trigger | Action |
|---------|--------|
| After `make guardian-qa-smoke` | Run [Smoke quality checklist](#smoke-quality-checklist-phase-143) on archived answers; then Settings → **Guardian feedback** |
| Weekly ops | Export last 7d down votes; triage invented data vs missed alerts |
| Before a demo | Confirm no open down votes on morning-walkthrough answers; no template leaks or fake URLs in last smoke archive |

Archived smoke JSON includes `feedback_review_prompt` reminding you to check this workflow. The eval runner also logs the same line when the archive is written.

---

## Smoke quality checklist (Phase 143)

**Heuristic 4/4 ≠ operator-trustworthy.** After smoke passes, spot-check each archived answer (`data/guardian_qa_runs/*.json` or Settings → **Guardian QA — last run**).

| Prompt | What to verify | Fail signals (thumbs-down + triage) |
|--------|----------------|--------------------------------------|
| `smoke-morning-walk` | Farm-specific walkthrough; no debug text | `## Your task`, echoed `Question:` block, **apology / “updated answer” tail**, instruction template at end of answer |
| `smoke-morning-walk` | Citations are labels, not fake links | `gr33n.com/`, **`gr33n-docs/`**, `https://gr33n.com/tasks`, markdown links to non-existent platform paths |
| `smoke-ec-ph` | Both EC **and** pH from docs | EC ranges only; no `ph` / pH targets; **off-topic tail** (endocrine, Lake Erie, unrelated field guides) |
| `smoke-cherry-forest` | On-topic forest-garden counsel | Invented farm (`secret mars dome`), off-topic filler |
| `smoke-unread-alerts` | Concrete seeded alerts | Generic “check your alerts” without humidity, stock, or photoperiod specifics |

**Dev turn debug** (local `import.meta.env.DEV`): confirm `leak_trimmed` and `citation_urls_sanitized` when WS1–2 hygiene fired on finalize.

**If any row fails:**

1. Open the smoke session in chat; **thumbs-down** with the closest reason chip (`invented_data`, `missed_alert`, `other`).
2. Settings → **Guardian feedback** — confirm the row appears; export CSV if sharing with agronomy.
3. Do **not** treat the run as demo-ready until WS1–4 fixes are deployed and smoke is re-run (Phase 143 WS6).

Eval heuristics (WS4) now encode several of these checks — failed rows show notes such as `instruction template leak`, `hallucinated gr33n.com citation URLs`, or `expected EC guidance and explicit pH targets`.

---

## In the UI (Settings)

1. Log in as **farm admin** on the demo farm.
2. Settings → **Guardian feedback — review queue** (farm counsel enabled).
3. Table shows **thumbs-down** rows from the last 7 or 30 days: question excerpt, reason chip, model, turn.
4. **Download CSV** for offline agronomy review or sharing with the team.

Operators submit feedback from the chat drawer via 👍/👎 on each assistant turn ([`GuardianTurnFeedback.vue`](../ui/src/components/GuardianTurnFeedback.vue)).

---

## CLI export (same data)

```bash
export TOKEN="<jwt from localStorage gr33n_token>"

# JSON
curl -s -H "Authorization: Bearer $TOKEN" \
  'http://127.0.0.1:8080/v1/chat/feedback/export?farm_id=1&since=7d' | jq .

# CSV
curl -s -H "Authorization: Bearer $TOKEN" \
  'http://127.0.0.1:8080/v1/chat/feedback/export?farm_id=1&since=7d&format=csv' \
  -o guardian-feedback-7d.csv
```

Columns: `session_id`, `turn_index`, `question`, `answer_excerpt`, `rating`, `reason`, `grounded`, `model`, `created_at`, `feedback_at`.

---

## Triage guide

| Reason (common) | Likely fix |
|-----------------|------------|
| Invented data | Persona honesty (133); check read-tool router fired (132); turn debug in dev |
| Missed alert | `walk_farm` / `summarize_unread_alerts` in turn debug; seed alerts present |
| Too slow | Runtime/timeouts (130); laptop tune; counsel vs quick model (138) |
| Other | Read full answer in session; compare to `data/guardian_qa_runs/` archive |

**Not in v1:** LLM-as-judge auto-scoring — human review only ([Phase 131 non-goals](plans/phase_131_guardian_qa_harness.plan.md)).

---

## Pair with QA smoke

Recommended order after Guardian changes:

1. `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`
2. Settings → **Guardian QA — last run** (Phase 140) — confirm heuristic pass/fail per step
3. [Smoke quality checklist](#smoke-quality-checklist-phase-143) — read full answers in the archive JSON
4. Manually thumbs-down any bad answers in the chat drawer during smoke
5. Settings → **Guardian feedback** — export or review down queue
6. *(Later)* Full manual walkthrough with turn debugger (`import.meta.env.DEV` / auth_test)

---

## Non-goals

- Automatic persona retraining from feedback
- Public multi-user ratings on the same turn
