# Browser E2E (Phase 117)

Playwright journeys against a **seeded dev stack** (`make dev-auth-test` or CI
`browser-e2e` lane). Set credentials via env when not using the demo seed defaults.

## Local

```bash
# Terminal 1 — API + UI (auth_test mode, seeded DB)
make dev-auth-test

# Terminal 2 — install once, then run journeys
cd e2e && npm ci && npx playwright install chromium
E2E_BASE_URL=http://127.0.0.1:5173 make e2e-browser
```

Optional overrides:

```bash
export E2E_DEV_EMAIL=dev@gr33n.local
export E2E_DEV_PASSWORD=devpassword
export E2E_BASE_URL=http://127.0.0.1:5173
```

## Journeys

1. `login-dashboard.spec.js` — login → Today dashboard
2. `task-create.spec.js` — tasks workspace create flow
3. `guardian-chat.spec.js` — Farm Guardian page shell (no live LLM required)

CI: manual **`browser-e2e`** job (`workflow_dispatch`) — same pattern as
`hardware-smoke` / `ollama-smoke`.
