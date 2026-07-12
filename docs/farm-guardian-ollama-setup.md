# Farm Guardian — Ollama setup runbook (Phase 27 WS1)

**Audience:** Operators standing up the on-farm inference server for **Farm Guardian** (`POST /v1/chat`) and RAG answer synthesis (`POST /farms/{id}/rag/answer`).

**Companion doc:** If you want to understand **what happens inside** a chat request once Ollama is up (UI → handler → RAG → snapshot → LLM → persistence) and the cost-guard rationale, read **[`farm-guardian-architecture.md`](farm-guardian-architecture.md)** alongside this runbook.

**Scope:** Single on-prem inference host running **Ollama** on the farm intranet, called by the gr33n Go API. This is the **Full mode** path described in [phase_27_farm_guardian_ai_layer.md](plans/phase_27_farm_guardian_ai_layer.md). For the **Lite mode** alternative (no LLM), set **`AI_ENABLED=false`** on the API and skip this whole document.

**Not in scope:** Kubernetes manifests, multi-node inference clusters, GPU pooling. Phase 27 deliberately stays on **Compose + systemd** (see [Phase 26 logging runbook](operator-logging-runbook.md) — same posture).

---

## 1. What you are deploying

```
┌──────────────────────────────┐         ┌────────────────────────────┐
│ gr33n API (Go, cmd/api)      │         │ Inference host             │
│                              │         │                            │
│ AI_ENABLED=true              │  HTTP   │ Ollama (systemd)           │
│ LLM_BASE_URL=…/v1   ────────────────▶  │ /v1/chat/completions       │
│ LLM_MODEL=…                  │         │ /v1/models                 │
│ LLM_API_KEY=  (empty intranet)         │                            │
└──────────────────────────────┘         └────────────────────────────┘
```

- **Ollama** speaks the OpenAI-compatible **`/v1/chat/completions`** endpoint, so the API's existing chat client (`internal/rag/llm`) talks to Ollama and to cloud providers (OpenAI, Mistral, etc.) with **only env-var changes** — no code change between dev laptop and production farm.
- The API performs a **startup probe** against **`GET {LLM_BASE_URL}/models`** when `AI_ENABLED=true` **and** `LLM_BASE_URL` + `LLM_MODEL` are set. If the probe fails, the API process exits with a clear error — no silent degradation.
- Ollama also exposes a native `/api/tags` endpoint that returns the same shape. The Go API uses the OpenAI-compatible `/v1/models` path for portability.

### 1.1 Where Farm Guardian's knowledge comes from

In Full mode the whole conversation stays on the farm intranet — no public-internet hop in the chat path:

```
┌───────────────────────────────── Farm intranet ────────────────────────────────┐
│                                                                                │
│  Pi clients ──HTTPS──▶ Go API (cmd/api)                                        │
│                          │                                                     │
│                          ├──▶ Postgres + pgvector + Timescale (your farm data) │
│                          │                                                     │
│                          └──▶ Ollama  (Llama 3.1 70B Q4, GPU box on intranet)  │
│                                                                                │
│  Browser ──HTTPS──▶ Vue UI ──▶ Go API (same as Pi)                             │
└────────────────────────────────────────────────────────────────────────────────┘
                       (no traffic leaves the LAN in Full mode)
```

The only **outside** calls a Full-mode install ever makes are:

- **First-time setup only** — `ollama pull llama3.1:70b-instruct-q4_K_M` downloads the model weights from `ollama.com`. After that the weights live in `/var/lib/ollama` and nothing in the chat path leaves the LAN.
- **OS / package updates** — outside the platform's responsibility.
- **Opt-in cloud LLM** — if you swap `LLM_BASE_URL` to OpenAI / Anthropic / Mistral, *those* requests leave the LAN. Production Full mode points at the on-farm Ollama and does not.

Farm Guardian's answers are stacked from three knowledge layers at request time:

| Layer | What it is | Source | Updated when |
|-------|------------|--------|---------------|
| **1. General knowledge** | Plant biology, hydroponics, agronomy terms, fertilizer chemistry, etc. | Baked into Llama 3.1 70B's weights at training time. | Frozen at model-release. Static until you pull a new tag. |
| **2. Per-farm RAG corpus** | This farm's `zones`, `sensors`, `schedules`, `rules`, `alerts`, `crop_cycles`, `fertigation_programs`, `automation_events`, `tasks` etc., chunked → embedded → stored in `pgvector`. | Phase 25 `rag-ingest` pipeline. | Whenever you re-run ingest (cron, or after schema changes). |
| **3. Live farm snapshot** | A fresh DB query at request time — zones, active crop cycles, unread-alerts count. | `internal/farmguardian/snapshot.go` (Phase 27 WS4 follow-up). | Every single chat turn. |

