---
name: Phase 63 — Guardian session memory & farm patterns
overview: >
  Guardian remembers what the operator asked about across sessions — not to build
  a profile, but to provide better context in the current conversation and surface
  "you asked about this last week" prompts that actually help.
  All memory is farm-scoped, operator-visible, and deletable.
todos:
  - id: ws1-session-summary
    content: "WS1: Go — session summary endpoint; topic tagging on close; farm-scoped storage"
    status: completed
  - id: ws2-memory-ui
    content: "WS2: Session list shows topic tags; 'You last asked about X' recent-topic chip"
    status: completed
  - id: ws3-context-injection
    content: "WS3: System prompt includes relevant prior session summary when topic matches"
    status: completed
  - id: ws4-delete
    content: "WS4: Delete session deletes summary; export all sessions to text"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: privacy note in operator-tour; phase-63-closure; OC-63"
    status: completed
isProject: false
---

# Phase 63 — Guardian session memory & farm patterns

## Status

**Shipped.** `session_summaries` table, close-on-new-session summaries, topic chips, route-matched continue chip, prior-summary system prompt injection, Settings clear/export.

---

## The one job

> **Guardian remembers you asked about VPD last Tuesday so it doesn't start from scratch today.**

---

## What "memory" means here

**Not:** a persistent user profile Guardian silently builds.
**Yes:** each closed session gets a 2–3 sentence summary (topic + outcome) stored server-side. When a new session opens on a related topic, that summary is injected into the system prompt.

---

## WS1 — Session summary backend

```
POST /chat/sessions/{id}/close
```
Triggers background job: LLM summarizes session into `session_summaries` table:

```sql
session_summaries (
  session_id, farm_id, operator_id,
  summary_text, topics[], created_at
)
```

Topics: `['alerts', 'feeding', 'comfort', 'grow', 'stock', 'money', 'setup']`

---

## WS2 — Session list UI

Session list item (already in sidebar) adds topic chips:

```
Walk me through adding zones…  [setup] [zones]  17 min ago
```

**Recent topic chip** above starters:
```
You recently asked about VPD — continue?  [Pick up where I left off]
```

Only shows if summary exists and current route relates to same topic.

---

## WS3 — Context injection

When new session opens + topic overlaps prior summary:

```
[Prior session context: 3 days ago you asked about high VPD in Flower Room.
Guardian suggested tightening humidity from 65% to 60%. Outcome: unknown.]
Address the current question with this in mind if relevant. Do not repeat this note to the user.
```

Injection is a `<context>` block, not visible in chat.

---

## WS4 — Operator control

- **Delete session** → deletes `session_summaries` row
- **Clear all memory** → `/settings/guardian` — delete all summaries for this farm
- **Export** → plain text file, all session summaries

Operator always in control. No silent memory accumulation.

---

## WS5 — Docs, tests, OC-63

- operator-tour "Guardian session memory" note + privacy section
- Test: summary created on session close; injected on related topic; delete removes summary
- `phase-63-closure.test.js`

---

## Definition of done

- [x] Sessions tagged with topics
- [x] Related prior summary injected in system prompt
- [x] Clear memory option in settings
- [x] OC-63 closed

---

## Boundary

- Memory is **per-farm**, not across farms
- No cross-operator memory (each operator on a farm has their own sessions)
- No vector DB / embeddings v1 — topic matching is tag-based keyword overlap
