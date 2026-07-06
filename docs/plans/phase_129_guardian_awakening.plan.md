---
name: Phase 129 — Guardian awakening (login-and-go operator UX)
overview: >
  Operators log in and use Guardian without Ollama rituals or .env surgery. The UI
  explains Quick chat vs Farm counsel, shows an awakening progress state while the
  backend prepares inference, surfaces readiness on login and in Settings, and ships
  idempotent laptop tune scripts. Runtime chat orchestration (embed unload on send,
  early SSE phases) is Phase 130 — this phase is what the grower sees and touches.
todos:
  - id: ws0-health-consolidation
    content: "WS0: Extend GET /v1/chat/health (Phase 37) — add ollama_loaded_models, chat_model_loaded, embed_blocks_chat, awakening_state, rag_corpus_ok; avoid duplicate /guardian/readiness unless alias"
    status: pending
  - id: ws1-warmup-api
    content: "WS1: POST /guardian/warmup {mode, farm_id} — async stir→ready; idempotent; returns 202 + poll health"
    status: pending
  - id: ws2-warmup-orchestration
    content: "WS2: Warmup backend — unload embed if blocking; preload chat model (phi3 farm_counsel / tinyllama quick) via Ollama generate+keep_alive; verify RAG chunk counts"
    status: pending
  - id: ws3-laptop-tune-script
    content: "WS3: scripts/tune-guardian-laptop.sh + make guardian-laptop-tune — cpu-16gb profile; LLM_TIMEOUT≥1500, RETRY=1; --apply; optional GUARDIAN_AUTO_TUNE on restart-local --serve"
    status: pending
  - id: ws4-awakening-ui
    content: "WS4: GuardianAwakeningPanel — poll /v1/chat/health; druid copy + checklist; Quick chat fallback; hide when ready"
    status: pending
  - id: ws5-context-mode-cards
    content: "WS5: Quick chat vs Farm counsel mode cards with layer chips (snapshot, read tools, RAG, chat); link connectivity-requirements"
    status: pending
  - id: ws6-login-badge
    content: "WS6: guardianReadiness store — post-login fetch; TopBar/GuardianNavLaunch asleep|stirring|ready|busy dot; background warmup once per session"
    status: pending
  - id: ws7-morning-walkthrough-cta
    content: "WS7: Farm counsel + awakening — Morning check starter pre-warms farm_counsel; dashboard chip opens drawer in walkthrough mode"
    status: pending
  - id: ws8-settings-surface
    content: "WS8: Settings Guardian card — readiness summary, corpus counts, Tune laptop link, last warmup error"
    status: pending
  - id: ws9-stack-scripts
    content: "WS9: check-local-stack + restart-local — probe Ollama /api/tags; warn if down; print guardian-laptop-tune hint on --serve"
    status: pending
  - id: ws10-lite-degrade
    content: "WS10: AI_ENABLED=false / Ollama down — mode cards show Lite path; no infinite awakening spinner"
    status: pending
  - id: ws11-docs-tests
    content: "WS11: Bootstrap + playbook rewrite; health/warmup handler tests; UI tests mode cards + awakening + morning CTA"
    status: pending
isProject: false
---

# Phase 129 — Guardian awakening (login-and-go operator UX)

**Status:** planned

**Blocks:** [Phase 128](phase_128_validate_phase127_guardian.plan.md) WS3 manual UI (morning walkthrough needs timeout + awakening first)

**Pairs with:** [Phase 130](phase_130_guardian_runtime_orchestration.plan.md) (embed unload on send, early SSE, grounded timeout env)

**Related:** [Phase 126](phase_126_guardian_cpu_efficiency.plan.md), [Phase 60](phase_60_guardian_morning_walkthrough.plan.md),
[local-operator-bootstrap.md](../local-operator-bootstrap.md), [connectivity-requirements.md](../connectivity-requirements.md)

---

## Problem (operator-reported)

After laptop reboot, grounded Guardian requires a **manual ritual** (.env, `ollama stop`, `ollama run`, new chat, 25 min wait). Morning walkthrough **did** run `walk_farm` but phi3 timed out at **777 s** with no first token — twice.

The backend is capable; the **operator journey** is not shippable.

---

## The one job

> **Log in → see if the Guardian is awake → pick Quick chat or Farm counsel → ask — no terminal.**

---

## Design principles

1. **Login never blocked** — dashboard loads; only Guardian surfaces wait/fallback.
2. **Modes are visible** — not a hidden checkbox + `<details>` dev essay.
3. **Extend, don't fork** — build on `GET /v1/chat/health` (Phase 37), `GET /guardian/models`, Phase 126 SSE `status`.
4. **Explicit tune** — `.env` changes via `make guardian-laptop-tune --apply`, not silent rewrite each request.
5. **Druid tone, not cosplay** — warm one-liners; checklist shows real subsystems.
6. **Honest about CPU** — Farm counsel may take many minutes; say so in the mode card.

---

## Gap analysis (what Phase 129 must close)

