---
name: Phase 134 — Guardian answer feedback & quality loop
overview: >
  Operators rate turns (thumbs + optional reason); metadata stored on conversation_turns
  for agronomy review. QA harness archives pair with feedback export. Feeds future
  persona/guide improvements — no LLM-as-judge in v1.
todos:
  - id: ws1-schema
    content: "WS1: Migration — conversation_turns.feedback_rating (up|down|null), feedback_reason text, feedback_at; optional operator_user_id"
    status: completed
  - id: ws2-api
    content: "WS2: PATCH /v1/chat/sessions/{id}/turns/{turn_index}/feedback {rating, reason}"
    status: completed
  - id: ws3-ui
    content: "WS3: GuardianChatPanel — thumbs on assistant messages; optional reason popover on down"
    status: completed
  - id: ws4-export
    content: "WS4: GET /v1/chat/feedback/export?farm_id= (admin) CSV/JSON for agronomy review; exclude PII beyond user id"
    status: completed
  - id: ws5-qa-link
    content: "WS5: guardian-qa JSON includes feedback_prompt field; doc runbook: review down votes after smoke"
    status: completed
  - id: ws6-tests
    content: "WS6: handler test feedback PATCH; vitest thumbs; smoke_phase134_test.go"
    status: completed
isProject: false
---

# Phase 134 — Guardian answer feedback loop

**Status:** **Shipped.** · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md)

---

## Problem

131 records automated runs; operators have no in-product signal when answers are wrong. Persona and field guides can't improve without structured negative examples.

---

## WS1 — Storage

```sql
ALTER TABLE gr33ncore.conversation_turns
  ADD COLUMN feedback_rating text CHECK (feedback_rating IN ('up','down')),
  ADD COLUMN feedback_reason text,
  ADD COLUMN feedback_at timestamptz;
```

One rating per turn per user (upsert on PATCH).

---

## WS2 — API

- Auth: turn must belong to user's session
- Farm member when grounded
- `reason` optional, max 500 chars, required for `down` in UI (soft required server-side)

---

## WS3 — UI

- Thumbs appear after stream completes (not during generating)
- Down → small textarea: "What was wrong?" (chips: "Invented data", "Missed alert", "Too slow", "Other")
- `data-test="chat-feedback-down"`

---

## WS4 — Export (admin / farm_manager)

Weekly review workflow:

```bash
curl -H "Authorization: Bearer …" \
  'http://127.0.0.1:8080/v1/chat/feedback/export?farm_id=1&since=7d'
```

Columns: `turn_index`, `question`, `answer_excerpt`, `rating`, `reason`, `grounded`, `model`, `created_at`

---

## Acceptance

- [x] Operator can thumbs-down morning walkthrough answer with reason
- [x] Export returns row linked to session
- [x] Feedback does not block chat or require network beyond API

---

## Non-goals (v1)

- LLM-as-judge auto-scoring
- Automatic persona retraining
- Public feedback / multi-user ratings on same turn
