# Guardian & real grows — readiness before you wire live plants

**Audience:** You — hooking gr33n to a **real** room, bench, or greenhouse for the first time. Demo seed is fun; live crops don't forgive wrong EC, wrong photoperiod, or a pump that runs because chat sounded confident.

**Companion docs:**
- [`farm-guardian-architecture.md`](farm-guardian-architecture.md) §8 — copilot vs actor vs automation
- [`guardian-change-requests-guide.md`](guardian-change-requests-guide.md) — Confirm workflow
- [`recommended-hardware-and-sizing.md`](recommended-hardware-and-sizing.md) — 8B vs 70B, GPU sizing
- [`local-operator-bootstrap.md`](local-operator-bootstrap.md) — dev stack + ingest commands
- [Phase 82](plans/archive/phase_82_guardian_crop_grounding_hardening.plan.md) (crop library + zero-chunk guardrails) · [Phase 83](plans/archive/phase_83_enterprise_agronomy_seed_pack.plan.md) (**shipped** — bootstrap + overrides) · [`phase-83-closure.md`](plans/phase-83-closure.md)

---

## The short version

| Risk | What protects you |
|------|-------------------|
| Guardian turns on a pump / light / dose | **Confirm** on every write — nothing actuates from chat alone |
| Wrong EC or feed advice | Structured **crop profiles** + ingested **field guides** (Phase 82/83) — not raw LLM memory |
| Chat cites docs that don't exist | Phase 82 **zero-chunk guardrail** — no fake `[1]` when RAG is empty |
| Automation runs while you're asleep | **Your rules** — review setpoints, test on bench, use `manual` policy until trusted |
| Small model hallucinates on edge cases | Run **readiness smokes on 8B + seed data** first; upgrade to 70B for prose quality, not safety of numbers |

**Guardian is a copilot, not autopilot.** Your automation rules are the always-on layer. Guardian proposes; you Confirm.

---

## Three knowledge layers (what to configure)

