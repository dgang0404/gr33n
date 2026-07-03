# Example environment (deprecated)

The **canonical template** is **[`.env.example`](../.env.example)** at the repo root.

The **grouped reference** with defaults and links is **[`environment-variables.md`](environment-variables.md)**.

Copy `.env.example` → `.env`, edit values, optionally add `.env.local` overrides. The API loads both automatically when started from the repo root.

**UI:** [`ui/.env.example`](../ui/.env.example) — typically `VITE_API_URL=http://localhost:8080`.

This file previously mirrored `.env.example` inline; that duplicate is retired in Phase 116 to avoid drift.
