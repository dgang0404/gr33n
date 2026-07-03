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
    status: done
  - id: ws2-login-rate-limit
    content: "WS2: Login rate limiting — per-IP + per-username sliding window on POST /auth/login (default 10/min, env override); 429 with Retry-After; audit log on lockout"
    status: done
  - id: ws3-jwt-query-string
    content: "WS3: Remove ?token= JWT — restrict query-string tokens to SSE endpoints only (EventSource cannot set headers), reject everywhere else; short-lived scoped SSE ticket as follow-up option"
    status: done
  - id: ws4-security-headers
    content: "WS4: Security headers middleware — X-Content-Type-Options, X-Frame-Options, Referrer-Policy, HSTS (behind TLS env flag), CSP report-only first"
    status: done
  - id: ws5-cost-guard-default
    content: "WS5: Chat abuse guard — enable CostGuardConfig defaults in prod builds (per-user daily token cap); env opt-out; 429 with clear operator message"
    status: done
  - id: ws6-password-change
    content: "WS6: DB-user password change — extend PATCH /auth/password to update auth.users bcrypt hash for the authenticated user (not just env-admin); require current password; audit log"
    status: done
  - id: ws7-shared-pi-key
    content: "WS7: Legacy PI_API_KEY deprecation — startup WARN when set; INSTALL.md migration steps to per-device keys; env PI_LEGACY_KEY_DISABLED=true kill switch"
    status: done
  - id: ws8-upload-sniffing
    content: "WS8: Upload magic-byte sniffing — validate receipt/zone-photo uploads with http.DetectContentType against allowlist, not client Content-Type"
    status: done
isProject: false
---

# Phase 113 — Security hardening (auth surface, rate limits, headers)

## Status

**Shipped** on `main` (July 2026).

---

## Acceptance

- [x] Fresh install: registering without invite fails 403 (mode=invite default); invite flow works end-to-end
- [x] 11th login attempt within a minute returns 429 with Retry-After
- [x] `?token=` rejected on non-SSE routes; SSE streams still work in UI
- [x] `curl -sI` shows nosniff/frame/referrer headers on API responses
- [x] Chat cost guard enabled by default in production (`GUARDIAN_COST_GUARD=off` opts out)
- [x] DB user can change own password; old password rejected afterwards; audit row written
- [x] Startup WARN when `PI_API_KEY` set; kill switch env rejects legacy key
- [x] Invalid upload content rejected by magic-byte sniffing
- [x] All new behavior covered in smoke tests; INSTALL.md + SECURITY.md updated
