# Guardian + Ollama — laptop and CPU playbook

**Audience:** Operators on a **16 GB CPU-only laptop** (or any box without a GPU) running local Ollama for Farm Guardian.

**Related:** [INSTALL.md](../INSTALL.md) · [local-operator-bootstrap.md](local-operator-bootstrap.md) (daily cheat sheet) · [farm-guardian-ollama-setup.md](farm-guardian-ollama-setup.md) · [Phase 126 plan](plans/phase_126_guardian_cpu_efficiency.plan.md)

---

## What “CPU” means in the UI (not a different model)

Farm Guardian uses the **same Ollama models** everywhere (`phi3:mini`, `llama3.1:8b`, etc.). The label **CPU** describes **where inference runs**, not a separate model SKU:

| `runtime_hint` (model selector) | Meaning |
|----------------------------------|---------|
| `cold — first message loads…` | Weights on disk; not in RAM yet. First turn loads the model (can take 1–5+ minutes on a laptop). |
| `loaded, CPU-only — expect slow replies` | Model is in RAM; Ollama reports **0 VRAM** → math runs on the **processor**, not a GPU. |
| `loaded on GPU` | Model uses GPU VRAM → much faster replies. |

The streaming banner **“Generating answer — running on CPU (no GPU)…”** is sent when **farm context is on** (`farm_id` set). It means: *this turn may take several minutes because inference is CPU-bound; wait for one reply before sending another.*

There is no “non-CPU model” in the dropdown — only **GPU vs CPU execution** for the tag you picked.

---

## What the UI does **not** do

| Action | UI / API behavior |
|--------|-------------------|
| **Switch model in “This chat” dropdown** | Next message uses the selected tag. **Does not** run `ollama stop` on other models. |
| **Change farm default** | Saves preference for the farm. **Does not** unload Ollama weights. |
| **Pull model** | `POST /guardian/models/pull` → Ollama `POST /api/pull` (blocks until done or timeout, default **600 s**). **One-time download** over the internet; not used on every chat turn. |
| **Stop button** | Aborts the in-flight HTTP/SSE chat request. **Does not** stop Ollama models in RAM. |

**Manual operator steps** (terminal on the Ollama host) when the box feels wedged or out of RAM:

```bash
# Stale CLI jobs from old terminals can block Ollama for hours — check first
pgrep -a 'ollama run'    # should be empty; kill any PIDs listed

# Free RAM before a long grounded chat session (after RAG ingest is done)
ollama stop phi3:mini
ollama stop rjmalagon/gte-qwen2-1.5b-instruct-embed-f16   # your EMBEDDING_MODEL tag

# If models hang on "Stopping..." for >30s
sudo systemctl restart ollama
```

Grounded chat **always** uses the **embedding** model briefly (RAG query vector), then the **chat** model. On a laptop, having **both** resident (~7 GB) plus 100% CPU often looks like a “timeout” in the UI.

---

## Switching models (tinyllama ↔ phi3 ↔ llama3.1:8b)

The **“This chat”** dropdown only changes which **tag** the API sends on the **next** `POST /v1/chat`. It does **not** download, delete, or unload models.

| Model (example) | Farm context **off** | Farm context **on** (grounded) | On your laptop |
|-------------------|----------------------|--------------------------------|----------------|
| **tinyllama** | Fast smoke, general Q&A | **Rejected** — advertised context 2048 &lt; 8192 minimum for grounded | Good for “hi”, bad for demo-farm RAG |
| **phi3:mini** | Works; ~10–60s first load on CPU | **Supported** — prompts trimmed to effective 4096 (Phase 126) | Default for quality grounded chat |
| **llama3.1:8b** (bigger) | Works if pulled | **Supported** if installed — slower/heavier on 16 GB CPU | Usually needs more RAM; GPU helps |

**What happens when you switch in the dropdown:**

1. **Next message** uses the new tag (e.g. tinyllama → phi3:mini).
2. Ollama **loads** that model from **local disk** if not already in RAM (cold load — minutes on CPU).
3. The **previous** model may **stay in RAM** until you `ollama stop <name>` or Ollama evicts it under memory pressure.
4. **Embedding model** still runs for grounded turns regardless of chat dropdown — that is a separate `EMBEDDING_MODEL` in `.env`, not the chat selector.

