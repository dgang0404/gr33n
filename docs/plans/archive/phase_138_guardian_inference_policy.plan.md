---
name: Phase 138 — Guardian inference policy & enterprise scale
overview: >
  Server farms: split embed vs chat hosts, per-farm Counsel/Quick model settings,
  token budget hints before send, and health that reports both inference endpoints.
  Laptop profile remains default in docs; this phase formalizes Profile C/D.
todos:
  - id: ws1-split-host-health
    content: "WS1: Health — llm_host + embedding_host reachability, loaded models per host; warmup targets chat host only"
    status: completed
  - id: ws2-farm-model-policy
    content: "WS2: farms.meta_data or settings — guardian_counsel_model, guardian_quick_model, grounded_timeout_seconds; UI Settings two dropdowns"
    status: completed
  - id: ws3-presolve-models
    content: "WS3: ResolveOutcome — Quick chat uses quick model; Farm counsel uses counsel model; farm default becomes counsel"
    status: completed
  - id: ws4-cost-estimate
    content: "WS4: GET /v1/chat/usage + last-turn avg — UI hint before send: ~Nk prompt tokens typical for farm counsel"
    status: completed
  - id: ws5-org-defaults
    content: "WS5 (optional): org-level guardian policy JSON — max models, allowed pull list for enterprise"
    status: cancelled
  - id: ws6-docs
    content: "WS6: recommended-hardware Profile C/D; hypothetical-enterprise-topology Guardian section; env vars"
    status: completed
  - id: ws7-tests
    content: "WS7: resolve model smoke; health dual-host mock; settings save counsel/quick"
    status: completed
isProject: false
---

# Phase 138 — Guardian inference policy

**Status:** **Shipped.** · **Depends on:** [129](phase_129_guardian_awakening.plan.md), [130](phase_130_guardian_runtime_orchestration.plan.md)

**Related:** [recommended-hardware-and-sizing.md](../recommended-hardware-and-sizing.md), [hypothetical-enterprise-topology.md](../hypothetical-enterprise-topology.md)

---

## WS1 — Split inference hosts

| Env | Health probe |
|-----|----------------|
| `LLM_BASE_URL` | chat completions /api/ps on chat host |
| `EMBEDDING_BASE_URL` | embed model on embed host (may differ) |

Warmup **never** loads embed on chat-only warm path unless single-host profile detected.

Laptop profile: same URL → existing contention rules (130).

---

## WS2 — Farm model policy

Settings per farm:

| Setting | Default laptop | Default server |
|---------|----------------|----------------|
| Counsel model | phi3:mini | llama3.1:8b |
| Quick model | tinyllama | llama3.1:8b or tiny |
| Grounded timeout | 1500 | 666 |

Stored via `PATCH /farms/{id}/settings` (`guardian_counsel_model`, `guardian_quick_model`, `guardian_grounded_timeout_seconds`).

---

## WS4 — Pre-send cost hint

Non-blocking UI line:

> Farm counsel typically uses ~3,800 prompt tokens on this farm (last 5 turns avg).

From `GET /v1/chat/usage` + session history stats. If near budget cap → amber link to Settings usage card.

---

## Acceptance

- [x] Server .env with different embed URL — health shows both reachable
- [x] Quick chat uses tinyllama while counsel uses phi3 on same farm
- [x] Cost hint appears on Farm counsel mode only

---

## Non-goals

- Auto-provision Ollama on second host
- Billing / chargeback per farm
- WS5 org-level policy JSON (deferred)
