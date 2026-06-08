---
name: Phase 37 — Guardian offline field assistant & trades guidance
overview: >
  Make Guardian genuinely useful to a non-IT operator at a remote site with no
  internet: walk them step-by-step through wiring the Pi (GPIO, relays, sensors),
  basic plumbing/irrigation hookups, and other hands-on trades work — fully offline
  on a local model + local doc corpus. Adds guided, confirm-per-step procedures, a
  curated field/trades knowledge pack for RAG, hard safety gating for mains-electrical
  and pressurized-water work, printable offline procedure cards, and **background
  Guardian chat** (streaming survives drawer close and route changes via a Pinia chat
  store — critical on slow local LLMs). Builds on existing Ollama offline inference,
  the Pi offline queue, and Phase 32 platform_doc RAG.
todos:
  - id: ws1-offline-readiness
    content: "WS1: Offline readiness — verify/instrument local-model + local-RAG path with no WAN; offline self-check (model reachable, corpus present); degrade gracefully"
    status: done
  - id: ws2-field-corpus
    content: "WS2: Field/trades corpus — authored guides (Pi wiring, relay/sensor hookup, irrigation/plumbing basics, electrical safety) + new RAG source_type field_guide; ingest + manifest"
    status: done
  - id: ws3-guided-procedures
    content: "WS3: Guided procedures — structured step-by-step playbooks (YAML) Guardian drives interactively: one step at a time, operator confirms/【needs help】, resume mid-procedure"
    status: done
  - id: ws4-safety-gating
    content: "WS4: Safety gating — classify steps (safe / caution / qualified-person-required); hard stop + escalate language for mains AC and pressurized water; never instruct unsafe work"
    status: done
  - id: ws5-diagnostics
    content: "WS5: Field diagnostics — guided 'sensor reads nothing / actuator won't fire' wiring + config troubleshooting using snapshot + procedure refs (no internet). Phase 65 adds summarize_device_health so Guardian reads platform GPIO/channel without operator narration."
    status: done
  - id: ws6-printable-cards
    content: "WS6: Printable/offline cards — export a procedure to a printable checklist (PDF/markdown) so a worker can follow it with the screen off or no device at the rig"
    status: done
  - id: ws7-guardian-wiring
    content: "WS7: Guardian wiring — field_guide RAG layer + procedure tools into prompt; persona: hands-on installer voice; cite procedure + step number"
    status: done
  - id: ws9-background-chat-store
    content: "WS9: Background Guardian chat — Pinia chat store owns SSE stream + transcript; navigate/close drawer without killing in-flight turns; global 'still thinking' indicator"
    status: done
  - id: ws8-docs-tests
    content: "WS8: Docs + tests — operator-tour 'first install with Guardian offline'; OpenAPI procedure endpoints; smokes for offline answer + procedure step flow + safety stop; Vitest WS9 background stream"
    status: done
isProject: false
---

# Phase 37 — Guardian offline field assistant & trades guidance

## Status

**Shipped (2026-06-03).** WS9 background chat; WS1 health + LLM-down degrade; WS2 field_guide corpus; WS3–WS6 procedures + print; WS4 safety; WS5 diagnostics YAMLs; WS7 field prompt + procedure chat; WS8 docs/smokes/OpenAPI. **OC-37E closed** (README + phase-14 + persona; run `make rag-ingest-platform-docs` / `make rag-ingest-field-guides` on each farm after doc/corpus edits). **v1 deferred:** WS5 live-snapshot branching; WS6 PDF print; automated WAN-blocked CI.

Depends on **Phase 27/29** (Guardian AI + agent layer), **Phase 32 WS8** (platform_doc RAG ingest), and existing **offline inference** (Ollama) + **Pi offline queue**. Phase 34 operator-stated facts shipped.

**Preconditions (exist today):**

