---
name: Phase 137 — Guardian counsel integration (nudges, vision, offline field)
overview: >
  Wire proactive nudges, vision uploads, and offline field mode into the Farm counsel
  / awakening story. Tap nudge → Farm counsel + warmup + context_ref. Photo questions
  declare vision warm-up cost. LLM down still offers procedures.
todos:
  - id: ws1-nudge-to-counsel
    content: "WS1: Nudge Review tap — set farm counsel on, POST warmup farm_counsel, open drawer with context_ref alert_id/category"
    status: pending
  - id: ws2-nudge-awakening
    content: "WS2: If nudge category critical — badge stirring until counsel ready (129 readiness store)"
    status: pending
  - id: ws3-vision-warmup
    content: "WS3: Health + awakening — vision_model_loaded; Farm counsel card note: zone photos may load vision model; warmup optional scope vision"
    status: pending
  - id: ws4-offline-field-banner
    content: "WS4: When llm_reachable=false — Lite banner in Guardian: procedures + static checklists work; procedure starters visible"
    status: pending
  - id: ws5-session-memory-doc
    content: "WS5: Mode card footnote — session memory is keyword tags only, not semantic; link Phase 63"
    status: pending
  - id: ws6-tests
    content: "WS6: vitest nudge→counsel; health vision flag; offline banner when LLM unreachable mock"
    status: pending
isProject: false
---

# Phase 137 — Guardian counsel integration

**Status:** planned · **Depends on:** [129](phase_129_guardian_awakening.plan.md), [Phase 61](phase_61_guardian_proactive_nudges.plan.md), [Phase 37](phase_37_guardian_offline_field_assistant.plan.md)

---

## WS1 — Nudge → Farm counsel

Today: nudge opens drawer with prefilled prompt.

After:

1. `useFarmContext = true` (Farm counsel)
2. `guardianReadiness.warmup('farm_counsel')`
3. `context_ref` from nudge payload (`alert_id`, `nudge_category`)
4. Starter message from nudge template

---

## WS3 — Vision contention

Extend health:

```json
"vision_model": "llava" ,
"vision_model_loaded": false
```

Farm counsel card:

> Zone photos use a separate vision model — first photo question may take extra time on CPU.

Warmup API optional `include_vision: true` (129 WS1 body extension).

---

## WS4 — Offline field mode

When `field_assistant.llm_reachable === false` && `procedures_available`:

```
The Guardian's voice is resting (Ollama unreachable).
Guided procedures and checklists still work offline.
```

Show procedure starters (`start procedure …`) — Phase 37.

---

## Acceptance

- [ ] Tap humidity nudge → Farm counsel on + warming + alert context
- [ ] LLM stopped → procedure starter still sendable (degraded path)
- [ ] Vision model status visible in Settings readiness

---

## Non-goals

- Push notifications for nudges
- Voice TTS changes (Phase 67 stays as-is)
