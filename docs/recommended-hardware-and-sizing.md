# Recommended hardware & sizing (gr33n platform)

**Audience:** Operators and IT choosing servers for a farm deployment — API, UI, PostgreSQL, automation worker, RAG, and Farm Guardian chat.

**Related:**
- [Farm Guardian + Ollama install](farm-guardian-ollama-setup.md) — inference host detail
- [Farm Guardian architecture](farm-guardian-architecture.md) — what runs on each chat turn
- [Local operator bootstrap](local-operator-bootstrap.md) — **laptop daily cheat sheet** + **server & frontier delta**
- [Guardian Ollama laptop playbook](guardian-ollama-laptop-playbook.md) — CPU contention, pull, timeouts
- [Raspberry Pi & deployment topology](raspberry-pi-and-deployment-topology.md) — edge vs central server

---

## Short answer: will it feel laggy?

**Yes, on underpowered hardware — especially the LLM step, not the RAG search.**

| Part of a grounded Guardian turn | Typical latency | What hurts if specs are low |
|----------------------------------|-----------------|-----------------------------|
| UI → API (auth, farm snapshot) | tens of ms | Weak CPU on API host only if DB is remote/slow |
| Embed question + pgvector search | ~100–500 ms | Slow Postgres disk; huge corpus without indexes |
| Build prompt + stream to LLM | **seconds to tens of seconds** | **No GPU, small RAM, or 70B model on a laptop** |
| First token (“time to first byte”) | 0.5–5+ s | Cold model load, CPU-only inference, busy GPU |
| Full answer (streaming) | 5–60+ s | Model size, `LLM_MAX_TOKENS`, GPU VRAM |

**RAG alone** (search / answer without chat) is lighter: one embedding call + vector query + optional synthesis. **Guardian chat** adds multi-turn history, live snapshot, rule-assisted proposals, and **streaming** — so it feels the heaviest part of the product.

**Lite mode** (`AI_ENABLED=false`): no chat, no RAG synthesis — only API, UI, DB, automation. Runs fine on a modest VM or NUC. Use this when you have no GPU and no cloud LLM budget.

---

## Deployment profiles

Pick one profile; you can split roles across machines (recommended for production).

### Profile A — Dev / demo (laptop or small VM)

**Goal:** One developer or demo farm; acceptable quality, not production load.

| Role | Spec | Notes |
|------|------|--------|
| All-in-one | 8+ CPU cores, **16 GB RAM**, 50 GB SSD | Same machine runs Compose Postgres + API + UI dev server + Ollama |
| GPU | Optional: **8–12 GB VRAM** or CPU-only | `.env`: `LLM_MODEL=llama3.1:8b` (see repo dev default). CPU-only works but first token is slow |
| Postgres | Docker `db` image or native PG16 | Timescale + PostGIS + pgvector required |
| Embeddings | Cloud API key **or** same Ollama host if it supports an embedding model | Without `EMBEDDING_API_KEY`, grounded chat/search returns empty RAG context (snapshot still works) |

**Expect:** Chat is usable for demos; 70B on this profile is usually impractical. Occasional “stutter” on first message after idle is normal (model load).

### Profile B — Small farm (Lite + optional cloud AI)

**Goal:** Real operations without on-prem GPU.

| Role | Spec |
|------|------|
| App + DB server | 4 vCPU, **8–16 GB RAM**, 100 GB SSD |
| GPU | None |
| LLM | **Cloud** OpenAI-compatible endpoint in `LLM_BASE_URL`, or stay in **Lite mode** |
| Embeddings | Cloud `EMBEDDING_*` for RAG |

**Expect:** Dashboard, alerts, automation, tasks — snappy. Guardian drawer shows Lite banner; no local chat.

### Profile C — Production Full Guardian (on-prem, recommended split)

**Goal:** Helpful Guardian + RAG on LAN without sending farm data to a cloud LLM.

