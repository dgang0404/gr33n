# Operator logging runbook — capture, retention, archival

**Audience:** Operators running the **Go API** (`cmd/api`) on Docker, systemd, or Kubernetes — anyone who needs **request traces**, **automation outcomes**, and **auth failures** beyond a scrolling terminal.

**Not in scope here:** TimescaleDB **hypertable retention** (sensor readings, time-series pruning). That deletes **database rows**, not application stdout. See **[workflow-guide.md](workflow-guide.md)** and DB operator docs for table-level policies. Phase 26 separates **log retention** from **data retention** deliberately.

---

## 1. Baseline — what the API emits today

The API uses Go **`log/slog`** to **stderr / stdout** (process standard streams). There is **no** built-in log database inside gr33n.

| Signal | When | Key fields |
|--------|------|------------|
| **`request`** | After each HTTP response | `request_id` (matches **`X-Request-ID`**), `method`, `path`, `status`, `duration_ms`, `auth`, optional `farm_id`, `user_id` |
| **`auth_rejected`** | Auth failure (optional) | `reason` — enabled with **`AUTH_DEBUG_LOG=true`** (never logs secrets) |
| **Automation** | Worker tick / runs | `automation worker tick failed` (`phase`, `err`); **`automation schedule run`** / **`automation rule run`** (`schedule_id` / `rule_id`, `farm_id`, `status`) |
| **RAG** (Phase 25+) | Embedding / answer errors | Structured **`slog`** on failures; success **`rag answer completed`** |

**Code:** `cmd/api/request_log.go`, `cmd/api/auth.go`, automation worker wiring in `cmd/api/`, RAG handler logs in `internal/handler/rag/`.

### Env vars (API process)

| Variable | Purpose |
|----------|---------|
| **`LOG_FORMAT=json`** | One JSON object per line — ship to Loki, Elastic, CloudWatch, Datadog, etc. Default is human-readable **text**. |
| **`AUTH_DEBUG_LOG=true`** | Adds **`auth_rejected`** lines with machine-readable **`reason`**. |

See **[INSTALL.md](../INSTALL.md)** § Optional: observability and **[operator-troubleshooting.md](operator-troubleshooting.md)** § Reading API logs.

---

## 2. Capture patterns

### 2a. Docker Compose (full stack)

Container logs are whatever writes to **stdout/stderr** inside the container. **`docker compose logs -f api`** tails them.

**Rotate container logs on disk** (json-file driver) so a noisy farm does not fill the host:

```yaml
# Example — merge under services.api (and optionally ui / db)
logging:
  driver: json-file
  options:
    max-size: "50m"
    max-file: "5"
```

The repo **`docker-compose.yml`** applies this pattern to **`api`**, **`ui`**, and **`db`** so default Compose runs get bounded json-file rotation without extra agents.

For **central aggregation**, add a **logging driver plugin** (e.g. Loki Docker driver) or run **Promtail** / **Fluent Bit** on the host reading **`/var/lib/docker/containers/...`** — vendor-specific; keep **`LOG_FORMAT=json`** first.

### 2b. systemd (bare metal / Pi hosting API only)

Use a **`gr33n-api.service`** unit with **`StandardOutput=journal`** / **`StandardError=journal`** (default when Type=simple). Logs land in **journald**.

```bash
journalctl -u gr33n-api -f
journalctl -u gr33n-api --since "1 hour ago" -o json-pretty
```

**Retention** is controlled by **`/etc/systemd/journald.conf`** (`SystemMaxUse=`, `MaxRetentionSec=`) — still unrelated to Postgres hypertable policies.

### 2c. Bare process (`make run`, dev)

Logs print to the terminal. Redirect if you need a file:

```bash
LOG_FORMAT=json AUTH_MODE=dev go run -tags dev ./cmd/api/ 2>&1 | tee -a /var/log/gr33n-api.jsonl
```

Rotate **`tee`** targets with **logrotate** or ship the file to object storage.

---

## 3. Aggregation stacks (optional)

Operators who want search and dashboards typically:

1. Emit **`LOG_FORMAT=json`** from every API replica.
2. Ship lines with **Promtail → Loki → Grafana**, **Fluent Bit → OpenSearch**, or a hosted agent.
3. Index on **`request_id`**, **`farm_id`**, **`path`**, **`status`**, **`schedule_id`**, **`rule_id`** as needed.

**Privacy:** Access logs include **paths** and ids — treat aggregated logs like **security-sensitive** data (RBAC on Grafana, encrypted buckets).

---

## 4. Archival export (compliance / cold storage)

Application logs are **not** a substitute for **audit_events** or finance trails in Postgres — but long-lived **operational** archives help after incidents.

| Source | Example export |
|--------|----------------|
| Docker | `docker logs gr33n-platform-api-1 2>&1 \| gzip -c > api-$(date -u +%Y%m%d).log.gz` |
| journald | `journalctl -u gr33n-api --since yesterday -o json > archive-$(date -u +%Y%m%d).jsonl` |
| Loki | Use **LogCLI** or Grafana **Explore → CSV/JSON** for a time window |

Push archives to **S3-compatible** cold storage with lifecycle rules (Glacier, etc.).

---

## 5. Correlation checklist

1. Note **`X-Request-ID`** from browser DevTools (Network tab) or client.
2. Search logs for that **`request_id`** across replicas (grep / Loki `{request_id="..."}`).
3. Pair with **`auth_rejected`** **`reason`** if status was **401/403**.
4. For automation, filter **`automation rule run`** **`rule_id`** / **`farm_id`** around the incident timestamp.

---

## 6. Related docs

| Doc | Use |
|-----|-----|
| [operator-troubleshooting.md](operator-troubleshooting.md) | Auth debug, reading **`request`** lines |
| [local-operator-bootstrap.md](local-operator-bootstrap.md) | Local URLs, Compose DB |
| [sit-in-operator-experience.md](workstreams/sit-in-operator-experience.md) | Logging workstream context |
| [phase_26_operator_tutorial_observability_rag.plan.md](plans/phase_26_operator_tutorial_observability_rag.plan.md) | Phase 26 WS2 scope |

---

*Phase 26 WS2 — operational log strategy (aggregation, retention vs DB, archival); complements sit-in structured **`slog`** baseline.*
