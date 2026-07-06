# Environment variables reference

**Canonical template:** [`.env.example`](../.env.example) at the repo root (copy to `.env`, then `.env.local` overrides).

The API loads `.env` and `.env.local` automatically when started from the repo root. Process environment always wins over files.

**UI (separate):** [`ui/.env.example`](../ui/.env.example) — typically `VITE_API_URL=http://localhost:8080`.

---

## Core API

| Variable | Default | Purpose |
|----------|---------|---------|
| `DATABASE_URL` | peer-auth DSN in dev | PostgreSQL connection string |
| `PORT` | `8080` | HTTP listen port |
| `AUTH_MODE` | `production` | `dev` \| `auth_test` \| `production` (dev/auth_test require `-tags dev` build) |
| `JWT_SECRET` | — | Required when `AUTH_MODE` ≠ `dev` |
| `PI_API_KEY` | — | Legacy shared Pi edge key (prefer per-device keys) |
| `PI_LEGACY_KEY_DISABLED` | `false` | When `true`, reject shared `PI_API_KEY` auth |
| `REGISTRATION_MODE` | invite in prod | `open` \| `invite` \| `closed` |
| `AUTH_LOGIN_MAX_PER_MINUTE` | `10` | Login rate limit per IP |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed browser origin |
| `LOG_FORMAT` | text | Set `json` for structured slog |
| `AUTH_DEBUG_LOG` | off | Log auth rejection reasons (no tokens) |
| `ADMIN_USERNAME` | `admin` | Env-admin username |
| `ADMIN_PASSWORD_HASH` | — | bcrypt hash (prefer `~/.gr33n/admin.hash`) |
| `ADMIN_BIND_USER_ID` | demo UUID | JWT `user_id` for env-admin |
| `ADMIN_BIND_EMAIL` | `dev@gr33n.local` | JWT email for env-admin |
| `OPENAPI_UI` | on in `-tags dev` | Serve `/openapi` Redoc browser (`true`/`false`) |

---

## Automation & device health

| Variable | Default | Purpose |
|----------|---------|---------|
| `AUTOMATION_SIMULATION_MODE` | `true` | When true, worker logs actions without GPIO |
| `AUTOMATION_COOLDOWN_SECONDS` | worker default | Min seconds between rule re-fires |
| `DEVICE_OFFLINE_AFTER_SECONDS` | `900` | Stale heartbeat → offline + alert |

---

## Farm Guardian / LLM

| Variable | Default | Purpose |
|----------|---------|---------|
| `AI_ENABLED` | `true` when unset | Master switch — `false` = Lite mode (no `/v1/chat`, no RAG synthesis) |
| `LLM_BASE_URL` | — | OpenAI-compatible chat endpoint (e.g. Ollama `http://127.0.0.1:11434/v1`) |
| `LLM_MODEL` | — | Default chat model id |
| `LLM_API_KEY` | — | Bearer token when provider requires it |
| `LLM_TIMEOUT_SECONDS` | `666` | HTTP timeout for chat completions |
| `LLM_MAX_TOKENS` | model default | Cap completion tokens |
| `LLM_TEMPERATURE` | model default | Sampling temperature |
| `LLM_RETRY_MAX_ATTEMPTS` | `3` | Retries on transient LLM failures |
| `LLM_RETRY_BACKOFF_MS` | `500` | Initial retry backoff |
| `LLM_VISION_MODEL` | — | Vision model for photo analysis |
| `LLM_VISION_BASE_URL` | `LLM_BASE_URL` | Vision API base |
| `LLM_VISION_API_KEY` | `LLM_API_KEY` | Vision API key |
| `GUARDIAN_LLM_PROPOSALS` | `false` | LLM-generated Confirm proposals |
| `GUARDIAN_COST_GUARD` | on in production | `off` disables token caps |
| `CHAT_COST_WINDOW_HOURS` | `24` | Rolling window for token caps |
| `CHAT_COST_MAX_TOKENS_PER_USER` | `200000` when guard on | Per-user cap → HTTP 429 |
| `CHAT_COST_MAX_TOKENS_PER_FARM` | `0` | Per-farm cap (0 = disabled) |
| `CHAT_SESSION_TTL_DAYS` | `30` | Prune old chat sessions (`0` = off) |
| `CHAT_SESSION_PRUNE_INTERVAL_HOURS` | `24` | Prune loop interval |
| `CHAT_SESSION_PRUNE_STARTUP_DELAY_SECONDS` | `30` | Delay before first prune |
| `GUARDIAN_OLLAMA_AUTO_PULL` | — | Auto-pull missing Ollama models |
| `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS` | — | Timeout for model pull |
| `GUARDIAN_OLLAMA_SHOW_CONCURRENCY` | — | Parallel `ollama show` during discovery |
| `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` | `max(1500, LLM_TIMEOUT_SECONDS)` | Grounded `/v1/chat` HTTP timeout floor (farm counsel on CPU) |
| `GUARDIAN_EARLY_SSE` | on | `0`/`false` disables early SSE phase status before prompt build |
| `GUARDIAN_INLINE_WARMUP_ON_SEND` | on | `0`/`false` skips inline chat preload on grounded send |
| `GUARDIAN_EVAL_TIMEOUT_SECONDS` | inherits grounded timeout | `cmd/guardian-eval` / `make guardian-qa-smoke` HTTP client timeout |
| `STT_BASE_URL` | — | Local speech-to-text (Whisper-compatible) — enables `/v1/chat/stt` |