- Offline inference: [`farm-guardian-ollama-setup.md`](../farm-guardian-ollama-setup.md), [`offline-or-intranet-deployment.md`](../offline-or-intranet-deployment.md) (`LLM_BASE_URL` / `EMBEDDING_BASE_URL` on LAN/loopback)
- Platform doc RAG + ingest: [`internal/rag/ingest/platform_docs.go`](../../internal/rag/ingest/platform_docs.go), [`docs/rag/platform-doc-manifest.yaml`](../rag/platform-doc-manifest.yaml), [`scripts/rag-ingest-platform-docs.sh`](../../scripts/rag-ingest-platform-docs.sh)
- Pi wiring reference: [`pi-integration-guide.md`](../pi-integration-guide.md), [`raspberry-pi-and-deployment-topology.md`](../raspberry-pi-and-deployment-topology.md), Pi offline SQLite queue ([`pi_client/gr33n_client.py`](../../pi_client/gr33n_client.py))
- Guardian RAG synthesis + grounding: [`internal/rag/synthesis/guardian.go`](../../internal/rag/synthesis/guardian.go), chat handler ([`internal/handler/chat/handler.go`](../../internal/handler/chat/handler.go))

**Today (gap):** Guardian can *cite* platform how-to docs online, but it cannot **walk a non-technical person through a physical install one step at a time**, has **no trades/plumbing/electrical knowledge pack**, no **safety gating** for dangerous steps, and **offline behavior for the field worker is unverified/undocumented** as a first-class mode.

**UX gap (WS9):** Chat state lives in `GuardianChatPanel.vue` component refs. Closing the slide-out drawer or navigating to another route **unmounts the panel** and drops the streaming UI (and often the partial reply). On a laptop running API + Ollama, slow turns make this painful. **Fix:** move stream + transcript into a **Pinia store** so the HTTP/SSE lifecycle is not tied to the route.

---

## Why this phase

The real deployment story: a grow site is **remote**, the **operator is not an IT person**, and there may be **no internet**. They need to physically wire a Pi to relays and sensors, hook up irrigation/plumbing, and recover when something doesn't read. Guardian should be the patient on-site guide — like a knowledgeable friend on the phone, except it works with the WAN unplugged.

| Today | After Phase 37 |
|-------|----------------|
| Cites docs (assumes reader is technical) | **Guides** a non-IT worker step-by-step, one action at a time |
| How-to only for gr33n software | Adds **trades knowledge**: Pi GPIO/relay/sensor wiring, irrigation/plumbing basics, electrical safety |
| Offline inference exists but unframed for field worker | **Offline-first field mode** with self-check + graceful degrade |
| No safety guardrails on physical steps | **Safety tiers**: hard stop + "get a qualified person" for mains AC / pressurized water |
| Answer = wall of text | **Procedure** = confirmable steps, resumable, printable |
| Chat dies when you leave the page | **Background stream** — turn keeps running; badge when drawer closed; reopen to see tokens |

**This is the offline counterpart to Phase 34:** Phase 34 lets the operator tell Guardian what it can't sense; Phase 37 lets Guardian tell the operator exactly what to do with their hands — and both work when Guardian is blind to hardware.

---

## Design principles

1. **Offline-first, not offline-maybe** — every field-assistant feature must work with **no WAN**: local model (`LLM_BASE_URL`), local embeddings, local corpus in Postgres. If the model is unreachable, Guardian still serves the **static procedure** + printable card.
2. **One step at a time** — guided procedures present a single step, wait for **done / didn't work / help**, then advance. No 12-step wall of text dumped at a stressed worker.
3. **Safety over completeness** — any step touching **mains AC, line-voltage, or pressurized/potable water** is gated. Guardian gives **low-voltage DC** wiring guidance (the Pi side) but **stops and escalates** for hazardous work. It never coaches an unsafe shortcut.
4. **Knowledge is curated, not hallucinated** — trades guidance comes from an **authored, reviewed `field_guide` corpus**, not the model's open weights. Procedures cite the guide + step.
5. **Reuse the rails** — new RAG `source_type='field_guide'` rides the Phase 32 ingest; procedures are data (YAML) + a thin driver; no new model stack.
6. **Honesty about blindness** — Guardian states it can't see the wiring; it asks the worker to describe/confirm, and labels operator observations (ties into Phase 34 `operator_provided`).
7. **No autonomous writes** — if a procedure ends in a config change (e.g. register the actuator), that's still a Confirm-gated proposal.
8. **Chat survives navigation** — streaming and transcript state live in a **Vue/Pinia store**, not in a mount-bound component. Closing the drawer or opening Zones must not abort an in-flight Guardian turn unless the operator explicitly cancels or sends a new message.

