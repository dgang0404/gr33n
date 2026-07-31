# Browser E2E (Phase 117 + 211.07)

Playwright journeys against a **seeded dev stack** (`make dev-auth-test` or CI
`browser-e2e` lane). Set credentials via env when not using the demo seed defaults.

## Local

```bash
# Terminal 1 — API + UI (auth_test mode, seeded DB)
make dev-auth-test

# Terminal 2 — install once, then run journeys
cd e2e && npm ci && npx playwright install chromium
E2E_BASE_URL=http://localhost:5173 make e2e-browser
```

Use **`localhost`** (not `127.0.0.1`) so the browser Origin matches API `CORS_ORIGIN`.

Optional overrides:

```bash
export E2E_DEV_EMAIL=dev@gr33n.local
export E2E_DEV_PASSWORD=devpassword
export E2E_BASE_URL=http://localhost:5173
export E2E_API_URL=http://127.0.0.1:8080
export E2E_FARM_ID=1
```

## Journeys

1. `login-dashboard.spec.js` — login → Today dashboard
2. `task-create.spec.js` — tasks workspace create flow
3. `guardian-chat.spec.js` — Farm Guardian chat shell + Pending tab chrome (no live LLM)
4. `guardian-pending.spec.js` — Confirm / Dismiss seeded pending proposals (no live LLM)

Pending Confirm/Dismiss seeds via `POST /v1/chat/proposals/seed-pending` (dev/auth_test
only). That path inserts a `create_task` proposal — same inbox the UI Pending tab reads.
No Ollama / phi3 required.

## Stagehand spike (deferred)

Phase 211.07 WS4 left **Stagehand / Midscene out of the default suite**. Plain Playwright
+ `data-test` hooks cover Confirm/Dismiss deterministically. Revisit Stagehand only if a
real selector-churn flaky flow shows up; keep/kill note belongs here when that happens.

CI: manual **`browser-e2e`** job (`workflow_dispatch`) — same pattern as
`hardware-smoke` / `ollama-smoke`.
