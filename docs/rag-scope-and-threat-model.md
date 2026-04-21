# RAG scope, threat model, and storage (Phase 24 WS1–WS2)

This document fixes **what** may enter the retrieval layer, **who** may query it, **what must not leave the farm** without explicit consent, and **where embeddings live** in Postgres. It informs WS3 (ingestion), WS4 (API), and WS5 (optional LLM). Implementation must stay aligned with `cmd/api/auth.go`, `internal/farmauthz/farmauthz.go`, and `internal/authctx/authctx.go`.

---

## 1. Goals

- Give operators **farm-scoped semantic search** over data they already store in gr33n (tasks, cycles, automation history, costs, etc.).
- Keep **farm isolation** as strong as the rest of the dashboard: one farm’s vectors and retrieved text must never satisfy another farm’s query.
- Make **optional LLM synthesis** (WS5) a deliberate, consent-gated path—not a silent data export.

Non-goals for the initial cut are unchanged from [Phase 24 — RAG retrieval system](plans/phase_24_rag_retrieval_system.plan.md): no training of foundation models on customer data; no shipment of farm payloads to third-party clouds without configuration and consent aligned with Insert Commons and audit patterns.

---

## 2. Threat model

### 2.1 Actors and proof

| Actor | Proof | Implication for RAG |
|--------|--------|----------------------|
| Dashboard user | HS256 JWT (`Authorization: Bearer` or `?token=` for SSE) | All **human-facing** retrieval and synthesis **must** use JWT paths and farm membership checks (see §3). |
| Pi / edge | `X-API-Key` == `PI_API_KEY` | **Not** a substitute for tenant-scoped RAG for operators. Only use Pi auth for RAG if we **explicitly** design a separate edge contract (we have not). Default: **no** Pi-key retrieval API. |
| Dev laptop | `-tags dev` **and** `AUTH_MODE=dev` | Full farm-authz skip—must never be exposed on the internet. |

### 2.2 Assets

| Asset | Risk if mishandled |
|--------|---------------------|
| Chunk text + metadata (farm_id, source table, record id) | Cross-tenant leakage via wrong `farm_id` filter or handler bug. |
| Embedding vectors | Same as chunk text—vectors are reversible “roughly” to semantics; treat like source text for isolation. |
| LLM prompts (WS5) | Could exfiltrate retrieved chunks if the provider is off-box; requires explicit trust and settings. |
| Logs (access, SSE) | Query strings with `?token=` must not be logged in full in production. |

### 2.3 Trust boundaries (summary)

- **JWT user + DB membership** is the source of truth for “may this user read this farm’s RAG index,” mirroring other farm APIs.
- **Context flags** (`FarmAuthzSkip`, `PiEdgeAuth`) are set only by auth middleware—**RAG handlers must not set them from request input.**
- **Farm id for authorization** must come from the URL path or from a server-resolved resource (e.g. load chunk by id, read `farm_id` from DB), not from unverified client-only body fields for listing or search.

---

## 3. Farm isolation (requirements)

These are **non-negotiable** for WS2–WS4:

1. **Every stored chunk** carries an explicit `farm_id` (and ideally `source_type` / `source_id` for audit and refresh).
2. **Every query** filters by the **authorized** farm id (the same farm the user proved membership for on that request).
3. **Dashboard routes** use `requireJWT` and `farmauthz.RequireFarmMember` (or stricter) consistent with existing `/farms/{id}/...` patterns; OpenAPI uses `bearerAuth` like other farm routes (`make audit-openapi` after changes).
4. **No mixing** of chunks across farms in index build, reindex, or hybrid search filters.
5. Optional **LLM** (WS5): prompts must include only chunks already authorized for that request’s farm—**no** “bring your own farm_id” in the synthesis request body without re-checking membership.

---

## 4. Data classes and embedding candidates

Classify sources by **sensitivity** and **utility** for operators. “Embed” means **eligible for indexing** after sanitization—not that every column is copied verbatim.

### 4.1 Sensitivity tiers

| Tier | Meaning | Handling |
|------|---------|----------|
| **Public-within-farm** | Operational facts the team already sees in-app (task titles, cycle name, automation status) | Allowed for v1 chunks after normal app-level access rules. |
| **Sensitive** | May contain PII or commercial detail (labor notes, free-text “cycle notes”, `automation_runs.details` JSONB, notification bodies) | **Sanitize or exclude** specific keys; never embed opaque JSON without a schema-aware allowlist. |
| **Secret / high-risk** | Webhook URLs, tokens, credentials, push tokens | **Never** embed; see §5. |

### 4.2 Candidate sources (by domain)

The list below ties to tables introduced or strengthened in Phase 20.95 and related work ([phase_20_95_rag_prep_and_housekeeping.plan.md](plans/phase_20_95_rag_prep_and_housekeeping.plan.md)). Prioritization for **v1 ingestion order** is a product call; a sensible default is **high-signal operator text first**, then structured rollups.

