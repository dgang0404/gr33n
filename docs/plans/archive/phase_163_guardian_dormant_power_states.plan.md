---
name: Phase 163 — Guardian dormant/wake power states
overview: >
  Phase 129 made Guardian auto-awaken on login, but there was no deliberate
  "go to sleep" control — operators running on solar, battery, or metered
  power at remote sites had no way to release the loaded model's RAM/CPU
  draw between sessions without a terminal (`ollama stop`). This phase adds
  an explicit `dormant` awakening state, Rest-now / Wake-now controls,
  optional auto-rest after idle minutes, and admin docs for full Ollama
  service stop. Future WS5 adds hand-drawn druid state artwork (no AI art).
todos:
  - id: ws1-dormant-backend
    content: "WS1: AwakeningStateDormant + RequestDormant() unload + POST /guardian/dormant"
    status: completed
  - id: ws2-settings-rest-button
    content: "WS2: Settings 'Rest now' button mirroring 'Awaken now'; awakening panel dormant copy"
    status: completed
  - id: ws3-auto-idle-dormant
    content: "WS3: GUARDIAN_AUTO_DORMANT_MINUTES + background loop + health idle countdown"
    status: completed
  - id: ws4-power-docs
    content: "WS4: Power-tier docs + scripts/guardian-power.sh admin helper"
    status: completed
  - id: ws5-state-art-hook
    content: "WS5: swappable art slot per Guardian state for hand-drawn (non-AI) druid artwork"
    status: completed
isProject: false
---

# Phase 163 — Guardian dormant/wake power states

**Status:** WS1–WS5 shipped

**Depends on:** [Phase 129](phase_129_guardian_awakening.plan.md)

---

## Shipped

| WS | Deliverable |
|----|-------------|
| **WS1** | `AwakeningStateDormant`, `RequestDormant()`, `POST /guardian/dormant` |
| **WS2** | Settings **Rest now** + **Awaken now**; awakening panel dormant copy |
| **WS3** | `GUARDIAN_AUTO_DORMANT_MINUTES` — background loop + health `idle_until_dormant_sec` + Settings countdown |
| **WS4** | Power-tier docs (`local-operator-bootstrap.md`, `environment-variables.md`) + `scripts/guardian-power.sh` |
| **WS5** | Hand-drawn druid art hook — `ui/public/assets/guardian/druid/` + manifest + `GuardianStateArt.vue` |

---

## Power tiers (operator mental model)

| Tier | Saves | Trigger |
|------|-------|---------|
| **Rest now** | Chat model RAM/VRAM | Settings button → `POST /guardian/dormant` |
| **Auto-rest** | Same, after idle | `GUARDIAN_AUTO_DORMANT_MINUTES=N` in `.env` |
| **Service stop** | Full Ollama process (admin) | `./scripts/guardian-power.sh sleep` — not in web API |

Awakening / chat / **Awaken now** clears dormant and resets the idle clock.

---

## WS3 — Auto-idle dormant ✅

- `GUARDIAN_AUTO_DORMANT_MINUTES` — `0` disables (default); e.g. `45` for solar sites.
- `NoteGuardianActivity(model)` on successful warmup + chat turn completion.
- `MaybeAutoDormant` — skips when busy, stirring, already dormant, or model already cold.
- `StartAutoDormantLoop` — 1-minute ticker in API when AI enabled (works with browser closed).
- Health exposes `auto_dormant_minutes` + `idle_until_dormant_sec` while `ready`.
- `ensureAwake` wakes from `dormant` when user sends a message.

---

## WS4 — Power docs + admin script ✅

- `scripts/guardian-power.sh {sleep|wake|status}` — `systemctl` wrapper for admins.
- Documented sudoers pattern in script header (optional, not shipped as config).
- `local-operator-bootstrap.md` — three-tier power table.

---

## WS5 — Druid artwork hook ✅

Hand-drawn, **non-AI** art per `awakening.state`. Six minimal SVG placeholders (watermarked) ship in `manifest.json`; artists replace per state when ready.

- `ui/src/lib/guardianStateArt.js` — manifest fetch + URL helpers
- `ui/src/components/GuardianStateArt.vue` — image slot (hidden until manifest + load succeed)
- Wired into `GuardianSettingsAwakeningCard` and `GuardianAwakeningPanel`
- Artist brief: `ui/public/assets/guardian/druid/README.md`

---

## Verification

```bash
go test ./internal/farmguardian/... ./internal/handler/chat/... -run 'Dormant|AutoDormant' -count=1
cd ui && npm test -- --run src/__tests__/guardian-settings-awakening.test.js src/__tests__/phase-163-ws5-guardian-state-art.test.js
```

**Try auto-rest locally:** add `GUARDIAN_AUTO_DORMANT_MINUTES=2` to `.env`, restart API, awaken Guardian, wait 2+ min idle — state should flip to `dormant` with auto-rest message.
