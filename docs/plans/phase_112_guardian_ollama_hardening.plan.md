---
name: Phase 112 — Guardian Ollama hardening (pull, context, CI E2E)
overview: >
  Close the Phase 111 out-of-scope gaps: enrich model discovery with Ollama /api/show
  context lengths so the 8192 grounded guardrail is meaningful; let farm admins pull
  missing models into the local Ollama runtime without SSH; add a dedicated CI lane
  with Ollama service + full E2E smokes (session override, audit, context reject,
  fallback) that today skip when no LLM is configured.
todos:
  - id: ws0-show-enrichment
    content: "WS0: Context enrichment — after /api/tags, POST /api/show per model (bounded concurrency); parse *.context_length from model_info; store in ModelCache; unit tests with fixture JSON"
    status: done
  - id: ws1-pull-api
    content: "WS1: Model pull — POST /guardian/models/pull {name}; farmauthz.RequireFarmAdmin; Ollama POST /api/pull; refresh cache on success; env GUARDIAN_OLLAMA_AUTO_PULL (default false) + GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS (default 600)"
    status: done
  - id: ws2-auto-pull-settings
    content: "WS2: Auto-pull on farm save — when GUARDIAN_OLLAMA_AUTO_PULL=true and admin PATCH /farms/{id}/settings names an unknown model, pull then persist; otherwise 400 with actionable message pointing at POST /guardian/models/pull"
    status: done
  - id: ws3-ui-pull
    content: "WS3: UI — GuardianModelSelector pull row for admins (model name + Pull button, progress/disabled state); show enriched context_window; toast on pull success/failure"
    status: done
  - id: ws4-openapi
    content: "WS4: OpenAPI — POST /guardian/models/pull; document enriched context_window; note pull is server-wide and admin-only"
    status: done
  - id: ws5-e2e-smokes
    content: "WS5: E2E smokes — cmd/api/smoke_phase112_ollama_e2e_test.go with //go:build ollama: session override → conversation_turns.llm_model; farm switch audit; phi3:mini grounded 400; missing-model fallback; pull tinyllama then discovery lists it"
    status: done
  - id: ws6-ci-lane
    content: "WS6: CI lane — .github/workflows/ci.yml job ollama-smoke (workflow_dispatch); Ollama service container; pull tinyllama + phi3:mini; go test -tags 'dev ollama' -run TestPhase112; document in INSTALL.md"
    status: done
isProject: false
---

# Phase 112 — Guardian Ollama hardening (pull, context, CI E2E)

## Status

**Shipped** on `main`. Builds on **Phase 111** (model cache, discovery, farm/session resolution).

**Preconditions (met on `main`):**
- [`internal/farmguardian/ollama_discovery.go`](../../internal/farmguardian/ollama_discovery.go) — `/api/tags` only; `context_window` always `0`
- [`internal/farmguardian/model_cache.go`](../../internal/farmguardian/model_cache.go) — `GuardianMinContextWindow = 8192`
- [`internal/handler/farm/guardian_settings.go`](../../internal/handler/farm/guardian_settings.go) — rejects models not in cache
- [`cmd/api/smoke_phase111_test.go`](../../cmd/api/smoke_phase111_test.go) — audit/session/guardrail tests skip without Ollama

---

## Why this phase

Phase 111 made model selection usable but left three operator/CI gaps:

| Gap today | After Phase 112 |
|-----------|-----------------|
| `context_window: 0` for almost all models → guardrail never fires | `/api/show` fills real context lengths; small models rejected on grounded chat |
| Unknown model on farm save → hard 400; operator must `ollama pull` over SSH | Admin **Pull** in UI or opt-in auto-pull on PATCH |
| Audit + session-override smokes skip in CI | Dedicated **ollama-smoke** job runs full E2E against real Ollama |

---

## Scope & design decisions

### WS0 — Context window enrichment (`/api/show`)

After `GET /api/tags`, enrich each discovered model:

```
POST {ollama_native}/api/show
{ "name": "llama3.2:latest" }
→ model_info: { "llama.context_length": 8192, ... }
```

**Parsing:** scan `model_info` for any key ending in `.context_length`; use the largest
positive integer found (some manifests expose multiple). If `/api/show` fails for one
model, log `WARN` and leave `context_window: 0` for that entry — do not fail the whole
discovery pass.

**Concurrency:** cap parallel `/api/show` calls at **4** (env override
`GUARDIAN_OLLAMA_SHOW_CONCURRENCY`) so startup stays polite on large model lists.

**When enrichment runs:** startup `RefreshFromEnv`, after successful pull (WS1), and on
manual `ModelCache.RefreshFromEnv` — same code path as tags.

**Guardrail behavior (unchanged logic, better data):**
- `context_window > 0 && < 8192` + grounded chat → **400** (Phase 111)
- `context_window == 0` → allow with warn (unchanged fallback for odd manifests)

---

### WS1 — Model pull API (server-wide, admin-only)

