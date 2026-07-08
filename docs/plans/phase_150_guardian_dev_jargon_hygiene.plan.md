---
name: Phase 150 — Guardian developer-jargon answer hygiene
overview: >
  Run #6's unread-alerts answer echoed raw developer strings verbatim from
  docs/local-operator-bootstrap.md — "`PATCH /alerts/{id}/acknowledge`",
  "proposal card → Confirm" — mid-sentence, not just as a trailing dump (which
  Phase 143's TrimSourceDump already catches). Phase 150 adds an answer-level
  redaction for literal HTTP-verb+path jargon so it can't reach farmer-facing
  chat regardless of which platform_doc chunk it came from.
todos:
  - id: ws1-redact-dev-jargon
    content: "WS1: RedactDevAPIJargon — strip METHOD /path tokens, collapse empty parens/dangling arrows"
    status: completed
  - id: ws2-wire-finalize
    content: "WS2: Wire into sanitizeAssistantAnswer finalize pipeline + TurnDebug fields"
    status: completed
  - id: ws3-drift-scorer
    content: "WS3: AnswerContainsDevAPIJargon as a smoke/regression hygiene failure"
    status: completed
isProject: false
---

# Phase 150 — Guardian developer-jargon answer hygiene

**Status:** **Shipped.** · **Depends on:** [143](phase_143_guardian_answer_quality.plan.md) · [148](phase_148_guardian_citation_claim_accuracy.plan.md)

---

## Why this phase

`docs/local-operator-bootstrap.md` is a **developer/operator setup doc** — it's correctly written with terminal commands and HTTP verbs for a human reading the markdown file. But it's also indexed as a `platform_doc` RAG source, and when cited for a farmer-facing question the model can paraphrase its literal API strings straight into the answer: `` `PATCH /alerts/{id}/acknowledge` `` mid-sentence in run #6's unread-alerts answer. Phase 143's `TrimSourceDump` only catches a full trailing `Sources:`-style dump, not an inline dev-jargon phrase embedded in otherwise-good prose.

## Workstreams

### WS1 — Redaction function ✅

**Shipped:** `internal/farmguardian/answer_leak.go` — `RedactDevAPIJargon` matches `` `?(GET|POST|PATCH|PUT|DELETE)\s+/path`? `` (optionally backtick-wrapped), removes it, and collapses the empty-parens / dangling-arrow artifacts left behind.

### WS2 — Finalize pipeline wiring ✅

**Shipped:** `internal/handler/chat/answer_finalize.go` — `sanitizeAssistantAnswer` calls `RedactDevAPIJargon` right after `TrimSourceDump`; `TurnDebug.DevJargonRedacted` / `DevJargonCharsRemoved` persist for dev inspection.

### WS3 — Drift scorer backstop ✅

**Shipped:** `topic_drift.go` `smokeAnswerHygieneNote` fails on `AnswerContainsDevAPIJargon` — a backstop in case any dev-jargon pattern reaches the archived answer before this ships everywhere, or a new corpus doc introduces the same leak.

---

## Acceptance

- [x] `RedactDevAPIJargon` unit test reproduces the exact run #6 sentence and removes the HTTP path while preserving surrounding farm content.
- [x] No redaction (no-op) when the answer has no dev-jargon pattern.
- [x] `AnswerContainsDevAPIJargon` used as an eval hygiene failure alongside the existing leak/dump checks.

## Non-goals

- Rewriting `local-operator-bootstrap.md` itself — it remains correct developer-facing documentation; the fix is at the answer boundary so it protects against *any* doc leaking dev jargon, not just this one.
- Blocking legitimate mentions of REST concepts in prose that aren't a literal `METHOD /path` token.
