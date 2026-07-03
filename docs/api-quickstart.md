# API quickstart — curl cookbook

Base URL: `http://localhost:8080` (or your deployment). Full spec: [openapi.yaml](../openapi.yaml) · browser UI: `http://localhost:8080/openapi` (dev builds or `OPENAPI_UI=true`).

---

## 1. Login (JWT)

```bash
TOKEN=$(curl -sS -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"YOUR_PASSWORD"}' | jq -r .token)

curl -sS http://localhost:8080/farms \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Use a registered user when `AUTH_MODE=auth_test` or `production`. Dev builds (`AUTH_MODE=dev`, `-tags dev`) may bypass auth — still use JWT for realistic testing.

---

## 2. Farm & zone CRUD

```bash
# Create farm
curl -sS -X POST http://localhost:8080/farms \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"North greenhouse","location_description":"Bench A"}' | jq .

# List zones
curl -sS http://localhost:8080/farms/1/zones \
  -H "Authorization: Bearer $TOKEN" | jq .

# Create zone
curl -sS -X POST http://localhost:8080/farms/1/zones \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Veg room","zone_type":"indoor"}' | jq .
```

Replace `farm_id` / paths with IDs from your responses.

---

## 3. Pi device key (edge client)

Per-device keys are preferred over the shared `PI_API_KEY`.

**Operator (JWT)** — register device, then read key from response or DB:

```bash
curl -sS -X POST http://localhost:8080/farms/1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"pi-bench-01","device_uid":"pi-bench-01"}' | jq .
```

**Edge calls** — use `X-API-Key` header (device key or legacy `PI_API_KEY`):

```bash
curl -sS -X PATCH http://localhost:8080/devices/1/status \
  -H "X-API-Key: YOUR_DEVICE_OR_PI_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"status":"online","client_version":"curl-smoke"}' | jq .
```

See [pi-integration-guide.md](pi-integration-guide.md) for heartbeat, sensor readings, and command queue.

---

## 4. Farm Guardian chat

Requires `AI_ENABLED=true` and reachable `LLM_BASE_URL` / `LLM_MODEL`.

```bash
curl -sS -X POST http://localhost:8080/v1/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "What should I check first on my humidity alert?",
    "farm_id": 1,
    "model": "llama3.1:8b"
  }' | jq .
```

Optional: list models (`GET /guardian/models`), pull a model (`POST /guardian/models/pull`), check health (`GET /v1/chat/health`).

Streaming responses use SSE — use `curl -N` or a client that reads event streams.

---

## Capabilities probe

```bash
curl -sS http://localhost:8080/capabilities \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Returns `ai_enabled`, `vision_chat_enabled`, `stt_local_enabled`.

---

## Further reading

- [environment-variables.md](environment-variables.md)
- [farm-guardian-architecture.md](farm-guardian-architecture.md)
- [CONTRIBUTING.md](../CONTRIBUTING.md) — OpenAPI parity when adding routes