```
POST /guardian/models/pull
{ "name": "llama3.2" }
→ 200 { "name": "llama3.2", "status": "success" }
→ 403 non-admin
→ 504 pull timeout
```

- **Scope:** server-wide — Ollama is one runtime (same rule as Phase 111 discovery).
- **RBAC:** `farmauthz.RequireFarmAdmin` on **any** farm the user admins, OR require
  `farm_id` query param and check admin on that farm. Simplest: authenticated user must
  be admin on **at least one** farm they belong to (server ops action). Alternative
  accepted in implementation: pass `farm_id` in body and use `RequireFarmAdmin(w,r,q,farmID)`.
- **Local only:** reject when `LLM_BASE_URL` is not a local inference URL
  (`farmguardian.IsLocalInferenceURL`) — never pull against cloud gateways.
- **Implementation:** `POST /api/pull` with `"stream": false` (or consume stream until
  `status: success`). Honor `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS` (default **600**).
- **After success:** `ModelCache.RefreshFromEnv(ctx)` so enriched metadata is immediate.
- **Audit (optional v1):** log `INFO`; defer `guardian_model_pulled` audit enum unless
  needed — pull is not farm-scoped. WS5 smokes don't require audit on pull.

**Env:**

| Variable | Default | Meaning |
|----------|---------|---------|
| `GUARDIAN_OLLAMA_AUTO_PULL` | `false` | When `true`, admin PATCH settings may pull missing model before save (WS2) |
| `GUARDIAN_OLLAMA_PULL_TIMEOUT_SECONDS` | `600` | Max wait for one pull |
| `GUARDIAN_OLLAMA_SHOW_CONCURRENCY` | `4` | Parallel `/api/show` during enrichment |

---

### WS2 — Auto-pull on farm settings save (opt-in)

When `GUARDIAN_OLLAMA_AUTO_PULL=true` and admin `PATCH /farms/{id}/settings` names a
model **not** in cache:

1. Attempt pull (same helper as WS1)
2. Refresh cache
3. If still missing → 400
4. Else persist `guardian_preferred_model` + audit as Phase 111

When auto-pull is **false** (default), keep Phase 111 behavior: **400** with message:
`model not loaded — use POST /guardian/models/pull or ollama pull on the server`.

Session-level `model` on `POST /v1/chat` does **not** auto-pull (avoid multi-minute
chat requests). Missing session model → fallback + `model_fallback: true` (Phase 111).

---

### WS3 — UI

Extend [`GuardianModelSelector.vue`](../../ui/src/components/GuardianModelSelector.vue):

- Admin-only row: text input + **Pull** button → `POST /guardian/models/pull`
- Disable Pull while in-flight; show error/success inline
- Model dropdown options show enriched `context_window` when &gt; 0
- Optional badge: `ctx 8192` / `fast` / `reasoning` (already partially there)

No change to non-admin read-only farm default display beyond richer metadata from GET.

---

### WS5 — Full E2E smokes (`//go:build ollama`)

New file: `cmd/api/smoke_phase112_ollama_e2e_test.go`

| Test | Asserts |
|------|---------|
| `TestPhase112_SessionOverride` | `POST /v1/chat` with `model` ≠ server default → `conversation_turns.llm_model` matches |
| `TestPhase112_FarmModelSwitchAudit` | Admin PATCH farm model → `guardian_model_changed` row (replaces skipped Phase 111 test) |
| `TestPhase112_ContextWindowGuardrail` | With `phi3:mini` in cache (4096 ctx), grounded chat + that model → **400** |
| `TestPhase112_FallbackOnMissingModel` | Farm pref = nonsense → chat returns `model_fallback: true`, turn uses env default |
| `TestPhase112_PullThenDiscover` | `POST /guardian/models/pull` tinyllama → `GET /guardian/models` lists it with `context_window > 0` or name present |
| `TestPhase112_ShowEnrichment` | At least one discovered model has `context_window > 0` after enrichment |

Tests **fail** (not skip) when `-tags ollama` is set and Ollama env is configured.
Without the tag, file is excluded from default `go test ./...`.

**Test hooks:** export `SetModelCacheForTest` or use httptest Ollama mock for unit
layer; E2E lane uses real Ollama only.

---

### WS6 — CI `ollama-smoke` lane

Mirror Phase 33 **hardware-smoke** pattern:

```yaml
ollama-smoke:
  if: github.event_name == 'workflow_dispatch'
  runs-on: ubuntu-latest
  services:
    ollama:
      image: ollama/ollama
      ports: ['11434:11434']
  steps:
    - checkout + setup-go
    - run: |
        curl -s http://localhost:11434/api/pull -d '{"name":"tinyllama"}'
        curl -s http://localhost:11434/api/pull -d '{"name":"phi3:mini"}'
      # wait until models listed in /api/tags
    - env:
        AI_ENABLED: "true"
        LLM_BASE_URL: "http://localhost:11434/v1"
        LLM_MODEL: "tinyllama"
      run: go test -tags 'dev ollama' ./cmd/api/ -run TestPhase112 -count=1 -v
```

