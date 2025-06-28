# 🔐 gr33n SECURITY POLICY

gr33n is built with the belief that food system tools should be **trusted, inspectable, and locally secured**—not beholden to distant servers or hidden dependencies.

This document outlines our approach to security in self-hosted, intranet, and offline-first environments.

---

## 🧭 Philosophy

- **Local trust > cloud trust**
- **Simplicity over obscurity**
- **Code is not magic—read it, own it, defend it**

Whether deployed in a cabin, co-op greenhouse, or airgapped industrial farm, gr33n values **resilience without surveillance.**

---

## 🔒 Application Security

| Feature | Status | Notes |
|--------|--------|-------|
| 🔐 User Auth | ✅ Supported | Use PostgreSQL native roles or Supabase `auth.users` locally |
| 🔑 Row-Level Security (RLS) | ✅ Required | Enforced on all user-specific tables across `gr33n_core` |
| 🔒 Role Separation | ⚙️ Recommended | Use `gr33n_admin`, `gr33n_operator`, and `gr33n_guest_inserter` roles with minimal permissions |
| 🔁 TLS Support | ✅ Local HTTPS | For LAN deployments, use Caddy or nginx to terminate HTTPS |
| 🔒 Encryption | ⚙️ Optional | Use LUKS/disk encryption for at-rest protection in sensitive deployments |

---

## 🛜 Off-Grid & Intranet Use

gr33n can be deployed without cloud dependencies:

- 💾 Run PostgreSQL + TimescaleDB locally
- 📡 Use static IPs or .local domains for intranet device access
- 📤 Optional insert sharing works via Git sync or USB stick transfers

gr33n **never “phones home.”** There are no hardcoded cloud services, telemetry pings, or remote update hooks.

---

## 🚨 Data Sharing + Inserts

For community-contributed insert statements:

- Inserts are staged into temporary tables
- `data_scrubber()` sanitizes input before promotion to production
- Insert Commons contributors must submit PRs via GitHub/GitLab with documented metadata

If you’re running a public gr33n node that accepts inserts, **enable sandbox roles and audit logging.**

---

## 🤝 Responsible Disclosure

Found a vulnerability in gr33n’s schema, API, or insert handling?

Please [open a security issue](https://github.com/YOUR_REPO/security) or email the maintainer with:
- Description of issue
- Affected module or table
- Reproduction steps or test data (if safe)

---

## 🧱 Future Work

- Automated `gr33n-scrub-bot` CI pipeline for insert validation
- Signed data packages using GPG or farm fingerprints
- Local firewall configuration guides for typical deployments

---

## 🧬 Security is Sovereignty

gr33n doesn’t outsource trust—it builds it from the soil up.

No telemetry. No black boxes.  
**Just code you can hold accountable.**
