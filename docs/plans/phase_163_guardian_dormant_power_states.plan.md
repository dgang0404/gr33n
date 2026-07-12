---
name: Phase 163 — Guardian dormant/wake power states
overview: >
  Phase 129 made Guardian auto-awaken on login, but there was no deliberate
  "go to sleep" control — operators running on solar, battery, or metered
  power at remote sites had no way to release the loaded model's RAM/CPU
  draw between sessions without a terminal (`ollama stop`). This phase adds
  an explicit `dormant` awakening state, a matching Rest-now / Wake-now
  button pair in Settings, and lays the groundwork (later WS, not this
  phase) for an idle auto-dormant timer and swappable Guardian state
  artwork (hand-drawn druid art later — no AI-generated art, ever, for
  this feature).
todos:
  - id: ws1-dormant-backend
    content: "WS1: AwakeningStateDormant + RequestDormant() unload + POST /guardian/dormant"
    status: completed
  - id: ws2-settings-rest-button
    content: "WS2: Settings 'Rest now' button mirroring 'Awaken now'; awakening panel dormant copy"
    status: completed
  - id: ws3-auto-idle-dormant
    content: "WS3: Optional auto-dormant after N idle minutes (GUARDIAN_AUTO_DORMANT_MINUTES) — deferred"
    status: pending
  - id: ws4-power-docs
    content: "WS4: docs — model-unload dormant vs stopping the Ollama service; optional admin power script + sudoers note — deferred"
    status: pending
  - id: ws5-state-art-hook
    content: "WS5 (future, non-goal for this phase): swappable art slot per Guardian state for hand-drawn (non-AI) druid artwork — not started"
    status: pending
isProject: false
---

# Phase 163 — Guardian dormant/wake power states

**Status:** WS1 + WS2 shipped · WS3–WS5 deferred (tracked here for later phases)

**Depends on:** [Phase 129](phase_129_guardian_awakening.plan.md) (awakening states, warmup, readiness store)

---

## Problem

Guardian's awakening flow (Phase 129) is one-directional: cold → stirring → ready. There is no operator-facing way to go the other direction — deliberately release the loaded chat (and vision) model's RAM/CPU footprint when Guardian won't be used for a while. Today that requires a terminal (`ollama stop <model>`), which defeats the "no terminal" promise of Phase 129 and is a real cost on:

- **Solar/battery sites** — every extra watt held by an idle LLM model matters.
- **Shared/low-RAM laptops** — freeing RAM for other work between Guardian sessions.
- **Multi-farm operators** — switching farms/models without waiting for the old model to time out on its own `keep_alive`.

At the same time, unloading and reloading the model on **every single chat turn** would make each turn pay the cold-start cost — that already exists as a failure mode this phase must NOT reintroduce. Dormant is an **explicit, operator-initiated** rest state, not a per-turn behavior.

---

## Design — Guardian awakening states (extended)

| State | Meaning | Entered by |
|-------|---------|-----------|
| `unavailable` | `AI_ENABLED=false` or Ollama unreachable | Config / Ollama down |
| **`dormant`** (**new**) | Operator explicitly asked Guardian to rest; chat model unloaded on purpose | `POST /guardian/dormant` |
| `sleeping` | Chat model not loaded, no explicit rest request (cold start / never warmed) | Default cold state |
| `stirring` | Warmup in progress | `POST /guardian/warmup` |
| `ready` | Chat model loaded | Warmup completed |
| `busy` | In-flight grounded chat | Chat turn running |

`dormant` and `sleeping` look similar (model not loaded) but have different **copy and intent**: `sleeping` says "hasn't been awakened yet"; `dormant` says "resting on purpose to save power — tap to wake." Calling `POST /guardian/warmup` (Awaken now, or any chat send that auto-warms) clears the dormant flag immediately.

---

## WS1 — Backend: dormant state + endpoint ✅

**Shipped:**