See also [INSTALL.md](../INSTALL.md) § Guardian, [farm-guardian-architecture.md](farm-guardian-architecture.md).

---

## RAG & embeddings

| Variable | Default | Purpose |
|----------|---------|---------|
| `EMBEDDING_API_KEY` | — | Embedding provider key (`rag-ingest` CLI) |
| `EMBEDDING_BASE_URL` | — | Embedding API base |
| `EMBEDDING_MODEL` | — | Embedding model name |
| `EMBEDDING_DIMENSION` | — | Vector dimension |
| `EMBEDDING_TIMEOUT_SECONDS` | — | Embedding HTTP timeout |
| `RAG_INGEST_UPDATED_AFTER` | — | Incremental ingest watermark (RFC3339) |
| `RAG_SYNTHESIS_MAX_PER_MINUTE` | `30` | Global rate limit on `POST .../rag/answer` |
| `RAG_SYNTHESIS_MAX_PER_MINUTE_PER_FARM` | `0` | Per-farm synthesis limit |
| `CROP_CATALOG_SOURCE` | `db` | `db` \| `yaml` for crop catalog |
| `AGRONOMY_FIELD_GUIDES_SOURCE` | `db` | Field guide source |
| `CROP_LIBRARY_PATH` | — | Legacy YAML crop library path |

---

## Push notifications (FCM)

| Variable | Default | Purpose |
|----------|---------|---------|
| `FCM_SERVICE_ACCOUNT_JSON` | — | Inline Firebase service account JSON |
| `GOOGLE_APPLICATION_CREDENTIALS` | — | Path to service account file |

See [push-notifications playbook](push-notifications-operator-playbook.md) if present, or INSTALL.md.

---

## File / receipt storage

| Variable | Default | Purpose |
|----------|---------|---------|
| `FILE_STORAGE_BACKEND` | `local` | `local` \| `s3` |
| `FILE_STORAGE_DIR` | `./data/files` | Local blob root |
| `FILE_STORAGE_SIGNED_URL_TTL_SECONDS` | `300` | Download URL TTL |
| `S3_BUCKET` | — | S3 bucket name |
| `S3_REGION` | — | AWS region |
| `S3_ENDPOINT` | — | Custom S3-compatible endpoint |
| `S3_PREFIX` | — | Key prefix inside bucket |
| `S3_ACCESS_KEY_ID` | — | S3 access key |
| `S3_SECRET_ACCESS_KEY` | — | S3 secret |
| `S3_USE_PATH_STYLE` | `false` | Path-style URLs (MinIO etc.) |
| `S3_DISABLE_HTTPS` | `false` | HTTP-only endpoints (dev/test) |

See [backup-restore-runbook.md](backup-restore-runbook.md).

---

## Insert Commons (optional)

| Variable | Default | Purpose |
|----------|---------|---------|
| `INSERT_COMMONS_INGEST_URL` | — | Farm API → receiver POST URL |
| `INSERT_COMMONS_SHARED_SECRET` | — | HMAC shared secret |
| `INSERT_COMMONS_PSEUDONYM_KEY` | — | Farm pseudonym salt |
| `INSERT_COMMONS_RECEIVER_LISTEN` | `:8765` | Receiver listen address |
| `INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH` | off | Local pilots only |
| `INSERT_COMMONS_RECEIVER_RETENTION_DAYS` | `90` | Receiver row retention |

---

## Security headers

| Variable | Default | Purpose |
|----------|---------|---------|
| `SECURITY_HSTS_ENABLED` | off | Send Strict-Transport-Security |
| `SECURITY_CSP_REPORT_ONLY` | off | CSP report-only mode |

---

## Domain / misc

| Variable | Default | Purpose |
|----------|---------|---------|
| `STRICT_PROGRAM_STAGE_MATCH` | off | Strict fertigation program stage matching |
| `DEV_SEED_PROFILE` | — | Dev reset profile (`small_indoor`, etc.) |
| `TIMESCALE_RETENTION_DAYS` | — | Dev DB retention helper |

---

## Parity check

```bash
make audit-env   # scripts/env_reference_parity.sh
```
