# What needs a connection — and to what

gr33n is **local-first**: the core loop (sensors, alerts, tasks, actuator control,
zone data) runs entirely on your LAN, no internet required. But a few features
reach further out. This page is the single place that says, per feature,
**which network it needs** — LAN-only, WAN/internet, or none at all.

Related: [offline-or-intranet-deployment.md](offline-or-intranet-deployment.md)
(how to lay out hosts for a no-WAN site), [environment-variables.md](environment-variables.md)
(the variables referenced below). Guardian **129–139** shipped awakening, runtime, QA, and inference profiles — see the [next-level roadmap](plans/archive/phase_129_139_guardian_next_level_roadmap.plan.md).

## Three connection tiers

| Tier | Meaning |
|------|---------|
| **None** | Runs entirely in the browser/Pi/API process; no network call at all |
| **LAN only** | Talks to another service on your network (or `127.0.0.1`/loopback) — never needs the public internet |
| **WAN required** | Needs to reach a service on the public internet |

A machine with **no internet uplink at all** — a barn on a LAN with no WAN
gateway — can run every "None" and "LAN only" row below. Nothing in that set
needs to leave the property.

## Feature matrix

| Feature | Tier | Detail |
|---------|------|--------|
| Dashboard, zones, sensors, alerts, tasks, schedules | None / LAN only | Browser (UI) ↔ API ↔ Postgres, all on your network |
| Actuator control (manual + automation rules) | LAN only | API → Pi edge client over your LAN (or loopback if API and Pi client share a host) |
| Pi sensor/actuator edge loop | LAN only | Pi client posts to the API's LAN/loopback address — never calls out to the internet itself |
| **Virtual Pi** (`/virtual-pi`) — board view, wiring edit, config.yaml download | LAN only | UI ↔ your API; generates config from DB wiring. **Notify Pi to reload** bumps `config_version` so platform-sync Pis refetch on LAN |
| PWA offline queue (Tasks create/status) | None while offline | Queues in browser SQLite (`offline_queue.db`); syncs to your API once it's reachable again — that sync is LAN, not WAN |
| **Guardian chat — using an already-installed model** | **LAN only** | API calls `LLM_BASE_URL` (e.g. `http://127.0.0.1:11434/v1` for Ollama). If Ollama runs on the same box or LAN, this is 100% local — no internet |
| **Guardian model switch (session/farm default dropdown)** | **None** | Only changes which already-downloaded model the next request uses — no network call happens at all until you actually send a chat message |
| **Guardian "Pull model into Ollama" (admin action)** | **WAN required** | Calls Ollama's `/api/pull`, which downloads model weights from Ollama's public registry. This is the one model-related action that needs internet — unless you've mirrored models internally (see air-gap notes below) |
| Guardian RAG retrieval (field guides, farm data grounding) | LAN only | Uses your local embedding service (`EMBEDDING_BASE_URL`) — same story as chat: local Ollama/LM Studio/vLLM = no internet |
| Guardian image understanding (`LLM_VISION_MODEL`, zone photo "Ask Guardian") | LAN only | Same local LLM endpoint as chat; no separate cloud dependency |
| Receipt photo → cost entry | LAN only | Same vision pipeline as above |
| Data Commons opt-in (aggregate sharing to Insert Commons) | **WAN required** | Explicit opt-in per farm; posts to `insertcommons.org`. Off by default — nothing leaves your server unless you turn this on |
| Commons catalog browsing (crop/recipe packs shipped with gr33n) | None | Seeded into your own Postgres at install time; browsing it is a local DB read, not a live fetch |
| Software updates (`git pull`), first-time `npm ci` / `go mod download` | WAN required | One-time or occasional — not part of day-to-day operation |
| Pulling container images / OS packages | WAN required | One-time setup step; mirror internally for a true air-gap (see [offline-or-intranet-deployment.md](offline-or-intranet-deployment.md)) |

## The short answer to "does switching the LLM need a connection?"

**Switching is free — pulling is not.**

- Picking a different model that's **already installed** in Ollama (the "This
  chat" / "Farm default" dropdowns in the model selector): no network call,
  works air-gapped.
- **Downloading a new model** you don't have yet (the "Pull model into Ollama"
  admin control, or running `ollama pull <model>` yourself): needs internet,
  because the weights come from Ollama's public model registry.

Once a model is pulled, it lives on disk — using it from then on is LAN-only
(or fully offline if the API and Ollama share a machine).

**Why the first reply after a restart is slow:** if the model selector shows a
**"cold"** hint, it means Ollama hasn't loaded that model into RAM/VRAM yet.
The first chat message triggers that load from local disk — it can take
anywhere from a few seconds to a couple of minutes depending on model size
and hardware, but it is 100% local (no internet, no download). Subsequent
messages are fast because the model stays warm in memory until Ollama evicts
it for inactivity or another model is requested.

**Grounded chat while Ollama is busy:** when farm context triggers RAG
embedding, Ollama may be slow on `GET /v1/models` even though the daemon is
healthy. The API uses a longer local probe (12s default), retries unreachable
probes after 2s, and treats a live `GET /api/ps` as reachable so grounded turns
are not rejected with a false "LLM unreachable" error.

## Air-gapped sites

If you have zero WAN access at the install site:

1. Pull every model you'll need **before** going on-site (or on a connected
   machine, then copy the Ollama model store over).
2. Everything in the "None" and "LAN only" rows above works unmodified.
3. Skip Data Commons opt-in — it has no offline equivalent by design.
4. See [offline-or-intranet-deployment.md](offline-or-intranet-deployment.md)
   for host layout and `LLM_BASE_URL` / `EMBEDDING_BASE_URL` wiring.

---

*Added 2026-07-03 in response to an operator question about whether switching
Guardian models needs internet.*