| Gap | Today | Phase 129 fix |
|-----|--------|----------------|
| Duplicate health APIs | `/v1/chat/health` has RAG counts; no loaded models | **WS0** extend health |
| No warmup | Cold phi3 after reboot | **WS1–2** POST warmup |
| Manual `.env` | 777s too low for grounded CPU | **WS3** tune script |
| Buried farm context | Checkbox + collapsed help | **WS5** mode cards |
| No pre-send UX | Blank until stream starts | **WS4** awakening panel (Phase 130 adds in-turn SSE) |
| Morning check disconnected | Starters exist; no warm path | **WS7** CTA + pre-warm |
| Settings blind | Health only in curl/tour | **WS8** Settings card |
| Ollama down after reboot | check-stack ignores Ollama | **WS9** stack probe |
| Lite / AI off | Generic errors | **WS10** degrade |
| Phase 128 blocked | Manual walkthrough times out | **129 + 130** then re-run checklist |

**Deferred to Phase 130 (not 129):** auto embed unload on each chat send, early SSE before snapshot/embed, eval HTTP 120s timeout, server-side chat busy lock.

---

## WS0 — Health consolidation (extend Phase 37)

Extend `GET /v1/chat/health?farm_id=` — **do not** add a parallel `/guardian/readiness` unless it is a thin alias.

New fields on response:

```json
{
  "ai_enabled": true,
  "field_assistant": { "...existing..." },
  "awakening": {
    "state": "sleeping|stirring|ready|busy|unavailable",
    "profile": "cpu_laptop|gpu_server|lite",
    "chat_model": "phi3:mini",
    "chat_model_loaded": false,
    "embed_model": "rjmalagon/gte-qwen2-...",
    "embed_loaded": true,
    "embed_blocks_chat": true,
    "ollama_loaded_models": ["..."],
    "rag_corpus_ok": true,
    "field_guide_chunks": 58,
    "platform_doc_chunks": 12,
    "messages": ["Embedding model is using RAM — awakening will make room for chat."],
    "warmup_in_progress": false,
    "last_warmup_error": ""
  }
}
```

`state` rules:

| state | Meaning |
|-------|---------|
| `unavailable` | `AI_ENABLED=false` or Ollama unreachable |
| `sleeping` | AI on, chat model not loaded |
| `stirring` | Warmup in progress |
| `ready` | Chat model loaded (or quick mode satisfied) |
| `busy` | In-flight grounded chat (Phase 130 adds server flag; until then UI `streaming` only) |

`rag_corpus_ok`: `field_guide_chunks > 0` OR platform chunks > 0 when farm counsel expected (warn with link to `make guardian-bootstrap-farm`).

Reuse: `farmguardian.EnrichModelRuntimeHints`, `BuildFieldAssistantHealth`, `ollama_ps.go`.

---

## WS1 — Warmup API

`POST /guardian/warmup` (JWT, farm member when `farm_id` set)

```json
{ "mode": "quick" | "farm_counsel", "farm_id": 1 }
```

- `202` + `{ "state": "stirring" }` when work started
- `200` + `{ "state": "ready" }` when already warm
- Idempotent while stirring
- Poll via `GET /v1/chat/health`

---

## WS2 — Warmup orchestration

**farm_counsel:**

1. Resolve model: farm default → first grounded-capable → `phi3:mini`
2. If `embed_blocks_chat` (CPU heuristic: embed loaded, chat not, `size_vram=0`): Ollama unload embed
3. Minimal `POST /api/generate` on chat model, `keep_alive: 30m`
4. Set in-memory warmup state until `ollama ps` confirms chat loaded

**quick:**

1. Resolve `LLM_MODEL` / `tinyllama`
2. Preload; skip embed unload unless RAM critical

**Not in warmup:** full RAG re-ingest (too heavy). Health reports chunk counts only.

---

## WS3 — Laptop tune script

`./scripts/tune-guardian-laptop.sh [--apply] [--profile cpu-16gb|gpu-server]`

| Check | cpu-16gb | gpu-server |
|-------|----------|------------|
| `LLM_TIMEOUT_SECONDS` | warn if `< 1500` | warn if `< 666` |
| `LLM_RETRY_MAX_ATTEMPTS` | `1` | `1` |
| `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` | suggest `1800` if unset | optional |

- Detect profile: `ollama ps` all `size_vram=0` → cpu_laptop
- `make guardian-laptop-tune` wrapper
- `restart-local.sh --serve`: if `GUARDIAN_AUTO_TUNE=1`, run tune `--apply` quietly
- `setup-first-clone.sh`: print tune hint after bootstrap

---

## WS4 — Awakening UI

`GuardianAwakeningPanel.vue` in drawer + `/chat`

| state | Copy (examples) |
|-------|-----------------|
| `sleeping` | "The Guardian rests. Awakening…" |
| `stirring` | Checklist: ☐ Field memories ☐ Live farm ☐ Voice |
| `ready` | Hidden |
| `unavailable` | "Guardian is in Lite mode" or "Ollama not reachable — start Ollama, then retry" |
| `failed` | Last error + **Try Quick chat** |

Flow: mount → `GET /v1/chat/health` → if `sleeping` and mode needs it → `POST /guardian/warmup` → poll 2s.