**Practical laptop pattern:**

- **Grounded demo farm** → `phi3:mini`, farm context on, one message at a time.
- **Quick ungrounded test** → `tinyllama`, farm context **off**.
- **Do not** expect to hop tinyllama ↔ phi3 ↔ llama3.1 rapidly on 16 GB CPU without stalls — each cold load competes for RAM.

---

## How pull works (UI and CLI)

Pull means: **download model weights from the internet once**, then they live under Ollama’s store (`~/.ollama/models` or `/usr/share/ollama`) forever. Chat after that is **offline** (local disk only).

### Path A — UI (farm admin)

**Pull is a text box, not a dropdown.** “This chat” and “Farm default” list models **already on disk** (`ollama list`). Pull is only for typing a **new** tag you want to download (e.g. a model **not** in those lists yet). If `llama3.1:8b` already appears in the chat dropdown, you do **not** pull it again — it is installed.

```
Browser  →  POST /guardian/models/pull  { "name": "llama3.1:8b", "farm_id": 1 }
         →  API checks farm admin JWT + LLM_BASE_URL is local Ollama
         →  Ollama POST /api/pull  { "name": "llama3.1:8b", "stream": false }
         →  (downloads layers from registry.ollama.com — needs internet)
         →  API refreshes GET /guardian/models  →  new tag appears in dropdown
```

- Button shows **“Pulling…”** until the HTTP request finishes or **times out** (default **600 s = 10 minutes**).
- **`llama3.1:8b` is ~4.9 GB** — on a 20 Mbps link that is ~30+ minutes; on 100 Mbps ~6–8 minutes. **10 minutes is often not enough** on home internet → UI shows timeout even while `ollama pull` continues in the background.
- **Fix for big models:** use the terminal (no API timeout) or raise in `.env`:

  ```bash
  GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS=3600   # 1 hour
  ```

  Then restart the API and retry UI pull, or prefer:

  ```bash
  ollama pull llama3.1:8b
  ```

- Requires **farm owner/manager** role (`canAdmin` in the selector).

### Path B — Terminal (most reliable for big models)

```bash
ollama pull llama3.1:8b    # or phi3:mini, tinyllama, etc.
ollama list                # confirm tag appears
```

Then refresh the Guardian page — `GET /guardian/models` reads **`ollama list`** / `/api/tags`; no second pull needed.

### When pull is **not** needed

If `ollama list` already shows the tag (you pulled during setup), the UI **Pull** box is only for **adding new** tags. Your laptop already has `phi3:mini` and `tinyllama` — switching dropdown uses those **local** weights; no download.

### Why pull can feel “broken”

| Issue | What’s going on |
|-------|------------------|
| Timeout at 10 min | Model &gt; ~2 GB on slow link; API gave up at 600 s |
| “Pull failed” immediately | No internet, wrong tag name, or not farm admin |
| Pulled but not in dropdown | Only **chat-capable** models appear (embedding-only tags hidden unless `?all=true`) |
| Pull “worked” but chat slow | Download done; **first chat** still **loads weights into RAM** (cold) — separate step |

---

## RAG bring-up (replicable sequence)

From the **repository root** (`cd ~/gr33n-platform`):

```bash
# 1. Stack up (Postgres + API + UI)
make restart-local          # db only, if needed
make dev-auth-test          # API :8080 + UI :5173 (blocks terminal)

# 2. Full Guardian corpus for demo farm 1 (needs EMBEDDING_API_KEY in .env)
make guardian-bootstrap-farm FARM_ID=1

# Or stepwise:
make rag-ingest-farm-operational FARM_ID=1
make rag-ingest-platform-docs

# 3. Verify chunks
PGPASSWORD=gr33n psql -h 127.0.0.1 -p 5433 -U gr33n -d gr33n -c \
  "SELECT source_type, count(*) FROM gr33ncore.rag_embedding_chunks WHERE farm_id=1 GROUP BY 1 ORDER BY 2 DESC;"
```

**Do not** run heavy RAG ingest and grounded chat **at the same time** — both hammer Ollama CPU.

After changing API code (e.g. Phase 126), **restart** API/UI so `GET /guardian/models` shows `effective_context_window` (e.g. `4096` for `phi3:mini`).

---

## Model pull vs fast dropdown (hardware tiers)