| Machine | Spec | Runs |
|---------|------|------|
| **App / data server** | 8 vCPU, **32 GB RAM**, 200+ GB SSD (NVMe preferred) | PostgreSQL (Timescale, PostGIS, pgvector), `cmd/api`, automation worker, static UI (nginx), file attachments |
| **Inference server** | **RTX 3090 / 4090 (24 GB VRAM)** or equivalent, **64 GB RAM**, 50 GB SSD | Ollama — `llama3.1:70b-instruct-q4_K_M` (Phase 27 default) |
| Network | Gigabit LAN between app host and inference host | `LLM_BASE_URL=http://ollama.farm.local:11434/v1` |

**Optional fourth box:** Pi / MQTT edge gateways (see [Pi integration guide](pi-integration-guide.md)) — minimal CPU; not for LLM.

**Expect:** First token often **1–3 s** with model kept warm (`OLLAMA_KEEP_ALIVE`); full answers stream smoothly. Under load (several operators chatting), add GPU headroom or cap tokens (`CHAT_COST_*`).

### Profile D — Single-box production (budget constrained)

**Goal:** One physical server only — trade model size for VRAM.

| Spec | Notes |
|------|--------|
| 12–16 CPU threads, **32–64 GB RAM**, 500 GB SSD | Postgres + API on same OS |
| GPU **12–16 GB VRAM** | Use **`llama3.1:8b`** or **`13b` Q4**, not 70B |
| Embeddings | Small local model or cloud API |

**Expect:** Guardian is **helpful for ops Q&A and proposals** but weaker on deep agronomy vs 70B. Monitor RAM: Postgres and Ollama compete — give Postgres at least 8–16 GB `shared_buffers` / OS cache headroom.

---

## Component-by-component requirements

### PostgreSQL (DB + RAG vectors)

| | Minimum | Recommended (Full + RAG) |
|--|---------|---------------------------|
| RAM | 4 GB dedicated | **16–32 GB** for farm DB + pgvector index in memory |
| CPU | 2 cores | 4–8 cores (Timescale + automation + vector search) |
| Disk | 20 GB | **100+ GB** SSD (sensor history, attachments, chunk growth) |
| Extensions | **TimescaleDB, PostGIS, pgvector** | Same — use repo `db/Dockerfile` or [install script](../scripts/install-system-deps-debian.sh) |

RAG ingest size scales with documents: rule of thumb **~1–4 KB per chunk row** + index. Thousands of chunks are fine on a NUC; millions need tuning (`lists` on ivfflat/hnsw) and RAM.

### gr33n API + automation worker

| | Notes |
|--|--------|
| CPU | 2–4 cores sufficient; Go process is not the bottleneck |
| RAM | **512 MB–2 GB** for API; worker adds little |
| Disk | Logs + receipt storage path if enabled |

Startup fails if `LLM_BASE_URL` is set but Ollama is down (`AI_ENABLED=true`) — size the inference host for uptime, not the API host.

### UI (Vue dashboard)

| | Notes |
|--|--------|
| Build | Dev: `npm run dev` on operator laptop. Prod: static files served by nginx/Caddy — **negligible** server cost |
| Browser | Any modern desktop; chat SSE is lightweight |

### Ollama / LLM (Farm Guardian + optional RAG answer)

| Model tier | VRAM (approx) | Quality / speed |
|------------|---------------|-----------------|
| **llama3.1:8b Q4** | 6–8 GB | Fastest on-prem; good for dev and small farms; more hallucination risk on edge cases |
| **llama3.1:70b Q4** | **22–24 GB** | Production default in docs; best general agronomy + instruction following |
| CPU-only 8b | 0 GPU, 16+ GB RAM | Works but **very laggy** (tens of seconds per reply) — not recommended for operators |

Tune timeouts: `LLM_TIMEOUT_SECONDS=666` (default); only lower if you prefer fast-fail over slow CPU answers.