| Domain | Tables / objects | Text useful for retrieval | Sanitization / notes |
|--------|------------------|---------------------------|----------------------|
| **Tasks** | `gr33ncore.tasks` | `title`, `description`, `task_type`, `status`, related module pointers | Exclude or tokenize assignee identifiers if displayed; `description` may be long—chunking strategy in WS3. |
| **Task labor** | `gr33ncore.task_labor_log` | `notes` | **Sensitive**—may name people; consider exclude v1 or strip. |
| **Crop cycles** | `gr33nfertigation.crop_cycles` | `name`, `strain_or_variety`, `yield_notes`, `cycle_notes`, stage + dates | Notes fields are **sensitive**; align with Phase 21 summary semantics when available ([phase_21_crop_cycle_analytics.plan.md](plans/phase_21_crop_cycle_analytics.plan.md)). |
| **Fertigation programs** | `gr33nfertigation.programs` | `name`, `description`, recipe linkage labels | `metadata` JSON—allowlist keys only if embedded. |
| **Automation runs** | `gr33ncore.automation_runs` | `status`, `message` | `details` JSONB may contain payloads—**allowlist** or omit until WS3 defines a scrubber. |
| **Automation rules / schedules** | `automation_rules`, `schedules`, `executable_actions` | Names, descriptions, human-readable condition summaries | **Exclude** `action_parameters` for action types that carry URLs/secrets (e.g. HTTP webhook parameters) unless scrubbed. |
| **Costs** | `gr33ncore.cost_transactions`, categories | Narrative memo fields, category + amount **aggregates** | Per-transaction amounts may be sensitive—product decision whether v1 embeds **line detail** vs **rollup text** only. |
| **Inputs / inventory** | `input_definitions`, batches | Names, SKU-like labels, low-stock context | Unit cost fields are commercial—align with operator expectations before embedding raw numbers. |
| **Alerts** | `gr33ncore.alerts_notifications` | Rendered subject/message text | May include sensor/device names—still farm-scoped; avoid recipient cross-links if stored. |

### 4.3 Explicit exclusions (must not embed)

- **Secrets and credentials**: webhook URLs with query tokens, API keys, shared secrets, `user_push_tokens`, raw `INSERT_COMMONS_*` keys.
- **Opaque JSON** without review: `automation_runs.details`, `executable_actions.action_parameters`, arbitrary `programs.metadata`—embed only after a **documented allowlist** in WS3.
- **Cross-farm aggregates** as “chunks” unless the product explicitly defines a new non-farm-scoped index (out of scope for v1).

---

## 5. Egress, LLMs, and Insert Commons

| Path | Policy |
|------|--------|
| **Embeddings + retrieval only** (no external LLM) | Vectors and chunk text stay in the same operational boundary as the app database (exact storage in WS2). |
| **Local / operator-controlled LLM** (e.g. LAN) | Same policy as above if no data leaves the operator network. |
| **Third-party LLM API** | **Explicit opt-in** per farm or deployment; document data flow, retention, and sub-processors; no payload by default ([insert-commons-pipeline-runbook.md](insert-commons-pipeline-runbook.md), [insert-commons-receiver-playbook.md](insert-commons-receiver-playbook.md) set expectations for minimal, consented sharing). |
| **Insert Commons / pseudonymized sharing** | **Separate** pipeline from interactive RAG; only **scrubbed aggregates** per existing contracts—never raw retrieval chunks unless a future phase explicitly merges those designs. |

---

## 6. Checklist — ship in v1 ingestion?

**Engineering defaults below** are the Phase 25 implementation baseline (aligned with [Phase 25 plan](plans/phase_25_rag_operations_and_expansion.plan.md) decisions). Product may refine priority later; ingest code must still respect sanitization rules in §4.2 regardless.

| Domain | Engineering default | Depends on / notes |
|--------|---------------------|-------------------|
| **Tasks** (`gr33ncore.tasks` text fields) | **Yes** | Implemented in Phase 24 — chunking/sanitize per §4.2 |
| **Task labor notes** (`task_labor_log.notes`) | **Later** | PII risk (§4.2); exclude or strip before any ingest |
| **Crop cycles** (`crop_cycles` names + strain + bounded notes) | **Yes** | Phase 25 ingestion expansion — align with Phase 21 summaries when present; sanitize sensitive notes |
| **Fertigation programs** (name, description; `metadata` allowlisted) | **Yes** | `sanitize.FertigationProgramMetadataForEmbed` — omits `steps` + sensitive keys; strings / numbers / bools / string arrays only |
| **Automation runs** (`status`, `message`; scrubbed `details`) | **Yes** | Implemented in Phase 24 — JSON scrubber for `details` |
| **Automation rules / schedules / executable_actions** (labels + scrubbed JSON) | **Yes** | `rag-ingest -schedules`, `-automation-rules`, `-executable-actions`; `action_parameters` via `sanitize.AutomationDetailsJSON`; metadata `module` **automation** |
| **Costs** (per-line narrative vs rollup text only) | **Later** | Commercial sensitivity (§4.2); rollup-only product call |
| **Inputs / inventory** (definitions, batches — no raw unit_cost unless approved) | **Later** | Operator expectation before raw commercial fields |
| **Alerts / notifications** (rendered subject/body text) | **Later** | Volume + noise |