See also: [local-operator-bootstrap.md](local-operator-bootstrap.md) § Server & frontier delta.

| Profile | RAM / GPU | Typical models pulled | Dropdown “switch” speed |
|---------|-----------|------------------------|-------------------------|
| **Laptop** | 16 GB, CPU only | `phi3:mini` + embed model | **Slow** — loading/switching can take minutes; keep **one** chat model loaded. |
| **Standard server** | 32 GB, optional GPU | `phi3:mini` + `llama3.1:8b` + embed | Moderate — may keep 1–2 chat models if RAM allows. |
| **Frontier / enterprise site** | 64 GB+, GPU | Several chat models pre-pulled | **Fast** — models already on disk and often warm in RAM; see [hypothetical-enterprise-topology.md](hypothetical-enterprise-topology.md). |

**Pull** (UI “Pull model into Ollama” or `ollama pull <tag>`) is a **background-style job**:

- Downloads GB-scale weights (**internet speed** — often 10–40+ minutes per large model).
- UI pull waits up to `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS` (default **600**); increase in `.env` for slow links.
- After pull, weights are **local** — chat no longer needs the internet.

Pre-pulling multiple models on a **nice server** makes the selector feel instant; on a **laptop**, prefer **one** default (`LLM_MODEL=phi3:mini`) and accept CPU latency.

---

## Sanity checks

```bash
# Ollama healthy (should return in ~10–15s on laptop)
curl -sf -m 60 http://127.0.0.1:11434/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"model":"phi3:mini","messages":[{"role":"user","content":"say hi"}],"stream":false,"max_tokens":15}'

# Phase 126 effective context (needs JWT)
TOKEN=$(curl -sf -X POST http://127.0.0.1:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"dev@gr33n.local","password":"devpassword"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
curl -sf -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/guardian/models \
  | python3 -c "import sys,json; m=next(x for x in json.load(sys.stdin)['available_models'] if x['name']=='phi3:mini'); print(m.get('effective_context_window'), m.get('runtime_hint'))"
```

---

## UI testing tips

1. **Warm-up:** ungrounded “hi” first (farm context off) loads `phi3:mini` from disk.
2. **Forest garden / off-farm plants:** turn **Use farm context** off — horticulture outside the demo farm does not need RAG.
3. **Grounded demo farm:** farm context **on**, farm **gr33n Demo Farm (id 1)**, wait for the amber **Generating…** banner — **one message at a time**.
4. If the banner never completes → run the **manual Ollama cleanup** above, then retry.

### Timeout profile (16 GB CPU laptop — validated)

Keep these in `.env` unless you prefer fast-fail over slow answers:

| Variable | Value | Why |
|----------|-------|-----|
| `LLM_TIMEOUT_SECONDS` | **666** | Cold `phi3:mini` on CPU can take many minutes; 666 s is the project default and enough for ungrounded + trimmed grounded turns |
| `LLM_RETRY_MAX_ATTEMPTS` | **1** | Avoids stacking two full generation runs after a transient blip |
| `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS` | **600** (default) | Pull only — use terminal for large models if the UI times out |

The Guardian chat UI has **no client-side fetch timeout**; only **Stop** aborts an in-flight stream. A hung “Generating…” banner is almost always Ollama contention or a stale `ollama run`, not the browser cutting off at 666 s.

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| “Generating…” for 5+ min, no text | CPU saturation; phi3 + embed both loaded; stale `ollama run` | Cleanup commands above; one chat at a time |
| `LLM request failed` / `llm_timeout` | Same + old API without Phase 126 trim | Restart API; verify `effective_context_window: 4096` |
| `llm_busy` | Embed + chat competing | `ollama stop <embed-model>` after ingest |
| RAG ingest `context deadline exceeded` | Ingest + chat + multiple models | Stop chat; ingest when Ollama idle |
| `tinyllama` grounded 400 | Grounded gate requires advertised context ≥ 8192 | Use `phi3:mini` for grounded, or farm context off with tinyllama |

---

## Changelog

| Date | Note |
|------|------|
| 2026-07-04 | Initial playbook — laptop validation, Phase 126, RAG bring-up, UI vs CLI cleanup |
| 2026-07-04 | Document validated 666 s timeout profile for 16 GB CPU laptop |
