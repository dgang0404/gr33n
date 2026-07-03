---
name: Phase 113 — Security hardening (auth surface, rate limits, headers)
overview: >
  Close the security gaps found in the July 2026 system audit: open self-registration,
  no login rate limiting, JWT accepted in query strings, missing security headers,
  chat cost guard disabled by default, DB-user password change missing, legacy shared
  PI_API_KEY, and weak upload content validation. Ordered by severity; each workstream
  is independently shippable.
todos:
  - id: ws1-registration-gate
    content: "WS1: Registration gating — REGISTRATION_MODE env (open|invite|closed, default invite); invite codes table + admin CRUD; smoke tests for each mode"
    status: pending
  - id: ws2-login-rate-limit
    content: "WS2: Login rate limiting — per-IP + per-username sliding window on POST /auth/login (default 10/min, env override); 429 with Retry-After; audit log on lockout"
    status: pending
  - id: ws3-jwt-query-string
    content: "WS3: Remove ?token= JWT — restrict query-string tokens to SSE endpoints only (EventSource cannot set headers), reject everywhere else; short-lived scoped SSE ticket as follow-up option"
    status: pending
  - id: ws4-security-headers
    content: "WS4: Security headers middleware — X-Content-Type-Options, X-Frame-Options, Referrer-Policy, HSTS (behind TLS env flag), CSP report-only first"
    status: pending
  - id: ws5-cost-guard-default
    content: "WS5: Chat abuse guard — enable CostGuardConfig defaults in prod builds (per-user daily token cap); env opt-out; 429 with clear operator message"
    status: pending
  - id: ws6-password-change
    content: "WS6: DB-user password change — extend PATCH /auth/password to update auth.users bcrypt hash for the authenticated user (not just env-admin); require current password; audit log"
    status: pending
  - id: ws7-shared-pi-key
    content: "WS7: Legacy PI_API_KEY deprecation — startup WARN when set; INSTALL.md migration steps to per-device keys; env PI_LEGACY_KEY_DISABLED=true kill switch"
    status: pending
  - id: ws8-upload-sniffing
    content: "WS8: Upload magic-byte sniffing — validate receipt/zone-photo uploads with http.DetectContentType against allowlist, not client Content-Type"
    status: pending
isProject: false
---

# Phase 113 — Security hardening (auth surface, rate limits, headers)

## Status

**Planned.** Sourced from the July 2026 comprehensive audit (schema/UI, Pi chain,
docs, security, tests). This phase covers the security findings; siblings cover the
rest ([114](phase_114_pi_edge_integrity.plan.md) Pi chain,
[115](phase_115_schema_utilization.plan.md) schema surfacing,
[116](phase_116_docs_refresh.plan.md) docs, [117](phase_117_test_depth.plan.md) tests,
[118](phase_118_guardian_model_capabilities.plan.md) Guardian model UX).

---

## Findings driving this phase

| # | Severity | Finding | Evidence |
|---|----------|---------|----------|
| 1 | High | `POST /auth/register` is public — anyone reaching the server can create an account | `cmd/api/routes.go` (public route), `internal/handler/auth/handler.go` |
| 2 | High | No brute-force protection on `POST /auth/login` (device keys have 120/min; login has none) | `cmd/api/routes.go` |
| 3 | High | `requireJWT` accepts `?token=` on every route — tokens leak into access logs, history, Referer | `cmd/api/auth.go` (~line 94) |
| 4 | High | Legacy shared `PI_API_KEY` authenticates all edge traffic when per-device keys absent | `cmd/api/pi_edge_auth.go` |
| 5 | Med | 24 h HS256 JWT, no refresh/revocation | `internal/handler/auth/handler.go` |
| 6 | Med | `PATCH /auth/password` only updates env-admin hash, not DB users | `internal/handler/auth/handler.go` |
| 7 | Med | No security headers (only CORS) | `cmd/api/cors.go` |
| 8 | Med | Upload validation trusts client `Content-Type`; no magic-byte check | `internal/handler/fileattach/handler.go`, `zone_photos.go` |
| 9 | Med | `/v1/chat` cost guard **disabled by default** — LLM abuse cap opt-in | `internal/farmguardian/cost_guard.go` |

Already good (no action): Pi per-device keys bcrypt-hashed with show-once plaintext;
sqlc/`$N` params everywhere (no SQL injection surface found); dev auth bypass double-gated
by `//go:build dev` + `AUTH_MODE=dev`; prod Docker image builds without `-tags dev`.

---

## Design notes

### WS1 — Registration gating

`REGISTRATION_MODE=open|invite|closed`, default **invite** (breaking change called out
in CHANGELOG/upgrade guide — existing open installs set `open` explicitly).
Invite codes: `auth.registration_invites` (code, created_by, expires_at, used_by,
used_at); owner/manager can mint codes from Settings. `closed` returns 403 with a
message pointing at the operator.

### WS2 — Login rate limiting

In-memory sliding window keyed by `(ip, username)` — this is a LAN-first single-node
app, no Redis needed. Defaults `AUTH_LOGIN_MAX_PER_MINUTE=10`. Lockout responses are
constant-time-ish (same body as bad-credentials plus Retry-After) to avoid user
enumeration via limiter behavior.

### WS3 — Query-string JWT

Audit which UI code paths rely on `?token=` (SSE `EventSource` cannot set an
Authorization header — that is the legitimate use). Restrict query-token acceptance to
an explicit allowlist of SSE routes; all other routes require the header. Follow-up
option (out of scope v1): one-time short-lived SSE tickets minted via authenticated
POST.

### WS5 — Cost guard defaults

Flip `CostGuardConfig` to enabled with generous defaults (e.g. 200k tokens/user/day)
so a stolen/shared login cannot melt a CPU-only Ollama box; `GUARDIAN_COST_GUARD=off`
opts out for trusted single-operator installs.

### Out of scope

- JWT refresh tokens / server-side revocation list (finding #5) — meaningful design
  work; schedule separately if multi-user internet-exposed installs become a target.
- TLS termination itself — stays a reverse-proxy concern; HSTS ships behind an env flag.
- CSP enforcement — ship report-only first, enforce after UI inline-style audit.

---

## Acceptance

- [ ] Fresh install: registering without invite fails 403 (mode=invite default); invite flow works end-to-end
- [ ] 11th login attempt within a minute returns 429 with Retry-After
- [ ] `?token=` rejected on non-SSE routes; SSE streams still work in UI
- [ ] `curl -sI` shows nosniff/frame/referrer headers on API responses
- [ ] Chat request over daily token cap returns 429; env opt-out restores old behavior
- [ ] DB user can change own password; old password rejected afterwards; audit row written
- [ ] Startup WARN when `PI_API_KEY` set; kill switch env rejects legacy key
- [ ] PNG renamed to `.pdf` (or vice versa) rejected on upload by content sniffing
- [ ] All new behavior covered in smoke tests; INSTALL.md + SECURITY.md updated

---

## Files expected to change

| Area | Files |
|------|-------|
| Auth | `cmd/api/auth.go`, `cmd/api/routes.go`, `internal/handler/auth/*` |
| Middleware | `cmd/api/cors.go` (or new `security_headers.go`), rate limiter |
| Schema | `db/schema/gr33n-schema-v2-FINAL.sql` + migration (invites table) |
| Guardian | `internal/farmguardian/cost_guard.go` |
| Edge | `cmd/api/pi_edge_auth.go` |
| Uploads | `internal/handler/fileattach/*` |
| Docs | `INSTALL.md`, `SECURITY.md`, `.env.example` |
| Tests | `cmd/api/smoke_phase113_*.go` |
