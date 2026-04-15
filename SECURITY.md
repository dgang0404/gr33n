# 🔐 gr33n Security Policy

gr33n is built with the belief that food system tools should be **trusted, inspectable, and locally secured** — not beholden to distant servers or hidden dependencies.

---

## 🧭 Philosophy

- **Local trust > cloud trust**
- **Simplicity over obscurity**
- **Code is not magic — read it, own it, defend it**

Whether deployed in a cabin, co-op greenhouse, or airgapped industrial farm, gr33n values **resilience without surveillance.**

---

## 🔒 Application Security

| Feature | Status | Notes |
|---------|--------|-------|
| User Auth | ✅ Supported | PostgreSQL peer auth (dev) or Supabase `auth.users` (hosted) |
| Row-Level Security (RLS) | ✅ Schema-ready | Enforced on all user-specific tables across `gr33ncore` |
| Role Separation | ⚙️ Recommended | Use `gr33n_admin`, `gr33n_operator`, `gr33n_guest_inserter` with minimal permissions |
| TLS / HTTPS | ✅ Local-ready | Use Caddy or nginx to terminate HTTPS on your LAN |
| At-rest Encryption | ⚙️ Optional | Use LUKS disk encryption for sensitive deployments |
| CORS | ✅ Configurable | Permissive in dev (`cors.go`), lock down for production |

---

## 🛜 Off-Grid & Intranet Use

gr33n runs entirely without cloud dependencies:

- Run PostgreSQL + TimescaleDB on any local machine or single-board computer
- Use static IPs or `.local` mDNS hostnames for intranet device access
- Optional insert sharing works via Git sync or USB stick transfers
- The Raspberry Pi client connects directly to your local API — no relay, no broker

gr33n **never phones home.** There are no hardcoded cloud services, telemetry pings, or remote update hooks anywhere in this codebase.

---

## 🍓 Raspberry Pi Client Security

The Pi client (`pi_client/gr33n_client.py`) communicates only with your local API:

- All requests go to the `api_base_url` defined in your `config.yaml` — point it at a local IP
- No credentials are stored on the Pi beyond the API base URL
- The systemd service runs as the `pi` user, not root
- GPIO access uses the `gpiozero` library — no kernel module hacks

For production use, rotate the API to a dedicated LAN interface and block WAN access at the router.

---

## 🧪 Automation simulation mode

For local development, automation can run without physical Pi/relay hardware:

- Set `AUTOMATION_SIMULATION_MODE=true` (default behavior in local dev)
- Schedule actions are still logged in `gr33ncore.automation_runs`
- Actuator commands are recorded as simulated actuator events
- Fertigation automation can create `gr33nfertigation.fertigation_events` records

Important: simulation mode validates logic and data flow, but does not physically switch lights, pumps, or valves.

---

## 🚨 Data Sharing + Inserts

For community-contributed insert statements (gr33n_inserts, coming soon):

- Inserts are staged into temporary tables before promotion
- `data_scrubber()` sanitizes input before it touches production data
- Contributors must submit PRs with documented metadata
- If running a public gr33n node that accepts inserts, enable sandbox roles and audit logging

---

## 🤝 Responsible Disclosure

Found a vulnerability in gr33n's schema, API, or insert handling?

Please [open a security issue](https://github.com/dgang0404/gr33n/security) or contact the maintainer directly with:
- Description of the issue
- Affected module, table, or endpoint
- Reproduction steps or test data (if safe to share)

We'll respond within 72 hours.

---

## 🧱 Future Work

- Automated `gr33n-scrub-bot` CI pipeline for insert validation
- Signed data packages using GPG or farm fingerprints
- Local firewall configuration guides for common deployment topologies
- JWT-based auth for multi-user local deployments

---

## 🧬 Security Is Sovereignty

gr33n doesn't outsource trust — it builds it from the soil up.

No telemetry. No black boxes.
**Just code you can hold accountable.**