### Embeddings (RAG retrieval)

Separate from chat LLM:

| Option | Hardware | Latency |
|--------|----------|---------|
| **Cloud** (`EMBEDDING_API_KEY` + OpenAI-compatible URL) | None on-prem; needs outbound HTTPS | Low per query; privacy policy decision |
| **Local** (same or second Ollama / vLLM embedding model) | Shares inference box CPU/GPU | Adds load on inference host |

Each grounded chat turn does **one embedding** of the user question, then pgvector search — cheap compared to generating 500+ completion tokens.

---

## What makes Guardian “actually helpful” vs frustrating

Helpful when:

1. **`AI_ENABLED=true`** and **`LLM_BASE_URL` + `LLM_MODEL`** point at a **working** backend (API startup probe passes).
2. **Grounded turns** use `farm_id` — live snapshot + (optional) RAG chunks. Ingest farm notes with `make rag-ingest-demo` or pipeline when `EMBEDDING_API_KEY` is set.
3. **Model large enough** for your questions (70B on GPU, or capable cloud model).
4. **Inference host on LAN** — not Wi‑Fi to a overloaded laptop across the farm.
5. **Cost caps** configured so one runaway session does not stall the GPU for everyone (`CHAT_COST_MAX_TOKENS_*`).

Frustrating when:

- 70B configured on **8 GB VRAM** → OOM or constant swapping.
- Postgres on **HDD** with huge `sensor_readings` hypertable → snapshot + rules slow everything.
- No embeddings ingested → Guardian only sees snapshot text (“0 chunks”) — still useful for alerts/tasks, weaker on “what did we log last week about EC?”
- **First message after lunch** slow → cold Ollama model; set `OLLAMA_KEEP_ALIVE=24h` on inference host.

---

## Sizing checklist (before go-live)

- [ ] Postgres has Timescale + PostGIS + **pgvector** (`make check-stack` or API boot).
- [ ] `GET /capabilities` → `ai_enabled: true` when you expect chat.
- [ ] `GET {LLM_BASE_URL}/models` from API host succeeds.
- [ ] Demo chat on `/chat` or Guardian drawer: first token **&lt; 5 s** with warm model.
- [ ] `make rag-ingest-demo` (or production ingest) if you need corpus-backed answers.
- [ ] Firewall: Ollama **not** on public internet ([Ollama setup §3.1](farm-guardian-ollama-setup.md)).
- [ ] Plan **Lite mode** fallback if GPU host is down (`AI_ENABLED=false` — farm still runs).

---

## Architecture sketch (typical production)

```
Operators (browser)
        │
        ▼
┌───────────────────┐     LAN      ┌────────────────────┐
│  App + Postgres   │ ────────────▶│  Ollama (GPU box)   │
│  API :8080        │  LLM_BASE_URL│  :11434 /v1         │
│  UI (static)      │              │  70B or 8B model    │
│  pgvector RAG     │              └────────────────────┘
└───────────────────┘
        │
        ▼ (optional)
   Pi / MQTT edge ── sensor readings, pending_command
```

---

## Summary table (copy-paste for IT)

| Tier | Use case | App+DB server | Inference (Ollama) | Helpful Guardian? |
|------|----------|---------------|----------------------|-------------------|
| **Minimum** | Lite only | 4 vCPU / 8 GB RAM | — | N/A (no chat) |
| **Dev** | Laptop demo | 16 GB RAM all-in-one | 8B Q4 GPU or CPU | Yes, with patience |
| **Recommended** | On-prem farm | 8 vCPU / 32 GB RAM | 24 GB VRAM, 70B Q4 | **Yes** |
| **Budget single box** | One server | 32 GB RAM + 12 GB GPU | Same box, **8B** model | Yes, scoped quality |

When in doubt, **split Postgres/API from GPU inference** and start with **8B on-prem** or **cloud LLM** before buying a 70B-class GPU server.