**Suggested implementation order for domains marked Yes:** tasks → automation runs → schedules / automation rules / executable actions → **crop cycles** → **fertigation programs** → remainder per product after §4.2 gates.

---

## 7. WS2 storage decision (implemented)

**Choice:** **pgvector** inside PostgreSQL (`CREATE EXTENSION vector`), same operational boundary as relational farm data — backups, replication, and access control stay unified; no separate vector SaaS for v1.

**Objects:**

| Artifact | Location |
|---------|----------|
| Migration (repeatable installs) | `db/migrations/20260518_phase24_rag_pgvector.sql` |
| Schema mirror | `db/schema/gr33n-schema-v2-FINAL.sql` (extension + table) |
| Docker dev DB | `db/Dockerfile` — TimescaleDB **pg16** image + pgvector **v0.8.0** built from source; `docker-compose.yml` `db` service builds this image |

**Table:** `gr33ncore.rag_embedding_chunks`

| Column | Role |
|--------|------|
| `farm_id` | Isolation — every query filters by authorized farm |
| `source_type`, `source_id`, `chunk_index` | Dedupe / upsert key per farm |
| `content_text` | Canonical snippet used for display and regeneration checks |
| `embedding` | `vector(1536)` — matches common OpenAI-compatible embedding widths; **WS3 must use one model per dimension** or add a migration |
| `model_id` | Embedding model identifier for audit and re-embed decisions |
| `metadata` | Optional hybrid filters (`module`, dates, zone ids) — **no secrets** |

**Indexes:** btree on `(farm_id)`, `(farm_id, source_type, source_id)`, and **HNSW** cosine on `embedding`. WS4 queries **must** include `WHERE farm_id = $authorized_farm` (see §3).

**Host installs:** bare-metal / manual Postgres needs the pgvector package before loading the schema — see [INSTALL.md](../INSTALL.md) §2c.

---

## 8. Hand-off to later work-streams (“§7 hand-offs”)

**What this means:** §7 in earlier drafts referred to **contracts between WS1 and downstream streams** — i.e. what storage, ingestion, HTTP, and LLM layers must respect from scope/threat-model + storage choices. Use this table when implementing WS3–WS6.

| WS | This document informs… |
|----|-------------------------|
| **WS2** (**done**) | pgvector + `rag_embedding_chunks`; `farm_id` + `(source_type, source_id, chunk_index)` dedupe; `metadata` for hybrid filters. |
| **WS3** (**done**) | Implemented: `internal/rag/sanitize`, `internal/rag/embed`, `internal/rag/ingest`, `cmd/rag-ingest`; sqlc `rag.sql`; ingestion pulls from checklist-approved domains only; automation `details` sanitized; **crop_cycle**, **fertigation_program**, **schedule**, **automation_rule**, **executable_action** documents (SQL `ListExecutableActionsByFarmForRAG`); embeddings must match `vector(1536)` unless `EMBEDDING_DIMENSION` matches a future migration. |
| **WS4** (**done**) | `GET`/`POST /farms/{id}/rag/search` — `requireJWT` + `RequireFarmMember`; farm id from path; OpenAPI `bearerAuth`; vector search always filters `farm_id`; optional `module` + `created_since`/`created_until` on chunk rows. |
| **WS5** (**done**) | `POST /farms/{id}/rag/answer` — same JWT + membership; retrieval uses only farm-filtered chunks; LLM sees **only** those numbered blocks; citations derived from `[n]` references mapped to `chunk_id` / `source_type` / `source_id`; optional **LLM** via env (see Phase 24 plan WS5); **rate limit** `RAG_SYNTHESIS_MAX_PER_MINUTE`. |
| **WS6** (**done**) | Dashboard **Monitor → Knowledge** (`/farm-knowledge`); smoke tests `cmd/api/smoke_rag_test.go`; glossary + §10.6 in `docs/workflow-guide.md`. |

---

## 9. References

- [Phase 24 — RAG retrieval system](plans/phase_24_rag_retrieval_system.plan.md)
- [Phase 20.95 — RAG-prep columns](plans/phase_20_95_rag_prep_and_housekeeping.plan.md)
- [Phase 21 — Crop cycle analytics](plans/phase_21_crop_cycle_analytics.plan.md)
- Insert Commons: [insert-commons-pipeline-runbook.md](insert-commons-pipeline-runbook.md), [insert-commons-receiver-playbook.md](insert-commons-receiver-playbook.md)
- Schema: `db/schema/gr33n-schema-v2-FINAL.sql`

---

*This file is the authoritative Phase 24 scope, threat-model, and WS2 storage record until amended by a planned phase update.*