- **Trigger:** `workflow_dispatch` only (pull time + disk). Document manual run in
  [`INSTALL.md`](../../INSTALL.md) beside hardware-smoke.
- **Not on every PR** — avoids 5–10 min model downloads on each push.
- Optional follow-up: nightly cron (out of scope for v1).

---

## Workstream detail

### WS0 — Context enrichment

**Deliverables:**
- `EnrichModelContextWindows(ctx, baseURL, models []ModelInfo, client, concurrency) []ModelInfo`
  in `ollama_discovery.go`
- `parseContextLength(modelInfo map[string]any) int`
- Wire into `DiscoverOllamaModels` or post-process in `ModelCache.RefreshFromEnv`
- Unit tests: fixture `model_info` maps for llama, gemma, missing key

**Verify:** `GET /guardian/models` returns `context_window: 8192` (or similar) for
installed llama3.x; grounded chat with phi3:mini returns 400.

---

### WS1 — Pull API

**Deliverables:**
- `PullOllamaModel(ctx, baseURL, name, client, timeout) error` in `ollama_discovery.go`
- `POST /guardian/models/pull` handler on chat handler or small `models_pull.go`
- Route in [`cmd/api/routes.go`](../../cmd/api/routes.go)

**Verify:** `curl -X POST /guardian/models/pull -d '{"name":"tinyllama"}'` as farm admin
→ model appears on next `GET /guardian/models`.

---

### WS2 — Auto-pull on PATCH

**Deliverables:** extend [`guardian_settings.go`](../../internal/handler/farm/guardian_settings.go)
- read `GUARDIAN_OLLAMA_AUTO_PULL`
- call pull helper when model missing before validation

**Verify:** with auto-pull true, PATCH unknown model → 200 after pull completes; with
false → 400 with pull hint.

---

### WS3 — UI

**Deliverables:** admin pull row + loading state in `GuardianModelSelector.vue`

**Verify:** Pull tinyllama from UI → dropdown includes it after refresh.

---

### WS4 — OpenAPI

**Deliverables:** path + schemas for pull request/response; note admin + server-wide scope.

---

### WS5 — E2E smokes

**Deliverables:** `smoke_phase112_ollama_e2e_test.go` with build tag `ollama`

**Verify:** `go test -tags 'dev ollama' ./cmd/api/ -run TestPhase112 -v` green with
local Ollama + phi3:mini + tinyllama pulled.

---

### WS6 — CI lane

**Deliverables:** `ollama-smoke` job in [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml);
INSTALL.md section

**Verify:** manual workflow run green on GitHub Actions.

---

## Acceptance

- [x] Discovery enriches `context_window` via `/api/show` for installed models
- [x] Grounded chat rejects models with known `context_window < 8192` (phi3:mini E2E)
- [x] `POST /guardian/models/pull` works for farm admins on local Ollama only
- [x] `GUARDIAN_OLLAMA_AUTO_PULL=true` enables pull-on-save for farm settings; default off
- [x] UI pull control for admins; enriched context shown in selector
- [x] `TestPhase112_*` pass under `-tags ollama` with real Ollama (no skips)
- [x] `ollama-smoke` CI job documented and runnable via workflow_dispatch
- [x] OpenAPI updated

---

## Out of scope

- **Cloud model pull** — pull API disabled when `LLM_BASE_URL` is not local
- **Embedding model pull** — `EMBEDDING_MODEL` unchanged
- **Pull on every PR** — too slow; workflow_dispatch only for CI
- **Public Ollama registry browse** — operator supplies model name manually
- **Async pull jobs / queue** — v1 uses synchronous pull with timeout (admin UX acceptable)

---

## Implementation order

WS0 (enrichment) → WS1 (pull API) → WS2 (auto-pull PATCH) → WS5 (smokes, needs WS0+WS1)
→ WS6 (CI, needs WS5) → WS3 (UI) → WS4 (OpenAPI)

WS3 can parallel WS5 after WS1.

---

## Files expected to change

| Area | Files |
|------|-------|
| Discovery / pull | `internal/farmguardian/ollama_discovery.go`, `model_cache.go`, `*_test.go` |
| Handlers | `internal/handler/chat/models_pull.go`, `internal/handler/farm/guardian_settings.go` |
| Routes | `cmd/api/routes.go` |
| UI | `ui/src/components/GuardianModelSelector.vue` |
| Tests | `cmd/api/smoke_phase112_ollama_e2e_test.go` |
| CI / docs | `.github/workflows/ci.yml`, `INSTALL.md`, `openapi.yaml` |

---

## Related

| Doc | Role |
|-----|------|
| [`phase_111_guardian_model_selector.plan.md`](phase_111_guardian_model_selector.plan.md) | Shipped foundation; out-of-scope items land here |
| [`phase_84_100_master_roadmap.plan.md`](phase_84_100_master_roadmap.plan.md) | Phase 111+ index |
| [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) | `hardware-smoke` pattern for WS6 |