Guardian combines four inputs on a grounded turn. Only one is automatic:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Ollama weights     General plant intuition (you pick 8B/70B) │
│ 2. RAG chunks         Field guides + your cycle/task notes       │
│ 3. crop_profiles      Exact EC/pH/VPD/DLI targets (DB tools)     │
│ 4. Live snapshot      Zones, active cycles, alerts (automatic)   │
└─────────────────────────────────────────────────────────────┘
```

| Layer | Who curates | You must… |
|-------|-------------|-----------|
| **Weights** | Meta / Ollama | `ollama pull` + set `LLM_MODEL`; see [Ollama setup](farm-guardian-ollama-setup.md) |
| **RAG** | Platform guides + **your** operational notes | Run ingest — **`make rag-ingest-field-guides`**, **`make rag-ingest-platform-docs`**, operational ingest (Phase 83 bootstrap) |
| **Profiles** | Platform seed (Phase 82 expands to ≥46 crops) | Assign the right crop to cycles/plants; optional farm overrides — **Settings → Crops & targets** or YAML (Phase 83) |
| **Snapshot** | Your live DB | Keep zone names, cycles, and sensors accurate |

**Today:** Platform ships **≥46** built-in crop profiles when Phase 82/84 migrations are applied; field guides and RAG chunks still require ingest. Treat numeric feed/light advice as **unverified** until bootstrap passes.

**After bootstrap (Phase 83):** run **`make guardian-bootstrap-farm FARM_ID=N`** and pass readiness smokes before trusting crop Q&A on live plants.

---

## Checklist — before live actuators touch water or lights

Use this order. Do not skip **bench** steps because the UI looks good.

### A. Infrastructure

- [ ] Postgres + API healthy (`make check-stack` or `GET /health`)
- [ ] Pi / edge reaches API; **`pending_command`** or **`device_commands`** queue tested on **bench** (no plants) — [Pi integration §9](pi-integration-guide.md)
- [ ] Ollama reachable from API host if using Full mode (`LLM_BASE_URL` probe passes on API start)
- [ ] `EMBEDDING_API_KEY` set if you want doc-backed answers (optional but strongly recommended for crop science)

### B. Knowledge (trust the advice)

- [ ] **`make guardian-bootstrap-farm FARM_ID=N`** — field guides + platform docs + operational domains (Phase 83)
- [ ] Or manually: **`make rag-ingest-field-guides`**, **`make rag-ingest-platform-docs`**, operational ingest
- [ ] Ask a test question: *"Compare cannabis and eggplant EC targets in mS/cm"* — metadata should show **chunks > 0** and/or tool block with **mS/cm** (never `% EC`)
- [ ] Ask: *"How should I feed ramps?"* — should refuse or redirect (unsupported crop), not invent 12/12 cannabis schedule

### C. Safety (trust the actuators)

- [ ] New automation rules start **`is_active: false`** — review in UI, enable one at a time
- [ ] Zone **`automation_policy`**: consider **`manual`** until sensors and setpoints are validated
- [ ] Guardian **Confirm** tested on harmless action first (ack alert, create task) before enqueue pump/light
- [ ] Physical **E-stop** / manual override for pumps and mains — software Confirm is not a substitute

### D. Model sizing (trust the synthesis)

- [ ] Dev/laptop: **`llama3.1:8b`** OK for **read** tools + structured targets if seed data is excellent
- [ ] Production agronomy prose: **≥14B**, docs recommend **70B Q4 on 24 GB VRAM** — [sizing profiles](recommended-hardware-and-sizing.md)
- [ ] Same readiness smokes on 8B first; upgrading model improves wording, not the pipeline

---

## What "grounded" means (and doesn't)

Chat metadata **`grounded · N chunks`** breaks down as:

| Label | Meaning |
|-------|---------|
| **grounded** | Live **farm snapshot** attached (zones, cycles, alerts) |
| **N chunks** | RAG found **N** document snippets — may be **0** even when grounded |

**Grounded ≠ "verified agronomy."** You can be grounded with **0 chunks** and still get plausible-sounding wrong EC. Phase 82 fixes UI honesty and handler guardrails; **you** still run ingest.

---

## When to trust chat vs the dashboard

| Question type | Trust first |
|---------------|-------------|
| "What's my EC right now?" | **Sensor reading** on zone / Water tab |
| "What should EC be for late flower?" | **Crop profile** + field guide chunk — after ingest |
| "Turn on the pump for 30s" | **Confirm card** — read duration, zone, actuator name |
| "Run veg lights 18/6" | **Lighting program** UI — Guardian proposes patches only after Confirm |
| "Why did humidity spike?" | Alerts + sensor history — Guardian summarizes, you verify chart |

**Rule of thumb:** if it moves water, power, or chemistry → **Confirm + your eyes**. If it explains → chat is fine after seed + ingest.

---

## Lite mode (no LLM)

Set **`AI_ENABLED=false`**. Farm still runs: sensors, rules, tasks, fertigation queue, dashboard. Guardian drawer shows Lite banner. **Zero risk from chat hallucination** — also zero chat help.

Many operators run **Lite on the Pi edge** and **Full on a LAN GPU box** when ready.

---

## Public demo vs your live farm

Starters cloning the repo see **demo seed** (`gr33n Demo Farm`) — not your genetics, your RO water, or your room layout.

For a **public README / video** that lands well:

1. Show **Confirm** before any actuator write — "this isn't a black box"
2. Show **local Ollama** or Lite mode — "your data stays on the LAN"
3. Show **crop profile numbers** matching chat (after bootstrap) — "structured, not vibes"
4. Show **`make guardian-bootstrap-farm`** + **Settings → Crops & targets** — enterprise bring-up is scripted

That honesty reads as **engineering**, not apology — and matches what AGPL operators expect.

---

## Roadmap tie-in

| Phase | What it gives your live grow |
|-------|------------------------------|
| **82** | ≥46 crop profiles, field guides, zero-chunk guardrails, multi-crop lookup, plant context bundle |
| **83** ✅ | `guardian-bootstrap-farm`, commons seed pack, farm EC overrides (UI + YAML), scheduled ingest, readiness smokes — [`phase-83-closure.md`](plans/phase-83-closure.md) |

**Suggested order for you:** migrate + parity check → **`make guardian-bootstrap-farm FARM_ID=N`** → smokes on 8B → optional EC overrides → wire actuators on bench → live plants.

---

## Related

| Doc | Topic |
|-----|--------|
| [operator tour §6](operator-tour.md#6-farm-guardian-change-requests-with-your-ok) | Guardian Confirm tour |
| [farm-guardian-ollama-setup.md](farm-guardian-ollama-setup.md) | Inference host |
| [rag-scope-and-threat-model.md](rag-scope-and-threat-model.md) | What gets embedded |
| [operator-troubleshooting.md](operator-troubleshooting.md) | When things go wrong |
| [phase-14 operator index](phase-14-operator-documentation.md) | All phase plans |