So when an operator asks *"why did zone 3 alert this morning?"*:

1. Llama already knows what *alert* / *zone* / *high-temp event* mean in general — **layer 1**.
2. RAG pulls the specific `automation_events` row for *your* zone 3 at 07:14 — **layer 2**.
3. The snapshot tells the model *"here are your 4 zones right now, 2 active cycles, 1 unread alert"* — **layer 3**.

The persona prompt (see `internal/farmguardian/persona.go`) explicitly instructs the model to **only draw on the farm data provided** and to say so when the answer isn't in the context — so it won't invent sensor readings.

### 1.2 Do you need to add agricultural training data?

**For v1, no.** Llama 3.1 70B already has solid baseline agricultural knowledge from its training set, and the per-farm RAG chunks supply the *specific* data this farm needs answered — sensor history, alert reasons, schedule definitions. The persona prompt keeps the model from extrapolating beyond what your DB says.

**Where it would help — future extension:**

- **Static reference corpus** — USDA / university extension docs, IPM diagnosis guides, deficiency-symptom tables, fertilizer compatibility charts, crop phenology calendars. These can be ingested as a **separate** RAG corpus alongside the per-farm DB chunks. The boundary is already documented in [`rag-scope-and-threat-model.md` §9](rag-scope-and-threat-model.md) (Phase 26 WS3): **farm DB RAG** (private, per-farm) vs **education corpus** (shared static reference) vs **operational logs** (never ingested).
- **Fine-tuning Llama on ag data** — heavier lift (training pipeline + curated dataset). Almost certainly overkill until you have real operators using v1 and can see specifically where the model is weak.

Recommendation: ship v1 as-is, see what farmers actually ask, and if you see repeated *"I don't know"* in a specific area (say, IPM diagnoses), curate a focused static reference corpus for that topic and run it through the Phase 25 ingest pipeline. Cheap, targeted, no fine-tuning required.

---

## 2. Hardware minimum (Full mode)

**Full-stack sizing** (API, UI, Postgres, RAG, automation, Pi edge): see **[recommended-hardware-and-sizing.md](recommended-hardware-and-sizing.md)** for deployment profiles and lag expectations.

| Resource | Recommended | Notes |
|----------|-------------|--------|
| GPU | **RTX 3090 (24 GB VRAM)** or equivalent | Required for Llama 3.1 70B Q4. Smaller models (7B/13B Q4) run on 12 GB cards if you accept lower quality. |
| RAM | **64 GB system RAM** | Ollama keeps the model resident; OS + page cache benefits from headroom. |
| Storage | **50 GB free** | Weights are large. Put `/var/lib/ollama` on an SSD. |
| OS | **Ubuntu 22.04 LTS** / **Debian 12** | Tested. Other distros work if NVIDIA drivers are happy. |
| Network | Intranet only | The farm API resolves a **DNS alias** (e.g. `ollama.farm.local`) — don't expose Ollama to the public internet. |

You can run a single smaller box for development (laptop with `ollama` on `localhost:11434`) and a separate dedicated GPU host for production. The API does **not** care — it follows `LLM_BASE_URL`.

---

## 3. Install Ollama on the inference host

```bash
# As an admin user with sudo
curl -fsSL https://ollama.com/install.sh | sh

# Confirm the service unit is installed
systemctl status ollama
```

The Ollama installer ships an **`ollama` systemd unit**. By default it binds to **`127.0.0.1:11434`** — that is not reachable from the API host, so we override it next.

### 3.1 Bind Ollama to the intranet interface

Create a systemd override so other hosts on the farm intranet (and only those) can reach Ollama:

