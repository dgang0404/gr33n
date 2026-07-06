---
name: Phase 133 — Guardian answer grounding & honesty (sources, trim, trust)
overview: >
  Operators see where each claim came from (live farm vs field guide vs RAG note),
  get warned when prompt trim reduces context, and citations distinguish curated guides
  from operator-ingested text. Persona rules enforce labeling; UI shows trim banner.
todos:
  - id: ws1-source-labeling-persona
    content: "WS1: Persona + synthesis rules — label Live farm / Field guide [n] / Farm note [n]; forbid presenting stale RAG as current sensor state"
    status: pending
  - id: ws2-trim-banner-ui
    content: "WS2: API returns trim_summary on chat done event; GuardianChatPanel amber banner + New chat CTA when history/RAG/snapshot trimmed"
    status: pending
  - id: ws3-citation-metadata
    content: "WS3: Citations include source_type (field_guide|platform_doc|operational); UI chip colors per type"
    status: pending
  - id: ws4-rag-trust-tiers
    content: "WS4: synthesis — operator-ingested chunks tagged untrusted_operational in prompt; field_guide/platform_doc trusted_curated"
    status: pending
  - id: ws5-mode-card-sources
    content: "WS5: Farm counsel mode card legend — what each layer means for truth (pairs 129 WS5)"
    status: pending
  - id: ws6-tests
    content: "WS6: Unit tests trim_summary payload; vitest citation chips; smoke grounded answer contains Live farm when snapshot used"
    status: pending
isProject: false
---

# Phase 133 — Guardian answer grounding & honesty

**Status:** planned · **Depends on:** [129](phase_129_guardian_awakening.plan.md) WS5, [132](phase_132_guardian_read_tool_router.plan.md)

**Related:** [rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md), Phase 97 structured truth

---

## Problem

| Issue | Example |
|-------|---------|
| RAG vs snapshot | "EC is high" from old note vs live reading |
| Silent trim | Turn 9 quality drops; operator unaware |
| Citation opacity | `[1]` without knowing if guide or random task note |

---

## WS1 — Source labeling (persona)

Add to grounded system block:

```
When answering:
- LIVE FARM DATA (snapshot, read tools): say "right now" / "on your farm today"
- FIELD GUIDE [n] / PLATFORM DOC [n]: say "per our field guide" / "per platform docs"
- Never state a sensor value from RAG alone — cross-check snapshot/read tools or say you only have a note
```

---

## WS2 — Trim visibility

Extend SSE `done` payload:

```json
{
  "trim_summary": {
    "history_turns": "20→8",
    "rag_top_k": "8→5",
    "snapshot_reduced": true,
    "effective_context_window": 4096
  }
}
```

UI: `data-test="chat-trim-warning"` — *"Long chat — earlier turns trimmed for phi3 CPU budget. Start a new chat for best results."*

Reuse existing `logPromptBudgetTrims` data from handler.

---

## WS3 — Citation metadata

`citations[]` in chat response already has fields — ensure `source_type` populated from `rag_embedding_chunks`. UI:

| type | Chip |
|------|------|
| `field_guide` | green "Field guide" |
| `platform_doc` | blue "Platform doc" |
| operational | zinc "Farm note" |

---

## WS4 — Trust tiers in prompt

When injecting operational RAG chunks, prefix block:

```
The following operator notes may be outdated — prefer LIVE FARM DATA for current state.
```

Field guides omit this prefix.

---

## Acceptance

- [ ] Grounded answer to "unread alerts" references live data explicitly
- [ ] Trim on 9th turn in long session shows banner once per turn when trim occurred
- [ ] Citation click/show includes source_type label
- [ ] QA smoke step 4 (EC/pH) cites field guide or platform doc type in JSON

---

## Non-goals

- Real-time sensor fusion into RAG index
- LLM-generated source attribution without retrieval metadata