- `internal/farmguardian/dormant.go` — `RequestDormant(ctx, llmBaseURL, chatModel, visionModel string) error` unloads the chat model (`keep_alive: 0` via the same Ollama `/api/generate` path Phase 130 uses for embed unload), best-effort unloads the vision model if set, and records an in-memory `dormantRequested` flag + timestamp (mirrors `warmupState` in `warmup.go`).
- `internal/farmguardian/awakening.go` — new `AwakeningStateDormant = "dormant"` constant; `BuildAwakeningHealth` reports `dormant` instead of `sleeping` when the model is unloaded **and** the dormant flag is set (checked after the existing `stirring`/`busy`/`ready` short-circuits, before the `sleeping` fallback).
- `internal/farmguardian/warmup.go` — `StartWarmup` calls `ClearDormantFlag()` before doing anything else, so any wake path (button, auto-warm on send, morning CTA) clears dormant.
- `internal/handler/chat/dormant.go` — `POST /guardian/dormant` (JWT; optional `farm_id` member check same as `PostWarmup`). Resolves the same chat/vision model Phase 129 warmup would have loaded (farm counsel/quick preference → env default), calls `RequestDormant`, returns `{"state": "dormant"}`.
- `cmd/api/routes.go` — registers `POST /guardian/dormant` next to `POST /guardian/warmup`.

**Not in WS1:** no change to per-turn chat behavior — dormant is only ever set by an explicit request, never by the chat pipeline itself.

---

## WS2 — UI: Rest now / Wake now ✅

**Shipped:**

- `ui/src/stores/guardianReadiness.js` — `restNow(farmId, mode)` action: `POST /guardian/dormant`, then re-fetch health. Mirrors `warmup()`.
- `ui/src/components/GuardianSettingsAwakeningCard.vue` — **Rest now** button next to **Awaken now**; disabled while `stirring`/`busy`/already `dormant`. `dormant` added to `stateLabel`/`stateBadgeClass` maps ("Resting", zinc badge).
- `ui/src/components/GuardianAwakeningPanel.vue` — dormant headline ("The Guardian rests to save power.") + message pointing at Settings → Rest now / Awaken now; distinct from the cold `sleeping` copy.

**Not in WS2:** no auto-triggered dormant from the UI (idle timers are WS3); no change to `GuardianNavLaunch` badge dot beyond reusing the existing `zinc` sleeping-like color for dormant.

---

## WS3 — Auto-idle dormant (deferred, not this phase)

Idea for a follow-up phase: `GUARDIAN_AUTO_DORMANT_MINUTES` env var; a lightweight ticker (or last-chat-activity timestamp check on the health poll) calls the same `RequestDormant` path after N idle minutes with no chat turns. Needs care to not fight the readiness store's own polling/auto-warm loop (`ensureAwake`) — likely gated to only arm after a session has been `ready` and then gone idle, never during active `stirring`.

---

## WS4 — Real power-off docs (deferred, not this phase)

`RequestDormant` only unloads the **model** from Ollama's RAM/VRAM — the **Ollama service itself** (and the machine) keeps drawing baseline power. For sites that want to cut Ollama's process entirely between sessions (bigger power win on solar), that still needs `systemctl stop ollama` / `start ollama`, which requires root and is intentionally **not** exposed over the web API (would let any JWT holder control a system service). Follow-up should ship:

- A doc section (`docs/farm-guardian-ollama-setup.md` or `local-operator-bootstrap.md`) explaining the two tiers: "Rest now" (RAM/VRAM only, API-driven, safe) vs. full service stop (terminal/cron/physical switch, admin-only).
- Optional: a small `scripts/guardian-power.sh {sleep|wake}` helper with a documented sudoers entry, for admins who want to wire it to a cron job or a physical low-power trigger — explicitly **not** reachable from the JWT-authenticated API.

---

## WS5 — Guardian state artwork (future, explicit non-goal for this phase)

Operator's stated intent for later: **hand-drawn, non-AI-generated** artwork per Guardian state (sleeping/dormant/stirring/ready/busy) — druid theme — to make the states feel natural instead of a status badge. This phase deliberately does **not** add any placeholder art (AI-generated or otherwise). When art assets exist, the hook point is the same `awakening.state` value already threaded through `guardianReadiness.js` / `GuardianAwakeningPanel.vue` / `GuardianSettingsAwakeningCard.vue` — a future phase can map `state → art asset` in one place without touching the state machine.

---

## Non-goals (this phase)

- Per-chat-turn model unload/reload (would reintroduce cold-start cost every message)
- Stopping the Ollama **service** (systemd) from the web API — security boundary, left to WS4 admin tooling
- Any AI-generated imagery for Guardian states
- Auto-dormant scheduling (WS3)

---

## Verification

```bash
go test ./internal/farmguardian/... ./internal/handler/chat/... -run Dormant -count=1
cd ui && npm test -- --run src/__tests__/guardian-settings-awakening.test.js
```