---

## Architecture

```
Remote site, no internet. Local box runs Postgres + API + UI + Ollama (LLM + embeddings).

Worker: "the temp sensor isn't showing any reading"
   └─► /v1/chat (farm scoped, OFFLINE)
        ├─ offline self-check: LLM_BASE_URL reachable? corpus present? → field mode
        ├─ RAG retrieve: field_guide chunks (sensor wiring) + platform_doc (Pi client)
        ├─ snapshot: device online? last_heartbeat? sensor row exists?
        └─ start guided procedure "diagnose-sensor-no-reading"
              Step 1/6: "Is the Pi powered (green LED on)?" → [yes]
              Step 2/6: "Find the sensor's 3 wires: red=3.3V, black=GND, yellow=data…" → [done]
              Step 3/6 [CAUTION]: "Check the data wire goes to GPIO 4 (physical pin 7)…"
              ...
              Step 6/6: resolved OR escalate ("if still nothing, the sensor may be faulty")

Model unreachable? → Guardian serves the static procedure text + 'Print checklist' (WS6).
Mains-AC step? → SAFETY STOP: "This needs a qualified electrician — do not proceed." (WS4)
```

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Offline readiness | offline self-check in chat handler; health endpoint flag; degrade path |
| **WS2** | Field/trades corpus | `docs/field-guides/*.md` (authored); `source_type='field_guide'`; manifest + ingest |
| **WS3** | Guided procedures | `docs/field-guides/procedures/*.yaml`; `internal/farmguardian/procedures/` driver |
| **WS4** | Safety gating | step `safety_tier`; hard-stop language; persona safety rules |
| **WS5** | Field diagnostics | snapshot-aware troubleshooting procedures |
| **WS6** | Printable cards | export procedure → markdown/PDF checklist endpoint + UI button |
| **WS7** | Guardian wiring | RAG layer + procedure tool in prompt; persona installer voice |
| **WS9** | Background chat store | Pinia `guardianChat` owns SSE + transcript; drawer/route safe; thinking indicator |
| **WS8** | Docs + tests | operator-tour offline install; OpenAPI; smokes; Vitest WS9 |

---

## Work-stream detail

### WS1 — Offline readiness & self-check

**Goal:** Field mode that is provably WAN-independent and degrades gracefully.

**Tasks:**

1. Startup/health: report whether `LLM_BASE_URL` + `EMBEDDING_BASE_URL` are LAN/loopback and reachable; expose in `/automation/worker/health`-style or a new `/v1/chat/health`.
2. Chat handler: detect "offline field mode" (no WAN / local endpoints) and prefer `field_guide` + `platform_doc` corpus; never block on an external call.
3. Graceful degrade: if the local model is down, return the **static procedure** content + printable card link instead of an error.
4. Document the exact env for a single-box field rig (Pi/NUC running API+UI+Ollama+Postgres).

**Acceptance:** With WAN blocked and local Ollama up, a field question returns a grounded guided answer; with the model also down, Guardian still returns the static procedure + print link (no hard failure).

### WS2 — Field / trades knowledge corpus

**Goal:** Curated, reviewed knowledge Guardian can ground on — not open-weight guesses.

**Tasks:**

1. Author `docs/field-guides/`:
   - `pi-wiring-basics.md` — power, GPIO pinout, common relay board (IN/VCC/GND), wiring a sensor (3-wire/I²C), grounding, polarity.
   - `relay-and-actuator-wiring.md` — low-voltage control side vs the **switched** load side (with safety boundary to WS4).
   - `sensor-install-and-calibration.md` — placement, EC/pH probe handling, calibration steps.
   - `irrigation-and-plumbing-basics.md` — tubing/fittings, drip vs flood, pump + reservoir, RO/well source notes, leak checks (non-pressurized guidance; pressurized/potable → escalate).
   - `electrical-safety.md` — DC vs AC, what a non-electrician may/may not do, lockout, when to call a pro.
   - `field-troubleshooting.md` — symptom → likely cause table.
2. New RAG `source_type='field_guide'`; extend `platform_docs.go` ingest (or sibling) + add a `field-guide-manifest.yaml`.
3. Each guide chunk carries metadata: `safety_tier`, `domain` (electrical/plumbing/sensor/pi), `requires_tools`.