```bash
sudo mkdir -p /etc/systemd/system/ollama.service.d
sudo tee /etc/systemd/system/ollama.service.d/override.conf > /dev/null <<'EOF'
[Service]
# Listen on every interface — restrict at the firewall/router instead.
Environment="OLLAMA_HOST=0.0.0.0:11434"

# Force GPU offload. Remove if you are running CPU-only for dev.
Environment="OLLAMA_NUM_GPU=1"

# Keep one model resident in VRAM for low first-token latency.
Environment="OLLAMA_KEEP_ALIVE=24h"
EOF

sudo systemctl daemon-reload
sudo systemctl restart ollama
sudo systemctl enable ollama
```

**Firewall the port.** Ollama has no authentication of its own — assume *anything that can reach :11434 can use your GPU*. On Ubuntu with `ufw`:

```bash
sudo ufw allow from 10.0.0.0/8 to any port 11434 proto tcp comment "ollama intranet only"
sudo ufw reload
```

Adjust the CIDR for your actual intranet. If you must route through a reverse proxy that adds auth, set **`LLM_API_KEY`** on the API host accordingly — the chat client adds `Authorization: Bearer …` only when that env var is non-empty.

### 3.2 Pull the target model

```bash
# Production model — Phase 27 default
ollama pull llama3.1:70b-instruct-q4_K_M

# Verify GPU offload + first-token speed
ollama run llama3.1:70b-instruct-q4_K_M "ping"
```

If you don't have a 24 GB+ card, use a smaller variant in dev — same env-var contract:

```bash
ollama pull llama3.1:8b-instruct-q4_K_M
```

### 3.3 Stable DNS alias

Register **`ollama.farm.local`** → `<inference host IP>` in your farm DNS, or fall back to a hosts entry on every host that calls the API:

```bash
# /etc/hosts on the gr33n API host (and on dev laptops)
10.0.0.42  ollama.farm.local
```

The API references the alias, **never** a raw IP — IPs change, aliases survive.

---

## 4. Configure the gr33n API

Set these env vars on the API host (or in `.env` / `.env.local` — see [`.env.example`](../.env.example)):

```bash
AI_ENABLED=true
LLM_BASE_URL=http://ollama.farm.local:11434/v1
LLM_MODEL=llama3.1:70b-instruct-q4_K_M
LLM_API_KEY=                     # leave empty for intranet Ollama
LLM_TIMEOUT_SECONDS=666          # default 666; lower only if you want faster fail on hung LLM
LLM_TEMPERATURE=0.2              # default 0.2; existing Phase 24/25 knob
LLM_MAX_TOKENS=1024              # default 1024; existing Phase 24/25 knob
```

For **local laptop development** without a GPU, point at a cloud endpoint with **zero code changes**:

```bash
AI_ENABLED=true
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4.1-mini
LLM_API_KEY=sk-...
```

Or at LM Studio / vLLM / Ollama on `localhost` — the chat client follows `LLM_BASE_URL` verbatim.

### 4.1 Restart the API and confirm the startup probe

```bash
# Compose
docker compose restart api

# Or bare process
systemctl restart gr33n-api
```

You should see something like:

```
AI_ENABLED=true (set AI_ENABLED=false for Lite mode — no synthesis or /v1/chat)
llm backend reachable base_url=http://ollama.farm.local:11434/v1
```

If the probe fails, the API **exits** — that is intentional. Common causes:

- `LLM_BASE_URL` typo (missing `/v1`, wrong port).
- Firewall blocking the API host.
- `OLLAMA_HOST=0.0.0.0:11434` not yet applied (systemd needs `daemon-reload`).
- DNS alias not resolvable from the API host.

### 4.2 Smoke-test from the API host

```bash
# Capability check (public — no JWT)
curl -sS http://localhost:8080/capabilities
# → {"ai_enabled":true}

# Chat smoke (JWT required). Replace TOKEN with a real bearer.
curl -sS -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message":"Give me a one-sentence summary of what Farm Guardian is for."}'
```

Expected status mapping (see [phase_27_farm_guardian_ai_layer.md](plans/phase_27_farm_guardian_ai_layer.md)):

| AI_ENABLED | LLM_BASE_URL + LLM_MODEL set | `POST /v1/chat` |
|------------|------------------------------|------------------|
| unset / true | yes | **200** + `{"answer":"…","llm_model":"…"}` |
| true | not set | **503** `Farm Guardian chat is not configured` |
| false / 0 / no / off | (any) | **503** `AI features are disabled on this installation` |

---

## 4.3 Optional — vision chat (Phase 30 WS6)

