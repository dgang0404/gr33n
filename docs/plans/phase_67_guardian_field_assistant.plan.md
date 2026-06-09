---
name: Phase 67 — Guardian hands-free field assistant
overview: >
  The "in the grow room with dirty gloves" assistant — voice in, voice out, and
  crop-profile-grounded photo diagnosis. Local-first to honor the LAN/offline ethos:
  browser speech as baseline, optional local whisper.cpp for fully-offline STT.
  Vision stays advisory ("hypothesis, not diagnosis") and is grounded on Phase 64. Wiring troubleshooting is grounded on Phase 65.
todos:
  - id: ws1-voice-in
    content: "WS1: Push-to-talk mic in Guardian panel — browser SpeechRecognition baseline"
    status: completed
  - id: ws2-voice-out
    content: "WS2: Optional TTS read-back for hands-free answers; toggle in settings"
    status: completed
  - id: ws3-local-stt
    content: "WS3: Optional local whisper.cpp STT service for fully-offline voice (LAN)"
    status: completed
  - id: ws4-photo-discoverability
    content: "WS4: Photo upload everywhere Guardian lives — /chat full page, slide-out, zone picker when context missing"
    status: completed
  - id: ws5-photo-diagnosis
    content: "WS5: Field photo diagnosis grounded on crop profile (Phase 64); deficiency/pest hypotheses"
    status: completed
  - id: ws6-glovebox-ux
    content: "WS6: One-handed mobile layout — big camera + mic targets, works on a phone in the room"
    status: completed
  - id: ws7-docs-tests
    content: "WS7: operator-tour § field assistant + vision discoverability; phase-67-closure; OC-67"
    status: completed
isProject: false
---

# Phase 67 — Guardian hands-free field assistant

## Status

**Shipped.** Push-to-talk mic, optional read-aloud, camera upload on `/chat` + slide-out with zone picker, crop-grounded vision prompts, local STT proxy (`STT_BASE_URL`), glovebox-sized controls.

---

## The one job

> **Standing in the grow room with wet gloves — ask out loud, hear the answer, snap a leaf photo, get a grounded hypothesis.**

---

## WS1 — Voice input

- Push-to-talk mic button in Guardian panel + compact slide-out
- **Baseline:** browser `SpeechRecognition` (works today, no install) — transcribes to the message box, user reviews, sends
- Visual "listening" state; tap to stop

---

## WS2 — Voice output (optional)

- Toggle in `/settings/guardian`: "Read answers aloud"
- Browser `SpeechSynthesis` baseline
- Reads the plain-language answer, not proposal JSON; stops on new input

---

## WS3 — Local STT (optional, fully offline)

- For farms that want **zero cloud speech**: optional local `whisper.cpp` service on the LAN (same spirit as local LLM via LM Studio)
- `STT_PROVIDER=local|browser`; falls back to browser if local unreachable
- Honors the offline-first promise — no audio leaves the LAN

---

## WS4 — Photo upload discoverability (fix today's one-way UX)

**Problem today:** "Zone photos (vision)" only appears when Guardian has **zone context** — i.e. you opened the slide-out from a zone page. On **Guardian full page** (`/chat`) or Dashboard, the upload block is hidden even when `vision_chat_enabled` is true. Farmers shouldn't have to discover a side door.

| Surface | Today | After Phase 67 |
|---------|-------|----------------|
| `/chat` full page | No photo UI | Camera / upload always visible when vision enabled |
| Slide-out (any page) | Photos only if `contextRef.type === 'zone'` | Same + zone picker if context missing |
| Zone detail | Reference photos (separate from chat) | Link: "Ask Guardian about this photo" |

**Deliverables:**
- Prominent **camera button** next to Send (full page + compact panel)
- **Zone picker** dropdown when no zone in context: "Which room is this photo from?" (required for attachment API)
- Optional: attach photo **without** pre-uploading to zone gallery — direct `POST` with message (one-shot field snap)
- Empty-state copy on `/chat`: "Snap a leaf photo — pick the room it came from"
- `v-nav-hint` from zone reference-photos section → Guardian with zone pre-selected

---

## WS5 — Field photo diagnosis (grounded)

- Reuse existing vision path (`vision_context.go`, `visionChatEnabled`)
- **Ground on Phase 64 crop profile:** "This is cannabis in flower — the yellowing between veins on older leaves is consistent with **magnesium** deficiency. Your EC is in range, so check pH lockout (target 5.8–6.2)."
- Cross-reference current sensor + targets so the hypothesis is farm-specific, not generic image-search
- **Stays advisory** — existing disclaimer kept: *"hypotheses, not certified diagnosis. Any change still needs Confirm."*

---

## WS6 — Glovebox UX

- One-handed mobile layout: large mic + camera buttons, thumb-reachable
- High-contrast for bright grow lights / sunlight
- Works on a phone walking the rows (ties to mobile-distribution backlog B4)

---

## WS7 — Docs, tests, OC-67

- operator-tour "Hands-free field assistant" section
- `phase-67-closure.test.js` — mic button renders; TTS toggle; vision disclaimer present
- Accessibility: voice is additive, never the only path (keyboard always works)

---

## Definition of done

- [x] Photo upload visible on `/chat` full page (not only zone slide-out)
- [x] Zone picker when context missing; one-shot camera attach works
- [x] Push-to-talk transcribes into Guardian
- [x] Optional read-aloud answers
- [x] Photo diagnosis cites the crop profile + current readings
- [x] Fully-offline STT path documented
- [x] OC-67 closed

---

## Boundary

- Voice/vision are **additive** — every action still works by typing and tapping
- Vision is **advisory only**; changes still go through Confirm
- No always-listening / wake-word — push-to-talk only (privacy + no ambient recording)
