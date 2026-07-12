# Field guides — curated knowledge for grounded Guardian

Markdown in this directory is ingested into RAG as the **`field_guide`** corpus (`make rag-ingest-field-guides`). Guardian cites these sources when they match the operator's question.

## When field guides are used

| Farm context | Embedding / RAG | Field guides in the prompt? |
|--------------|-------------------|----------------------------|
| **Off** | Not run | **No** — tinyllama, phi3, and larger chat models answer from general training only |
| **On** + farm selected | Runs if `EMBEDDING_MODEL` is configured | **Yes, when chunks match** the question (plus live farm snapshot, platform docs, farm rows) |

The **chat model** (phi3:mini, tinyllama, llama3.1:8b, …) does not change *whether* RAG runs — only how well it reads and cites the retrieved chunks. A larger local model is usually more reliable; tinyllama is not offered for grounded chat (context window too small).

## What belongs here

- **Supported crops** — `crop-*-nutrition.md`, nursery guides, care sheets aligned with the crop catalog
- **Field install** — Pi GPIO, relays, sensors, irrigation basics, electrical safety boundaries
- **Fertigation triage** — `fertigation-troubleshooting.md` (schedule vs manual, pump/EC checks)
- **Demo farm map** — `demo-farm-pi-layout.md` (seed device names, relay channels — farm 1)
- **Procedures** — `procedures/*.yaml` for step-by-step confirm flows
- **Honest gaps** — `crop-unsupported-*.md` when gr33n does not ship targets (woodland, mushroom, etc.)

Author for **reviewed, citeable facts** — not open-ended LLM guesses. After edits, re-ingest: `make rag-ingest-field-guides`.

## Should we add more guides?

**Yes, incrementally**, when operators repeat the same grounded questions and catalog/RAG gaps show up in eval or support.

**No**, for one-off home-garden questions with farm context off — those never hit this corpus. Example: forest-garden understory (cherry + goldenrod + blackberries) is off-farm horticulture unless you add a dedicated reviewed guide *and* operators query with **farm context on**.

Existing related files:

- `crop-chrysanthemum-care.md` — short-day mums (demo farm veg/bloom); photoperiod, humidity, harvest cues
- `crop-marigold-care.md` · `crop-geranium-care.md` — bedding flowers (Phase 172 catalog)
- `crop-cherry-nursery.md` — sweet cherry **nursery/production**, not backyard forest garden
- `crop-unsupported-woodland.md` — ramps/ginseng; explains gr33n does not invent woodland feed schedules
- `crop-strawberry-nutrition.md` — bench crop targets, not wild bramble management

A new guide (e.g. `crop-forest-garden-understory.md`) would only help after ingest + grounded questions that retrieve it.

## See also

- [crop-knowledge-operator-runbook.md](../crop-knowledge-operator-runbook.md) — catalog vs RAG vs read tools
- [farm-guardian-persona-platform-context.md](../farm-guardian-persona-platform-context.md) — grounding stack
- [operator-tour.md §6d](../operator-tour.md) — offline field assistant