**Do not block Send** for Quick chat when tinyllama path is ready; block Farm counsel send only while `stirring` (optional soft block with "Awaken first" — product choice: prefer soft warn + auto-warm on send in Phase 130).

---

## WS5 — Context mode cards

Replace `Use farm context` checkbox in `GuardianChatPanel`:

### Quick chat
- **Layers:** Chat model only
- **Speed:** Fast (seconds–few min on CPU)
- **Use for:** Off-farm horticulture, general Q&A
- **No:** snapshot, read tools, RAG, field guides
- Triggers warmup `mode=quick` on select

### Farm counsel
- **Layers:** chips — Snapshot · Read tools · RAG · Chat model
- **Speed:** Slow on CPU (many minutes per turn)
- **Use for:** Morning walkthrough, alerts, zones, demo farm
- **Requires:** grounded-capable model (phi3+)
- Triggers warmup `mode=farm_counsel` on select
- Link: [connectivity-requirements.md](../connectivity-requirements.md) § cold models

Keep `GuardianModelSelector` under collapsible **Voice & models**.

---

## WS6 — Login badge

`ui/src/stores/guardianReadiness.js`:

- After `capabilities.fetch()` + auth, `fetchHealth(farmId)` once per session
- If `sleeping` && `ai_enabled`, background `warmup(farm_counsel)` when farm selected
- `GuardianNavLaunch` + TopBar: dot `zinc|amber|green|red`

---

## WS7 — Morning walkthrough CTA

Wire [Phase 60](phase_60_guardian_morning_walkthrough.plan.md) starters to awakening:

- Dashboard **Morning check** chip → open drawer, set Farm counsel, fire warmup, prefill starter with `guardian_mode: morning_walkthrough`
- Chat panel starters: same pre-warm before send
- Copy: "Morning check uses Farm counsel — the Guardian reads your farm first"

Unblocks Phase 128 prompt #1-style validation without operator knowing the stack.

---

## WS8 — Settings surface

Settings → Guardian card (alongside usage):

- Awakening state + last check time
- RAG corpus: field guide / platform chunk counts
- Button: **Awaken now** → POST warmup
- Link: `make guardian-laptop-tune` doc anchor
- Admin: pull model (existing)

---

## WS9 — Stack scripts

`check-local-stack.sh`:

- `curl -sf http://127.0.0.1:11434/api/tags` → ok/warn
- If `AI_ENABLED=true` in `.env` and Ollama down → warn with `systemctl start ollama` hint

`restart-local.sh --serve`:

- After API up, optional curl `/v1/chat/health` if JWT smoke token available (defer) OR print "open Guardian to awaken"

---

## WS10 — Lite degrade

| Condition | UI |
|-----------|-----|
| `ai_enabled=false` | Hide awakening; "Lite mode — Pi and dashboard only" |
| Ollama unreachable | Actionable banner; no spinner > 30s |
| No grounded models | Farm counsel card disabled; link to pull phi3 |
| `rag_corpus_ok=false` | Amber on Farm counsel: "Field memories not ingested — run bootstrap" |

---

## WS11 — Docs & tests

**Docs:** Replace ritual sections in `local-operator-bootstrap.md` and `guardian-ollama-laptop-playbook.md` with:

```bash
make guardian-laptop-tune ARGS="--apply"   # once per machine
make restart-local-serve
# Login → Guardian awakens automatically
```

**Tests:**

| Area | Test |
|------|------|
| Health awakening fields | handler unit |
| Warmup idempotent | handler integration (mock Ollama) |
| Mode cards | vitest |
| Awakening poll | vitest mock health |
| Morning CTA pre-warm | vitest |

---

## Non-goals (Phase 129)

- Auto `.env` on every API boot
- GPU farm-wide recommendation engine
- Cloud LLM routing
- `make guardian-eval` in awakening path
- Full in-turn SSE phases (**Phase 130**)
- Auto embed unload on chat send (**Phase 130**)

---

## Acceptance (laptop profile)

1. Reboot → `make restart-local-serve` → login → Guardian badge goes amber → green ≤ 5 min (phi3 preload).
2. Farm counsel → Morning check starter → `walk_farm` in logs → completes without manual `ollama stop` (with Phase 130 timeout).
3. Quick chat works without waiting for phi3.
4. Settings shows corpus counts; warns if field guides = 0.
5. `make guardian-laptop-tune` prints fix when `LLM_TIMEOUT_SECONDS=777`.
6. Phase 128 manual checklist passable after 129+130.

---

## Implementation order

| Order | WS | Rationale |
|-------|-----|-----------|
| 1 | WS3 | Immediate timeout relief |
| 2 | WS0 + WS2 | Health + warmup backend |
| 3 | WS1 | Warmup endpoint |
| 4 | WS5 + WS4 | Mode cards + awakening UI |
| 5 | WS6 + WS7 | Login badge + morning CTA |
| 6 | WS8 + WS9 + WS10 | Settings, stack, degrade |
| 7 | WS11 | Docs + tests |

**Ship 129 WS3+WS0+WS2 first** for your laptop validation; UI can follow in same PR or fast follow.