Zone reference photos (WS5) can be attached to a grounded chat turn when a **multimodal** model is configured. Text-only tags such as `llama3.1:8b` cannot interpret pixels.

| Env var | Purpose |
|---------|---------|
| `LLM_VISION_MODEL` | Vision model id (e.g. `llava`, `llama3.2-vision`) — **required** to enable vision |
| `LLM_VISION_BASE_URL` | Optional; defaults to `LLM_BASE_URL` |
| `LLM_VISION_API_KEY` | Optional; defaults to `LLM_API_KEY` |

Example (same Ollama host as text chat):

```bash
ollama pull llava
# .env
LLM_VISION_MODEL=llava
# LLM_VISION_BASE_URL=http://ollama.farm.local:11434/v1   # omit to reuse LLM_BASE_URL
```

`GET /capabilities` returns `vision_chat_enabled: true` when `LLM_VISION_MODEL` and a base URL resolve. The Guardian drawer (zone **Ask Guardian** context) shows photo chips; `POST /v1/chat` accepts `attachment_ids` (zone `file_attachments` ids, max 3).

**Operator expectations:** vision output is **advisory** — flag possible issues, suggest checks, prefer **`create_task`** proposals over silent config changes. See [farm-guardian-architecture.md §8.4](farm-guardian-architecture.md#84-vision-and-zone-photos--limits).

CI smokes skip live vision unless `GR33N_VISION_TEST=1`.

---

## 5. Operational hygiene

| Concern | Action |
|---------|--------|
| **Model upgrades** | Pull the new tag, restart Ollama, update `LLM_MODEL`, restart the API. The startup probe will catch typos. |
| **Disk pressure** | `ollama list` / `ollama rm <tag>` to retire old weights. Put `/var/lib/ollama` on a dedicated SSD volume so it can't fill the root partition. |
| **GPU monitoring** | `nvidia-smi` and `journalctl -u ollama -f`. If you run the Phase 26 Loki overlay, Promtail will pick up the Ollama unit's journal automatically. |
| **First-token latency** | Tune `OLLAMA_KEEP_ALIVE=24h` (above) — without it the model unloads after a few minutes and the next request pays the cold-start tax. |
| **Idle power (Phase 163)** | **Rest now** / `GUARDIAN_AUTO_DORMANT_MINUTES` unload the warm model from RAM via the API. For full process stop on solar sites, admins use [`scripts/guardian-power.sh`](../scripts/guardian-power.sh) — not exposed in the web UI. |
| **Concurrent requests** | Ollama serializes per model. If you start seeing tail latency on a busy farm, scale **vertically** (bigger GPU) or split RAG synthesis off to a separate Ollama instance behind a different `LLM_BASE_URL` for each consumer — both still OpenAI-compatible. |
| **Switching to cloud** | Flip `LLM_BASE_URL` + `LLM_MODEL` + `LLM_API_KEY`. No code change. Useful for outage failover during GPU maintenance. |

---

## 6. Switching back to Lite mode

If the inference host is down for maintenance and you don't have a cloud fallback configured:

```bash
# Cleanly degrade — no broken UI, no errors mid-session.
echo "AI_ENABLED=false" >> .env.local
docker compose restart api   # or: systemctl restart gr33n-api
```

After restart:

- **`GET /capabilities`** returns `{"ai_enabled": false}`.
- The UI **Settings → AI features** chip flips to **Lite — AI disabled**.
- **Farm knowledge → Ask (LLM)** is disabled with a clean explanation.
- **`POST /v1/chat`** and **`POST /farms/{id}/rag/answer`** return **503** with the same message.

All operational features (schedules, rules, tasks, alerts, fertigation, sensors) keep working exactly as before — this is the Lite-mode contract from the Phase 27 plan.

---

## 7. References

- [Phase 27 — Farm Guardian AI layer](plans/phase_27_farm_guardian_ai_layer.md) — WS1 lives here.
- [Phase 26 — Operator logging runbook](operator-logging-runbook.md) — Compose + systemd logging posture (RAG / chat / automation `slog` lines).
- [RAG scope and threat model](rag-scope-and-threat-model.md) — §9 boundary between static guide, DB RAG, and ops logs.
- [Ollama documentation](https://ollama.com/library/llama3.1) — official model list and runtime knobs.

---

*Phase 27 WS1 v1 — Compose + systemd only; no Kubernetes track.*