**Acceptance:** Ingest loads field guides as `field_guide` chunks; retrieval on "how do I wire the relay" returns the relay guide with `safety_tier` metadata.

### WS3 — Guided procedures (confirm-per-step)

**Goal:** Interactive, resumable, one-step-at-a-time walkthroughs.

**Tasks:**

1. Procedure schema `docs/field-guides/procedures/*.yaml`:
   ```yaml
   id: wire-pi-relay-light
   title: Wire a Pi to a relay for a grow light
   domain: pi
   offline_ok: true
   steps:
     - n: 1
       safety_tier: safe
       say: "Unplug the grow light from the wall before touching anything."
       confirm: "Light is unplugged?"
     - n: 2
       safety_tier: caution
       say: "Connect relay IN to GPIO 17 (physical pin 11), VCC to 5V (pin 2), GND to GND (pin 6)."
       confirm: "All three control wires connected?"
       ref: relay-and-actuator-wiring.md#control-side
     - n: 3
       safety_tier: qualified_person_required
       say: "Wiring the light's AC mains to the relay's switched side is line-voltage work."
       stop_unless_qualified: true
   ```
2. Driver in `internal/farmguardian/procedures/`: load, present step `n`, accept `done` / `failed` / `help`, advance or branch; persist progress (session-scoped, resumable after a break) in proposal/session meta.
3. Procedures reference `field_guide` sections for the "why/visual."
4. Read-only by default; a terminating "register this actuator in gr33n" step emits a normal Confirm-gated proposal (reuse Phase 32 tools), not a silent write.

**Acceptance:** Starting `wire-pi-relay-light` yields step 1 only; confirming advances to step 2; a `qualified_person_required` step halts (see WS4); progress resumes after an unrelated turn.

### WS4 — Safety gating

**Goal:** Guardian never coaches dangerous work; it escalates.

**Tasks:**

1. `safety_tier` per step/guide: `safe` (low-voltage DC, dry, hand-tight), `caution` (verify power off, double-check polarity), `qualified_person_required` (mains AC, line voltage, pressurized/potable water, gas).
2. On `qualified_person_required`: Guardian **stops the procedure**, states plainly why, and advises a licensed electrician/plumber — even when asked to continue.
3. Persona/system rules (offline too): never give step-by-step mains wiring or pressurized-plumbing instructions; DC control side + "what to ask the pro" only.
4. Safety lines are part of the **authored corpus** so they're consistent and reviewable, not model-improvised.

**Acceptance:** A request to "just tell me how to wire the 120V to the relay" returns the safety stop + escalation, not instructions; smoke asserts the stop.

### WS5 — Field diagnostics

**Goal:** Recover common field failures with no internet.

**Tasks:**

1. Diagnostic procedures: `diagnose-sensor-no-reading`, `diagnose-actuator-wont-fire`, `diagnose-pi-offline`.
2. Use the live snapshot (device `last_heartbeat`, sensor exists, actuator `pending_command` stuck) + `field_guide` wiring checks to branch steps.
3. Tie into Pi realities: offline queue backlog, API key mismatch (401), wrong `farm_id`, GPIO pin mismatch.

**Acceptance:** "actuator won't turn on" walks: command queued? Pi online? relay wired to the right GPIO? — each step grounded; ends in resolution or escalation.

### WS6 — Printable / offline procedure cards

**Goal:** A worker can follow steps at the rig with no screen handy.

**Tasks:**

1. Export endpoint: `GET /field-guides/procedures/{id}/print` → printable markdown/PDF checklist (steps, safety call-outs, pinout).
2. UI "Print checklist" button on a procedure; bundle a few core cards into the offline deployment so they exist even if the model is down.
3. Include a simple pinout diagram reference (static asset) in the wiring cards.

**Acceptance:** Print export renders all steps + safety tiers; available even when the LLM endpoint is offline (static render path).

### WS7 — Guardian wiring (prompt + persona)

**Tasks:**

1. Add `field_guide` to retrieval; extend `GuardianRAGInstructions` so field/install/troubleshooting questions prefer `field_guide` + `platform_doc`, and Guardian offers to **start a guided procedure**.
2. Procedure-control as Guardian capability: start / next / repeat / stop a procedure from chat.
3. Persona update: patient on-site installer voice for a non-IT worker; plain language, no jargon without explaining it; always surface safety tier; cite `procedure#step` / `field_guide#section`.

