# Guardian feedback review — operator runbook (Phase 141)

**Audience:** Farm admins and agronomy reviewers closing the Guardian quality loop after smoke tests or field use.

**Related:** [Phase 134 plan](plans/phase_134_guardian_answer_feedback.plan.md) · [Phase 131 QA harness](plans/phase_131_guardian_qa_harness.plan.md) · [ci-guardian-qa.md](ci-guardian-qa.md)

---

## When to review

| Trigger | Action |
|---------|--------|
| After `make guardian-qa-smoke` | Open Settings → **Guardian feedback**; check for down votes on the 4 smoke prompts |
| Weekly ops | Export last 7d down votes; triage invented data vs missed alerts |
| Before a demo | Confirm no open down votes on morning-walkthrough answers |

Archived smoke JSON includes `feedback_review_prompt` reminding you to check this workflow.

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
2. Settings → **Guardian QA — last run** (Phase 140) — confirm 4/4 heuristic pass
3. Manually thumbs-down any bad answers in the chat drawer during smoke
4. Settings → **Guardian feedback** — export or review down queue
5. *(Later)* Full manual walkthrough with turn debugger (`import.meta.env.DEV` / auth_test)

---

## Non-goals

- Automatic persona retraining from feedback
- Public multi-user ratings on the same turn
