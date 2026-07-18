---
name: Phase 57 — Per-device Pi API keys
overview: >
  Security track deferred from Phase 51 "52+" — each edge device authenticates with
  its own scoped API key instead of sharing farm JWT. Platform issues/revokes keys;
  Pi stores locally; config sync and telemetry use device credential.
todos:
  - id: ws1-schema
    content: "WS1: device_api_keys table — hash, label, device_id, revoked_at, last_used"
    status: completed
  - id: ws2-platform-ui
    content: "WS2: Device detail — issue key, show once, revoke; operator copy"
    status: completed
  - id: ws3-pi-agent
    content: "WS3: Pi agent reads key from env/file; migrates from shared secret"
    status: completed
  - id: ws4-auth-middleware
    content: "WS4: Accept device key header on device-scoped routes; audit log"
    status: completed
  - id: ws5-docs-tests
    content: "WS5: pi-sequent-hat-setup + operator guide; security smokes; OC-57"
    status: completed
isProject: false
---

# Phase 57 — Per-device Pi API keys

## Status

**Shipped.** Follows [Phase 51](phase_51_pi_config_sync.plan.md) config sync (WS6 shipped).

**When to ship:** Multi-Pi farms in production or shared operator accounts on one farm.

---

## The one job

> **If one Pi is compromised or rotated, revoke one key — not the whole farm.**

---

## WS1 — Schema

```sql
device_api_keys (
  id, device_id, key_hash, label,
  created_at, revoked_at, last_used_at
)
```

- Plain key shown once at creation
- Store bcrypt/argon hash only

---

## WS2 — Platform UI

- Device setup wizard step: "Copy API key to Pi"
- Device card: **Rotate key** / **Revoke**
- Warning when device still on legacy shared auth

---

## WS3 — Pi agent

- `GR33N_DEVICE_API_KEY` or `/etc/gr33n/device.key`
- Fallback to legacy during migration window (logged deprecation)

---

## WS4 — Auth middleware

- Header: `X-Device-Key` or `Authorization: Device <key>`
- Scope: device_id must match URL `{id}` on pi-config, telemetry ingest
- Rate limit per key

---

## WS5 — Docs, tests, OC-57

- Update [pi-sequent-hat-setup.md](../pi-sequent-hat-setup.md) § credentials
- Integration: issue → telemetry → revoke → 401
- Phase 51 closure test note: keys replace shared-secret path

---

**Enterprise boundary:** Org-level key management / multi-farm admin deferred — [Phase 59](../enterprise-tier-boundary.md).

---

## Definition of done

- [x] New Pi provisioned with per-device key only
- [x] Revoke stops telemetry within TTL
- [x] No plaintext keys in DB
- [x] OC-57 closed