**Acceptance:** "help me wire the Pi to a light" offers the procedure and runs it step-by-step with citations; persona/platform docs mirror the new capability + safety rules.

### WS9 — Background Guardian chat (Pinia store)

**Goal:** Slow local LLMs (70B on a laptop, field NUC) must not lose work when the operator checks Zones, actuators, or automation while Guardian is still streaming.

**Why a store (not a “small fix”):** Today `send()` / `consumeSSE()` / `streamingText` / `transcript` live in [`GuardianChatPanel.vue`](../../ui/src/components/GuardianChatPanel.vue). The drawer uses `v-if` on the panel ([`GuardianDrawer.vue`](../../ui/src/components/GuardianDrawer.vue)); `/chat` mounts a second panel instance. Unmount ends the UI loop; there is no shared session. **A Pinia store is the clean cut** — one stream owner, many views.

**Tasks:**

1. **`ui/src/stores/guardianChat.js`** (or `.ts` if the UI migrates):
   - State: `activeSessionId`, `streaming`, `streamingText`, `error`, `transcriptBySession` (or server-synced session list + current transcript), `pendingUserMessage`, `lastFarmId`.
   - Actions: `sendMessage({ message, farmId, contextRef, attachmentIds })` — owns `fetch` + `ReadableStream` reader; **`AbortController` only** on explicit `cancelStream()` or new `sendMessage` (not on component unmount).
   - On `done` event: append turn to transcript, `refreshSessions()`, update `guardianProposals` pending count if proposals returned.
2. **Refactor `GuardianChatPanel.vue`** to a thin view: `storeToRefs(guardianChat)` + bind inputs; no local `streaming` / `transcript` refs for the active turn.
3. **Single store across surfaces:** drawer compact layout and `/chat` full layout both use the same store (no duplicate streams). Consider `KeepAlive` on `/chat` only if needed after store lands — store-first.
4. **Global indicator:** extend [`guardianPanel`](../../ui/src/stores/guardianPanel.js) or TopBar / drawer tab — when `guardianChat.streaming`, show **“Guardian thinking…”** on the robot tab even if drawer is closed; click reopens to live `streamingText`.
5. **Drawer lifecycle:** change drawer body to **`v-show`** (or keep panel mounted while `streaming`) so closing backdrop does not destroy the stream consumer; if drawer stays `v-if`, store must still hold stream (required either way).
6. **Operator constraints:** while `streaming`, disable second Send on same session (or queue — v1: disable). Other farm pages remain usable (read-only navigation OK).
7. **Vitest:** mount panel → start send (mock SSE) → unmount panel → assert store still `streaming` and deltas append → remount → assert `streamingText` visible → done event completes transcript.

**Acceptance:** Start a long streamed reply from the drawer, navigate to **Zones → Greenhouse**, confirm tokens still accumulate in the store and reopening the drawer shows the live partial answer; completed turn appears in session history. Explicit Cancel aborts fetch. New message aborts previous stream.

**Not in WS9 v1:** second concurrent chat per farm; background Confirm execution; push notifications when done.

### WS8 — Docs + tests

**Tasks:**

- `operator-tour.md` — "First install with Guardian, offline" (single-box rig, start wiring procedure, hit a safety stop, finish, register actuator via Confirm).
- `offline-or-intranet-deployment.md` — add field-assistant mode + which corpus/procedures ship.
- `farm-guardian-architecture.md` — knowledge layers gain **field_guide**; add §7.x guided procedures + safety tiers; phase ledger.
- OpenAPI: procedure list/step/print endpoints + `field_guide` source type.
- Smokes: offline grounded answer; procedure step advance; safety stop; printable export with model down.
- **WS9:** `guardian-chat-background.test.js` — stream survives unmount; thinking badge when drawer closed (per WS9 acceptance).

**Acceptance:** Docs added to phase-14 index + manifest; `go test` + Vitest green; offline + safety paths asserted; WS9 Vitest green.

---

## Knowledge layers after this phase (for architecture doc)

| Layer | Source | Used for |
|-------|--------|----------|
| Live snapshot | DB now | "right now" farm state |
| Operational RAG | farm tasks/cycles/alerts | farm-specific recall |
| platform_doc RAG | curated operator docs (Phase 32) | gr33n how-to / troubleshooting |
| **field_guide RAG** (new) | authored trades/wiring/plumbing guides | **physical install + repair, offline** |
| Guided procedures (new) | YAML playbooks | **step-by-step hands-on walkthroughs** |
| Operator-stated facts (Phase 34) | the worker | things no sensor can see |

---

## Out of scope (this phase)

- Step-by-step **mains AC** or **pressurized/potable plumbing** instructions (escalate to a pro — by design)
- Computer-vision "look at my wiring" photo analysis (future; worker still describes/confirms)
- Region-specific electrical/plumbing **code compliance** (point to local code + licensed trades)
- Auto-ordering parts / BOM generation (future)

---

## Recommended order

**WS9 (background chat store) can ship in parallel** with WS1 — frontend-only, no field corpus dependency; high value on slow laptops running the full stack.

WS1 (offline proof) → WS2 (corpus) → WS4 (safety, before any step content ships) → WS3 (procedures) → WS5 (diagnostics) → WS6 (print) → WS7 (Guardian wiring) → **WS9** (if not already merged) → WS8 (docs/tests). WS4 intentionally precedes WS3 content authoring.

---

## Definition of done (phase ship)

- [x] Field-assistant features verified on LAN; graceful degrade when model is down (`go test` smokes + manual WAN-off walkthrough per operator-tour §6d)
- [x] `field_guide` corpus authored; ingest via `make rag-ingest-field-guides` (retrieval + safety metadata — run on farm after deploy)
- [x] Guided procedures run one step at a time, resumable, with citations
- [x] Safety gating hard-stops mains-AC / pressurized-water steps and escalates
- [x] Field diagnostics YAMLs (sensor / actuator / Pi-offline) — static branches; live snapshot branching deferred
- [x] Printable procedure cards — markdown via `GET …/print` (PDF deferred)
- [x] Persona = non-IT installer voice; no autonomous writes (terminal config = Confirm)
- [x] **Background chat:** Pinia store owns SSE stream; drawer/route change does not kill in-flight turns; thinking indicator when drawer closed
- [x] operator-tour + offline-deployment + architecture + OpenAPI + tests updated

---

## Using this plan in a new chat

> Implement Phase 37 from `docs/plans/phase_37_guardian_offline_field_assistant.plan.md`. **WS9:** move Guardian SSE + transcript into a Pinia store (`guardianChat`) so closing the drawer or changing routes does not abort slow streams — refactor `GuardianChatPanel` to read the store. Start WS1 offline self-check + WS2 field_guide corpus (reuse Phase 32 platform_docs ingest with a new source_type). Author WS4 safety tiers before WS3 procedure content. Guided procedures are confirm-per-step and resumable. Guardian must NEVER give step-by-step mains-AC or pressurized-water instructions — it stops and tells the worker to get a qualified person. Everything must work offline on a local model + local corpus; no autonomous writes.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_34_guardian_pr_iteration.plan.md](phase_34_guardian_pr_iteration.plan.md) | Operator-stated facts (the worker is Guardian's eyes) |
| [pi-integration-guide.md](../pi-integration-guide.md) | Pi client, routes, offline queue, GPIO |
| [farm-guardian-ollama-setup.md](../farm-guardian-ollama-setup.md) | Local offline model |
| [offline-or-intranet-deployment.md](../offline-or-intranet-deployment.md) | No-WAN topology (extend with field mode) |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | platform_doc RAG ingest reused for field_guide |
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Wiring a light relay = a Phase 37 procedure |
| [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md) | WS8 + end-of-37 sweep (OC-37, OC-37E) |
| [phase_36_greenhouse_climate.plan.md](phase_36_greenhouse_climate.plan.md) | Wiring shade motor / fan = Phase 37 procedures |
| [phase_33_guardian_polish_and_enterprise_ops.plan.md](phase_33_guardian_polish_and_enterprise_ops.plan.md) | Guardian UX polish lineage; WS9 completes “slow model” ergonomics |
| [`GuardianChatPanel.vue`](../../ui/src/components/GuardianChatPanel.vue) | Current mount-bound stream — WS9 migrates to store |
